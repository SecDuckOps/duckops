package graphx

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/SecDuckOps/duckops/internal/graphx/analysis/blast"
	"github.com/SecDuckOps/duckops/internal/graphx/analysis/community"
	"github.com/SecDuckOps/duckops/internal/graphx/analysis/threat"
	ctxengine "github.com/SecDuckOps/duckops/internal/graphx/engine/context"
	"github.com/SecDuckOps/duckops/internal/graphx/engine/incremental"
	"github.com/SecDuckOps/duckops/internal/graphx/engine/watch"
	"github.com/SecDuckOps/duckops/internal/graphx/ir"
	"github.com/SecDuckOps/duckops/internal/graphx/parser"
	"github.com/SecDuckOps/duckops/internal/graphx/storage"
)

type Engine struct {
	repoID    string
	repoPath  string
	store     *storage.JSONStore
	registry  *parser.Registry
	parser    *MultiLanguageParser

	threatEngine  *threat.ThreatModelEngine
	blastEngine   *blast.BlastRadiusEngine
	communityEng  *community.Engine
	contextEngine *ctxengine.ContextEngine
	watcher       *watch.Watcher
	incremental   *incremental.Engine

	mu         sync.RWMutex
	initialized bool
	status     EngineStatus
}

type EngineStatus struct {
	State         string    `json:"state"`
	NodesCount    int       `json:"nodes_count"`
	EdgesCount    int       `json:"edges_count"`
	LastUpdated   time.Time `json:"last_updated"`
	LastAnalyzed  time.Time `json:"last_analyzed"`
	WatchEnabled  bool      `json:"watch_enabled"`
	Errors        []string  `json:"errors,omitempty"`
}

type Config struct {
	RepoID    string
	RepoPath  string
	StoragePath string
	WatchEnabled bool
	Parallelism int
}

func NewEngine(cfg Config) (*Engine, error) {
	store, err := storage.NewJSONStore(cfg.StoragePath)
	if err != nil {
		return nil, err
	}

	e := &Engine{
		repoID:    cfg.RepoID,
		repoPath:  cfg.RepoPath,
		store:     store,
		registry:  parser.NewRegistry(),
		parser:    NewMultiLanguageParser(),
		status:    EngineStatus{State: "initializing"},
	}

	e.threatEngine = threat.NewThreatModelEngine(store)
	e.blastEngine = blast.NewBlastRadiusEngine(store)
	e.communityEng = community.NewCommunityEngine(store)
	e.contextEngine = ctxengine.NewContextEngine(store)

	return e, nil
}

func (e *Engine) Initialize(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.initialized {
		return nil
	}

	if err := e.store.Init(); err != nil {
		return err
	}

	if err := e.parseRepository(ctx); err != nil {
		e.status.Errors = append(e.status.Errors, err.Error())
	}

	e.enrichSecurityMetadata()
	e.detectTrustBoundaries()
	e.buildCrossReferences()

	e.initialized = true
	e.status.State = "ready"
	e.status.LastAnalyzed = time.Now()

	if e.watcher != nil {
		e.watcher.Start()
		e.status.WatchEnabled = true
	}

	return nil
}

func (e *Engine) parseRepository(ctx context.Context) error {
	files, err := e.findCodeFiles()
	if err != nil {
		return err
	}

	e.status.State = "parsing"
	
	for _, file := range files {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		program, err := e.parser.ParseFile(file)
		if err != nil {
			e.status.Errors = append(e.status.Errors, "parse error: "+err.Error())
			continue
		}

		if err := e.storeFromProgram(file, program); err != nil {
			e.status.Errors = append(e.status.Errors, "store error: "+err.Error())
		}
	}

	e.status.NodesCount, e.status.EdgesCount, _ = e.store.GetStats()
	e.status.LastUpdated = time.Now()

	return nil
}

func (e *Engine) findCodeFiles() ([]string, error) {
	extensions := []string{".go", ".py", ".ts", ".tsx", ".js", ".jsx", ".rs", ".java", ".rb", ".php", ".kt", ".swift"}
	
	files := make([]string, 0)
	
	for _, ext := range extensions {
		pattern := filepath.Join(e.repoPath, "**", "*"+ext)
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		files = append(files, matches...)
	}

	excludePatterns := []string{"**/vendor/**", "**/node_modules/**", "**/*.test.go", "**/*_test.py", "**/test/**"}
	filtered := make([]string, 0)
	
outer:
	for _, f := range files {
		for _, pat := range excludePatterns {
			matched, _ := filepath.Match(pat, f)
			if matched {
				continue outer
			}
		}
		filtered = append(filtered, f)
	}
	
	return filtered, nil
}

func (e *Engine) storeFromProgram(filePath string, program *ir.Program) error {
	program.FilePath = filePath
	program.RepoID = e.repoID

	for _, el := range program.Elements {
		switch elem := el.(type) {
		case *ir.Function:
			node := e.functionToNode(filePath, elem)
			e.store.AddNode(node)
			
		case *ir.Class:
			node := e.classToNode(filePath, elem)
			e.store.AddNode(node)
			
		case *ir.Import:
			node := e.importToNode(filePath, elem)
			e.store.AddNode(node)
		}
	}

	for _, el := range program.Elements {
		if fn, ok := el.(*ir.Function); ok {
			for _, call := range fn.Calls {
				edge := e.createCallEdge(fn, call)
				if edge != nil {
					e.store.AddEdge(edge)
				}
			}
		}
	}

	return nil
}

func (e *Engine) functionToNode(filePath string, fn *ir.Function) *storage.Node {
	node := storage.NewNode(e.repoID, fn.Name, storage.NodeTypeFunction)
	node.FullyQualifiedName = fn.FullyQualifiedName
	node.FilePath = filePath
	node.Language = ir.DetectLanguage(filePath)
	node.IsExported = fn.IsExported
	node.Metadata = map[string]interface{}{
		"start_line":   fn.StartLine,
		"end_line":     fn.EndLine,
		"is_async":     fn.IsAsync,
		"is_private":   fn.IsPrivate,
		"parameters":   fn.Parameters,
		"return_type":  fn.ReturnType,
		"security_tags": fn.SecurityTags,
		"calls":        fn.Calls,
	}

	node.SecuritySensitivity = e.calculateSecuritySensitivity(fn)
	
	if fn.IsExported && !fn.IsPrivate {
		node.InternetExposed = true
	}
	
	node.RiskScore = e.calculateRiskScore(fn)
	node.BlastRadiusScore = e.calculateBlastRadius(fn)
	
	return node
}

func (e *Engine) classToNode(filePath string, c *ir.Class) *storage.Node {
	node := storage.NewNode(e.repoID, c.Name, storage.NodeTypeClass)
	node.FullyQualifiedName = c.FullyQualifiedName
	node.FilePath = filePath
	node.Language = ir.DetectLanguage(filePath)
	node.IsExported = true
	node.Metadata = map[string]interface{}{
		"start_line":   c.StartLine,
		"end_line":     c.EndLine,
		"methods":      c.Methods,
		"fields":       c.Fields,
		"parent":       c.Parent,
		"security_tags": c.SecurityTags,
	}
	
	return node
}

func (e *Engine) importToNode(filePath string, imp *ir.Import) *storage.Node {
	node := storage.NewNode(e.repoID, imp.Source, storage.NodeType("import"))
	node.FilePath = filePath
	node.Metadata = map[string]interface{}{
		"kind":        imp.Kind,
		"alias":       imp.Alias,
		"is_dynamic":  imp.IsDynamic,
	}

	if ir.IsSecuritySensitive(imp.Source) {
		node.AuthSensitive = true
	}

	return node
}

func (e *Engine) createCallEdge(fn *ir.Function, calledFunc string) *storage.Edge {
	edge := storage.NewEdge(e.repoID, fn.ID, calledFunc, storage.EdgeType("calls"))
	edge.Confidence = 0.7
	edge.ConfidenceLevel = "INFERRED"
	return edge
}

func (e *Engine) calculateSecuritySensitivity(fn *ir.Function) float64 {
	sensitivity := 0.0
	
	for _, tag := range fn.SecurityTags {
		switch tag {
		case "auth":
			sensitivity += 0.8
		case "admin":
			sensitivity += 0.7
		case "api":
			sensitivity += 0.5
		case "public_api":
			sensitivity += 0.6
		case "database":
			sensitivity += 0.5
		}
	}

	if fn.IsExported {
		sensitivity += 0.3
	}
	
	return sensitivity
}

func (e *Engine) calculateRiskScore(fn *ir.Function) float64 {
	score := 0.0
	
	if fn.IsExported {
		score += 3.0
	}
	
	for _, tag := range fn.SecurityTags {
		switch tag {
		case "auth":
			score += 4.0
		case "admin":
			score += 3.5
		case "api":
			score += 2.0
		case "handles_pii":
			score += 3.0
		}
	}
	
	if len(fn.Calls) > 10 {
		score += 2.0
	}
	
	return score
}

func (e *Engine) calculateBlastRadius(fn *ir.Function) float64 {
	return float64(len(fn.Calls)) * 0.5
}

func (e *Engine) enrichSecurityMetadata() {
	nodes, _ := e.store.GetAllNodes()
	
	for _, node := range nodes {
		e.detectPIIHandling(node)
		e.detectTokenHandling(node)
		e.detectAuthSensitivity(node)
		e.store.UpdateNode(node)
	}
}

func (e *Engine) detectPIIHandling(node *storage.Node) {
	if node.Type != storage.NodeTypeFunction {
		return
	}
	
	metadata := node.Metadata
	if metadata == nil {
		return
	}
	
	name := strings.ToLower(node.Name)
	if strings.Contains(name, "email") || strings.Contains(name, "phone") ||
		strings.Contains(name, "ssn") || strings.Contains(name, "address") ||
		strings.Contains(name, "name") {
		node.PIIHandling = true
	}
}

func (e *Engine) detectTokenHandling(node *storage.Node) {
	if node.Type != storage.NodeTypeFunction {
		return
	}
	
	name := strings.ToLower(node.Name)
	if strings.Contains(name, "token") || strings.Contains(name, "jwt") ||
		strings.Contains(name, "session") || strings.Contains(name, "auth") {
		node.AuthSensitive = true
	}
}

func (e *Engine) detectAuthSensitivity(node *storage.Node) {
	name := strings.ToLower(node.Name)
	if strings.Contains(name, "password") || strings.Contains(name, "credential") ||
		strings.Contains(name, "secret") || strings.Contains(name, "apikey") {
		node.AuthSensitive = true
	}
}

func (e *Engine) detectTrustBoundaries() {
	nodes, _ := e.store.GetNodesByType(e.repoID, storage.NodeTypeAPI)
	
	for _, node := range nodes {
		node.AuthSensitive = true
		node.Metadata["trust_boundary"] = "public"
		e.store.UpdateNode(node)
	}
}

func (e *Engine) buildCrossReferences() {
	nodes, _ := e.store.GetAllNodes()
	nodeNameIndex := make(map[string][]*storage.Node)
	
	for _, node := range nodes {
		if node.Type == storage.NodeTypeFunction {
			for _, calledName := range e.getCalledNames(node) {
				nodeNameIndex[calledName] = append(nodeNameIndex[calledName], node)
			}
		}
	}
	
	for _, node := range nodes {
		if node.Type != storage.NodeTypeFunction {
			continue
		}
		
		calledNames := e.getCalledNames(node)
		for _, calledName := range calledNames {
			targets := nodeNameIndex[calledName]
			for _, target := range targets {
				if node.ID != target.ID {
					edge := storage.NewEdge(e.repoID, node.ID, target.ID, storage.EdgeType("references"))
					edge.Confidence = 0.6
					edge.ConfidenceLevel = "INFERRED"
					e.store.AddEdge(edge)
				}
			}
		}
	}
}

func (e *Engine) getCalledNames(node *storage.Node) []string {
	if node.Metadata == nil {
		return nil
	}
	
	if calls, ok := node.Metadata["calls"]; ok {
		if callList, ok := calls.([]string); ok {
			return callList
		}
	}
	
	return nil
}

func (e *Engine) EnableWatchMode(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.watcher = watch.NewWatcher(nil, nil)

	e.incremental = incremental.NewIncrementalEngine(e.store)

	e.watcher.Start()
	e.status.WatchEnabled = true

	return nil
}

func (e *Engine) handleFileEvents(ctx context.Context, events []watch.Event) {
	for _, event := range events {
		if event.Type == watch.EventDelete {
			e.store.RemoveNodesByFile(event.Path)
			continue
		}

		program, err := e.parser.ParseFile(event.Path)
		if err != nil {
			continue
		}

		e.storeFromProgram(event.Path, program)
	}

	e.enrichSecurityMetadata()
	e.status.LastUpdated = time.Now()
}

func (e *Engine) GetStatus() EngineStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	status := e.status
	status.NodesCount, status.EdgesCount, _ = e.store.GetStats()
	
	return status
}

func (e *Engine) GetThreatModel() (*threat.ThreatModel, error) {
	return e.threatEngine.GenerateModel(e.repoID)
}

func (e *Engine) GetBlastRadius(nodeID string) (*blast.BlastResult, error) {
	return e.blastEngine.Analyze(nodeID)
}

func (e *Engine) GetContext(req *ctxengine.ContextRequest) (*ctxengine.ContextResponse, error) {
	return e.contextEngine.GenerateContext(req)
}

func (e *Engine) SearchByName(name string) []*storage.Node {
	return e.store.SearchByName(name)
}

func (e *Engine) GetFunctions() []*storage.Node {
	nodes, _ := e.store.GetNodesByType(e.repoID, storage.NodeTypeFunction)
	return nodes
}

func (e *Engine) GetAPIs() []*storage.Node {
	nodes, _ := e.store.GetNodesByType(e.repoID, storage.NodeTypeAPI)
	return nodes
}

func (e *Engine) Shutdown() {
	if e.watcher != nil {
		e.watcher.Stop()
	}
	e.status.State = "shutdown"
}