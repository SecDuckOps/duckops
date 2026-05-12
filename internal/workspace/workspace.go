// Package workspace defines the Workspace interface used by all
// frontends (TUI, CLI) to interact with a running workspace. Two
// implementations exist: one wrapping a local app.App instance and one
// wrapping the HTTP client SDK.
package workspace

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/catwalk/pkg/catwalk"
	mcptools "github.com/SecDuckOps/duckops/internal/agent/tools/mcp"
	"github.com/SecDuckOps/duckops/internal/config"
	"github.com/SecDuckOps/duckops/internal/history"
	"github.com/SecDuckOps/duckops/internal/lsp"
	"github.com/SecDuckOps/duckops/internal/message"
	"github.com/SecDuckOps/duckops/internal/oauth"
	"github.com/SecDuckOps/duckops/internal/permission"
	"github.com/SecDuckOps/duckops/internal/session"
)

// LSPClientInfo holds information about an LSP client's state. This is
// the frontend-facing type; implementations translate from the
// underlying app or proto representation.
type LSPClientInfo struct {
	Name            string
	State           lsp.ServerState
	Error           error
	DiagnosticCount int
	ConnectedAt     time.Time
}

// LSPEventType represents the type of LSP event.
type LSPEventType string

const (
	LSPEventStateChanged       LSPEventType = "state_changed"
	LSPEventDiagnosticsChanged LSPEventType = "diagnostics_changed"
)

// LSPEvent represents an LSP event forwarded to the TUI.
type LSPEvent struct {
	Type            LSPEventType
	Name            string
	State           lsp.ServerState
	Error           error
	DiagnosticCount int
}

// AgentModel holds the model information exposed to the UI.
type AgentModel struct {
	CatwalkCfg catwalk.Model
	ModelCfg   config.SelectedModel
}

// Workspace is the main abstraction consumed by the TUI and CLI. It
// groups every operation a frontend needs to perform against a running
// workspace, regardless of whether the workspace is in-process or
// remote.
type Workspace interface {
	// Sessions
	CreateSession(ctx context.Context, title string) (session.Session, error)
	GetSession(ctx context.Context, sessionID string) (session.Session, error)
	ListSessions(ctx context.Context) ([]session.Session, error)
	SaveSession(ctx context.Context, sess session.Session) (session.Session, error)
	DeleteSession(ctx context.Context, sessionID string) error
	CreateAgentToolSessionID(messageID, toolCallID string) string
	ParseAgentToolSessionID(sessionID string) (messageID string, toolCallID string, ok bool)

	// Messages
	ListMessages(ctx context.Context, sessionID string) ([]message.Message, error)
	ListUserMessages(ctx context.Context, sessionID string) ([]message.Message, error)
	ListAllUserMessages(ctx context.Context) ([]message.Message, error)

	// Agent
	AgentRun(ctx context.Context, sessionID, prompt string, attachments ...message.Attachment) error
	AgentCancel(sessionID string)
	AgentIsBusy() bool
	AgentIsSessionBusy(sessionID string) bool
	AgentModel() AgentModel
	AgentIsReady() bool
	AgentQueuedPrompts(sessionID string) int
	AgentQueuedPromptsList(sessionID string) []string
	AgentClearQueue(sessionID string)
	AgentSummarize(ctx context.Context, sessionID string) error
	UpdateAgentModel(ctx context.Context) error
	InitCoderAgent(ctx context.Context) error
	GetDefaultSmallModel(providerID string) config.SelectedModel

	// Permissions
	PermissionGrant(perm permission.PermissionRequest)
	PermissionGrantPersistent(perm permission.PermissionRequest)
	PermissionDeny(perm permission.PermissionRequest)
	PermissionSkipRequests() bool
	PermissionSetSkipRequests(skip bool)

	// FileTracker
	FileTrackerRecordRead(ctx context.Context, sessionID, path string)
	FileTrackerLastReadTime(ctx context.Context, sessionID, path string) time.Time
	FileTrackerListReadFiles(ctx context.Context, sessionID string) ([]string, error)

	// History
	ListSessionHistory(ctx context.Context, sessionID string) ([]history.File, error)

	// LSP
	LSPStart(ctx context.Context, path string)
	LSPStopAll(ctx context.Context)
	LSPGetStates() map[string]LSPClientInfo
	LSPGetDiagnosticCounts(name string) lsp.DiagnosticCounts

	// Config (read-only data)
	Config() *config.Config
	WorkingDir() string
	Resolver() config.VariableResolver

	// Config mutations (proxied to server in client mode)
	UpdatePreferredModel(scope config.Scope, modelType config.SelectedModelType, model config.SelectedModel) error
	SetCompactMode(scope config.Scope, enabled bool) error
	SetProviderAPIKey(scope config.Scope, providerID string, apiKey any) error
	SetConfigField(scope config.Scope, key string, value any) error
	RemoveConfigField(scope config.Scope, key string) error
	ImportCopilot() (*oauth.Token, bool)
	RefreshOAuthToken(ctx context.Context, scope config.Scope, providerID string) error

	// Project lifecycle
	ProjectNeedsInitialization() (bool, error)
	MarkProjectInitialized() error
	InitializePrompt() (string, error)

	// MCP operations (server-side in client mode)
	MCPGetStates() map[string]mcptools.ClientInfo
	MCPRefreshPrompts(ctx context.Context, name string)
	MCPRefreshResources(ctx context.Context, name string)
	RefreshMCPTools(ctx context.Context, name string)
	ReadMCPResource(ctx context.Context, name, uri string) ([]MCPResourceContents, error)
	GetMCPPrompt(clientID, promptID string, args map[string]string) (string, error)
	EnableDockerMCP(ctx context.Context) error
	DisableDockerMCP() error

	// Events
	Subscribe(program *tea.Program)
	Shutdown()

	// GraphX - Security Knowledge Graph
	GraphXInit(ctx context.Context) error
	GraphXGetStatus() GraphXStatus
	GraphXGetContext(query string, securityFocused bool) (*GraphXContext, error)
	GraphXGetThreatModel() (*GraphXThreatModel, error)
	GraphXGetBlastRadius(nodeID string) (*GraphXBlastResult, error)
	GraphXSearchNodes(name string) []GraphXNode
	GraphXWatch(enabled bool) error
	GraphXShutdown()
}

type GraphXStatus struct {
	State        string `json:"state"`
	NodesCount   int    `json:"nodes_count"`
	EdgesCount   int    `json:"edges_count"`
	WatchEnabled bool   `json:"watch_enabled"`
	LastUpdated  string `json:"last_updated"`
}

type GraphXContext struct {
	Summary    string     `json:"summary"`
	FilePaths  []string   `json:"file_paths"`
	Nodes      []GraphXNode `json:"nodes"`
	TokenCount int        `json:"token_count"`
}

type GraphXNode struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	FilePath  string   `json:"file_path"`
	RiskScore float64  `json:"risk_score"`
	Tags      []string `json:"tags"`
}

type GraphXThreatModel struct {
	Summary       *ThreatSummary `json:"summary"`
	EntryPoints   []GraphXNode   `json:"entry_points"`
	TrustBoundaries []string     `json:"trust_boundaries"`
	AttackPaths   []AttackPath   `json:"attack_paths"`
	ExecutionFlows []ExecutionFlow `json:"execution_flows"`
}

type ThreatSummary struct {
	Spoofing   int `json:"spoofing"`
	Tampering  int `json:"tampering"`
	Repudiation int `json:"repudiation"`
	Disclosure int `json:"disclosure"`
	DoS        int `json:"dos"`
	EoP        int `json:"eop"`
	Total      int `json:"total"`
}

type AttackPath struct {
	Source     string   `json:"source"`
	Target     string   `json:"target"`
	Path       []string `json:"path"`
	Complexity string   `json:"complexity"`
	Impact     string   `json:"impact"`
}

type ExecutionFlow struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	EntryPoint    string     `json:"entry_point"`
	Steps         []FlowStep `json:"steps"`
	AuthRequired  bool       `json:"auth_required"`
	HandlesPII    bool       `json:"handles_pii"`
	RiskLevel     string     `json:"risk_level"`
}

type FlowStep struct {
	NodeID   string `json:"node_id"`
	NodeName string `json:"node_name"`
	Type     string `json:"type"`
	Action   string `json:"action"`
}

type GraphXBlastResult struct {
	AffectedNodes  []GraphXNode `json:"affected_nodes"`
	AuthFlows      []string     `json:"auth_flows"`
	ImpactedAPIs   []string     `json:"impacted_apis"`
	RiskScore      float64      `json:"risk_score"`
	CriticalPath   []string     `json:"critical_path"`
}

// MCPResourceContents holds the contents of an MCP resource.
type MCPResourceContents struct {
	URI      string `json:"uri"`
	MIMEType string `json:"mime_type,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     []byte `json:"blob,omitempty"`
}
