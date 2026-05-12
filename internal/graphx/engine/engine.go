package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/SecDuckOps/duckops/internal/graphx/config"
	"github.com/SecDuckOps/duckops/internal/graphx/ir"
	"github.com/SecDuckOps/duckops/internal/graphx/parser"
	"github.com/SecDuckOps/duckops/internal/graphx/storage"
)

type Engine struct {
	cfg        *config.Config
	store      *storage.JSONStore
	registry   *parser.Registry
	repoID     string
	mu         sync.RWMutex
}

func NewEngine(cfg *config.Config) (*Engine, error) {
	dbPath := cfg.Graph.Path
	if !filepath.IsAbs(dbPath) {
		dbPath = filepath.Join(cfg.Repository.Path, dbPath)
	}

	store, err := storage.NewJSONStore(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create store: %w", err)
	}

	e := &Engine{
		cfg:      cfg,
		store:    store,
		registry: parser.NewRegistry(),
	}

	e.registerParsers()

	if err := e.initRepository(); err != nil {
		return nil, fmt.Errorf("failed to init repository: %w", err)
	}

	return e, nil
}

func (e *Engine) registerParsers() {
	e.registry.Register(parser.NewPythonParser())
	e.registry.Register(parser.NewGoParser())
	e.registry.Register(parser.NewTypeScriptParser())
	e.registry.Register(parser.NewRustParser())
}

func (e *Engine) initRepository() error {
	repoPath, err := filepath.Abs(e.cfg.Repository.Path)
	if err != nil {
		return err
	}

	e.repoID = generateRepoID(repoPath)
	e.cfg.Repository.RootPath = repoPath

	return nil
}

func (e *Engine) Build(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	files, err := e.collectSourceFiles()
	if err != nil {
		return fmt.Errorf("failed to collect files: %w", err)
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, e.cfg.Parser.Parallel)
	errors := make([]error, 0)
	var mu sync.Mutex

	for _, file := range files {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if err := e.processFile(f); err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
			}
		}(file)
	}

	wg.Wait()

	if len(errors) > 0 {
		return fmt.Errorf("processing errors: %v", errors[:min(5, len(errors))])
	}

	return nil
}

func (e *Engine) collectSourceFiles() ([]string, error) {
	var files []string
	extensions := map[string]bool{
		".py":    true,
		".go":    true,
		".ts":    true,
		".tsx":   true,
		".js":    true,
		".jsx":   true,
		".rs":    true,
		".java":  true,
		".rb":    true,
		".php":   true,
		".kt":    true,
		".swift": true,
		".tf":    true,
		".yaml":  true,
		".yml":   true,
	}

	err := filepath.Walk(e.cfg.Repository.RootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			for _, ignore := range e.cfg.Repository.Ignore {
				if matchPattern(path, ignore) {
					return filepath.SkipDir
				}
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !extensions[ext] {
			return nil
		}

		relPath, err := filepath.Rel(e.cfg.Repository.RootPath, path)
		if err != nil {
			return nil
		}

		for _, ignore := range e.cfg.Repository.Ignore {
			if matchPattern(relPath, ignore) {
				return nil
			}
		}

		files = append(files, path)
		return nil
	})

	return files, err
}

func (e *Engine) processFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	p := e.registry.ForFile(path)
	if p == nil {
		return nil
	}

	program, err := p.Parse(content)
	if err != nil {
		return err
	}

	program.RepoID = e.repoID
	program.FilePath = path
	program.SourceHash = hashContent(content)

	if err := e.importProgramToStore(program); err != nil {
		return err
	}

	return nil
}

func (e *Engine) importProgramToStore(program *ir.Program) error {
	now := time.Now().UnixMilli()
	fileNode := storage.NewNode(e.repoID, program.FilePath, storage.NodeTypeFile)
	fileNode.Language = program.Language
	fileNode.ContentHash = program.SourceHash
	fileNode.SecuritySensitivity = 0.1
	fileNode.UpdatedAt = now

	if err := e.store.AddNode(fileNode); err != nil {
		return err
	}

	var functions []*storage.Node

	for _, el := range program.Elements {
		switch elem := el.(type) {
		case *ir.Function:
			node := storage.NewNode(e.repoID, elem.Name, storage.NodeTypeFunction)
			node.FullyQualifiedName = elem.FullyQualifiedName
			node.FilePath = program.FilePath
			node.Language = program.Language
			node.IsExported = elem.IsExported
			node.SecuritySensitivity = calculateSensitivity(elem.SecurityTags, elem.Name)

			lowerName := strings.ToLower(elem.Name)
			authKeywords := []string{"auth", "login", "logout", "password", "credential", "token", "session", "jwt", "oauth", "verify", "validate", "permission", "role", "access", "admin"}
			for _, kw := range authKeywords {
				if strings.Contains(lowerName, kw) {
					node.AuthSensitive = true
					break
				}
			}
			node.Metadata = map[string]interface{}{
				"start_line":    elem.StartLine,
				"security_tags": elem.SecurityTags,
				"calls":         elem.Calls,
			}
			node.UpdatedAt = now

			if err := e.store.AddNode(node); err != nil {
				return err
			}
			functions = append(functions, node)

		case *ir.Class:
			node := storage.NewNode(e.repoID, elem.Name, storage.NodeTypeClass)
			node.FullyQualifiedName = elem.FullyQualifiedName
			node.FilePath = program.FilePath
			node.Language = program.Language
			node.SecuritySensitivity = calculateSensitivity(elem.SecurityTags, elem.Name)

			lowerName := strings.ToLower(elem.Name)
			if strings.Contains(lowerName, "auth") || strings.Contains(lowerName, "session") || strings.Contains(lowerName, "user") {
				node.AuthSensitive = true
			}
			node.UpdatedAt = now

			if err := e.store.AddNode(node); err != nil {
				return err
			}

		case *ir.Import:
			impNode := storage.NewNode(e.repoID, elem.Source, storage.NodeTypeFile)
			impNode.Type = storage.NodeType("import")
			impNode.Metadata = map[string]interface{}{
				"kind": elem.Kind,
			}
			impNode.UpdatedAt = now

			if err := e.store.AddNode(impNode); err != nil {
				return err
			}

			edge := storage.NewEdge(e.repoID, fileNode.ID, impNode.ID, storage.EdgeType("imports"))
			edge.Confidence = 0.9
			edge.ConfidenceLevel = "EXTRACTED"
			edge.UpdatedAt = now

			if err := e.store.AddEdge(edge); err != nil {
				return err
			}
		}
	}

	if len(functions) > 1 {
		for i := 0; i < len(functions)-1; i++ {
			if functions[i].IsExported || functions[i+1].IsExported {
				edge := storage.NewEdge(e.repoID, functions[i].ID, functions[i+1].ID, storage.EdgeType("calls"))
				edge.Confidence = 0.7
				edge.ConfidenceLevel = "INFERRED"
				edge.UpdatedAt = now
				_ = e.store.AddEdge(edge)
			}
		}
	}

	return nil
}

func calculateSensitivity(tags []string, name string) float64 {
	lowerName := strings.ToLower(name)

	authKeywords := []string{"auth", "login", "logout", "password", "credential", "token", "session", "jwt", "oauth", "verify", "validate", "permission", "role", "access", "security"}
	adminKeywords := []string{"admin", "root", "sudo", "elevate", "privilege"}
	apiKeywords := []string{"handler", "endpoint", "route", "controller", "api", "http"}
	dbKeywords := []string{"query", "sql", "database", "db", "store", "repository"}
	piiKeywords := []string{"pii", "personal", "email", "phone", "address", "ssn", "credit", "payment"}

	for _, kw := range authKeywords {
		if strings.Contains(lowerName, kw) {
			return 0.9
		}
	}
	for _, kw := range adminKeywords {
		if strings.Contains(lowerName, kw) {
			return 0.85
		}
	}
	for _, kw := range apiKeywords {
		if strings.Contains(lowerName, kw) {
			return 0.7
		}
	}
	for _, kw := range dbKeywords {
		if strings.Contains(lowerName, kw) {
			return 0.6
		}
	}
	for _, kw := range piiKeywords {
		if strings.Contains(lowerName, kw) {
			return 0.8
		}
	}

	if containsTag(tags, "auth") || containsTag(tags, "admin") {
		return 0.9
	}
	if containsTag(tags, "api") {
		return 0.7
	}
	if containsTag(tags, "database") {
		return 0.6
	}
	return 0.1
}

func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if strings.ToLower(t) == strings.ToLower(tag) {
			return true
		}
	}
	return false
}

func (e *Engine) GetStore() *storage.JSONStore {
	return e.store
}

func (e *Engine) GetRepoID() string {
	return e.repoID
}

func (e *Engine) Close() error {
	return e.store.Close()
}

func generateRepoID(path string) string {
	h := sha256.Sum256([]byte(path))
	return "repo_" + hex.EncodeToString(h[:8])
}

func hashContent(content []byte) string {
	h := sha256.Sum256(content)
	return hex.EncodeToString(h[:])
}

func matchPattern(path, pattern string) bool {
	pattern = strings.TrimPrefix(pattern, "**/")
	pattern = strings.TrimPrefix(pattern, "*")
	return strings.HasSuffix(path, pattern) || path == pattern
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}