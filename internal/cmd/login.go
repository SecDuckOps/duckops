package cmd

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/SecDuckOps/duckops/internal/client"
	"github.com/SecDuckOps/duckops/internal/config"
	"github.com/SecDuckOps/duckops/internal/oauth"
	"github.com/SecDuckOps/duckops/internal/oauth/copilot"
	"github.com/SecDuckOps/duckops/internal/oauth/hyper"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/x/ansi"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var loginCmd = &cobra.Command{
	Aliases: []string{"auth"},
	Use:     "login [platform]",
	Short:   "Login duckops to a platform",
	Long: `Login duckops to a specified platform.
The platform should be provided as an argument.
Available platforms are: duckops, copilot.`,
	Example: `
# Authenticate with duckops
duckops login

# Authenticate with GitHub Copilot
duckops login copilot

# Force re-authentication even if already logged in
duckops login -f copilot
  `,
	ValidArgs: []cobra.Completion{
		// "hyper",
		"copilot",
		"github",
		"github-copilot",
		"duckops",
	},
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		provider := "duckops"
		if len(args) > 0 {
			provider = args[0]
		}

		if provider == "duckops" {
			force, _ := cmd.Flags().GetBool("force")
			pat, _ := cmd.Flags().GetString("token")
			return loginDuckOps(force, pat)
		}

		c, ws, cleanup, err := connectToServer(cmd)
		if err != nil {
			return err
		}
		defer cleanup()

		progressEnabled := ws.Config.Options.Progress == nil || *ws.Config.Options.Progress
		if progressEnabled && supportsProgressBar() {
			_, _ = fmt.Fprintf(os.Stderr, ansi.SetIndeterminateProgressBar)
			defer func() { _, _ = fmt.Fprintf(os.Stderr, ansi.ResetProgressBar) }()
		}

		force, _ := cmd.Flags().GetBool("force")
		switch provider {
		case "hyper":
			return loginHyper(c, ws.ID, force)
		case "copilot", "github", "github-copilot":
			return loginCopilot(c, ws.ID, force)
		default:
			return fmt.Errorf("unknown platform: %s", provider)
		}
	},
}

func init() {
	loginCmd.Flags().BoolP("force", "f", false, "Force re-authentication even if already logged in")
	loginCmd.Flags().StringP("token", "t", "", "Personal Access Token")
	loginCmd.Flags().Bool("pat-only", false, "Skip email/password flow and use PAT directly")
}

func loginHyper(c *client.Client, wsID string, force bool) error {
	ctx := getLoginContext()

	if !force {
		cfg, err := c.GetConfig(ctx, wsID)
		if err == nil && cfg != nil {
			if pc, ok := cfg.Providers.Get("hyper"); ok && pc.OAuthToken != nil {
				fmt.Println("You are already logged in to Hyper.")
				fmt.Println("Use --force to re-authenticate.")
				return nil
			}
		}
	}

	resp, err := hyper.InitiateDeviceAuth(ctx)
	if err != nil {
		return err
	}

	if clipboard.WriteAll(resp.UserCode) == nil {
		fmt.Println("The following code should be on clipboard already:")
	} else {
		fmt.Println("Copy the following code:")
	}

	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Render(resp.UserCode))
	fmt.Println()
	fmt.Println("Press enter to open this URL, and then paste it there:")
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Hyperlink(resp.VerificationURL, "id=hyper").Render(resp.VerificationURL))
	fmt.Println()
	waitEnter()
	if err := browser.OpenURL(resp.VerificationURL); err != nil {
		fmt.Println("Could not open the URL. You'll need to manually open the URL in your browser.")
	}

	fmt.Println("Exchanging authorization code...")
	refreshToken, err := hyper.PollForToken(ctx, resp.DeviceCode, resp.ExpiresIn)
	if err != nil {
		return err
	}

	fmt.Println("Exchanging refresh token for access token...")
	token, err := hyper.ExchangeToken(ctx, refreshToken)
	if err != nil {
		return err
	}

	fmt.Println("Verifying access token...")
	introspect, err := hyper.IntrospectToken(ctx, token.AccessToken)
	if err != nil {
		return fmt.Errorf("token introspection failed: %w", err)
	}
	if !introspect.Active {
		return fmt.Errorf("access token is not active")
	}

	if err := cmp.Or(
		c.SetConfigField(ctx, wsID, config.ScopeGlobal, "providers.hyper.api_key", token.AccessToken),
		c.SetConfigField(ctx, wsID, config.ScopeGlobal, "providers.hyper.oauth", token),
	); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("You're now authenticated with Hyper!")
	return nil
}

func loginCopilot(c *client.Client, wsID string, force bool) error {
	loginCtx := getLoginContext()

	if !force {
		cfg, err := c.GetConfig(loginCtx, wsID)
		if err == nil && cfg != nil {
			if pc, ok := cfg.Providers.Get("copilot"); ok && pc.OAuthToken != nil {
				fmt.Println("You are already logged in to GitHub Copilot.")
				fmt.Println("Use --force to re-authenticate.")
				return nil
			}
		}
	}

	diskToken, hasDiskToken := copilot.RefreshTokenFromDisk()
	var token *oauth.Token

	switch {
	case hasDiskToken:
		fmt.Println("Found existing GitHub Copilot token on disk. Using it to authenticate...")

		t, err := copilot.RefreshToken(loginCtx, diskToken)
		if err != nil {
			return fmt.Errorf("unable to refresh token from disk: %w", err)
		}
		token = t
	default:
		fmt.Println("Requesting device code from GitHub...")
		dc, err := copilot.RequestDeviceCode(loginCtx)
		if err != nil {
			return err
		}

		fmt.Println()
		fmt.Println("Open the following URL and follow the instructions to authenticate with GitHub Copilot:")
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Hyperlink(dc.VerificationURI, "id=copilot").Render(dc.VerificationURI))
		fmt.Println()
		fmt.Println("Code:", lipgloss.NewStyle().Bold(true).Render(dc.UserCode))
		fmt.Println()
		fmt.Println("Waiting for authorization...")

		t, err := copilot.PollForToken(loginCtx, dc)
		if err == copilot.ErrNotAvailable {
			fmt.Println()
			fmt.Println("GitHub Copilot is unavailable for this account. To signup, go to the following page:")
			fmt.Println()
			fmt.Println(lipgloss.NewStyle().Hyperlink(copilot.SignupURL, "id=copilot-signup").Render(copilot.SignupURL))
			fmt.Println()
			fmt.Println("You may be able to request free access if eligible. For more information, see:")
			fmt.Println()
			fmt.Println(lipgloss.NewStyle().Hyperlink(copilot.FreeURL, "id=copilot-free").Render(copilot.FreeURL))
		}
		if err != nil {
			return err
		}
		token = t
	}

	if err := cmp.Or(
		c.SetConfigField(loginCtx, wsID, config.ScopeGlobal, "providers.copilot.api_key", token.AccessToken),
		c.SetConfigField(loginCtx, wsID, config.ScopeGlobal, "providers.copilot.oauth", token),
	); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("You're now authenticated with GitHub Copilot!")
	return nil
}

func loginDuckOps(force bool, pat string) error {
	ctx := getLoginContext()

	if !force {
		if existing, _ := duckopsAPIKey(); existing != "" {
			fmt.Println("You are already logged in to DuckOps.")
			fmt.Println("Use --force to re-authenticate.")
			return nil
		}
	}

	serverHost := cmp.Or(os.Getenv("DUCKOPS_SERVER_HOST"), "192.168.1.14")
	serverPort := cmp.Or(os.Getenv("DUCKOPS_SERVER_PORT"), "8000")
	serverURL := fmt.Sprintf("http://%s:%s", serverHost, serverPort)

	dashHost := cmp.Or(os.Getenv("DUCKOPS_DASHBOARD_HOST"), "localhost")
	dashPort := cmp.Or(os.Getenv("DUCKOPS_DASHBOARD_PORT"), "3000")

	var patToken string

	if pat != "" {
		patToken = strings.TrimSpace(pat)
	} else {
		dashURL := fmt.Sprintf("http://%s:%s/settings/tokens", dashHost, dashPort)
		fmt.Println("Opening browser to generate a Personal Access Token...")
		fmt.Println()
		fmt.Println("If the browser does not open, visit:")
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Hyperlink(dashURL, "id=duckops-login").Render(dashURL))
		fmt.Println()
		fmt.Println("Sign in and generate a token, then paste it below.")
		fmt.Println()

		if err := browser.OpenURL(dashURL); err != nil {
			fmt.Println("Could not open the browser. Please open the URL manually.")
		}

		fmt.Print("Paste your Personal Access Token: ")
		inputBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return fmt.Errorf("failed to read token: %w", err)
		}
		patToken = strings.TrimSpace(string(inputBytes))
	}

	if !strings.HasPrefix(patToken, "duck_pat_") {
		return fmt.Errorf("invalid token format: must start with 'duck_pat_'")
	}

	if pat == "" {
		fmt.Println("Token generated successfully!")
		fmt.Println()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, serverURL+"/api/v1/agents/ping", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+patToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to DuckOps server at %s: %w", serverURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token verification failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	// Extract user email from ping response
	var pingResp struct {
		Data struct {
			User struct {
				Email string `json:"email"`
			} `json:"user"`
		} `json:"data"`
	}
	userEmail := ""
	if err := json.NewDecoder(resp.Body).Decode(&pingResp); err == nil {
		userEmail = pingResp.Data.User.Email
	}

	// Save PAT to auth.json
	if err := writeDuckOpsAuth(patToken, userEmail); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	// Clean up legacy auth fields from duckops.json
	if st, err := config.Load(config.GlobalWorkspaceDir(), "", false); err == nil {
		_ = st.RemoveConfigField(config.ScopeGlobal, "duckops_api_key")
		_ = st.RemoveConfigField(config.ScopeGlobal, "duckops_user_email")
	}

	fmt.Println()
	fmt.Printf("You're now authenticated with DuckOps as %s!\n", userEmail)
	return nil
}

func getLoginContext() context.Context {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	go func() {
		<-ctx.Done()
		cancel()
		os.Exit(1)
	}()
	return ctx
}

func waitEnter() {
	_, _ = fmt.Scanln()
}

// writeDuckOpsAuth writes duckops authentication credentials to auth.json
// in the provider-centric format.
func writeDuckOpsAuth(patToken, userEmail string) error {
	authPath := config.GlobalAuthConfig()

	// Read existing auth.json or start fresh
	var auth map[string]any
	if data, err := os.ReadFile(authPath); err == nil {
		json.Unmarshal(data, &auth)
	}
	if auth == nil {
		auth = make(map[string]any)
	}

	duckopsEntry := map[string]any{
		"type": "api",
		"key":  patToken,
	}
	if userEmail != "" {
		duckopsEntry["email"] = userEmail
	}
	auth["duckops"] = duckopsEntry

	data, err := json.MarshalIndent(auth, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal auth config: %w", err)
	}
	data = append(data, '\n')

	if err := os.MkdirAll(filepath.Dir(authPath), 0o700); err != nil {
		return fmt.Errorf("failed to create auth config directory: %w", err)
	}
	if err := os.WriteFile(authPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write auth config: %w", err)
	}

	return nil
}
