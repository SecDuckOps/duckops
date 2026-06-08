package cmd

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	toolServerTimeout  int
	toolServerPort     int
	toolServerHost     string
	toolServerToken    string
	toolServerCertFile string
	toolServerKeyFile  string
)

func init() {
	toolServerCmd.Flags().IntVarP(&toolServerTimeout, "timeout", "", 600, "Tool execution timeout in seconds")
	toolServerCmd.Flags().IntVarP(&toolServerPort, "port", "", 48081, "Port to listen on")
	toolServerCmd.Flags().StringVar(&toolServerHost, "host", "0.0.0.0", "Host to bind to")
	toolServerCmd.Flags().StringVar(&toolServerToken, "token", "", "Bearer token for authentication (env: TOOL_SERVER_TOKEN)")
	toolServerCmd.Flags().StringVar(&toolServerCertFile, "cert", "", "TLS certificate file")
	toolServerCmd.Flags().StringVar(&toolServerKeyFile, "key", "", "TLS key file")
	rootCmd.AddCommand(toolServerCmd)
}

type ToolServer struct {
	port           int
	host           string
	token          string
	requestTimeout time.Duration
	mux            *http.ServeMux
	mu             sync.Mutex
	nextTaskID     uint64
	agentTasks     map[string]agentTask
}

type agentTask struct {
	id     uint64
	cancel context.CancelFunc
}

type ToolExecutionRequest struct {
	AgentID  string                 `json:"agent_id"`
	ToolName string                 `json:"tool_name"`
	Kwargs   map[string]interface{} `json:"kwargs"`
}

type ToolExecutionResponse struct {
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

type SecurityScanRequest struct {
	AgentID string   `json:"agent_id"`
	Path    string   `json:"path,omitempty"`
	Target  string   `json:"target,omitempty"`
	Tools   []string `json:"tools,omitempty"`
}

type SecurityScanResult struct {
	Tool   string      `json:"tool"`
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

var toolServer *ToolServer

func newToolServer(port int, host, token string, timeout int) *ToolServer {
	ts := &ToolServer{
		port:           port,
		host:           host,
		token:          token,
		requestTimeout: time.Duration(timeout) * time.Second,
		agentTasks:     make(map[string]agentTask),
	}

	ts.mux = http.NewServeMux()
	ts.mux.HandleFunc("POST /execute", ts.handleExecute)
	ts.mux.HandleFunc("POST /scan", ts.handleScan)
	ts.mux.HandleFunc("POST /register_agent", ts.handleRegisterAgent)
	ts.mux.HandleFunc("GET /health", ts.handleHealth)
	ts.mux.HandleFunc("GET /verify", ts.handleVerify)

	return ts
}

var toolServerCmd = &cobra.Command{
	Use:   "tool-server",
	Short: "Run DuckOps tool server (HTTP mode)",
	Long: `Starts an HTTP server that provides security scanning tools.
Used inside a sandbox container to execute bandit, gosec, semgrep, etc.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		token := toolServerToken
		if token == "" {
			token = generateToken(32)
			fmt.Fprintf(os.Stderr, "[INFO] Generated token: %s\n", token)
		}

		toolServer = newToolServer(toolServerPort, toolServerHost, token, toolServerTimeout)

		addr := fmt.Sprintf("%s:%d", toolServerHost, toolServerPort)
		srv := &http.Server{
			Addr:              addr,
			Handler:           toolServer,
			ReadHeaderTimeout: 10 * time.Second,
		}

		fmt.Fprintf(os.Stderr, "[INFO] Starting tool server on %s\n", addr)
		if toolServerCertFile != "" && toolServerKeyFile != "" {
			fmt.Fprintf(os.Stderr, "[INFO] Using HTTPS\n")
			go func() {
				if err := srv.ListenAndServeTLS(toolServerCertFile, toolServerKeyFile); err != nil && err != http.ErrServerClosed {
					slog.Error("HTTPS server error", "error", err)
				}
			}()
		} else {
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					slog.Error("HTTP server error", "error", err)
				}
			}()
		}

		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)
		<-sigch

		fmt.Fprintf(os.Stderr, "[INFO] Shutting down tool server...\n")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(ctx)
	},
}

func (ts *ToolServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if ts.token != "" {
		if r.URL.Path == "/health" {
			ts.mux.ServeHTTP(w, r)
			return
		}
		auth := strings.TrimSpace(r.Header.Get("Authorization"))
		if !strings.HasPrefix(auth, "Bearer ") {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeJSON(w, http.StatusUnauthorized, map[string]string{"detail": "Invalid authentication scheme"})
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		if token != ts.token {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeJSON(w, http.StatusUnauthorized, map[string]string{"detail": "Invalid token"})
			return
		}
	}
	ts.mux.ServeHTTP(w, r)
}

func (ts *ToolServer) handleExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ToolExecutionResponse{Error: "failed to read request body"})
		return
	}
	defer r.Body.Close()

	var req ToolExecutionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ToolExecutionResponse{Error: "invalid request body"})
		return
	}

	req.AgentID = strings.TrimSpace(req.AgentID)
	req.ToolName = strings.TrimSpace(req.ToolName)
	if req.AgentID == "" || req.ToolName == "" {
		writeJSON(w, http.StatusBadRequest, ToolExecutionResponse{Error: "agent_id and tool_name are required"})
		return
	}
	if req.Kwargs == nil {
		req.Kwargs = make(map[string]interface{})
	}

	fmt.Fprintf(os.Stderr, "[DEBUG] Execute: agent=%s tool=%s kwargs=%v\n", req.AgentID, req.ToolName, req.Kwargs)

	execCtx, cancel := context.WithTimeout(r.Context(), ts.requestTimeout)
	taskID := ts.replaceAgentTask(req.AgentID, cancel)
	defer ts.finishAgentTask(req.AgentID, taskID)

	result, err := executeTool(execCtx, req.ToolName, req.Kwargs)
	if err != nil {
		writeJSON(w, http.StatusOK, ToolExecutionResponse{Error: fmt.Sprintf("Tool execution error: %v", err)})
		return
	}

	writeJSON(w, http.StatusOK, ToolExecutionResponse{Result: result})
}

func (ts *ToolServer) handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ToolExecutionResponse{Error: "failed to read request body"})
		return
	}
	defer r.Body.Close()

	var req SecurityScanRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, ToolExecutionResponse{Error: "invalid request body"})
		return
	}

	req.AgentID = strings.TrimSpace(req.AgentID)
	if req.AgentID == "" {
		req.AgentID = "security-scan"
	}
	req.Path = strings.TrimSpace(req.Path)
	if req.Path == "" {
		req.Path = "."
	}
	req.Target = strings.TrimSpace(req.Target)
	if req.Target == "" {
		req.Target = req.Path
	}
	if len(req.Tools) == 0 {
		req.Tools = defaultSecurityScanTools()
	}

	taskID := ts.replaceAgentTask(req.AgentID, func() {})
	defer ts.finishAgentTask(req.AgentID, taskID)

	perToolTimeout := ts.requestTimeout / time.Duration(max(1, len(req.Tools)/3))
	if perToolTimeout < 120*time.Second {
		perToolTimeout = 120 * time.Second
	}

	results := make([]SecurityScanResult, 0, len(req.Tools))
	for _, tool := range req.Tools {
		tool = strings.TrimSpace(tool)
		if tool == "" {
			continue
		}
		toolCtx, toolCancel := context.WithTimeout(r.Context(), perToolTimeout)
		result, err := executeTool(toolCtx, tool, defaultToolKwargs(tool, req.Path, req.Target))
		toolCancel()
		item := SecurityScanResult{Tool: tool, Result: result}
		if err != nil {
			item.Error = fmt.Sprintf("Tool execution error: %v", err)
		}
		results = append(results, item)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"agent_id": req.AgentID,
		"results":  results,
	})
}

func (ts *ToolServer) handleRegisterAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	agentID := strings.TrimSpace(r.URL.Query().Get("agent_id"))
	writeJSON(w, http.StatusOK, map[string]string{
		"status":   "registered",
		"agent_id": agentID,
	})
}

func (ts *ToolServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	ts.mu.Lock()
	agents := make([]string, 0, len(ts.agentTasks))
	for agentID := range ts.agentTasks {
		agents = append(agents, agentID)
	}
	ts.mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":          "healthy",
		"sandbox_mode":    fmt.Sprintf("%t", os.Getenv("DUCKOPS_SANDBOX_MODE") == "true"),
		"environment":     "sandbox",
		"auth_configured": fmt.Sprintf("%t", ts.token != ""),
		"active_agents":   len(agents),
		"agents":          agents,
	})
}

func (ts *ToolServer) handleVerify(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "verified"})
}

func (ts *ToolServer) replaceAgentTask(agentID string, cancel context.CancelFunc) uint64 {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if oldTask, ok := ts.agentTasks[agentID]; ok {
		oldTask.cancel()
	}
	ts.nextTaskID++
	taskID := ts.nextTaskID
	ts.agentTasks[agentID] = agentTask{id: taskID, cancel: cancel}
	return taskID
}

func (ts *ToolServer) finishAgentTask(agentID string, taskID uint64) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if current, ok := ts.agentTasks[agentID]; ok && current.id == taskID {
		delete(ts.agentTasks, agentID)
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func defaultSecurityScanTools() []string {
	return []string{
		"sast__bandit",
		"sast__gosec",
		"sast__semgrep",
		"sca__trivy_fs",
		"secrets__gitleaks",
		"secrets__trufflehog",
	}
}

func defaultToolKwargs(toolName, path, target string) map[string]interface{} {
	switch toolName {
	case "sast__gosec", "sast__semgrep", "sast__eslint", "secrets__trufflehog":
		return map[string]interface{}{"target": target}
	default:
		return map[string]interface{}{"path": path}
	}
}

func executeTool(ctx context.Context, toolName string, kwargs map[string]interface{}) (interface{}, error) {
	switch toolName {
	case "sast__bandit":
		path, err := stringKwarg(kwargs, "path", ".")
		if err != nil {
			return nil, err
		}
		return runTool(ctx, "bandit", []string{"-r", "-f", "json", path})
	case "sast__gosec":
		// Use ./... to scan project packages only, avoiding the module cache.
		return runTool(ctx, "gosec", []string{"-fmt", "json", "-concurrency", "4", "./..."})
	case "sast__semgrep":
		target, err := stringKwarg(kwargs, "target", ".")
		if err != nil {
			return nil, err
		}
		return runTool(ctx, "semgrep", []string{"--config=p/default", "--json", target})
	case "sast__eslint":
		target, err := stringKwarg(kwargs, "target", ".")
		if err != nil {
			return nil, err
		}
		return runTool(ctx, "npx", []string{"eslint", "--format", "json", target})
	case "sast__retire":
		path, err := stringKwarg(kwargs, "path", ".")
		if err != nil {
			return nil, err
		}
		return runTool(ctx, "npx", []string{"retire", "--output", "json", "--path", path})
	case "sca__syft_sbom":
		format := "json"
		if f, ok := kwargs["format"].(string); ok {
			format = f
		}
		path, err := stringKwarg(kwargs, "path", ".")
		if err != nil {
			return nil, err
		}
		return runTool(ctx, "syft", []string{"packages", path, "-o", format + ".json"})
	case "sca__trivy_fs":
		path, err := stringKwarg(kwargs, "path", ".")
		if err != nil {
			return nil, err
		}
		return runTool(ctx, "trivy", []string{"fs", "--skip-db-update", "--offline-scan", "--format", "json", path})
	case "secrets__gitleaks":
		format := "json"
		if f, ok := kwargs["format"].(string); ok {
			format = f
		}
		path, err := stringKwarg(kwargs, "path", ".")
		if err != nil {
			return nil, err
		}
		return runTool(ctx, "gitleaks", []string{"detect", "--source", path, "-f", format})
	case "secrets__trufflehog":
		scanType := "filesystem"
		if s, ok := kwargs["scan_type"].(string); ok {
			scanType = s
		}
		target, err := stringKwarg(kwargs, "target", ".")
		if err != nil {
			return nil, err
		}
		var args []string
		if scanType == "git" {
			args = []string{"--no-update", "git", target}
		} else {
			args = []string{"--no-update", "filesystem", target}
		}
		return runTool(ctx, "trufflehog", args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

func stringKwarg(kwargs map[string]interface{}, key, fallback string) (string, error) {
	raw, ok := kwargs[key]
	if !ok || raw == nil {
		return fallback, nil
	}
	value, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("%s must be a string", key)
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback, nil
	}
	return value, nil
}

func runTool(ctx context.Context, name string, args []string) (interface{}, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		code := exitCode(err)

		if isSoftToolError(name, code) {
			return formatToolOutput(output), nil
		}

		return string(output), err
	}

	return formatToolOutput(output), nil
}

func isSoftToolError(name string, code int) bool {
	if name == "gitleaks" && code == 1 {
		return true
	}
	if name == "semgrep" && code == 7 {
		return true
	}
	return false
}

func formatToolOutput(output []byte) interface{} {
	if json.Valid(output) {
		var prettyJSON bytes.Buffer
		_ = json.Indent(&prettyJSON, output, "", "  ")
		return string(prettyJSON.Bytes())
	}
	return string(output)
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}

func generateToken(length int) string {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "default-token"
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}
