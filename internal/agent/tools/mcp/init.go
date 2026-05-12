// Package mcp provides functionality for managing Model Context Protocol (MCP)
// clients within the DuckOps application.
package mcp

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/SecDuckOps/duckops/internal/config"
	"github.com/SecDuckOps/duckops/internal/csync"
	"github.com/SecDuckOps/duckops/internal/home"
	"github.com/SecDuckOps/duckops/internal/permission"
	"github.com/SecDuckOps/duckops/internal/pubsub"
	"github.com/SecDuckOps/duckops/internal/version"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func parseLevel(level mcp.LoggingLevel) slog.Level {
	switch level {
	case "info":
		return slog.LevelInfo
	case "notice":
		return slog.LevelInfo
	case "warning":
		return slog.LevelWarn
	default:
		return slog.LevelDebug
	}
}

// ClientSession wraps an mcp.ClientSession with a context cancel function so
// that the context created during session establishment is properly cleaned up
// on close.
type ClientSession struct {
	*mcp.ClientSession
	cancel context.CancelFunc
}

// Close cancels the session context and then closes the underlying session.
func (s *ClientSession) Close() error {
	s.cancel()
	return s.ClientSession.Close()
}

var (
	sessions = csync.NewMap[string, *ClientSession]()
	states   = csync.NewMap[string, ClientInfo]()
	broker   = pubsub.NewBroker[Event]()
	initOnce sync.Once
	initDone = make(chan struct{})
)

// State represents the current state of an MCP client
type State int

const (
	StateDisabled State = iota
	StateStarting
	StateConnected
	StateError
)

func (s State) String() string {
	switch s {
	case StateDisabled:
		return "disabled"
	case StateStarting:
		return "starting"
	case StateConnected:
		return "connected"
	case StateError:
		return "error"
	default:
		return "unknown"
	}
}

// EventType represents the type of MCP event
type EventType uint

const (
	EventStateChanged EventType = iota
	EventToolsListChanged
	EventPromptsListChanged
	EventResourcesListChanged
)

// Event represents an event in the MCP system
type Event struct {
	Type   EventType
	Name   string
	State  State
	Error  error
	Counts Counts
}

// Counts number of available tools, prompts, etc.
type Counts struct {
	Tools     int
	Prompts   int
	Resources int
}

// ClientInfo holds information about an MCP client's state
type ClientInfo struct {
	Name        string
	State       State
	Error       error
	Client      *ClientSession
	Counts      Counts
	ConnectedAt time.Time
}

// SubscribeEvents returns a channel for MCP events
func SubscribeEvents(ctx context.Context) <-chan pubsub.Event[Event] {
	return broker.Subscribe(ctx)
}

// GetStates returns the current state of all MCP clients
func GetStates() map[string]ClientInfo {
	return states.Copy()
}

// GetState returns the state of a specific MCP client
func GetState(name string) (ClientInfo, bool) {
	return states.Get(name)
}

// Close closes all MCP clients. This should be called during application shutdown.
func Close(ctx context.Context) error {
	var wg sync.WaitGroup
	for name, session := range sessions.Seq2() {
		wg.Go(func() {
			done := make(chan error, 1)
			go func() {
				done <- session.Close()
			}()
			select {
			case err := <-done:
				if err != nil &&
					!errors.Is(err, io.EOF) &&
					!errors.Is(err, context.Canceled) &&
					err.Error() != "signal: killed" {
					slog.Warn("Failed to shutdown MCP client", "name", name, "error", err)
				}
			case <-ctx.Done():
			}
		})
	}
	wg.Wait()
	broker.Shutdown()
	return nil
}

// Initialize initializes MCP clients based on the provided configuration.
func Initialize(ctx context.Context, permissions permission.Service, cfg *config.ConfigStore) {
	slog.Info("Initializing MCP clients")
	var wg sync.WaitGroup
	// Initialize states for all configured MCPs
	for name, m := range cfg.Config().MCP {
		if m.Disabled {
			updateState(name, StateDisabled, nil, nil, Counts{})
			slog.Debug("Skipping disabled MCP", "name", name)
			continue
		}

		// Set initial starting state
		wg.Add(1)
		go func(name string, m config.MCPConfig) {
			defer func() {
				wg.Done()
				if r := recover(); r != nil {
					var err error
					switch v := r.(type) {
					case error:
						err = v
					case string:
						err = fmt.Errorf("panic: %s", v)
					default:
						err = fmt.Errorf("panic: %v", v)
					}
					updateState(name, StateError, err, nil, Counts{})
					slog.Error("Panic in MCP client initialization", "error", err, "name", name)
				}
			}()

			if err := initClient(ctx, cfg, name, m, cfg.Resolver()); err != nil {
				slog.Debug("Failed to initialize MCP client", "name", name, "error", err)
			}
		}(name, m)
	}
	wg.Wait()
	initOnce.Do(func() { close(initDone) })
}

// WaitForInit blocks until MCP initialization is complete.
// If Initialize was never called, this returns immediately.
func WaitForInit(ctx context.Context) error {
	select {
	case <-initDone:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// InitializeSingle initializes a single MCP client by name.
func InitializeSingle(ctx context.Context, name string, cfg *config.ConfigStore) error {
	m, exists := cfg.Config().MCP[name]
	if !exists {
		return fmt.Errorf("mcp '%s' not found in configuration", name)
	}

	if m.Disabled {
		updateState(name, StateDisabled, nil, nil, Counts{})
		slog.Debug("Skipping disabled MCP", "name", name)
		return nil
	}

	return initClient(ctx, cfg, name, m, cfg.Resolver())
}

// initClient initializes a single MCP client with the given configuration.
func initClient(ctx context.Context, cfg *config.ConfigStore, name string, m config.MCPConfig, resolver config.VariableResolver) error {
	updateState(name, StateStarting, nil, nil, Counts{})

	workspaceRoot, _ := cfg.Resolver().ResolveValue("")

	session, err := createSession(ctx, name, m, resolver)
	if err != nil {
		return err
	}

	tools, err := getTools(ctx, session)
	if err != nil {
		slog.Error("Error listing tools", "error", err)
		updateState(name, StateError, err, nil, Counts{})
		session.Close()
		return err
	}

	prompts, err := getPrompts(ctx, session)
	if err != nil {
		slog.Error("Error listing prompts", "error", err)
		updateState(name, StateError, err, nil, Counts{})
		session.Close()
		return err
	}

	toolCount := updateTools(cfg, name, tools)
	updatePrompts(name, prompts)
	sessions.Set(name, session)

	updateState(name, StateConnected, nil, session, Counts{
		Tools:   toolCount,
		Prompts: len(prompts),
	})

	if m.Type == "stdio" || m.Type == "" {
		if err := setupPathMapping(name, m, workspaceRoot); err != nil {
			slog.Warn("Failed to setup path mapping for MCP", "name", name, "error", err)
		}
	}

	return nil
}

func setupPathMapping(name string, m config.MCPConfig, workspaceRoot string) error {
	mapping := make(PathMapping)

	ws := GetActiveWorkspace()
	if ws.ProjectRoot != "" && ws.WorkspaceRoot != "" {
		projectName := filepath.Base(ws.ProjectRoot)
		mount := filepath.Join(ws.WorkspaceRoot, projectName)
		mapping[ws.ProjectRoot] = mount
		if ws.GitRoot != "" && ws.GitRoot != ws.ProjectRoot {
			mapping[ws.GitRoot] = mount
		}
	}

	if workspaceRoot != "" && mapping[workspaceRoot] == "" {
		mount := "/workspace"
		if m.WorkspaceMount != "" {
			mount = m.WorkspaceMount
		}
		mapping[workspaceRoot] = mount
	}

	if len(m.PathMapping) > 0 {
		for host, container := range m.PathMapping {
			mapping[host] = container
		}
	}

	if home, err := os.UserHomeDir(); err == nil && mapping[home] == "" {
		mapping[home] = "/workspace"
	}

	if len(mapping) > 0 {
		SetPathMapping(name, mapping)
	}

	return nil
}

// DisableSingle disables and closes a single MCP client by name.
func DisableSingle(cfg *config.ConfigStore, name string) error {
	session, ok := sessions.Get(name)
	if ok {
		if err := session.Close(); err != nil &&
			!errors.Is(err, io.EOF) &&
			!errors.Is(err, context.Canceled) &&
			err.Error() != "signal: killed" {
			slog.Warn("Error closing MCP session", "name", name, "error", err)
		}
		sessions.Del(name)
	}

	// Clear tools and prompts for this MCP.
	updateTools(cfg, name, nil)
	updatePrompts(name, nil)

	// Update state to disabled.
	updateState(name, StateDisabled, nil, nil, Counts{})

	slog.Info("Disabled mcp client", "name", name)
	return nil
}

func getOrRenewClient(ctx context.Context, cfg *config.ConfigStore, name string) (*ClientSession, error) {
	sess, ok := sessions.Get(name)
	if !ok {
		return nil, fmt.Errorf("mcp '%s' not available", name)
	}

	m := cfg.Config().MCP[name]
	state, _ := states.Get(name)

	timeout := mcpTimeout(m)
	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	err := sess.Ping(pingCtx, nil)
	if err == nil {
		return sess, nil
	}
	updateState(name, StateError, maybeTimeoutErr(err, timeout), nil, state.Counts)

	sess, err = createSession(ctx, name, m, cfg.Resolver())
	if err != nil {
		return nil, err
	}

	updateState(name, StateConnected, nil, sess, state.Counts)
	sessions.Set(name, sess)
	return sess, nil
}

// updateState updates the state of an MCP client and publishes an event
func updateState(name string, state State, err error, client *ClientSession, counts Counts) {
	info := ClientInfo{
		Name:   name,
		State:  state,
		Error:  err,
		Client: client,
		Counts: counts,
	}
	switch state {
	case StateConnected:
		info.ConnectedAt = time.Now()
	case StateError:
		sessions.Del(name)
	}
	states.Set(name, info)

	// Publish state change event
	broker.Publish(pubsub.UpdatedEvent, Event{
		Type:   EventStateChanged,
		Name:   name,
		State:  state,
		Error:  err,
		Counts: counts,
	})
}

func createSession(ctx context.Context, name string, m config.MCPConfig, resolver config.VariableResolver) (*ClientSession, error) {
	timeout := mcpTimeout(m)
	mcpCtx, cancel := context.WithCancel(ctx)
	cancelTimer := time.AfterFunc(timeout, cancel)

	transport, err := createTransport(mcpCtx, m, resolver)
	if err != nil {
		updateState(name, StateError, err, nil, Counts{})
		slog.Error("Error creating MCP client", "error", err, "name", name)
		cancel()
		cancelTimer.Stop()
		return nil, err
	}

	client := mcp.NewClient(
		&mcp.Implementation{
			Name:    "duckops",
			Version: version.Version,
			Title:   "DuckOps",
		},
		&mcp.ClientOptions{
			ToolListChangedHandler: func(context.Context, *mcp.ToolListChangedRequest) {
				broker.Publish(pubsub.UpdatedEvent, Event{
					Type: EventToolsListChanged,
					Name: name,
				})
			},
			PromptListChangedHandler: func(context.Context, *mcp.PromptListChangedRequest) {
				broker.Publish(pubsub.UpdatedEvent, Event{
					Type: EventPromptsListChanged,
					Name: name,
				})
			},
			ResourceListChangedHandler: func(context.Context, *mcp.ResourceListChangedRequest) {
				broker.Publish(pubsub.UpdatedEvent, Event{
					Type: EventResourcesListChanged,
					Name: name,
				})
			},
			LoggingMessageHandler: func(ctx context.Context, req *mcp.LoggingMessageRequest) {
				level := parseLevel(req.Params.Level)
				slog.Log(ctx, level, "MCP log", "name", name, "logger", req.Params.Logger, "data", req.Params.Data)
			},
		},
	)

	session, err := client.Connect(mcpCtx, transport, nil)
	if err != nil {
		err = maybeStdioErr(err, transport)
		updateState(name, StateError, maybeTimeoutErr(err, timeout), nil, Counts{})
		slog.Error("MCP client failed to initialize", "error", err, "name", name)
		cancel()
		cancelTimer.Stop()
		return nil, err
	}

	cancelTimer.Stop()
	slog.Debug("MCP client initialized", "name", name)
	return &ClientSession{session, cancel}, nil
}

// maybeStdioErr if a stdio mcp prints an error in non-json format, it'll fail
// to parse, and the cli will then close it, causing the EOF error.
// so, if we got an EOF err, and the transport is STDIO, we try to exec it
// again with a timeout and collect the output so we can add details to the
// error.
// this happens particularly when starting things with npx, e.g. if node can't
// be found or some other error like that.
func maybeStdioErr(err error, transport mcp.Transport) error {
	if !errors.Is(err, io.EOF) {
		return err
	}
	ct, ok := transport.(*mcp.CommandTransport)
	if !ok {
		return err
	}
	if err2 := stdioCheck(ct.Command); err2 != nil {
		err = errors.Join(err, err2)
	}
	return err
}

func maybeTimeoutErr(err error, timeout time.Duration) error {
	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("timed out after %s", timeout)
	}
	return err
}

func createTransport(ctx context.Context, m config.MCPConfig, resolver config.VariableResolver) (mcp.Transport, error) {
	switch m.Type {
	case config.MCPStdio:
		command, err := resolver.ResolveValue(m.Command)
		if err != nil {
			return nil, fmt.Errorf("invalid mcp command: %w", err)
		}
		if strings.TrimSpace(command) == "" {
			return nil, fmt.Errorf("mcp stdio config requires a non-empty 'command' field")
		}
		args, err := m.ResolvedArgs(resolver)
		if err != nil {
			return nil, err
		}
		envs, err := m.ResolvedEnv(resolver)
		if err != nil {
			return nil, err
		}
		cmd := exec.CommandContext(ctx, home.Long(command), args...)
		cmd.Env = append(os.Environ(), envs...)
		return &mcp.CommandTransport{
			Command: cmd,
		}, nil
	case config.MCPHttp:
		url, err := m.ResolvedURL(resolver)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(url) == "" {
			return nil, fmt.Errorf("mcp http config requires a non-empty 'url' field")
		}
		headers, err := m.ResolvedHeaders(resolver)
		if err != nil {
			return nil, err
		}
		client := &http.Client{
			Transport: &headerRoundTripper{
				headers: headers,
			},
		}
		return &mcp.StreamableClientTransport{
			Endpoint:   url,
			HTTPClient: client,
		}, nil
	case config.MCPSSE:
		url, err := m.ResolvedURL(resolver)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(url) == "" {
			return nil, fmt.Errorf("mcp sse config requires a non-empty 'url' field")
		}
		headers, err := m.ResolvedHeaders(resolver)
		if err != nil {
			return nil, err
		}
		client := &http.Client{
			Transport: &headerRoundTripper{
				headers: headers,
			},
		}
		return &mcp.SSEClientTransport{
			Endpoint:   url,
			HTTPClient: client,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported mcp type: %s", m.Type)
	}
}

type headerRoundTripper struct {
	headers map[string]string
}

func (rt headerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range rt.headers {
		req.Header.Set(k, v)
	}
	return http.DefaultTransport.RoundTrip(req)
}

func mcpTimeout(m config.MCPConfig) time.Duration {
	return time.Duration(cmp.Or(m.Timeout, 15)) * time.Second
}

func stdioCheck(old *exec.Cmd) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	cmd := exec.CommandContext(ctx, old.Path, old.Args...)
	cmd.Env = old.Env
	out, err := cmd.CombinedOutput()
	if err == nil || errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return nil
	}
	return fmt.Errorf("%w: %s", err, string(out))
}

type PathMapping map[string]string

var pathMappings = csync.NewMap[string, PathMapping]()

type WorkspaceContext struct {
	ProjectRoot           string
	WorkspaceRoot         string
	GitRoot              string
	ContainerProjectRoot  string
}

var activeWorkspace = csync.NewValue(WorkspaceContext{})

func SetActiveWorkspace(ctx WorkspaceContext) {
	activeWorkspace.Set(ctx)
}

func GetActiveWorkspace() WorkspaceContext {
	return activeWorkspace.Get()
}

func SetPathMapping(mcpName string, mapping PathMapping) {
	pathMappings.Set(mcpName, mapping)
}

func TranslateToolPaths(mcpName string, toolName string, args map[string]interface{}) map[string]interface{} {
	mapping, ok := pathMappings.Get(mcpName)
	if !ok || len(mapping) == 0 {
		mapping = buildProjectOnlyMapping()
		if len(mapping) == 0 {
			return args
		}
	}

	translated := make(map[string]interface{}, len(args))
	for k, v := range args {
		if str, ok := v.(string); ok {
			translated[k] = translatePath(str, mapping)
		} else if strSlice, ok := v.([]string); ok {
			out := make([]string, len(strSlice))
			for i, s := range strSlice {
				out[i] = translatePath(s, mapping)
			}
			translated[k] = out
		} else {
			translated[k] = v
		}
	}
	return translated
}

func buildProjectOnlyMapping() PathMapping {
	ws := GetActiveWorkspace()
	if ws.ProjectRoot == "" || ws.WorkspaceRoot == "" {
		return nil
	}
	mapping := make(PathMapping)
	projectName := filepath.Base(ws.ProjectRoot)
	containerRoot := filepath.Join(ws.WorkspaceRoot, projectName)
	mapping[ws.ProjectRoot] = containerRoot
	return mapping
}

func translatePath(path string, mapping PathMapping) string {
	if path == "" {
		return path
	}

	for hostPrefix, containerPrefix := range mapping {
		if strings.HasPrefix(path, hostPrefix) {
			return containerPrefix + path[len(hostPrefix):]
		}
	}

	expanded := expandTilde(path)
	absPath, err := filepath.Abs(expanded)
	if err != nil {
		return path
	}

	for hostPrefix, containerPrefix := range mapping {
		rel, err := filepath.Rel(hostPrefix, absPath)
		if err != nil {
			continue
		}
		if !strings.HasPrefix(rel, "..") {
			continue
		}

		parts := strings.Split(rel, string(filepath.Separator))
		upCount := 0
		for _, p := range parts {
			if p == ".." {
				upCount++
			} else {
				break
			}
		}

		if upCount > 0 {
			remain := filepath.Join(parts[upCount:]...)
			return filepath.Join(containerPrefix, remain)
		}
	}

	return path
}

func expandTilde(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == "~" {
		return home
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	return path
}

func DetectGitRoot(dir string) string {
	if dir == "" {
		dir, _ = os.Getwd()
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}

	for {
		gitPath := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func DetectProjectRoot(dir string) string {
	if dir == "" {
		dir, _ = os.Getwd()
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}

	gitRoot := DetectGitRoot(absDir)
	if gitRoot != "" {
		return gitRoot
	}

	return absDir
}

func DetectWorkspaceRoot() string {
	home, _ := os.UserHomeDir()
	candidates := []string{
		"/workspace",
		"/app/workspace",
		filepath.Join(home, "workspace"),
		filepath.Join(home, "Projects"),
		filepath.Join(home, "projects"),
	}

	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}
	return "/workspace"
}

func NormalizePath(path string) string {
	expanded := expandTilde(path)
	abs, err := filepath.Abs(expanded)
	if err != nil {
		return path
	}
	clean := filepath.Clean(abs)
	realPath, err := filepath.EvalSymlinks(clean)
	if err != nil {
		return clean
	}
	return realPath
}

func TranslateProjectPath(path string, projectRoot, workspaceRoot string) string {
	if path == "" || projectRoot == "" || workspaceRoot == "" {
		return path
	}

	normalized := NormalizePath(path)

	if strings.HasPrefix(normalized, projectRoot) {
		rel, err := filepath.Rel(projectRoot, normalized)
		if err != nil {
			return path
		}
		return filepath.Join(workspaceRoot, rel)
	}

	return path
}

func InitWorkspaceContext(dir string) WorkspaceContext {
	containerCwd, _ := os.Getwd()
	if dir != "" {
		containerCwd = dir
	}

	hostProjectRoot := containerCwd
	if strings.HasPrefix(containerCwd, "/workspace") {
		home, _ := os.UserHomeDir()
		if home != "" {
			hostProjectRoot = filepath.Join(home, containerCwd[len("/workspace"):])
		}
	}

	gitRoot := DetectGitRoot(hostProjectRoot)
	if gitRoot != "" {
		hostProjectRoot = gitRoot
	}

	workspaceRoot := DetectWorkspaceRoot()
	containerProjectRoot := "/workspace"
	home, _ := os.UserHomeDir()
	if home != "" && strings.HasPrefix(hostProjectRoot, home) {
		containerProjectRoot = filepath.Join("/workspace", hostProjectRoot[len(home):])
	}

	ctx := WorkspaceContext{
		ProjectRoot:          hostProjectRoot,
		WorkspaceRoot:        workspaceRoot,
		GitRoot:             gitRoot,
		ContainerProjectRoot: containerProjectRoot,
	}

	SetActiveWorkspace(ctx)

	if hostProjectRoot != "" {
		mapping := make(PathMapping)
		projectName := filepath.Base(hostProjectRoot)
		if projectName == "" || projectName == "/" {
			projectName = "project"
		}
		mapping[hostProjectRoot] = filepath.Join(workspaceRoot, projectName)
		SetPathMapping("default", mapping)
	}

	return ctx
}
