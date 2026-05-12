package community

import (
	"math"
	"math/rand"
	"sort"

	"github.com/SecDuckOps/duckops/internal/graphx/storage"
)

type Engine struct {
	store *storage.JSONStore
}

func NewCommunityEngine(store *storage.JSONStore) *Engine {
	return &Engine{store: store}
}

type Cluster struct {
	ID       int
	Nodes    []*storage.Node
	Edges    []*storage.Edge
	Modularity float64
}

type Partition struct {
	Clusters []Cluster
	Score    float64
}

func (e *Engine) DetectCommunities(repoID string, maxClusters int) (*Partition, error) {
	nodes, edges, err := e.store.Query(&storage.NodeQuery{
		RepoID:    repoID,
		WithEdges: true,
		Limit:     1000,
	})
	if err != nil {
		return nil, err
	}

	adj := e.buildAdjacencyList(nodes, edges)

	partition := &Partition{
		Clusters: e.labelPropagation(nodes, adj),
	}

	partition.Score = e.calculateModularity(nodes, edges, partition.Clusters)

	sort.Slice(partition.Clusters, func(i, j int) bool {
		return len(partition.Clusters[i].Nodes) > len(partition.Clusters[j].Nodes)
	})

	if len(partition.Clusters) > maxClusters {
		partition.Clusters = partition.Clusters[:maxClusters]
	}

	return partition, nil
}

func (e *Engine) buildAdjacencyList(nodes []*storage.Node, edges []*storage.Edge) map[string]map[string]bool {
	adj := make(map[string]map[string]bool)

	for _, n := range nodes {
		adj[n.ID] = make(map[string]bool)
	}

	for _, edge := range edges {
		if adj[edge.SourceID] != nil {
			adj[edge.SourceID][edge.TargetID] = true
		}
		if adj[edge.TargetID] != nil {
			adj[edge.TargetID][edge.SourceID] = true
		}
	}

	return adj
}

func (e *Engine) labelPropagation(nodes []*storage.Node, adj map[string]map[string]bool) []Cluster {
	labels := make(map[string]int)

	for i, n := range nodes {
		labels[n.ID] = i
	}

	iterations := 0
	maxIterations := 50

	for iterations < maxIterations {
		changed := false

		shuffled := make([]*storage.Node, len(nodes))
		copy(shuffled, nodes)
		sort.Slice(shuffled, func(i, j int) bool {
			return rand.Float64() < 0.5
		})

		for _, n := range shuffled {
			if adj[n.ID] == nil {
				continue
			}

			labelCounts := make(map[int]int)
			for neighbor := range adj[n.ID] {
				if label, ok := labels[neighbor]; ok {
					labelCounts[label]++
				}
			}

			if len(labelCounts) == 0 {
				continue
			}

			var maxCount int
			var bestLabel int
			for label, count := range labelCounts {
				if count > maxCount {
					maxCount = count
					bestLabel = label
				}
			}

			if bestLabel != labels[n.ID] {
				labels[n.ID] = bestLabel
				changed = true
			}
		}

		if !changed {
			break
		}
		iterations++
	}

	clusterMap := make(map[int]*Cluster)
	for i, n := range nodes {
		label := labels[n.ID]
		if clusterMap[label] == nil {
			clusterMap[label] = &Cluster{ID: label}
		}
		clusterMap[label].Nodes = append(clusterMap[label].Nodes, nodes[i])
	}

	var clusters []Cluster
	for _, c := range clusterMap {
		clusters = append(clusters, *c)
	}

	return clusters
}

func (e *Engine) calculateModularity(nodes []*storage.Node, edges []*storage.Edge, clusters []Cluster) float64 {
	if len(clusters) == 0 {
		return 0
	}

	nodeCluster := make(map[string]int)
	for _, c := range clusters {
		for _, n := range c.Nodes {
			nodeCluster[n.ID] = c.ID
		}
	}

	m := float64(len(edges))
	if m == 0 {
		return 0
	}

	var Q float64

	for _, edge := range edges {
		if nodeCluster[edge.SourceID] == nodeCluster[edge.TargetID] {
			Q += 1.0
		}
	}

	return Q / m
}

func (e *Engine) FindArchitecturalPatterns(repoID string) (map[string][]string, error) {
	partition, err := e.DetectCommunities(repoID, 10)
	if err != nil {
		return nil, err
	}

	patterns := make(map[string][]string)

	for i, cluster := range partition.Clusters {
		if len(cluster.Nodes) > 20 {
			patterns["large_services"] = append(patterns["large_services"], "cluster_"+string(rune(i)))
		}

		var exportedCount, authCount int
		for _, n := range cluster.Nodes {
			if n.IsExported {
				exportedCount++
			}
			if n.AuthSensitive {
				authCount++
			}
		}

		if exportedCount > len(cluster.Nodes)/2 {
			patterns["public_apis"] = append(patterns["public_apis"], "cluster_"+string(rune(i)))
		}
		if authCount > 0 {
			patterns["auth_boundaries"] = append(patterns["auth_boundaries"], "cluster_"+string(rune(i)))
		}
	}

	return patterns, nil
}

func (e *Engine) FindCoupledClusters(repoID string) ([][2]int, error) {
	partition, err := e.DetectCommunities(repoID, 20)
	if err != nil {
		return nil, err
	}

	var coupled [][2]int
	clusterSlice := partition.Clusters

	for i := 0; i < len(clusterSlice); i++ {
		for j := i + 1; j < len(clusterSlice); j++ {
			if e.areClustersCoupled(&clusterSlice[i], &clusterSlice[j]) {
				coupled = append(coupled, [2]int{i, j})
			}
		}
	}

	return coupled, nil
}

func (e *Engine) areClustersCoupled(a, b *Cluster) bool {
	nodeIDs := make(map[string]bool)
	for _, n := range a.Nodes {
		nodeIDs[n.ID] = true
	}

	for _, n := range b.Nodes {
		edges, _ := e.store.GetEdgesByNode(n.ID)
		for _, edge := range edges {
			if nodeIDs[edge.TargetID] || nodeIDs[edge.SourceID] {
				return true
			}
		}
	}

	return false
}

type BetweennessResult struct {
	NodeID     string
	Score      float64
	Centrality float64
}

func (e *Engine) CalculateBetweenness(repoID string) ([]BetweennessResult, error) {
	nodes, _, err := e.store.Query(&storage.NodeQuery{
		RepoID: repoID,
		Limit:  100,
	})
	if err != nil {
		return nil, err
	}

	results := make([]BetweennessResult, len(nodes))

	nodeMap := make(map[string]int)
	for i, n := range nodes {
		nodeMap[n.ID] = i
	}

	for i, n := range nodes {
		score := e.calculateNodeBetweenness(n.ID, nodes)
		results[i] = BetweennessResult{
			NodeID: n.ID,
			Score:  score,
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

func (e *Engine) calculateNodeBetweenness(nodeID string, nodes []*storage.Node) float64 {
	nodeMap := make(map[string]bool)
	for _, n := range nodes {
		nodeMap[n.ID] = true
	}

	edges, _ := e.store.GetEdgesByNode(nodeID)
	outgoingCount := 0
	for _, edge := range edges {
		if edge.SourceID == nodeID && nodeMap[edge.TargetID] {
			outgoingCount++
		}
	}

	normalizedScore := float64(outgoingCount) / math.Max(1, float64(len(nodes)-1))
	return normalizedScore
}

func (e *Engine) CalculateDegreeCentrality(repoID string) (map[string]float64, error) {
	nodes, edges, err := e.store.Query(&storage.NodeQuery{
		RepoID:    repoID,
		WithEdges: true,
		Limit:     1000,
	})
	if err != nil {
		return nil, err
	}

	degree := make(map[string]int)
	for _, n := range nodes {
		degree[n.ID] = 0
	}

	for _, e := range edges {
		degree[e.SourceID]++
		degree[e.TargetID]++
	}

	maxDegree := 0
	for _, d := range degree {
		if d > maxDegree {
			maxDegree = d
		}
	}

	centrality := make(map[string]float64)
	for id, d := range degree {
		if maxDegree > 0 {
			centrality[id] = float64(d) / float64(maxDegree)
		}
	}

	return centrality, nil
}