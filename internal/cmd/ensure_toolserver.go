package cmd

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/SecDuckOps/duckops/internal/home"
)

const (
	defaultToolServerPort   = 48081
	toolServerTokenFilename = "tool-server.token"
	toolServerPidFilename   = "tool-server.pid"
	toolServerReadyTimeout  = 5 * time.Second
	toolServerPollInterval  = 200 * time.Millisecond
)

func duckopsGlobalDir() string {
	return filepath.Join(home.Dir(), ".duckops")
}

func toolServerTokenPath() string {
	return filepath.Join(duckopsGlobalDir(), toolServerTokenFilename)
}

func toolServerPidPath() string {
	return filepath.Join(duckopsGlobalDir(), toolServerPidFilename)
}

func readToolServerToken() (string, error) {
	// Primary: ~/.duckops/tool-server.token
	data, err := os.ReadFile(toolServerTokenPath())
	if err == nil {
		return strings.TrimSpace(string(data)), nil
	}

	// Fallback: TOOL_SERVER_TOKEN env var (set by container entrypoint)
	if envToken := os.Getenv("TOOL_SERVER_TOKEN"); envToken != "" {
		return envToken, nil
	}

	// Fallback: container runtime token file
	runtimeHome := os.Getenv("DUCKOPS_RUNTIME_HOME")
	if runtimeHome == "" {
		runtimeHome = "/var/lib/duckops/home"
	}
	if data, err := os.ReadFile(filepath.Join(runtimeHome, ".tool_server_token")); err == nil {
		return strings.TrimSpace(string(data)), nil
	}

	return "", fmt.Errorf("token not found in any location")
}

func writeToolServerToken(token string) error {
	return os.WriteFile(toolServerTokenPath(), []byte(token+"\n"), 0o600)
}

func readToolServerPid() (int, error) {
	data, err := os.ReadFile(toolServerPidPath())
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
}

func writeToolServerPid(pid int) error {
	return os.WriteFile(toolServerPidPath(), []byte(fmt.Sprintf("%d\n", pid)), 0o600)
}

func toolServerHealthCheck(port int) error {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/health", port))
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}
	return nil
}

// verifyTokenResult returns one of:
//
//	"valid"   — /verify returned 200 (token is correct)
//	"invalid" — /verify returned 401 (token is wrong)
//	"unknown" — /verify returned 404 or connection error (old server, can't tell)
func verifyTokenResult(port int, token string) string {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/verify", port), nil)
	if err != nil {
		return "unknown"
	}
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "unknown"
	}
	resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		return "valid"
	case http.StatusNotFound:
		return "unknown"
	default:
		return "invalid"
	}
}

func resolveToolServerPort() int {
	if v := os.Getenv("TOOL_SERVER_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			return p
		}
	}
	return defaultToolServerPort
}

func EnsureToolServer() (port int, token string, ok bool) {
	port = resolveToolServerPort()
	baseDir := duckopsGlobalDir()

	if toolServerHealthCheck(port) == nil {
		cached, _ := readCachedToolServerToken()
		envToken := strings.TrimSpace(os.Getenv("TOOL_SERVER_TOKEN"))
		if token, ok := resolveVerifiedToolServerToken(port, envToken, cached); ok {
			if cached != "" && token != cached {
				slog.Info("Synced tool server token with running sandbox")
			}
			return port, token, true
		}

		// Can't recover the token. Proceed without auth — the server is running
		// but we can't authenticate. This is degraded but functional.
		slog.Warn("Tool server is running but token is unknown, proceeding without authentication")
		return port, "", true
	}

	// If running outside the sandbox container, start the tool server in Docker.
	if os.Getenv("DUCKOPS_SANDBOX_MODE") != "true" {
		return EnsureToolServerDocker(port)
	}

	if err := os.MkdirAll(baseDir, 0o700); err != nil {
		slog.Error("Failed to create ~/.duckops directory", "error", err)
		return port, "", false
	}

	token = generateToken(32)
	if err := writeToolServerToken(token); err != nil {
		slog.Error("Failed to write tool server token", "error", err)
		return port, "", false
	}

	pid, err := startDetachedToolServer(port, token)
	if err != nil {
		slog.Error("Failed to start tool server", "error", err)
		return port, "", false
	}

	if err := writeToolServerPid(pid); err != nil {
		slog.Warn("Failed to write tool server PID", "error", err)
	}

	deadline := time.Now().Add(toolServerReadyTimeout)
	for time.Now().Before(deadline) {
		if toolServerHealthCheck(port) == nil {
			slog.Info("Tool server started", "pid", pid, "port", port)
			return port, token, true
		}
		time.Sleep(toolServerPollInterval)
	}

	slog.Error("Tool server did not become ready within timeout",
		"pid", pid, "port", port, "timeout", toolServerReadyTimeout)
	return port, "", false
}

func sandboxImageName() string {
	if img := os.Getenv("DUCKOPS_SANDBOX_IMAGE"); img != "" {
		return img
	}
	return "duckops-sandbox:latest"
}

func EnsureToolServerDocker(port int) (int, string, bool) {
	port, token, ok := EnsureDockerSandbox(port)
	if ok {
		saveLastGoodImage(sandboxImageName())
	}
	return port, token, ok
}

func startDetachedToolServer(port int, token string) (int, error) {
	exe, err := os.Executable()
	if err != nil {
		return 0, fmt.Errorf("failed to get executable path: %w", err)
	}

	cmd := exec.Command(exe, "tool-server",
		"--port", fmt.Sprintf("%d", port),
		"--token", token,
		"--timeout", "600",
	)
	detachProcess(cmd)

	logDir := filepath.Join(duckopsGlobalDir(), "tool-server-logs")
	if err := os.MkdirAll(logDir, 0o700); err != nil {
		return 0, fmt.Errorf("failed to create log directory: %w", err)
	}

	stdout, err := os.Create(filepath.Join(logDir, "stdout.log"))
	if err != nil {
		return 0, fmt.Errorf("failed to create stdout log: %w", err)
	}
	defer stdout.Close()
	cmd.Stdout = stdout

	stderr, err := os.Create(filepath.Join(logDir, "stderr.log"))
	if err != nil {
		return 0, fmt.Errorf("failed to create stderr log: %w", err)
	}
	defer stderr.Close()
	cmd.Stderr = stderr

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start tool server: %w", err)
	}

	if err := cmd.Process.Release(); err != nil {
		return 0, fmt.Errorf("failed to detach tool server: %w", err)
	}

	return cmd.Process.Pid, nil
}
