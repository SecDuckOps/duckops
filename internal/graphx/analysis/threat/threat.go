package threat

import (
	"sort"
	"strings"

	"github.com/SecDuckOps/duckops/internal/graphx/storage"
)

type ThreatModelEngine struct {
	store *storage.JSONStore
}

type Threat struct {
	ID          string
	Category    string
	Severity    string
	Title       string
	Description string
	TargetNode  string
	AttackPaths [][]string
	STRIDE      []string
}

type ThreatModel struct {
	RepositoryID    string
	EntryPoints     []*storage.Node
	TrustBoundaries []string
	Threats         []Threat
	AttackPaths     []AttackPath
	ExecutionFlows  []ExecutionFlow
	Summary         *ModelSummary
}

type AttackPath struct {
	Source     string
	Target     string
	Path       []string
	Complexity string
	Impact     string
}

type ExecutionFlow struct {
	ID          string
	Name        string
	EntryPoint  string
	Steps       []FlowStep
	AuthRequired bool
	HandlesPII  bool
	RiskLevel   string
}

type FlowStep struct {
	NodeID   string
	NodeName string
	Type     string
	Action   string
}

type ModelSummary struct {
	Spoofing              int
	Tampering            int
	Repudiation          int
	InformationDisclosure int
	DenialOfService      int
	ElevationOfPrivilege  int
	Total                 int
}

func NewThreatModelEngine(store *storage.JSONStore) *ThreatModelEngine {
	return &ThreatModelEngine{store: store}
}

func (e *ThreatModelEngine) GenerateModel(repoID string) (*ThreatModel, error) {
	model := &ThreatModel{
		RepositoryID: repoID,
	}

	entryPoints, err := e.findEntryPoints(repoID)
	if err != nil {
		return nil, err
	}
	model.EntryPoints = entryPoints

	model.TrustBoundaries = e.detectTrustBoundaries(repoID)

	model.Threats, _ = e.analyzeSTRIDE(repoID, entryPoints)

	model.AttackPaths, _ = e.findAttackPaths(repoID)

	model.ExecutionFlows = e.traceExecutionFlows(repoID, entryPoints)

	model.Summary = e.calculateSummary(model.Threats)

	return model, nil
}

func (e *ThreatModelEngine) findEntryPoints(repoID string) ([]*storage.Node, error) {
	apis, err := e.store.GetNodesByType(repoID, storage.NodeTypeAPI)
	if err != nil {
		return nil, err
	}

	return apis, nil
}

func (e *ThreatModelEngine) analyzeSTRIDE(repoID string, entryPoints []*storage.Node) ([]Threat, error) {
	threats := make([]Threat, 0)

	for _, ep := range entryPoints {
		spoofing := Threat{
			ID:          "spoofing-" + ep.ID,
			Category:    "Spoofing",
			Severity:    "Medium",
			Title:       "Authentication Bypass",
			Description: "Attacker could bypass authentication on " + ep.Name,
			TargetNode:  ep.ID,
			STRIDE:      []string{"Spoofing"},
		}
		threats = append(threats, spoofing)

		tampering := Threat{
			ID:          "tampering-" + ep.ID,
			Category:    "Tampering",
			Severity:    "High",
			Description: "Data could be tampered in transit to " + ep.Name,
			TargetNode:  ep.ID,
			STRIDE:      []string{"Tampering"},
		}
		threats = append(threats, tampering)

		if ep.AuthSensitive {
			eop := Threat{
				ID:          "eop-" + ep.ID,
				Category:    "Elevation of Privilege",
				Severity:    "Critical",
				Title:       "Privilege Escalation",
				Description: "Attacker could elevate privileges through " + ep.Name,
				TargetNode:  ep.ID,
				STRIDE:      []string{"Elevation of Privilege"},
			}
			threats = append(threats, eop)
		}

		disclosure := Threat{
			ID:          "disclosure-" + ep.ID,
			Category:    "Information Disclosure",
			Severity:    "Medium",
			Description: "Sensitive information may be exposed through " + ep.Name,
			TargetNode:  ep.ID,
			STRIDE:      []string{"Information Disclosure"},
		}
		threats = append(threats, disclosure)
	}

	return threats, nil
}

func (e *ThreatModelEngine) findAttackPaths(repoID string) ([]AttackPath, error) {
	entryPoints, _ := e.findEntryPoints(repoID)

	attackPaths := make([]AttackPath, 0)

	for _, ep := range entryPoints {
		paths := e.traceAttackPaths(ep.ID, 5)
		attackPaths = append(attackPaths, paths...)
	}

	sort.Slice(attackPaths, func(i, j int) bool {
		return attackPaths[i].Complexity == "low"
	})

	return attackPaths, nil
}

func (e *ThreatModelEngine) traceAttackPaths(startNodeID string, maxDepth int) []AttackPath {
	paths := make([]AttackPath, 0)

	visited := make(map[string]bool)
	visited[startNodeID] = true

	var dfs func(current string, path []string, depth int)
	dfs = func(current string, path []string, depth int) {
		if depth >= maxDepth {
			return
		}

		edges, _ := e.store.GetOutgoingEdges(current)
		for _, edge := range edges {
			if visited[edge.TargetID] {
				continue
			}
			visited[edge.TargetID] = true

			newPath := append([]string{}, path...)
			newPath = append(newPath, edge.TargetID)

			targetNode, _ := e.store.GetNode(edge.TargetID)
			if targetNode != nil && targetNode.AuthSensitive {
				paths = append(paths, AttackPath{
					Source:     startNodeID,
					Target:     edge.TargetID,
					Path:       newPath,
					Complexity: e.assessPathComplexity(depth),
					Impact:     "High",
				})
			}

			dfs(edge.TargetID, newPath, depth+1)
			visited[edge.TargetID] = false
		}
	}

	dfs(startNodeID, []string{startNodeID}, 0)

	return paths
}

func (e *ThreatModelEngine) assessPathComplexity(depth int) string {
	switch {
	case depth <= 2:
		return "low"
	case depth <= 4:
		return "medium"
	default:
		return "high"
	}
}

func (e *ThreatModelEngine) calculateSummary(threats []Threat) *ModelSummary {
	summary := &ModelSummary{}

	for _, t := range threats {
		for _, s := range t.STRIDE {
			switch s {
			case "Spoofing":
				summary.Spoofing++
			case "Tampering":
				summary.Tampering++
			case "Repudiation":
				summary.Repudiation++
			case "Information Disclosure":
				summary.InformationDisclosure++
			case "Denial of Service":
				summary.DenialOfService++
			case "Elevation of Privilege":
				summary.ElevationOfPrivilege++
			}
		}
	}

	summary.Total = len(threats)

	return summary
}

func (e *ThreatModelEngine) detectTrustBoundaries(repoID string) []string {
	boundaries := make(map[string]bool)

	nodes, _ := e.store.GetNodesByRepo(repoID)
	for _, node := range nodes {
		if node.Type == storage.NodeTypeAPI {
			boundaries["public_api:"+node.ID] = true
		}
		if node.AuthSensitive {
			boundaries["auth_sensitive:"+node.ID] = true
		}
		if metadata := node.Metadata; metadata != nil {
			if tb, ok := metadata["trust_boundary"]; ok {
				if tbStr, ok := tb.(string); ok {
					boundaries["boundary:"+tbStr+"_"+node.ID] = true
				}
			}
		}
	}

	result := make([]string, 0, len(boundaries))
	for b := range boundaries {
		result = append(result, b)
	}

	return result
}

func (e *ThreatModelEngine) traceExecutionFlows(repoID string, entryPoints []*storage.Node) []ExecutionFlow {
	flows := make([]ExecutionFlow, 0)

	for _, ep := range entryPoints {
		flow := e.traceSingleFlow(ep)
		if flow != nil {
			flows = append(flows, *flow)
		}
	}

	return flows
}

func (e *ThreatModelEngine) traceSingleFlow(entryPoint *storage.Node) *ExecutionFlow {
	visited := make(map[string]bool)
	steps := make([]FlowStep, 0)

	var dfs func(nodeID string, depth int)
	dfs = func(nodeID string, depth int) {
		if depth > 10 || visited[nodeID] {
			return
		}
		visited[nodeID] = true

		node, err := e.store.GetNode(nodeID)
		if err != nil {
			return
		}

		step := FlowStep{
			NodeID:   node.ID,
			NodeName: node.Name,
			Type:     string(node.Type),
			Action:   e.inferAction(node),
		}
		steps = append(steps, step)

		edges, _ := e.store.GetOutgoingEdges(nodeID)
		for _, edge := range edges {
			if !visited[edge.TargetID] {
				dfs(edge.TargetID, depth+1)
			}
		}
	}

	dfs(entryPoint.ID, 0)

	if len(steps) == 0 {
		return nil
	}

	flow := &ExecutionFlow{
		ID:         "flow_" + entryPoint.ID,
		Name:       "Flow from " + entryPoint.Name,
		EntryPoint: entryPoint.ID,
		Steps:      steps,
	}

	flow.AuthRequired = e.checkAuthRequirement(steps)
	flow.HandlesPII = e.checkPIIHandling(steps)
	flow.RiskLevel = e.assessFlowRisk(flow)

	return flow
}

func (e *ThreatModelEngine) inferAction(node *storage.Node) string {
	name := strings.ToLower(node.Name)

	if strings.Contains(name, "read") || strings.Contains(name, "get") || strings.Contains(name, "fetch") {
		return "read"
	}
	if strings.Contains(name, "write") || strings.Contains(name, "create") || strings.Contains(name, "post") {
		return "write"
	}
	if strings.Contains(name, "delete") || strings.Contains(name, "remove") {
		return "delete"
	}
	if strings.Contains(name, "auth") || strings.Contains(name, "login") || strings.Contains(name, "verify") {
		return "authenticate"
	}
	if strings.Contains(name, "exec") || strings.Contains(name, "run") || strings.Contains(name, "execute") {
		return "execute"
	}

	return "process"
}

func (e *ThreatModelEngine) checkAuthRequirement(steps []FlowStep) bool {
	for _, step := range steps {
		if step.Action == "authenticate" {
			return true
		}
		node, _ := e.store.GetNode(step.NodeID)
		if node != nil && node.AuthSensitive {
			return true
		}
	}
	return false
}

func (e *ThreatModelEngine) checkPIIHandling(steps []FlowStep) bool {
	for _, step := range steps {
		node, _ := e.store.GetNode(step.NodeID)
		if node != nil && node.PIIHandling {
			return true
		}
	}
	return false
}

func (e *ThreatModelEngine) assessFlowRisk(flow *ExecutionFlow) string {
	riskScore := 0.0

	if flow.AuthRequired {
		riskScore += 3.0
	}
	if flow.HandlesPII {
		riskScore += 2.5
	}

	for _, step := range flow.Steps {
		if step.Action == "execute" || step.Action == "authenticate" {
			riskScore += 1.0
		}
	}

	switch {
	case riskScore >= 6.0:
		return "critical"
	case riskScore >= 4.0:
		return "high"
	case riskScore >= 2.0:
		return "medium"
	default:
		return "low"
	}
}