package blast

import (
	"container/heap"
	"math"
	"sort"

	"github.com/SecDuckOps/duckops/internal/graphx/storage"
)

type BlastRadiusEngine struct {
	store *storage.JSONStore
	opts  *BlastOpts
}

type BlastOpts struct {
	MaxDepth        int
	WeightAuth      float64
	WeightAPI       float64
	WeightTrust     float64
	IncludeEdgeTypes []storage.EdgeType
	ExcludeNodeIDs  []string
}

func DefaultBlastOpts() *BlastOpts {
	return &BlastOpts{
		MaxDepth:    10,
		WeightAuth:  2.0,
		WeightAPI:   1.5,
		WeightTrust: 2.5,
	}
}

type BlastResult struct {
	AffectedNodes   []*storage.Node
	AffectedEdges   []*storage.Edge
	AuthFlows       []string
	ImpactedAPIs    []string
	TrustBoundaries []string
	RiskScore       float64
	CriticalPath    []string
}

func NewBlastRadiusEngine(store *storage.JSONStore) *BlastRadiusEngine {
	return &BlastRadiusEngine{
		store: store,
		opts:  DefaultBlastOpts(),
	}
}

func (e *BlastRadiusEngine) Analyze(nodeID string) (*BlastResult, error) {
	node, err := e.store.GetNode(nodeID)
	if err != nil {
		return nil, err
	}

	result := &BlastResult{
		RiskScore: node.RiskScore,
	}

	visited := make(map[string]bool)
	queue := &priorityQueue{}
	heap.Init(queue)

	heap.Push(queue, &queueItem{
		nodeID:  nodeID,
		depth:   0,
		weight:  1.0,
		priority: 1.0,
	})

	affectedNodes := make([]*storage.Node, 0)
	affectedEdges := make([]*storage.Edge, 0)

	for queue.Len() > 0 {
		item := heap.Pop(queue).(*queueItem)

		if visited[item.nodeID] {
			continue
		}
		visited[item.nodeID] = true

		n, err := e.store.GetNode(item.nodeID)
		if err != nil {
			continue
		}
		affectedNodes = append(affectedNodes, n)

		if n.AuthSensitive {
			result.AuthFlows = append(result.AuthFlows, n.ID)
		}
		if n.Type == storage.NodeTypeAPI {
			result.ImpactedAPIs = append(result.ImpactedAPIs, n.ID)
		}

		edges, err := e.store.GetOutgoingEdges(item.nodeID)
		if err != nil {
			continue
		}

		for _, edge := range edges {
			affectedEdges = append(affectedEdges, edge)

			if item.depth+1 < e.opts.MaxDepth {
				weight := e.calculateWeight(edge.EdgeType)
				newPriority := item.priority * weight
				heap.Push(queue, &queueItem{
					nodeID:  edge.TargetID,
					depth:   item.depth + 1,
					weight:  item.weight * weight,
					priority: newPriority,
				})
			}
		}
	}

	result.AffectedNodes = affectedNodes
	result.AffectedEdges = affectedEdges
	result.RiskScore = e.calculateRiskScore(affectedNodes)
	result.CriticalPath = e.findCriticalPath(affectedNodes)

	return result, nil
}

func (e *BlastRadiusEngine) calculateWeight(edgeType storage.EdgeType) float64 {
	switch edgeType {
	case storage.EdgeType("auth_dep"):
		return e.opts.WeightAuth
	case storage.EdgeType("api_call"):
		return e.opts.WeightAPI
	case storage.EdgeType("trust_boundary"):
		return e.opts.WeightTrust
	default:
		return 1.0
	}
}

func (e *BlastRadiusEngine) calculateRiskScore(nodes []*storage.Node) float64 {
	if len(nodes) == 0 {
		return 0.0
	}

	var totalRisk float64
	var maxRisk float64

	for _, n := range nodes {
		risk := n.RiskScore
		totalRisk += risk
		if risk > maxRisk {
			maxRisk = risk
		}
	}

	avgRisk := totalRisk / float64(len(nodes))

	return math.Min(10.0, (avgRisk*0.4)+(maxRisk*0.6))
}

func (e *BlastRadiusEngine) findCriticalPath(nodes []*storage.Node) []string {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].RiskScore > nodes[j].RiskScore
	})

	criticalPath := make([]string, 0, min(10, len(nodes)))
	for _, n := range nodes[:min(10, len(nodes))] {
		if n.RiskScore > 0.5 {
			criticalPath = append(criticalPath, n.ID)
		}
	}

	return criticalPath
}

func (e *BlastRadiusEngine) FindAffectedAuthFlows(nodeID string) ([]string, error) {
	result, err := e.Analyze(nodeID)
	if err != nil {
		return nil, err
	}
	return result.AuthFlows, nil
}

func (e *BlastRadiusEngine) FindImpactedAPIs(nodeID string) ([]string, error) {
	result, err := e.Analyze(nodeID)
	if err != nil {
		return nil, err
	}
	return result.ImpactedAPIs, nil
}

func (e *BlastRadiusEngine) CalculateRiskScore(nodeID string) (float64, error) {
	result, err := e.Analyze(nodeID)
	if err != nil {
		return 0.0, err
	}
	return result.RiskScore, nil
}

type queueItem struct {
	nodeID   string
	depth    int
	weight   float64
	priority float64
	index    int
}

type priorityQueue []*queueItem

func (pq priorityQueue) Len() int { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool { return pq[i].priority > pq[j].priority }
func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}
func (pq *priorityQueue) Push(x interface{}) {
	item := x.(*queueItem)
	item.index = len(*pq)
	*pq = append(*pq, item)
}
func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[:n-1]
	return item
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}