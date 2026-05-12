package diff

import (
	"sort"
	"time"

	"github.com/SecDuckOps/duckops/internal/graphx/storage"
)

type Engine struct {
	store *storage.JSONStore
}

func NewEngine(store *storage.JSONStore) *Engine {
	return &Engine{store: store}
}

type Snapshot struct {
	RepoID     string    `json:"repo_id"`
	Timestamp  time.Time `json:"timestamp"`
	NodeCount  int       `json:"node_count"`
	EdgeCount  int       `json:"edge_count"`
	Hash       string    `json:"hash"`
	Nodes      []string  `json:"nodes"`
	APIs       []string  `json:"apis"`
	AuthNodes  []string  `json:"auth_nodes"`
}

type DiffResult struct {
	AddedNodes   []string `json:"added_nodes"`
	RemovedNodes []string `json:"removed_nodes"`
	ModifiedNodes []string `json:"modified_nodes"`
	AddedAPIs    []string `json:"added_apis"`
	RemovedAPIs  []string `json:"removed_apis"`
	NewAuthExposed []string `json:"new_auth_exposed"`
	SecurityDrift SecurityDrift `json:"security_drift"`
}

type SecurityDrift struct {
	NewInternetExposed int  `json:"new_internet_exposed"`
	NewAuthSensitive  int  `json:"new_auth_sensitive"`
	RiskIncrease      float64 `json:"risk_increase"`
}

func (e *Engine) CreateSnapshot(repoID string) (*Snapshot, error) {
	nodes, _, err := e.store.Query(&storage.NodeQuery{
		RepoID: repoID,
		Limit:  10000,
	})
	if err != nil {
		return nil, err
	}

	snap := &Snapshot{
		RepoID:    repoID,
		Timestamp: time.Now(),
		NodeCount: len(nodes),
		Nodes:     make([]string, len(nodes)),
		APIs:      []string{},
		AuthNodes: []string{},
	}

	nodeMap := make(map[string]*storage.Node)
	for i, n := range nodes {
		snap.Nodes[i] = n.ID
		nodeMap[n.ID] = n

		if n.Type == storage.NodeTypeAPI {
			snap.APIs = append(snap.APIs, n.ID)
		}
		if n.AuthSensitive {
			snap.AuthNodes = append(snap.AuthNodes, n.ID)
		}
	}

	edges, _, err := e.store.Query(&storage.NodeQuery{
		RepoID:    repoID,
		WithEdges: true,
		Limit:     10000,
	})
	if err == nil {
		snap.EdgeCount = len(edges)
	}

	return snap, nil
}

func (e *Engine) CompareSnapshots(before, after *Snapshot) *DiffResult {
	result := &DiffResult{
		AddedNodes:    []string{},
		RemovedNodes:  []string{},
		ModifiedNodes: []string{},
		AddedAPIs:     []string{},
		RemovedAPIs:   []string{},
	}

	beforeSet := make(map[string]bool)
	for _, id := range before.Nodes {
		beforeSet[id] = true
	}

	afterSet := make(map[string]bool)
	for _, id := range after.Nodes {
		afterSet[id] = true
	}

	for _, id := range after.Nodes {
		if !beforeSet[id] {
			result.AddedNodes = append(result.AddedNodes, id)
		}
	}

	for _, id := range before.Nodes {
		if !afterSet[id] {
			result.RemovedNodes = append(result.RemovedNodes, id)
		}
	}

	beforeAPIs := make(map[string]bool)
	for _, id := range before.APIs {
		beforeAPIs[id] = true
	}
	afterAPIs := make(map[string]bool)
	for _, id := range after.APIs {
		afterAPIs[id] = true
	}

	for _, id := range after.APIs {
		if !beforeAPIs[id] {
			result.AddedAPIs = append(result.AddedAPIs, id)
		}
	}

	for _, id := range before.APIs {
		if !afterAPIs[id] {
			result.RemovedAPIs = append(result.RemovedAPIs, id)
		}
	}

	result.SecurityDrift = e.calculateSecurityDrift(before, after)

	sort.Strings(result.AddedNodes)
	sort.Strings(result.RemovedNodes)
	sort.Strings(result.AddedAPIs)
	sort.Strings(result.RemovedAPIs)

	return result
}

func (e *Engine) calculateSecurityDrift(before, after *Snapshot) SecurityDrift {
	drift := SecurityDrift{}

	beforeNodes, _, _ := e.store.Query(&storage.NodeQuery{
		RepoID: before.RepoID,
		Limit: 10000,
	})

	afterNodes, _, _ := e.store.Query(&storage.NodeQuery{
		RepoID: after.RepoID,
		Limit: 10000,
	})

	beforeMap := make(map[string]*storage.Node)
	for _, n := range beforeNodes {
		beforeMap[n.ID] = n
	}

	afterMap := make(map[string]*storage.Node)
	for _, n := range afterNodes {
		afterMap[n.ID] = n
	}

	var beforeExposed, afterExposed int
	var beforeAuth, afterAuth int
	var beforeRisk, afterRisk float64

	for _, n := range beforeNodes {
		if n.InternetExposed {
			beforeExposed++
		}
		if n.AuthSensitive {
			beforeAuth++
		}
		beforeRisk += n.RiskScore
	}

	for _, n := range afterNodes {
		_, wasBefore := beforeMap[n.ID]
		if n.InternetExposed && !wasBefore {
			drift.NewInternetExposed++
		}
		if n.InternetExposed {
			afterExposed++
		}
		if n.AuthSensitive && !wasBefore {
			drift.NewAuthSensitive++
		}
		if n.AuthSensitive {
			afterAuth++
		}
		afterRisk += n.RiskScore
	}

	if len(beforeNodes) > 0 {
		drift.RiskIncrease = (afterRisk - beforeRisk) / float64(len(beforeNodes))
	}

	return drift
}

func (e *Engine) GenerateReport(diff *DiffResult) string {
	_ = diff
	return ""
}