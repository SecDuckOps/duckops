package platform

import "time"

type AgentRegistration struct {
	AgentID      string   `json:"agent_id"`
	Hostname     string   `json:"hostname"`
	Version      string   `json:"version"`
	Platform     string   `json:"platform"`
	Capabilities []string `json:"capabilities"`
}

type HeartbeatRequest struct {
	AgentID        string   `json:"agent_id"`
	CPU            float64  `json:"cpu,omitempty"`
	Memory         float64  `json:"memory,omitempty"`
	ActivePipelines []string `json:"active_pipelines,omitempty"`
	Status         string   `json:"status"`
}

type DisconnectRequest struct {
	AgentID string `json:"agent_id"`
	Status  string `json:"status,omitempty"`
}

type ScanPayload struct {
	WorkspaceID string `json:"workspace_id,omitempty"`
	ProjectID   string `json:"project_id,omitempty"`
	ScanID      string `json:"scan_id"`
	Type        string `json:"type"`
	Environment string `json:"environment"`
	Status      string `json:"status"`
	StartedAt   string `json:"started_at,omitempty"`
	CompletedAt string `json:"completed_at,omitempty"`
}

type VulnerabilityPayload struct {
	ScanID       string  `json:"scan_id,omitempty"`
	WorkspaceID  string  `json:"workspace_id,omitempty"`
	ProjectID    string  `json:"project_id,omitempty"`
	CVEID        string  `json:"cve_id,omitempty"`
	Title        string  `json:"title"`
	Severity     string  `json:"severity"`
	CVSSScore    float64 `json:"cvss_score,omitempty"`
	PackageName  string  `json:"package_name,omitempty"`
	PackageVer   string  `json:"package_version,omitempty"`
	FixedVersion string  `json:"fixed_version,omitempty"`
	SourceTool   string  `json:"source_tool,omitempty"`
	Status       string  `json:"status,omitempty"`
	DetectedAt   string  `json:"detected_at,omitempty"`
}

type FindingPayload struct {
	ScanID       string `json:"scan_id,omitempty"`
	WorkspaceID  string `json:"workspace_id,omitempty"`
	ProjectID    string `json:"project_id,omitempty"`
	Tool         string `json:"tool"`
	Severity     string `json:"severity"`
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	FilePath     string `json:"file_path,omitempty"`
	LineStart    int    `json:"line_start,omitempty"`
	LineEnd      int    `json:"line_end,omitempty"`
	RuleID       string `json:"rule_id,omitempty"`
	CWEID        string `json:"cwe_id,omitempty"`
	CVEID        string `json:"cve_id,omitempty"`
	CVSSScore    float64 `json:"cvss_score,omitempty"`
	PackageName  string `json:"package_name,omitempty"`
	PackageVer   string `json:"package_version,omitempty"`
	FixedVersion string `json:"fixed_version,omitempty"`
	SourceTool   string `json:"source_tool,omitempty"`
	RawDetails   any    `json:"raw_details,omitempty"`
}

type PipelineEventPayload struct {
	EventType    string `json:"event_type"`
	PipelineID   string `json:"pipeline_id,omitempty"`
	WorkspaceID  string `json:"workspace_id,omitempty"`
	ProjectID    string `json:"project_id,omitempty"`
	BranchName   string `json:"branch_name,omitempty"`
	CommitHash   string `json:"commit_hash,omitempty"`
	Status       string `json:"status"`
	DurationSec  int    `json:"duration_seconds,omitempty"`
	TriggerType  string `json:"trigger_type,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
	Timestamp    string `json:"timestamp"`
}

type BulkScansRequest struct {
	Scans []ScanPayload `json:"scans"`
}

type BulkVulnerabilitiesRequest struct {
	Vulnerabilities []VulnerabilityPayload `json:"vulnerabilities"`
}

type BulkFindingsRequest struct {
	Findings []FindingPayload `json:"findings"`
}

type SessionPayload struct {
	SessionID        string `json:"session_id"`
	ParentSessionID  string `json:"parent_session_id,omitempty"`
	Title            string `json:"title"`
	MessageCount     int64  `json:"message_count"`
	PromptTokens     int64  `json:"prompt_tokens"`
	CompletionTokens int64  `json:"completion_tokens"`
	Cost             float64 `json:"cost"`
	SummaryMessageID string `json:"summary_message_id,omitempty"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

type BulkSessionsRequest struct {
	Sessions []SessionPayload `json:"sessions"`
}

type ServerInfoResponse struct {
	Status string `json:"status"`
	Data   struct {
		Version   string            `json:"version"`
		Endpoints map[string]string `json:"endpoints"`
	} `json:"data"`
}

type AgentRegisterResponse struct {
	Status string `json:"status"`
	Data   struct {
		Agent struct {
			UUID      string `json:"uuid"`
			AgentID   string `json:"agent_id"`
			Status    string `json:"status"`
		} `json:"agent"`
	} `json:"data"`
}

type QueueItem struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Payload   []byte    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
	Retries   int       `json:"retries"`
}

const (
	EventPipelineCreated   = "PipelineCreated"
	EventPipelineStarted   = "PipelineStarted"
	EventPipelineRunning   = "PipelineRunning"
	EventPipelineSucceeded = "PipelineSucceeded"
	EventPipelineFailed    = "PipelineFailed"
	EventPipelineCancelled = "PipelineCancelled"
	EventPipelineTimedOut  = "PipelineTimedOut"
)
