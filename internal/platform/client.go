package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

const (
	defaultBaseURL    = "http://127.0.0.1:8080/api/v1"
	defaultTimeout    = 30 * time.Second
	maxRetries        = 5
	baseBackoff       = 500 * time.Millisecond
	maxBackoff        = 30 * time.Second
	jitterFactor      = 0.2
	heartbeatInterval = 30 * time.Second
)

type Client struct {
	BaseURL string
	PAT     string
	http    *http.Client
	agentID string
	queue   *Queue

	logf func(format string, args ...any)
}

type ClientOption func(*Client)

func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) {
		c.http = hc
	}
}

func WithLogger(logf func(string, ...any)) ClientOption {
	return func(c *Client) {
		c.logf = logf
	}
}

func WithQueueDir(dir string) ClientOption {
	return func(c *Client) {
		q, err := NewQueue(dir)
		if err == nil {
			c.queue = q
		}
	}
}

func NewClient(pat string, opts ...ClientOption) *Client {
	c := &Client{
		BaseURL: defaultBaseURL,
		PAT:     pat,
		http: &http.Client{
			Timeout: defaultTimeout,
		},
		logf: func(format string, args ...any) {},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) log(format string, args ...any) {
	c.logf("[platform] "+format, args...)
}

// --- HTTP internals with retry ---

func (c *Client) do(ctx context.Context, method, path string, body any) ([]byte, error) {
	if c.PAT == "" {
		return nil, &PlatformError{Code: ErrAuthFailed, Message: "PAT not set — run 'duckops login' first"}
	}

	url := c.BaseURL + "/" + strings.TrimPrefix(path, "/")

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, &PlatformError{Code: ErrInvalidPayload, Message: "failed to marshal request", Wrapped: err}
		}
		reqBody = bytes.NewReader(data)
	}

	return c.retryDo(ctx, method, url, reqBody)
}

func (c *Client) retryDo(ctx context.Context, method, url string, body io.Reader) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := computeBackoff(attempt)
			c.log("retry attempt %d/%d after %v", attempt, maxRetries, backoff)

			select {
			case <-ctx.Done():
				return nil, &PlatformError{Code: ErrTimeout, Message: "request cancelled", Wrapped: ctx.Err()}
			case <-time.After(backoff):
			}
		}

		resp, err := c.doOnce(ctx, method, url, body)
		if err != nil {
			lastErr = err

			if isRetryable(err) {
				continue
			}
			return nil, err
		}

		return resp, nil
	}

	return nil, &PlatformError{Code: ErrMaxRetries, Message: fmt.Sprintf("max retries exceeded (%d)", maxRetries), Wrapped: lastErr}
}

func (c *Client) doOnce(ctx context.Context, method, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, &PlatformError{Code: ErrInvalidPayload, Message: "failed to create request", Wrapped: err}
	}

	masked := maskPAT(c.PAT)
	req.Header.Set("Authorization", "Bearer "+c.PAT)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "duckops-agent/1.0")

	c.log("%s %s (PAT: %s)", method, url, masked)

	resp, err := c.http.Do(req)
	if err != nil {
		c.log("request failed: %v", err)
		return nil, &PlatformError{
			Code: ErrOffline, Message: fmt.Sprintf("cannot reach %s", c.BaseURL),
			Wrapped: err,
		}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return respBody, nil
	}

	return nil, classifyError(fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody)), resp.StatusCode)
}

func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	var pe *PlatformError
	if asPlatformError(err, &pe) {
		switch pe.Code {
		case ErrRateLimited, ErrServerError, ErrOffline, ErrTimeout:
			return true
		case ErrAuthFailed, ErrNotFound, ErrBadRequest, ErrInvalidPayload:
			return false
		}
	}

	return false
}

func computeBackoff(attempt int) time.Duration {
	exp := math.Pow(2, float64(attempt))
	backoff := float64(baseBackoff) * exp
	if backoff > float64(maxBackoff) {
		backoff = float64(maxBackoff)
	}

	jitter := (rand.Float64()*2 - 1) * jitterFactor * backoff
	return time.Duration(backoff + jitter)
}

func maskPAT(pat string) string {
	if len(pat) < 12 {
		return "***"
	}
	return pat[:4] + "****" + pat[len(pat)-4:]
}

// --- API methods ---

func (c *Client) GetServerInfo(ctx context.Context) (*ServerInfoResponse, error) {
	data, err := c.do(ctx, http.MethodGet, "/server/info", nil)
	if err != nil {
		return nil, err
	}
	var resp ServerInfoResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, &PlatformError{Code: ErrInvalidPayload, Message: "failed to parse server info", Wrapped: err}
	}
	return &resp, nil
}

func (c *Client) VerifyPAT(ctx context.Context) error {
	data, err := c.do(ctx, http.MethodGet, "/agents/ping", nil)
	if err != nil {
		return err
	}
	var result struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return &PlatformError{Code: ErrInvalidPayload, Message: "failed to parse ping response", Wrapped: err}
	}
	if result.Status != "success" {
		return &PlatformError{Code: ErrAuthFailed, Message: "PAT verification failed"}
	}
	return nil
}

func (c *Client) RegisterAgent(ctx context.Context, req AgentRegistration) error {
	data, err := c.do(ctx, http.MethodPost, "/agents/register", req)
	if err != nil {
		return c.handleQueue("agent_register", req, err)
	}
	var resp AgentRegisterResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return &PlatformError{Code: ErrInvalidPayload, Message: "failed to parse register response", Wrapped: err}
	}
	c.agentID = resp.Data.Agent.AgentID
	c.log("agent registered: %s (id=%s)", resp.Data.Agent.AgentID, resp.Data.Agent.UUID)
	return nil
}

func (c *Client) Heartbeat(ctx context.Context, req HeartbeatRequest) error {
	_, err := c.do(ctx, http.MethodPost, "/agents/heartbeat", req)
	if err != nil {
		return c.handleQueue("heartbeat", req, err)
	}
	return nil
}

func (c *Client) DisconnectAgent(ctx context.Context, agentID string) error {
	req := DisconnectRequest{
		AgentID: agentID,
		Status:  "disconnected",
	}
	_, err := c.do(ctx, http.MethodPost, "/agents/disconnect", req)
	if err != nil {
		return c.handleQueue("agent_disconnect", req, err)
	}
	return nil
}

func (c *Client) UploadScan(ctx context.Context, scan ScanPayload) error {
	_, err := c.do(ctx, http.MethodPost, "/scans", scan)
	if err != nil {
		return c.handleQueue("scan", scan, err)
	}
	return nil
}

func (c *Client) UploadScansBulk(ctx context.Context, scans []ScanPayload) error {
	if len(scans) == 0 {
		return nil
	}
	req := BulkScansRequest{Scans: scans}
	_, err := c.do(ctx, http.MethodPost, "/bulk/scans", req)
	if err != nil {
		return c.handleQueue("bulk_scans", scans, err)
	}
	return nil
}

func (c *Client) UploadVulnerability(ctx context.Context, vuln VulnerabilityPayload) error {
	_, err := c.do(ctx, http.MethodPost, "/vulnerabilities", vuln)
	if err != nil {
		return c.handleQueue("vulnerability", vuln, err)
	}
	return nil
}

func (c *Client) UploadVulnerabilitiesBulk(ctx context.Context, vulns []VulnerabilityPayload) error {
	if len(vulns) == 0 {
		return nil
	}
	req := BulkVulnerabilitiesRequest{Vulnerabilities: vulns}
	_, err := c.do(ctx, http.MethodPost, "/bulk/vulnerabilities", req)
	if err != nil {
		return c.handleQueue("bulk_vulnerabilities", vulns, err)
	}
	return nil
}

func (c *Client) UploadPipelineEvent(ctx context.Context, event PipelineEventPayload) error {
	_, err := c.do(ctx, http.MethodPost, "/pipelines", event)
	if err != nil {
		return c.handleQueue("pipeline_event", event, err)
	}
	return nil
}

func (c *Client) UploadFinding(ctx context.Context, finding FindingPayload) error {
	_, err := c.do(ctx, http.MethodPost, "/findings", finding)
	if err != nil {
		return c.handleQueue("finding", finding, err)
	}
	return nil
}

func (c *Client) UploadFindingsBulk(ctx context.Context, findings []FindingPayload) error {
	if len(findings) == 0 {
		return nil
	}
	req := BulkFindingsRequest{Findings: findings}
	_, err := c.do(ctx, http.MethodPost, "/bulk/findings", req)
	if err != nil {
		return c.handleQueue("bulk_findings", findings, err)
	}
	return nil
}

func (c *Client) UploadSession(ctx context.Context, session SessionPayload) error {
	_, err := c.do(ctx, http.MethodPost, "/sessions", session)
	if err != nil {
		return c.handleQueue("session", session, err)
	}
	return nil
}

func (c *Client) UploadSessionsBulk(ctx context.Context, sessions []SessionPayload) error {
	if len(sessions) == 0 {
		return nil
	}
	req := BulkSessionsRequest{Sessions: sessions}
	_, err := c.do(ctx, http.MethodPost, "/bulk/sessions", req)
	if err != nil {
		return c.handleQueue("bulk_sessions", sessions, err)
	}
	return nil
}

// --- Agent-triggered scan lifecycle ---

type TriggerScanPayload struct {
	Type        string `json:"type"`
	Environment string `json:"environment"`
	WorkspaceID string `json:"workspace_id,omitempty"`
	ProjectID   string `json:"project_id,omitempty"`
}

type TriggerScanResponse struct {
	Status string `json:"status"`
	Data   struct {
		Scan struct {
			UUID    string `json:"uuid"`
			ScanID  string `json:"scan_id"`
			Type    string `json:"type"`
			Status  string `json:"status"`
		} `json:"scan"`
	} `json:"data"`
}

type UpdateScanPayload struct {
	Status        string `json:"status,omitempty"`
	CurrentStage  string `json:"current_stage,omitempty"`
	SecurityScore int    `json:"security_score,omitempty"`
	StartedAt     string `json:"started_at,omitempty"`
	CompletedAt   string `json:"completed_at,omitempty"`
}

type UploadResultsPayload struct {
	Findings       []FindingPayload       `json:"findings,omitempty"`
	Vulnerabilities []VulnerabilityPayload `json:"vulnerabilities,omitempty"`
}

type UpdateScanResponse struct {
	Status string `json:"status"`
	Data   struct {
		Scan struct {
			UUID    string `json:"uuid"`
			ScanID  string `json:"scan_id"`
			Status  string `json:"status"`
		} `json:"scan"`
	} `json:"data"`
}

func (c *Client) TriggerScan(ctx context.Context, payload TriggerScanPayload) (*TriggerScanResponse, error) {
	data, err := c.do(ctx, http.MethodPost, "/agents/"+c.agentID+"/scans", payload)
	if err != nil {
		return nil, err
	}
	var resp TriggerScanResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, &PlatformError{Code: ErrInvalidPayload, Message: "failed to parse trigger scan response", Wrapped: err}
	}
	return &resp, nil
}

func (c *Client) UpdateScan(ctx context.Context, scanID string, payload UpdateScanPayload) (*UpdateScanResponse, error) {
	data, err := c.do(ctx, http.MethodPatch, "/agents/"+c.agentID+"/scans/"+scanID, payload)
	if err != nil {
		return nil, err
	}
	var resp UpdateScanResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, &PlatformError{Code: ErrInvalidPayload, Message: "failed to parse update scan response", Wrapped: err}
	}
	return &resp, nil
}

func (c *Client) UploadScanResults(ctx context.Context, scanID string, payload UploadResultsPayload) error {
	_, err := c.do(ctx, http.MethodPost, "/agents/"+c.agentID+"/scans/"+scanID+"/results", payload)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) SetAgentID(id string) {
	c.agentID = id
}

func (c *Client) AgentID() string {
	return c.agentID
}

func (c *Client) Queue() *Queue {
	return c.queue
}

func (c *Client) handleQueue(itemType string, payload any, err error) error {
	if IsOffline(err) && c.queue != nil {
		data, marshalErr := json.Marshal(payload)
		if marshalErr != nil {
			return err
		}
		if qErr := c.queue.Enqueue(itemType, data); qErr != nil {
			c.log("failed to enqueue %s: %v", itemType, qErr)
		} else {
			c.log("enqueued %s for later delivery", itemType)
		}
	}
	return err
}
