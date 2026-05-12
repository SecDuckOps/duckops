package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type JSONStore struct {
	mu     sync.RWMutex
	nodes  map[string]*Node
	edges  map[string]*Edge
	path   string
}

func NewJSONStore(path string) (*JSONStore, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	store := &JSONStore{
		nodes: make(map[string]*Node),
		edges: make(map[string]*Edge),
		path:  path,
	}

	if err := store.load(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return store, nil
}

func (s *JSONStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}

	type StoreData struct {
		Nodes map[string]*Node `json:"nodes"`
		Edges map[string]*Edge `json:"edges"`
	}

	var storeData StoreData
	if err := json.Unmarshal(data, &storeData); err != nil {
		return err
	}

	s.nodes = storeData.Nodes
	s.edges = storeData.Edges

	if s.nodes == nil {
		s.nodes = make(map[string]*Node)
	}
	if s.edges == nil {
		s.edges = make(map[string]*Edge)
	}

	return nil
}

func (s *JSONStore) save() error {
	data, err := json.MarshalIndent(map[string]interface{}{
		"nodes": s.nodes,
		"edges": s.edges,
	}, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0644)
}

func (s *JSONStore) AddNode(node *Node) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nodes[node.ID] = node
	return s.save()
}

func (s *JSONStore) AddEdge(edge *Edge) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.edges[edge.ID] = edge
	return s.save()
}

func (s *JSONStore) GetNode(id string) (*Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	node, ok := s.nodes[id]
	if !ok {
		return nil, fmt.Errorf("node not found: %s", id)
	}
	return node, nil
}

func (s *JSONStore) GetNodesByRepo(repoID string) ([]*Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Node
	for _, node := range s.nodes {
		if node.RepoID == repoID {
			result = append(result, node)
		}
	}
	return result, nil
}

func (s *JSONStore) GetEdgesByNode(nodeID string) ([]*Edge, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Edge
	for _, edge := range s.edges {
		if edge.SourceID == nodeID || edge.TargetID == nodeID {
			result = append(result, edge)
		}
	}
	return result, nil
}

func (s *JSONStore) GetOutgoingEdges(nodeID string) ([]*Edge, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Edge
	for _, edge := range s.edges {
		if edge.SourceID == nodeID {
			result = append(result, edge)
		}
	}
	return result, nil
}

func (s *JSONStore) GetIncomingEdges(nodeID string) ([]*Edge, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Edge
	for _, edge := range s.edges {
		if edge.TargetID == nodeID {
			result = append(result, edge)
		}
	}
	return result, nil
}

func (s *JSONStore) GetNodesByType(repoID string, nodeType NodeType) ([]*Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Node
	for _, node := range s.nodes {
		if node.RepoID == repoID && node.Type == nodeType {
			result = append(result, node)
		}
	}
	return result, nil
}

func (s *JSONStore) GetNodesByFile(repoID, filePath string) ([]*Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Node
	for _, node := range s.nodes {
		if node.RepoID == repoID && node.FilePath == filePath {
			result = append(result, node)
		}
	}
	return result, nil
}

type NodeQuery struct {
	RepoID          string
	NodeTypes       []NodeType
	NamePattern     string
	MinRiskScore    float64
	AuthSensitive   *bool
	InternetExposed *bool
	WithEdges       bool
	Limit           int
	Offset          int
}

func (s *JSONStore) Query(q *NodeQuery) ([]*Node, []*Edge, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var nodes []*Node

	for _, node := range s.nodes {
		if q.RepoID != "" && node.RepoID != q.RepoID {
			continue
		}
		if len(q.NodeTypes) > 0 {
			found := false
			for _, nt := range q.NodeTypes {
				if node.Type == nt {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		if q.NamePattern != "" && !contains(node.Name, q.NamePattern) {
			continue
		}
		if q.MinRiskScore > 0 && node.RiskScore < q.MinRiskScore {
			continue
		}
		if q.AuthSensitive != nil && node.AuthSensitive != *q.AuthSensitive {
			continue
		}
		if q.InternetExposed != nil && node.InternetExposed != *q.InternetExposed {
			continue
		}

		nodes = append(nodes, node)

		if q.Limit > 0 && len(nodes) >= q.Limit {
			break
		}
	}

	var edges []*Edge
	if q.WithEdges {
		nodeIDs := make(map[string]bool)
		for _, n := range nodes {
			nodeIDs[n.ID] = true
		}
		for _, e := range s.edges {
			if nodeIDs[e.SourceID] || nodeIDs[e.TargetID] {
				edges = append(edges, e)
			}
		}
	}

	return nodes, edges, nil
}

func (s *JSONStore) DeleteNode(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.nodes, id)

	var edgesToDelete []string
	for id, edge := range s.edges {
		if edge.SourceID == id || edge.TargetID == id {
			edgesToDelete = append(edgesToDelete, id)
		}
	}
	for _, id := range edgesToDelete {
		delete(s.edges, id)
	}

	return s.save()
}

func (s *JSONStore) DeleteEdgesByNode(nodeID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var edgesToDelete []string
	for id, edge := range s.edges {
		if edge.SourceID == nodeID || edge.TargetID == nodeID {
			edgesToDelete = append(edgesToDelete, id)
		}
	}
	for _, id := range edgesToDelete {
		delete(s.edges, id)
	}

	return s.save()
}

func (s *JSONStore) CountNodes(repoID string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, node := range s.nodes {
		if node.RepoID == repoID {
			count++
		}
	}
	return count, nil
}

func (s *JSONStore) CountEdges(repoID string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, edge := range s.edges {
		if edge.RepoID == repoID {
			count++
		}
	}
	return count, nil
}

func (s *JSONStore) GetHighRiskNodes(repoID string, threshold float64) ([]*Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Node
	for _, node := range s.nodes {
		if node.RepoID == repoID && node.RiskScore >= threshold {
			result = append(result, node)
		}
	}
	return result, nil
}

func (s *JSONStore) GetAuthSensitiveNodes(repoID string) ([]*Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Node
	for _, node := range s.nodes {
		if node.RepoID == repoID && node.AuthSensitive {
			result = append(result, node)
		}
	}
	return result, nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[:len(substr)] == substr || contains(s[1:], substr)))
}

func (s *JSONStore) Close() error {
	return nil
}

func (s *JSONStore) Init() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.nodes == nil {
		s.nodes = make(map[string]*Node)
	}
	if s.edges == nil {
		s.edges = make(map[string]*Edge)
	}
	
	return s.save()
}

func (s *JSONStore) GetAllNodes() ([]*Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	result := make([]*Node, 0, len(s.nodes))
	for _, node := range s.nodes {
		result = append(result, node)
	}
	return result, nil
}

func (s *JSONStore) UpdateNode(node *Node) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.nodes[node.ID] = node
	return s.save()
}

func (s *JSONStore) SearchByName(name string) []*Node {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var result []*Node
	lowerName := strings.ToLower(name)
	
	for _, node := range s.nodes {
		if strings.Contains(strings.ToLower(node.Name), lowerName) ||
			strings.Contains(strings.ToLower(node.FullyQualifiedName), lowerName) {
			result = append(result, node)
		}
	}
	return result
}

func (s *JSONStore) RemoveNodesByFile(filePath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	var nodesToDelete []string
	for id, node := range s.nodes {
		if node.FilePath == filePath {
			nodesToDelete = append(nodesToDelete, id)
		}
	}
	
	for _, id := range nodesToDelete {
		delete(s.nodes, id)
		
		var edgesToDelete []string
		for edgeID, edge := range s.edges {
			if edge.SourceID == id || edge.TargetID == id {
				edgesToDelete = append(edgesToDelete, edgeID)
			}
		}
		for _, edgeID := range edgesToDelete {
			delete(s.edges, edgeID)
		}
	}
	
	return s.save()
}

func (s *JSONStore) GetStats() (int, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	nodeCount := 0
	edgeCount := 0
	
	for _, node := range s.nodes {
		if node.RepoID == s.nodes[node.ID].RepoID {
			nodeCount++
		}
	}
	
	for _, edge := range s.edges {
		if edge.RepoID == s.edges[edge.ID].RepoID {
			edgeCount++
		}
	}
	
	return nodeCount, edgeCount, nil
}