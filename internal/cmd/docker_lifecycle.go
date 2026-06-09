package cmd

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	containerName    = "duckops-sandbox"
	containerPort    = 48081
	lifecycleTimeout = 30 * time.Second
	pullTimeout      = 120 * time.Second
)

type containerConfig struct {
	Image   string
	Port    int
	Token   string
	Env     map[string]string
	Volumes map[string]string
}

func defaultSandboxConfig(port int) containerConfig {
	imageName := os.Getenv("DUCKOPS_SANDBOX_IMAGE")
	if imageName == "" {
		imageName = "ghcr.io/usestrix/strix-sandbox:0.1.13"
	}
	cwd, _ := os.Getwd()
	return containerConfig{
		Image: imageName,
		Port:  port,
		Env: map[string]string{
			"TOOL_SERVER_PORT": fmt.Sprintf("%d", containerPort),
		},
		Volumes: map[string]string{
			cwd: "/workspace/duck",
		},
	}
}

// EnsureDockerSandbox is the full lifecycle entry point.
// It handles image pull, container management, health checks, and rollback.
func EnsureDockerSandbox(port int) (int, string, bool) {
	cfg := defaultSandboxConfig(port)
	cfg.Token = generateToken(32)

	// Phase 1: Ensure the image exists locally.
	if !ensureImage(cfg.Image) {
		return port, "", false
	}

	// Phase 2: Check container state and act accordingly.
	state := inspectContainer(containerName)

	// Resolve the actual host port for any existing container.
	if p, ok := resolveContainerHostPort(containerName, containerPort); ok {
		port = p
	}

	switch state {
	case "running":
		cached, _ := readCachedToolServerToken()
		if token, ok := resolveVerifiedToolServerToken(port, cached); ok {
			slog.Info("Docker sandbox already running", "container", containerName, "port", port)
			return port, token, true
		}
		slog.Warn("Running container but token could not be verified, proceeding without auth")
		return port, "", true

	case "stopped":
		slog.Info("Container exists but stopped, starting", "container", containerName)
		if err := exec.Command("docker", "start", containerName).Run(); err != nil {
			slog.Error("Failed to start existing container", "error", err)
			return removeAndRecreate(cfg)
		}
		if waitForHealth(port) {
			token := recoverTokenFromContainer(containerName)
			if token != "" {
				writeToolServerToken(token)
			}
			return port, token, true
		}
		slog.Warn("Started container failed health check, recreating")
		exec.Command("docker", "rm", "-f", containerName).Run()
		return createAndWait(cfg)

	case "config_changed":
		slog.Info("Container config changed, recreating", "container", containerName)
		exec.Command("docker", "rm", "-f", containerName).Run()
		return createAndWait(cfg)

	default: // "missing"
		return createAndWait(cfg)
	}
}

// ensureImage checks if the image exists locally and pulls if needed.
// Returns true if an image is available after the operation.
func ensureImage(image string) bool {
	// Check if image exists locally.
	if imageExistsLocally(image) {
		slog.Debug("Image found locally", "image", image)
		return true
	}

	slog.Info("Image not found locally, pulling", "image", image)
	return pullImage(image)
}

func imageExistsLocally(image string) bool {
	cmd := exec.Command("docker", "images", "--format", "{{.Repository}}:{{.Tag}}")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if strings.TrimSpace(line) == image {
			return true
		}
	}
	return false
}

func pullImage(image string) bool {
	cmd := exec.Command("docker", "pull", image)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		slog.Error("Failed to pull image", "image", image, "error", err, "stderr", stderr.String())
		// If a local copy exists despite pull failure, log and continue.
		if imageExistsLocally(image) {
			slog.Warn("Pull failed but local image exists, using local copy", "image", image)
			return true
		}
		return false
	}
	return true
}

// containerState is one of: "running", "stopped", "missing", "config_changed"
func inspectContainer(name string) string {
	// Check if container exists at all.
	cmd := exec.Command("docker", "ps", "-a", "--filter", "name="+name, "--format", "{{.Status}}")
	out, err := cmd.Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		return "missing"
	}
	status := strings.TrimSpace(string(out))

	// Check if container is running.
	runningCheck := exec.Command("docker", "ps", "--filter", "name="+name, "--filter", "status=running", "--format", "{{.ID}}")
	runningOut, _ := runningCheck.Output()
	if strings.TrimSpace(string(runningOut)) != "" {
		// Running — check if config matches.
		if containerConfigMatches(name) {
			return "running"
		}
		return "config_changed"
	}

	// Stopped — check if config matches (for potential restart).
	if strings.HasPrefix(status, "Exited") {
		if containerConfigMatches(name) {
			return "stopped"
		}
		return "config_changed"
	}

	return "missing"
}

// containerConfigMatches checks if the running container's config matches what we expect.
func containerConfigMatches(name string) bool {
	cfg := defaultSandboxConfig(resolveToolServerPort())

	// Check image name.
	imageCmd := exec.Command("docker", "inspect", name, "--format", "{{.Config.Image}}")
	imageOut, err := imageCmd.Output()
	if err != nil || strings.TrimSpace(string(imageOut)) != cfg.Image {
		return false
	}

	// Check port mapping.
	portCmd := exec.Command("docker", "inspect", name, "--format", "{{range $p, $conf := .NetworkSettings.Ports}}{{$p}}{{end}}")
	portOut, _ := portCmd.Output()
	expectedPort := fmt.Sprintf("%d/tcp", containerPort)
	if !strings.Contains(string(portOut), expectedPort) {
		return false
	}

	return true
}

func recoverTokenFromContainer(name string) string {
	getToken := exec.Command("docker", "exec", name, "cat", "/var/lib/duckops/home/.tool_server_token")
	tokenBytes, err := getToken.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(tokenBytes))
}

func createAndWait(cfg containerConfig) (int, string, bool) {
	if err := writeToolServerToken(cfg.Token); err != nil {
		slog.Error("Failed to write tool server token", "error", err)
		return cfg.Port, "", false
	}

	args := buildDockerRunArgs(cfg)
	cmd := exec.Command("docker", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		slog.Error("Failed to start Docker sandbox", "error", err, "stderr", stderr.String())
		return cfg.Port, "", false
	}

	// Resolve the actual host port mapped to the container.
	hostPort, ok := resolveContainerHostPort(containerName, containerPort)
	if !ok {
		slog.Error("Failed to resolve container host port")
		exec.Command("docker", "rm", "-f", containerName).Run()
		return cfg.Port, "", false
	}
	cfg.Port = hostPort

	if waitForHealth(cfg.Port) {
		slog.Info("Docker sandbox started", "port", cfg.Port)
		return cfg.Port, cfg.Token, true
	}

	slog.Error("Docker sandbox did not become healthy within timeout", "port", cfg.Port)

	// Optional rollback: if we detect this is a new image that failed, try last known good.
	rollbackImage := readLastGoodImage()
	if rollbackImage != "" && rollbackImage != cfg.Image {
		slog.Info("New image failed health check, attempting rollback", "image", rollbackImage)
		exec.Command("docker", "rm", "-f", containerName).Run()
		cfg.Image = rollbackImage
		args := buildDockerRunArgs(cfg)
		rollbackCmd := exec.Command("docker", args...)
		var rbStderr bytes.Buffer
		rollbackCmd.Stderr = &rbStderr
		if rollbackCmd.Run() == nil {
			if hostPort, ok := resolveContainerHostPort(containerName, containerPort); ok {
				cfg.Port = hostPort
			}
			if waitForHealth(cfg.Port) {
				slog.Info("Rollback successful", "image", rollbackImage)
				return cfg.Port, cfg.Token, true
			}
		}
		slog.Error("Rollback also failed", "image", rollbackImage)
	}

	return cfg.Port, "", false
}

func removeAndRecreate(cfg containerConfig) (int, string, bool) {
	exec.Command("docker", "rm", "-f", containerName).Run()
	return createAndWait(cfg)
}

func buildDockerRunArgs(cfg containerConfig) []string {
	args := []string{
		"run", "--rm", "-d",
		"--name", containerName,
		"-p", fmt.Sprintf("%d", containerPort),
	}

	for hostPath, containerPath := range cfg.Volumes {
		args = append(args, "-v", fmt.Sprintf("%s:%s", hostPath, containerPath))
	}

	args = append(args, "-e", fmt.Sprintf("TOOL_SERVER_TOKEN=%s", cfg.Token))
	for k, v := range cfg.Env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	args = append(args, cfg.Image)
	return args
}

func resolveContainerHostPort(name string, containerPort int) (int, bool) {
	cmd := exec.Command("docker", "port", name, fmt.Sprintf("%d/tcp", containerPort))
	out, err := cmd.Output()
	if err != nil {
		return 0, false
	}
	// Output format: "0.0.0.0:34501\n" or "0.0.0.0:34501\n[::]:34501\n"
	line := strings.TrimSpace(string(out))
	if idx := strings.LastIndex(line, ":"); idx >= 0 {
		if port, err := strconv.Atoi(line[idx+1:]); err == nil {
			return port, true
		}
	}
	return 0, false
}

func waitForHealth(port int) bool {
	deadline := time.Now().Add(lifecycleTimeout)
	for time.Now().Before(deadline) {
		if toolServerHealthCheck(port) == nil {
			return true
		}
		time.Sleep(toolServerPollInterval)
	}
	return false
}

// ─── Last Known Good Image Tracking ──────────────────────────────

func lastGoodImagePath() string {
	return filepath.Join(duckopsGlobalDir(), "last-good-image.txt")
}

func saveLastGoodImage(image string) {
	os.WriteFile(lastGoodImagePath(), []byte(image+"\n"), 0o644)
}

func readLastGoodImage() string {
	data, err := os.ReadFile(lastGoodImagePath())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// ─── Container Log Retrieval ─────────────────────────────────────

func containerLogs(name string, tail int) string {
	cmd := exec.Command("docker", "logs", "--tail", fmt.Sprintf("%d", tail), name)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("failed to get logs: %v", err)
	}
	return string(out)
}
