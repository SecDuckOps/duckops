package container

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"
)

const DefaultToolPort = 48081

type ToolClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
	timeout    time.Duration
}

type ExecuteRequest struct {
	AgentID  string                 `json:"agent_id"`
	ToolName string                 `json:"tool_name"`
	Kwargs   map[string]interface{} `json:"kwargs"`
}

type ExecuteResponse struct {
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

func NewToolClient(baseURL, token string, timeout time.Duration) *ToolClient {
	if timeout == 0 {
		timeout = 120 * time.Second
	}
	return &ToolClient{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: timeout + 30*time.Second,
		},
		timeout: timeout,
	}
}

func (c *ToolClient) Execute(ctx context.Context, toolName string, kwargs map[string]interface{}) (string, error) {
	if kwargs == nil {
		kwargs = make(map[string]interface{})
	}

	reqBody := ExecuteRequest{
		AgentID:  "duckops-host",
		ToolName: toolName,
		Kwargs:   kwargs,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/execute", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var execResp ExecuteResponse
	if err := json.Unmarshal(respBody, &execResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if execResp.Error != "" {
		return "", fmt.Errorf("tool error: %s", execResp.Error)
	}

	if execResp.Result == nil {
		return "", nil
	}

	switch v := execResp.Result.(type) {
	case string:
		return v, nil
	default:
		resultJSON, _ := json.MarshalIndent(v, "", "  ")
		return string(resultJSON), nil
	}
}

func (c *ToolClient) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: status %d", resp.StatusCode)
	}
	return nil
}

type ContainerInfo struct {
	ID        string
	HostPort  int
	Token     string
	Network   string
	IPAddress string
}

func GetContainerInfo(ctx context.Context, containerID string) (*ContainerInfo, error) {
	cmd := exec.CommandContext(ctx, "docker", "inspect", containerID, "--format", "{{json .NetworkSettings.Ports}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	var ports map[string][]map[string]string
	if err := json.Unmarshal(output, &ports); err != nil {
		return nil, fmt.Errorf("failed to parse ports: %w", err)
	}

	info := &ContainerInfo{ID: containerID}

	portKey := fmt.Sprintf("%d/tcp", DefaultToolPort)
	if bindings, ok := ports[portKey]; ok && len(bindings) > 0 {
		hostPortStr := bindings[0]["HostPort"]
		fmt.Sscanf(hostPortStr, "%d", &info.HostPort)
	}

	info.Token = getContainerToken()

	return info, nil
}

func getContainerToken() string {
	token := os.Getenv("TOOL_SERVER_TOKEN")
	if token != "" {
		return token
	}

	runtimeHome := os.Getenv("DUCKOPS_RUNTIME_HOME")
	if runtimeHome == "" {
		runtimeHome = os.Getenv("HOME") + "/.duckops/runtime"
	}

	tokenFile := runtimeHome + "/.tool_server_token"
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return ""
	}
	return string(bytes.TrimSpace(data))
}

func RunToolInContainer(ctx context.Context, containerID, toolName string, kwargs map[string]interface{}) (string, error) {
	info, err := GetContainerInfo(ctx, containerID)
	if err != nil {
		return "", err
	}

	if info.HostPort == 0 {
		info.HostPort = DefaultToolPort
	}

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", info.HostPort)
	client := NewToolClient(baseURL, info.Token, 120*time.Second)

	return client.Execute(ctx, toolName, kwargs)
}

func HealthCheckContainer(ctx context.Context, containerID string) error {
	info, err := GetContainerInfo(ctx, containerID)
	if err != nil {
		return err
	}

	if info.HostPort == 0 {
		info.HostPort = DefaultToolPort
	}

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", info.HostPort)
	client := NewToolClient(baseURL, info.Token, 30*time.Second)

	return client.HealthCheck(ctx)
}