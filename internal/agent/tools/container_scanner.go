package tools

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/SecDuckOps/duckops/internal/container"
)

type ContainerScannerTool struct{}

func NewContainerScannerTool() *ContainerScannerTool {
	return &ContainerScannerTool{}
}

func (t *ContainerScannerTool) Name() string {
	return "container_scanner"
}

func (t *ContainerScannerTool) Schema() ToolSchema {
	return ToolSchema{
		Name:        "container_scanner",
		Description: "Run security scanners (bandit, gosec, semgrep, etc.) inside an isolated Docker container",
		Parameters: map[string]string{
			"action":       "Action: execute, health",
			"container_id": "Docker container ID with tool server",
			"tool":         "Tool name: sast__bandit, sast__gosec, sast__semgrep, sca__trivy_fs, secrets__gitleaks",
			"path":         "Target path to scan (host path, will be translated to container path)",
			"target":       "Alias for path parameter",
			"format":       "Output format (json, text)",
			"scan_type":    "Scan type for trufflehog (filesystem, git)",
			"timeout":      "Execution timeout in seconds",
		},
	}
}

func (t *ContainerScannerTool) ExecuteRaw(ctx context.Context, input map[string]interface{}) (Result, error) {
	action, _ := input["action"].(string)

	switch action {
	case "execute":
		return t.execute(ctx, input)
	case "health":
		return t.health(ctx, input)
	default:
		return t.execute(ctx, input)
	}
}

func (t *ContainerScannerTool) execute(ctx context.Context, input map[string]interface{}) (Result, error) {
	containerID, _ := input["container_id"].(string)
	if containerID == "" {
		return Result{Success: false, Error: "container_id is required"}, nil
	}

	toolName, _ := input["tool"].(string)
	if toolName == "" {
		return Result{Success: false, Error: "tool name is required"}, nil
	}

	path := getString(input, "path")
	if path == "" {
		path = getString(input, "target")
	}

	timeout := 120
	if to, ok := input["timeout"].(float64); ok {
		timeout = int(to)
	}

	kwargs := make(map[string]interface{})
	if path != "" {
		kwargs["path"] = path
		kwargs["target"] = path
	}

	if format := getString(input, "format"); format != "" {
		kwargs["format"] = format
	}

	if scanType := getString(input, "scan_type"); scanType != "" {
		kwargs["scan_type"] = scanType
	}

	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	result, err := container.RunToolInContainer(execCtx, containerID, toolName, kwargs)
	if err != nil {
		return Result{Success: false, Error: fmt.Sprintf("Tool execution failed: %v", err)}, nil
	}

	return Result{Success: true, Data: result}, nil
}

func (t *ContainerScannerTool) health(ctx context.Context, input map[string]interface{}) (Result, error) {
	containerID, _ := input["container_id"].(string)
	if containerID == "" {
		return Result{Success: false, Error: "container_id is required"}, nil
	}

	if err := container.HealthCheckContainer(ctx, containerID); err != nil {
		return Result{Success: false, Error: fmt.Sprintf("Health check failed: %v", err)}, nil
	}

	return Result{Success: true, Data: map[string]interface{}{
		"status":       "healthy",
		"container_id": containerID,
		"transport":    "HTTP",
	}}, nil
}

type Result struct {
	Success bool
	Data    interface{}
	Error   string
}

type ToolSchema struct {
	Name        string
	Description string
	Parameters  map[string]string
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func init() {
	slog.Debug("Container scanner tool registered")
}