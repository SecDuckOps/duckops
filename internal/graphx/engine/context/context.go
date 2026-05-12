package context

import (
	"fmt"
	"sort"
	"strings"

	"github.com/SecDuckOps/duckops/internal/graphx/storage"
)

type ContextEngine struct {
	store *storage.JSONStore
}

func NewContextEngine(store *storage.JSONStore) *ContextEngine {
	return &ContextEngine{store: store}
}

type ContextRequest struct {
	RepoID          string
	Query           string
	FocusNodes      []string
	SecurityFocused bool
	MaxTokens       int
	MaxFiles        int
}

type ContextResponse struct {
	Nodes       []*storage.Node
	Edges       []*storage.Edge
	FilePaths   []string
	Summary     string
	TokenCount  int
}

type ContextBlock struct {
	Type      string
	Content   string
	TokenCost int
}

func (e *ContextEngine) GenerateContext(req *ContextRequest) (*ContextResponse, error) {
	var nodes []*storage.Node
	var edges []*storage.Edge

	if len(req.FocusNodes) > 0 {
		for _, nodeID := range req.FocusNodes {
			node, err := e.store.GetNode(nodeID)
			if err != nil {
				continue
			}
			nodes = append(nodes, node)

			nodeEdges, err := e.store.GetEdgesByNode(nodeID)
			if err == nil {
				edges = append(edges, nodeEdges...)
			}
		}
	} else if req.Query != "" {
		result, _, err := e.store.Query(&storage.NodeQuery{
			RepoID:      req.RepoID,
			NamePattern: req.Query,
			Limit:       req.MaxFiles,
		})
		if err == nil {
			nodes = result
		}
	} else {
		result, _, err := e.store.Query(&storage.NodeQuery{
			RepoID: req.RepoID,
			Limit:  req.MaxFiles,
		})
		if err == nil {
			nodes = result
		}
	}

	if req.SecurityFocused {
		nodes = e.filterSecurityRelevant(nodes)
	}

	nodes = e.prioritizeNodes(nodes, req.FocusNodes)

	resp := &ContextResponse{
		Nodes:      nodes,
		Edges:      edges,
		TokenCount: e.estimateTokens(nodes, edges),
	}

	resp.Summary = e.generateSummary(resp)
	resp.FilePaths = e.extractFilePaths(nodes)

	return resp, nil
}

func (e *ContextEngine) filterSecurityRelevant(nodes []*storage.Node) []*storage.Node {
	var filtered []*storage.Node
	for _, n := range nodes {
		if n.AuthSensitive || n.RiskScore > 0.5 || n.SecuritySensitivity > 0.5 || n.InternetExposed {
			filtered = append(filtered, n)
		}
	}
	if len(filtered) == 0 {
		return nodes[:min(10, len(nodes))]
	}
	return filtered
}

func (e *ContextEngine) prioritizeNodes(nodes []*storage.Node, focusNodes []string) []*storage.Node {
	focusSet := make(map[string]bool)
	for _, id := range focusNodes {
		focusSet[id] = true
	}

	sort.Slice(nodes, func(i, j int) bool {
		aScore := nodes[i].RiskScore*10 + nodes[i].SecuritySensitivity
		bScore := nodes[j].RiskScore*10 + nodes[j].SecuritySensitivity

		if focusSet[nodes[i].ID] {
			aScore += 100
		}
		if focusSet[nodes[j].ID] {
			bScore += 100
		}

		return aScore > bScore
	})

	return nodes
}

func (e *ContextEngine) extractFilePaths(nodes []*storage.Node) []string {
	paths := make(map[string]bool)
	for _, n := range nodes {
		if n.FilePath != "" {
			paths[n.FilePath] = true
		}
	}

	result := make([]string, 0, len(paths))
	for p := range paths {
		result = append(result, p)
	}
	return result
}

func (e *ContextEngine) estimateTokens(nodes []*storage.Node, edges []*storage.Edge) int {
	avgNodeSize := 50
	avgEdgeSize := 30

	return len(nodes)*avgNodeSize + len(edges)*avgEdgeSize
}

func (e *ContextEngine) generateSummary(resp *ContextResponse) string {
	var b strings.Builder

	b.WriteString("## Knowledge Graph Context\n\n")

	b.WriteString("### Statistics\n")
	b.WriteString(fmt.Sprintf("- Total nodes: %d\n", len(resp.Nodes)))
	b.WriteString(fmt.Sprintf("- Total edges: %d\n", len(resp.Edges)))
	b.WriteString(fmt.Sprintf("- Estimated tokens: %d\n", resp.TokenCount))

	if len(resp.FilePaths) > 0 {
		b.WriteString("\n### Key Files\n")
		for i, p := range resp.FilePaths {
			if i >= 10 {
				break
			}
			b.WriteString("- " + p + "\n")
		}
	}

	var authSensitive, highRisk int
	for _, n := range resp.Nodes {
		if n.AuthSensitive {
			authSensitive++
		}
		if n.RiskScore > 0.7 {
			highRisk++
		}
	}

	if authSensitive > 0 || highRisk > 0 {
		b.WriteString("\n### Security Summary\n")
		if authSensitive > 0 {
			b.WriteString(fmt.Sprintf("- Auth-sensitive components: %d\n", authSensitive))
		}
		if highRisk > 0 {
			b.WriteString(fmt.Sprintf("- High-risk components: %d\n", highRisk))
		}
	}

	return b.String()
}

func (e *ContextEngine) OptimizeForWindow(ctx *ContextResponse, maxTokens int) *ContextResponse {
	if ctx.TokenCount <= maxTokens {
		return ctx
	}

	targetCount := (maxTokens / 50) * 4 / 5

	if len(ctx.Nodes) > targetCount {
		ctx.Nodes = ctx.Nodes[:targetCount]
		ctx.TokenCount = e.estimateTokens(ctx.Nodes, ctx.Edges)
	}

	return ctx
}

func (e *ContextEngine) BuildSecurityContext(repoID string, change string) (*ContextResponse, error) {
	affectedNodes, err := e.findAffectedByChange(repoID, change)
	if err != nil {
		return nil, err
	}

	return e.GenerateContext(&ContextRequest{
		RepoID:          repoID,
		FocusNodes:      affectedNodes,
		SecurityFocused: true,
		MaxFiles:        20,
		MaxTokens:       8000,
	})
}

func (e *ContextEngine) findAffectedByChange(repoID, filePath string) ([]string, error) {
	nodes, _, err := e.store.Query(&storage.NodeQuery{
		RepoID:    repoID,
		NamePattern: "",
		Limit:     1000,
	})
	if err != nil {
		return nil, err
	}

	var affected []string
	for _, n := range nodes {
		if n.FilePath == filePath {
			affected = append(affected, n.ID)

			edges, _ := e.store.GetOutgoingEdges(n.ID)
			for _, e := range edges {
				affected = append(affected, e.TargetID)
			}
		}
	}

	return affected, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}