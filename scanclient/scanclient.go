package scanclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	BaseURL    string
	PAT        string
	HTTPClient *http.Client
	AgentID    string
	Hostname   string
	Version    string
	Platform   string
}

type ScanResult struct {
	ScanID         string           `json:"scanId,omitempty"`
	Findings       []Finding        `json:"findings,omitempty"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities,omitempty"`
}

type Finding struct {
	Tool           string `json:"tool"`
	Severity       string `json:"severity"`
	Title          string `json:"title"`
	Description    string `json:"description,omitempty"`
	FilePath       string `json:"file_path,omitempty"`
	LineStart      int    `json:"line_start,omitempty"`
	LineEnd        int    `json:"line_end,omitempty"`
	RuleID         string `json:"rule_id,omitempty"`
	CWEID          string `json:"cwe_id,omitempty"`
	CVEID          string `json:"cve_id,omitempty"`
	CVSSScore      float64 `json:"cvss_score,omitempty"`
	PackageName    string `json:"package_name,omitempty"`
	PackageVersion string `json:"package_version,omitempty"`
	FixedVersion   string `json:"fixed_version,omitempty"`
	Status         string `json:"status,omitempty"`
}

type Vulnerability struct {
	CVEID         string  `json:"cve_id,omitempty"`
	Title         string  `json:"title,omitempty"`
	Severity      string  `json:"severity"`
	CVSSScore     float64 `json:"cvss_score,omitempty"`
	PackageName   string  `json:"package_name,omitempty"`
	Version       string  `json:"package_version,omitempty"`
	FixedVersion  string  `json:"fixed_version,omitempty"`
	Status        string  `json:"status,omitempty"`
}

type ScanPayload struct {
	WorkspaceID string `json:"workspace_id"`
	ProjectID   string `json:"project_id,omitempty"`
	Type        string `json:"type"`
	Environment string `json:"environment"`
}

type ScanStatusUpdate struct {
	Status       string `json:"status,omitempty"`
	CurrentStage string `json:"current_stage,omitempty"`
	SecurityScore int   `json:"security_score,omitempty"`
}

func New(baseURL, pat string) *Client {
	return &Client{
		BaseURL:    baseURL,
		PAT:        pat,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		Version:    "1.0.0",
		Platform:   "linux",
	}
}

func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.PAT)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respData))
	}

	return respData, nil
}

func (c *Client) RegisterAgent() (string, error) {
	hostname := c.Hostname
	if hostname == "" {
		hostname = "unknown-agent"
	}
	body := map[string]interface{}{
		"agent_id": c.AgentID,
		"hostname": hostname,
		"version":  c.Version,
		"platform": c.Platform,
	}
	resp, err := c.doRequest("POST", "/api/v1/agents/register", body)
	if err != nil {
		return "", fmt.Errorf("register agent: %w", err)
	}
	var result struct {
		Data struct {
			Agent struct {
				UUID string `json:"uuid"`
			} `json:"agent"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}
	return result.Data.Agent.UUID, nil
}

func (c *Client) Heartbeat() error {
	body := map[string]interface{}{
		"agent_id": c.AgentID,
		"status":   "online",
	}
	_, err := c.doRequest("POST", "/api/v1/agents/heartbeat", body)
	return err
}

func (c *Client) StartHeartbeatLoop(interval time.Duration, stop <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := c.Heartbeat(); err != nil {
				fmt.Printf("[scanclient] heartbeat error: %v\n", err)
			}
		case <-stop:
			return
		}
	}
}

func (c *Client) CreateScan(workspaceID, projectID, scanType, environment string) (string, error) {
	payload := ScanPayload{
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		Type:        scanType,
		Environment: environment,
	}
	resp, err := c.doRequest("POST", "/api/v1/scans", payload)
	if err != nil {
		return "", fmt.Errorf("create scan: %w", err)
	}
	var result struct {
		Data struct {
			Scan struct {
				UUID string `json:"uuid"`
			} `json:"scan"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", fmt.Errorf("parse scan response: %w", err)
	}
	return result.Data.Scan.UUID, nil
}

func (c *Client) UpdateScanStatus(scanUUID, status, currentStage string, score int) error {
	update := ScanStatusUpdate{
		Status:        status,
		CurrentStage:  currentStage,
		SecurityScore: score,
	}
	_, err := c.doRequest("PATCH", "/api/v1/scans/"+scanUUID, update)
	return err
}

func (c *Client) SendScanResults(scanUUID string, results ScanResult) error {
	_, err := c.doRequest("POST", "/api/v1/scans/"+scanUUID+"/results", results)
	return err
}

func (c *Client) RunScan(workspaceID, projectID, scanType, environment string, execute func() (ScanResult, error)) error {
	scanUUID, err := c.CreateScan(workspaceID, projectID, scanType, environment)
	if err != nil {
		return fmt.Errorf("create scan: %w", err)
	}

	fmt.Printf("[scanclient] scan created: %s\n", scanUUID)

	if err := c.UpdateScanStatus(scanUUID, "running", "", 0); err != nil {
		return fmt.Errorf("update to running: %w", err)
	}

	fmt.Printf("[scanclient] scan running: %s\n", scanUUID)

	result, err := execute()
	if err != nil {
		c.UpdateScanStatus(scanUUID, "failed", "", 0)
		return fmt.Errorf("scan execution failed: %w", err)
	}

	result.ScanID = scanUUID

	if err := c.SendScanResults(scanUUID, result); err != nil {
		return fmt.Errorf("send results: %w", err)
	}

	fmt.Printf("[scanclient] scan completed: %s (findings=%d, vulns=%d)\n",
		scanUUID, len(result.Findings), len(result.Vulnerabilities))

	return nil
}
