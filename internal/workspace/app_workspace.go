package workspace

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/SecDuckOps/duckops/internal/agent"
	mcptools "github.com/SecDuckOps/duckops/internal/agent/tools/mcp"
	"github.com/SecDuckOps/duckops/internal/app"
	"github.com/SecDuckOps/duckops/internal/commands"
	"github.com/SecDuckOps/duckops/internal/config"
	"github.com/SecDuckOps/duckops/internal/graphx"
	ctxengine "github.com/SecDuckOps/duckops/internal/graphx/engine/context"
	"github.com/SecDuckOps/duckops/internal/history"
	"github.com/SecDuckOps/duckops/internal/lsp"
	"github.com/SecDuckOps/duckops/internal/message"
	"github.com/SecDuckOps/duckops/internal/oauth"
	"github.com/SecDuckOps/duckops/internal/permission"
	"github.com/SecDuckOps/duckops/internal/session"
)

// AppWorkspace implements the Workspace interface by delegating
// directly to an in-process [app.App] instance.
type AppWorkspace struct {
	app   *app.App
	store *config.ConfigStore
	graphx *graphx.Engine
}

// NewAppWorkspace creates a new AppWorkspace wrapping the given app
// and config store.
func NewAppWorkspace(a *app.App, store *config.ConfigStore) *AppWorkspace {
	return &AppWorkspace{
		app:   a,
		store: store,
	}
}

// -- Sessions --

func (w *AppWorkspace) CreateSession(ctx context.Context, title string) (session.Session, error) {
	return w.app.Sessions.Create(ctx, title)
}

func (w *AppWorkspace) GetSession(ctx context.Context, sessionID string) (session.Session, error) {
	return w.app.Sessions.Get(ctx, sessionID)
}

func (w *AppWorkspace) ListSessions(ctx context.Context) ([]session.Session, error) {
	return w.app.Sessions.List(ctx)
}

func (w *AppWorkspace) SaveSession(ctx context.Context, sess session.Session) (session.Session, error) {
	return w.app.Sessions.Save(ctx, sess)
}

func (w *AppWorkspace) DeleteSession(ctx context.Context, sessionID string) error {
	return w.app.Sessions.Delete(ctx, sessionID)
}

func (w *AppWorkspace) CreateAgentToolSessionID(messageID, toolCallID string) string {
	return w.app.Sessions.CreateAgentToolSessionID(messageID, toolCallID)
}

func (w *AppWorkspace) ParseAgentToolSessionID(sessionID string) (string, string, bool) {
	return w.app.Sessions.ParseAgentToolSessionID(sessionID)
}

// -- Messages --

func (w *AppWorkspace) ListMessages(ctx context.Context, sessionID string) ([]message.Message, error) {
	return w.app.Messages.List(ctx, sessionID)
}

func (w *AppWorkspace) ListUserMessages(ctx context.Context, sessionID string) ([]message.Message, error) {
	return w.app.Messages.ListUserMessages(ctx, sessionID)
}

func (w *AppWorkspace) ListAllUserMessages(ctx context.Context) ([]message.Message, error) {
	return w.app.Messages.ListAllUserMessages(ctx)
}

// -- Agent --

func (w *AppWorkspace) AgentRun(ctx context.Context, sessionID, prompt string, attachments ...message.Attachment) error {
	if w.app.AgentCoordinator == nil {
		return errors.New("agent coordinator not initialized")
	}
	_, err := w.app.AgentCoordinator.Run(ctx, sessionID, prompt, attachments...)
	return err
}

func (w *AppWorkspace) AgentCancel(sessionID string) {
	if w.app.AgentCoordinator != nil {
		w.app.AgentCoordinator.Cancel(sessionID)
	}
}

func (w *AppWorkspace) AgentIsBusy() bool {
	if w.app.AgentCoordinator == nil {
		return false
	}
	return w.app.AgentCoordinator.IsBusy()
}

func (w *AppWorkspace) AgentIsSessionBusy(sessionID string) bool {
	if w.app.AgentCoordinator == nil {
		return false
	}
	return w.app.AgentCoordinator.IsSessionBusy(sessionID)
}

func (w *AppWorkspace) AgentModel() AgentModel {
	if w.app.AgentCoordinator == nil {
		return AgentModel{}
	}
	m := w.app.AgentCoordinator.Model()
	return AgentModel{
		CatwalkCfg: m.CatwalkCfg,
		ModelCfg:   m.ModelCfg,
	}
}

func (w *AppWorkspace) AgentIsReady() bool {
	return w.app.AgentCoordinator != nil
}

func (w *AppWorkspace) AgentQueuedPrompts(sessionID string) int {
	if w.app.AgentCoordinator == nil {
		return 0
	}
	return w.app.AgentCoordinator.QueuedPrompts(sessionID)
}

func (w *AppWorkspace) AgentQueuedPromptsList(sessionID string) []string {
	if w.app.AgentCoordinator == nil {
		return nil
	}
	return w.app.AgentCoordinator.QueuedPromptsList(sessionID)
}

func (w *AppWorkspace) AgentClearQueue(sessionID string) {
	if w.app.AgentCoordinator != nil {
		w.app.AgentCoordinator.ClearQueue(sessionID)
	}
}

func (w *AppWorkspace) AgentSummarize(ctx context.Context, sessionID string) error {
	if w.app.AgentCoordinator == nil {
		return errors.New("agent coordinator not initialized")
	}
	return w.app.AgentCoordinator.Summarize(ctx, sessionID)
}

func (w *AppWorkspace) UpdateAgentModel(ctx context.Context) error {
	return w.app.UpdateAgentModel(ctx)
}

func (w *AppWorkspace) InitCoderAgent(ctx context.Context) error {
	return w.app.InitCoderAgent(ctx)
}

func (w *AppWorkspace) GetDefaultSmallModel(providerID string) config.SelectedModel {
	return w.app.GetDefaultSmallModel(providerID)
}

// -- Permissions --

func (w *AppWorkspace) PermissionGrant(perm permission.PermissionRequest) {
	w.app.Permissions.Grant(perm)
}

func (w *AppWorkspace) PermissionGrantPersistent(perm permission.PermissionRequest) {
	w.app.Permissions.GrantPersistent(perm)
}

func (w *AppWorkspace) PermissionDeny(perm permission.PermissionRequest) {
	w.app.Permissions.Deny(perm)
}

func (w *AppWorkspace) PermissionSkipRequests() bool {
	return w.app.Permissions.SkipRequests()
}

func (w *AppWorkspace) PermissionSetSkipRequests(skip bool) {
	w.app.Permissions.SetSkipRequests(skip)
}

// -- FileTracker --

func (w *AppWorkspace) FileTrackerRecordRead(ctx context.Context, sessionID, path string) {
	w.app.FileTracker.RecordRead(ctx, sessionID, path)
}

func (w *AppWorkspace) FileTrackerLastReadTime(ctx context.Context, sessionID, path string) time.Time {
	return w.app.FileTracker.LastReadTime(ctx, sessionID, path)
}

func (w *AppWorkspace) FileTrackerListReadFiles(ctx context.Context, sessionID string) ([]string, error) {
	return w.app.FileTracker.ListReadFiles(ctx, sessionID)
}

// -- History --

func (w *AppWorkspace) ListSessionHistory(ctx context.Context, sessionID string) ([]history.File, error) {
	return w.app.History.ListBySession(ctx, sessionID)
}

// -- LSP --

func (w *AppWorkspace) LSPStart(ctx context.Context, path string) {
	w.app.LSPManager.Start(ctx, path)
}

func (w *AppWorkspace) LSPStopAll(ctx context.Context) {
	w.app.LSPManager.StopAll(ctx)
}

func (w *AppWorkspace) LSPGetStates() map[string]LSPClientInfo {
	states := app.GetLSPStates()
	result := make(map[string]LSPClientInfo, len(states))
	for k, v := range states {
		result[k] = LSPClientInfo{
			Name:            v.Name,
			State:           v.State,
			Error:           v.Error,
			DiagnosticCount: v.DiagnosticCount,
			ConnectedAt:     v.ConnectedAt,
		}
	}
	return result
}

func (w *AppWorkspace) LSPGetDiagnosticCounts(name string) lsp.DiagnosticCounts {
	state, ok := app.GetLSPState(name)
	if !ok || state.Client == nil {
		return lsp.DiagnosticCounts{}
	}
	return state.Client.GetDiagnosticCounts()
}

// -- Config (read-only) --

func (w *AppWorkspace) Config() *config.Config {
	return w.store.Config()
}

func (w *AppWorkspace) WorkingDir() string {
	return w.store.WorkingDir()
}

func (w *AppWorkspace) Resolver() config.VariableResolver {
	return w.store.Resolver()
}

// -- Config mutations --

func (w *AppWorkspace) UpdatePreferredModel(scope config.Scope, modelType config.SelectedModelType, model config.SelectedModel) error {
	return w.store.UpdatePreferredModel(scope, modelType, model)
}

func (w *AppWorkspace) SetCompactMode(scope config.Scope, enabled bool) error {
	return w.store.SetCompactMode(scope, enabled)
}

func (w *AppWorkspace) SetProviderAPIKey(scope config.Scope, providerID string, apiKey any) error {
	return w.store.SetProviderAPIKey(scope, providerID, apiKey)
}

func (w *AppWorkspace) SetConfigField(scope config.Scope, key string, value any) error {
	return w.store.SetConfigField(scope, key, value)
}

func (w *AppWorkspace) RemoveConfigField(scope config.Scope, key string) error {
	return w.store.RemoveConfigField(scope, key)
}

func (w *AppWorkspace) ImportCopilot() (*oauth.Token, bool) {
	return w.store.ImportCopilot()
}

func (w *AppWorkspace) RefreshOAuthToken(ctx context.Context, scope config.Scope, providerID string) error {
	return w.store.RefreshOAuthToken(ctx, scope, providerID)
}

// -- Project lifecycle --

func (w *AppWorkspace) ProjectNeedsInitialization() (bool, error) {
	return config.ProjectNeedsInitialization(w.store)
}

func (w *AppWorkspace) MarkProjectInitialized() error {
	return config.MarkProjectInitialized(w.store)
}

func (w *AppWorkspace) InitializePrompt() (string, error) {
	return agent.InitializePrompt(w.store)
}

// -- MCP operations --

func (w *AppWorkspace) MCPGetStates() map[string]mcptools.ClientInfo {
	return mcptools.GetStates()
}

func (w *AppWorkspace) MCPRefreshPrompts(ctx context.Context, name string) {
	mcptools.RefreshPrompts(ctx, name)
}

func (w *AppWorkspace) MCPRefreshResources(ctx context.Context, name string) {
	mcptools.RefreshResources(ctx, name)
}

func (w *AppWorkspace) RefreshMCPTools(ctx context.Context, name string) {
	mcptools.RefreshTools(ctx, w.store, name)
}

func (w *AppWorkspace) ReadMCPResource(ctx context.Context, name, uri string) ([]MCPResourceContents, error) {
	contents, err := mcptools.ReadResource(ctx, w.store, name, uri)
	if err != nil {
		return nil, err
	}
	result := make([]MCPResourceContents, len(contents))
	for i, c := range contents {
		result[i] = MCPResourceContents{
			URI:      c.URI,
			MIMEType: c.MIMEType,
			Text:     c.Text,
			Blob:     c.Blob,
		}
	}
	return result, nil
}

func (w *AppWorkspace) GetMCPPrompt(clientID, promptID string, args map[string]string) (string, error) {
	return commands.GetMCPPrompt(w.store, clientID, promptID, args)
}

func (w *AppWorkspace) EnableDockerMCP(ctx context.Context) error {
	mcpConfig, err := w.store.PrepareDockerMCPConfig()
	if err != nil {
		return err
	}

	if err := mcptools.InitializeSingle(ctx, config.DockerMCPName, w.store); err != nil {
		disableErr := mcptools.DisableSingle(w.store, config.DockerMCPName)
		delete(w.store.Config().MCP, config.DockerMCPName)
		return fmt.Errorf("failed to start docker MCP: %w", errors.Join(err, disableErr))
	}

	if err := w.store.PersistDockerMCPConfig(mcpConfig); err != nil {
		disableErr := mcptools.DisableSingle(w.store, config.DockerMCPName)
		delete(w.store.Config().MCP, config.DockerMCPName)
		return fmt.Errorf("docker MCP started but failed to persist configuration: %w", errors.Join(err, disableErr))
	}

	return nil
}

func (w *AppWorkspace) DisableDockerMCP() error {
	if err := mcptools.DisableSingle(w.store, config.DockerMCPName); err != nil {
		return fmt.Errorf("failed to disable docker MCP: %w", err)
	}
	return w.store.DisableDockerMCP()
}

// -- Lifecycle --

func (w *AppWorkspace) Subscribe(program *tea.Program) {
	w.app.Subscribe(program)
}

func (w *AppWorkspace) Shutdown() {
	w.app.Shutdown()
}

// App returns the underlying app.App instance.
func (w *AppWorkspace) App() *app.App {
	return w.app
}

// Store returns the underlying config store.
func (w *AppWorkspace) Store() *config.ConfigStore {
	return w.store
}

func (w *AppWorkspace) initGraphX(ctx context.Context) error {
	if w.graphx != nil {
		return nil
	}

	repoPath := w.store.WorkingDir()
	repoID := filepath.Base(repoPath)
	storagePath := filepath.Join(repoPath, ".graphx", "graph.json")

	cfg := graphx.Config{
		RepoID:       repoID,
		RepoPath:     repoPath,
		StoragePath:  storagePath,
		WatchEnabled: false,
	}

	engine, err := graphx.NewEngine(cfg)
	if err != nil {
		return err
	}

	if err := engine.Initialize(ctx); err != nil {
		return err
	}

	w.graphx = engine
	return nil
}

func (w *AppWorkspace) GraphXInit(ctx context.Context) error {
	return w.initGraphX(ctx)
}

func (w *AppWorkspace) GraphXGetStatus() GraphXStatus {
	if w.graphx == nil {
		return GraphXStatus{State: "not_initialized"}
	}

	status := w.graphx.GetStatus()
	return GraphXStatus{
		State:        status.State,
		NodesCount:   status.NodesCount,
		EdgesCount:   status.EdgesCount,
		WatchEnabled: status.WatchEnabled,
		LastUpdated:  status.LastUpdated.Format(time.RFC3339),
	}
}

func (w *AppWorkspace) GraphXGetContext(query string, securityFocused bool) (*GraphXContext, error) {
	if w.graphx == nil {
		if err := w.initGraphX(context.Background()); err != nil {
			return nil, err
		}
	}

	req := &ctxengine.ContextRequest{
		RepoID:          filepath.Base(w.store.WorkingDir()),
		Query:           query,
		SecurityFocused: securityFocused,
		MaxFiles:        20,
	}

	resp, err := w.graphx.GetContext(req)
	if err != nil {
		return nil, err
	}

	nodes := make([]GraphXNode, len(resp.Nodes))
	for i, n := range resp.Nodes {
		tags := []string{}
		if n.Metadata != nil {
			if t, ok := n.Metadata["security_tags"]; ok {
				if ts, ok := t.([]string); ok {
					tags = ts
				}
			}
		}
		nodes[i] = GraphXNode{
			ID:        n.ID,
			Name:      n.Name,
			Type:      string(n.Type),
			FilePath:  n.FilePath,
			RiskScore: n.RiskScore,
			Tags:      tags,
		}
	}

	return &GraphXContext{
		Summary:    resp.Summary,
		FilePaths:  resp.FilePaths,
		Nodes:      nodes,
		TokenCount: resp.TokenCount,
	}, nil
}

func (w *AppWorkspace) GraphXGetThreatModel() (*GraphXThreatModel, error) {
	if w.graphx == nil {
		if err := w.initGraphX(context.Background()); err != nil {
			return nil, err
		}
	}

	model, err := w.graphx.GetThreatModel()
	if err != nil {
		return nil, err
	}

	entryPoints := make([]GraphXNode, len(model.EntryPoints))
	for i, ep := range model.EntryPoints {
		entryPoints[i] = GraphXNode{
			ID:       ep.ID,
			Name:     ep.Name,
			Type:     string(ep.Type),
			FilePath: ep.FilePath,
			RiskScore: ep.RiskScore,
		}
	}

	attackPaths := make([]AttackPath, len(model.AttackPaths))
	for i, ap := range model.AttackPaths {
		attackPaths[i] = AttackPath{
			Source:     ap.Source,
			Target:     ap.Target,
			Path:       ap.Path,
			Complexity: ap.Complexity,
			Impact:     ap.Impact,
		}
	}

	executionFlows := make([]ExecutionFlow, len(model.ExecutionFlows))
	for i, ef := range model.ExecutionFlows {
		steps := make([]FlowStep, len(ef.Steps))
		for j, s := range ef.Steps {
			steps[j] = FlowStep{
				NodeID:   s.NodeID,
				NodeName: s.NodeName,
				Type:     s.Type,
				Action:   s.Action,
			}
		}
		executionFlows[i] = ExecutionFlow{
			ID:           ef.ID,
			Name:         ef.Name,
			EntryPoint:   ef.EntryPoint,
			Steps:        steps,
			AuthRequired: ef.AuthRequired,
			HandlesPII:   ef.HandlesPII,
			RiskLevel:    ef.RiskLevel,
		}
	}

	var summary *ThreatSummary
	if model.Summary != nil {
		summary = &ThreatSummary{
			Spoofing:   model.Summary.Spoofing,
			Tampering:  model.Summary.Tampering,
			Repudiation: model.Summary.Repudiation,
			Disclosure: model.Summary.InformationDisclosure,
			DoS:        model.Summary.DenialOfService,
			EoP:        model.Summary.ElevationOfPrivilege,
			Total:      model.Summary.Total,
		}
	}

	return &GraphXThreatModel{
		Summary:        summary,
		EntryPoints:    entryPoints,
		TrustBoundaries: model.TrustBoundaries,
		AttackPaths:    attackPaths,
		ExecutionFlows: executionFlows,
	}, nil
}

func (w *AppWorkspace) GraphXGetBlastRadius(nodeID string) (*GraphXBlastResult, error) {
	if w.graphx == nil {
		if err := w.initGraphX(context.Background()); err != nil {
			return nil, err
		}
	}

	result, err := w.graphx.GetBlastRadius(nodeID)
	if err != nil {
		return nil, err
	}

	affectedNodes := make([]GraphXNode, len(result.AffectedNodes))
	for i, n := range result.AffectedNodes {
		affectedNodes[i] = GraphXNode{
			ID:        n.ID,
			Name:      n.Name,
			Type:      string(n.Type),
			FilePath:  n.FilePath,
			RiskScore: n.RiskScore,
		}
	}

	return &GraphXBlastResult{
		AffectedNodes: affectedNodes,
		AuthFlows:     result.AuthFlows,
		ImpactedAPIs:  result.ImpactedAPIs,
		RiskScore:     result.RiskScore,
		CriticalPath:  result.CriticalPath,
	}, nil
}

func (w *AppWorkspace) GraphXSearchNodes(name string) []GraphXNode {
	if w.graphx == nil {
		return nil
	}

	nodes := w.graphx.SearchByName(name)
	result := make([]GraphXNode, len(nodes))
	for i, n := range nodes {
		result[i] = GraphXNode{
			ID:        n.ID,
			Name:      n.Name,
			Type:      string(n.Type),
			FilePath:  n.FilePath,
			RiskScore: n.RiskScore,
		}
	}
	return result
}

func (w *AppWorkspace) GraphXWatch(enabled bool) error {
	if w.graphx == nil {
		return errors.New("graphx not initialized")
	}

	if enabled {
		return w.graphx.EnableWatchMode(context.Background())
	}

	w.graphx.Shutdown()
	return nil
}

func (w *AppWorkspace) GraphXShutdown() {
	if w.graphx != nil {
		w.graphx.Shutdown()
		w.graphx = nil
	}
}

// Compile-time check that AppWorkspace implements Workspace.
var _ Workspace = (*AppWorkspace)(nil)
