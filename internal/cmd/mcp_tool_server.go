package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/SecDuckOps/duckops/internal/version"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

var (
	mcpToolServerEndpoint  string
	mcpToolServerToken     string
	mcpToolServerTimeout   int
	mcpToolServerWorkspace string
)

func init() {
	mcpToolServerCmd.Flags().StringVar(&mcpToolServerEndpoint, "endpoint", getenvDefault("DUCKOPS_TOOL_SERVER_URL", "http://127.0.0.1:48081"), "DuckOps tool-server URL")
	mcpToolServerCmd.Flags().StringVar(&mcpToolServerToken, "token", firstNonEmpty(os.Getenv("DUCKOPS_TOOL_SERVER_TOKEN"), os.Getenv("TOOL_SERVER_TOKEN")), "Bearer token for DuckOps tool-server")
	mcpToolServerCmd.Flags().StringVar(&mcpToolServerWorkspace, "workspace", getenvDefault("DUCKOPS_CONTAINER_WORKSPACE", "/workspace"), "Workspace path inside the DuckOps sandbox container")
	mcpToolServerCmd.Flags().IntVar(&mcpToolServerTimeout, "timeout", 180, "HTTP request timeout in seconds")
	rootCmd.AddCommand(mcpToolServerCmd)
}

func parsePort(endpoint string) int {
	if !strings.Contains(endpoint, ":") {
		return defaultToolServerPort
	}
	parts := strings.Split(endpoint, ":")
	if len(parts) == 0 {
		return defaultToolServerPort
	}
	p, err := strconv.Atoi(strings.TrimRight(parts[len(parts)-1], "/"))
	if err != nil {
		return defaultToolServerPort
	}
	return p
}

func resolveToolServerConfig() (endpoint, token string) {
	endpoint = mcpToolServerEndpoint
	port := parsePort(endpoint)

	cached, _ := readCachedToolServerToken()
	envToken := firstNonEmpty(
		strings.TrimSpace(mcpToolServerToken),
		os.Getenv("DUCKOPS_TOOL_SERVER_TOKEN"),
		os.Getenv("TOOL_SERVER_TOKEN"),
	)

	if toolServerHealthCheck(port) == nil {
		if token, ok := resolveVerifiedToolServerToken(port, envToken, cached); ok {
			return endpoint, token
		}
		slog.Warn("Tool server is reachable but no valid token was found, re-ensuring sandbox")
	}

	p, t, ok := EnsureToolServer()
	if ok {
		endpoint = fmt.Sprintf("http://localhost:%d", p)
		if t != "" {
			return endpoint, t
		}
		if token, ok := resolveVerifiedToolServerToken(p, envToken, cached); ok {
			return endpoint, token
		}
	}

	return endpoint, ""
}

var mcpToolServerCmd = &cobra.Command{
	Use:   "mcp-tool-server",
	Short: "Run a stdio MCP adapter for the DuckOps security tool server",
	RunE: func(cmd *cobra.Command, _ []string) error {
		endpoint, token := resolveToolServerConfig()
		mcpToolServerEndpoint = endpoint
		mcpToolServerToken = token

		client := newToolServerHTTPClient(endpoint, token, mcpToolServerWorkspace, time.Duration(mcpToolServerTimeout)*time.Second)
		server := mcp.NewServer(&mcp.Implementation{
			Name:    "duckops-security",
			Title:   "DuckOps Security Sandbox",
			Version: version.Version,
		}, nil)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "duckops_security_scan",
			Title:       "Run DuckOps security scan",
			Description: "Run the default DuckOps security scan preset in the configured sandbox container.",
		}, client.securityScan)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "duckops_execute_tool",
			Title:       "Run DuckOps scanner",
			Description: "Run a single bundled DuckOps scanner by tool name.",
		}, client.executeTool)

		return server.Run(cmd.Context(), &mcp.StdioTransport{})
	},
}

type toolServerHTTPClient struct {
	endpoint  string
	token     string
	workspace string
	client    *http.Client
}

type mcpSecurityScanInput struct {
	AgentID string   `json:"agent_id,omitempty" jsonschema:"AI agent identifier,default=duckops-mcp"`
	Path    string   `json:"path,omitempty" jsonschema:"Filesystem path inside the sandbox container,default=/workspace"`
	Target  string   `json:"target,omitempty" jsonschema:"Target path inside the sandbox container for tools that use target,default=/workspace"`
	Tools   []string `json:"tools,omitempty" jsonschema:"Optional scanner names to run. Empty runs the default preset."`
}

type mcpExecuteToolInput struct {
	AgentID  string         `json:"agent_id,omitempty" jsonschema:"AI agent identifier,default=duckops-mcp"`
	ToolName string         `json:"tool_name" jsonschema:"Scanner name such as sast__semgrep or sca__trivy_fs"`
	Kwargs   map[string]any `json:"kwargs,omitempty" jsonschema:"Scanner arguments passed to the DuckOps tool server"`
}

func newToolServerHTTPClient(endpoint, token, workspace string, timeout time.Duration) *toolServerHTTPClient {
	return &toolServerHTTPClient{
		endpoint:  strings.TrimRight(endpoint, "/"),
		token:     token,
		workspace: cleanContainerWorkspace(workspace),
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *toolServerHTTPClient) securityScan(ctx context.Context, _ *mcp.CallToolRequest, input mcpSecurityScanInput) (*mcp.CallToolResult, any, error) {
	if strings.TrimSpace(input.AgentID) == "" {
		input.AgentID = "duckops-mcp"
	}
	if strings.TrimSpace(input.Path) == "" {
		input.Path = c.workspace
	}
	if strings.TrimSpace(input.Target) == "" {
		input.Target = input.Path
	}

	result, err := c.post(ctx, "/scan", input)
	if err != nil {
		return toolError(err), nil, nil
	}
	return textResult(result), nil, nil
}

func cleanContainerWorkspace(workspace string) string {
	workspace = strings.TrimSpace(workspace)
	if workspace == "" {
		return "/workspace"
	}
	return strings.TrimRight(workspace, "/")
}

func (c *toolServerHTTPClient) executeTool(ctx context.Context, _ *mcp.CallToolRequest, input mcpExecuteToolInput) (*mcp.CallToolResult, any, error) {
	input.AgentID = strings.TrimSpace(input.AgentID)
	if input.AgentID == "" {
		input.AgentID = "duckops-mcp"
	}
	input.ToolName = strings.TrimSpace(input.ToolName)
	if input.ToolName == "" {
		return toolError(fmt.Errorf("tool_name is required")), nil, nil
	}
	if input.Kwargs == nil {
		input.Kwargs = map[string]any{}
	}

	result, err := c.post(ctx, "/execute", input)
	if err != nil {
		return toolError(err), nil, nil
	}
	return textResult(result), nil, nil
}

func (c *toolServerHTTPClient) post(ctx context.Context, path string, payload any) (string, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+path, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(c.token) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(c.token))
	}

	slog.Debug("tool server request", "path", path, "payload_size", len(body))

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("call DuckOps tool server: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	slog.Debug("tool server response", "path", path, "status", resp.StatusCode, "response_size", len(data))

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("tool server returned %s: %s", resp.Status, strings.TrimSpace(string(data)))
	}
	return prettyJSON(data), nil
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

func toolError(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
		IsError: true,
	}
}

func prettyJSON(data []byte) string {
	if !json.Valid(data) {
		return string(data)
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "  "); err != nil {
		return string(data)
	}
	return buf.String()
}

func getenvDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
