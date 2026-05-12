package semantic

import (
	"math"
	"sort"
	"strings"

	"github.com/SecDuckOps/duckops/internal/graphx/storage"
)

type Engine struct {
	store    *storage.JSONStore
	embeddings map[string][]float32
}

func NewEngine(store *storage.JSONStore) *Engine {
	return &Engine{
		store:     store,
		embeddings: make(map[string][]float32),
	}
}

type SearchResult struct {
	Node     *storage.Node
	Score    float32
	Keywords []string
}

func (e *Engine) Search(query string, limit int) ([]SearchResult, error) {
	queryLower := strings.ToLower(query)
	queryWords := strings.Fields(queryLower)

	nodes, _, err := e.store.Query(&storage.NodeQuery{
		Limit: 1000,
	})
	if err != nil {
		return nil, err
	}

	var results []SearchResult

	for _, n := range nodes {
		score := e.calculateScore(n, queryWords)

		if score > 0.1 {
			results = append(results, SearchResult{
				Node:     n,
				Score:    score,
				Keywords: e.extractKeywords(n, queryWords),
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

func (e *Engine) calculateScore(n *storage.Node, queryWords []string) float32 {
	var score float32

	nameLower := strings.ToLower(n.Name)
	pathLower := strings.ToLower(n.FilePath)

	for _, word := range queryWords {
		if strings.Contains(nameLower, word) {
			score += 1.0
		}
		if strings.Contains(pathLower, word) {
			score += 0.5
		}
		if n.FullyQualifiedName != "" && strings.Contains(strings.ToLower(n.FullyQualifiedName), word) {
			score += 0.8
		}
	}

	if n.AuthSensitive {
		score *= 1.5
	}
	if n.RiskScore > 0.7 {
		score *= 1.3
	}
	if n.SecuritySensitivity > 0.5 {
		score *= 1.4
	}
	if n.InternetExposed {
		score *= 1.2
	}

	return score
}

func (e *Engine) extractKeywords(n *storage.Node, queryWords []string) []string {
	var keywords []string

	nameLower := strings.ToLower(n.Name)
	for _, word := range queryWords {
		if strings.Contains(nameLower, word) {
			keywords = append(keywords, word)
		}
	}

	if n.AuthSensitive {
		keywords = append(keywords, "auth")
	}
	if n.RiskScore > 0.7 {
		keywords = append(keywords, "high-risk")
	}
	if n.InternetExposed {
		keywords = append(keywords, "internet-exposed")
	}
	if n.Type == storage.NodeTypeAPI {
		keywords = append(keywords, "api")
	}

	return keywords
}

func (e *Engine) SemanticSearch(query string, limit int) ([]SearchResult, error) {
	queryVec := e.embedQuery(query)

	nodes, _, err := e.store.Query(&storage.NodeQuery{
		Limit: 1000,
	})
	if err != nil {
		return nil, err
	}

	var results []SearchResult

	for _, n := range nodes {
		nodeVec, exists := e.embeddings[n.ID]
		if !exists {
			continue
		}

		similarity := cosineSimilarity(queryVec, nodeVec)
		if similarity > 0.3 {
			results = append(results, SearchResult{
				Node:  n,
				Score: similarity,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

func (e *Engine) embedQuery(query string) []float32 {
	words := strings.Fields(strings.ToLower(query))
	vec := make([]float32, 128)

	for i, word := range words {
		if i >= len(vec) {
			break
		}
		vec[i] = float32(len(word)) / 100.0
	}

	return normalize(vec)
}

func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dot float32
	var aMag, bMag float32

	for i := range a {
		dot += a[i] * b[i]
		aMag += a[i] * a[i]
		bMag += b[i] * b[i]
	}

	if aMag == 0 || bMag == 0 {
		return 0
	}

	return dot / (float32(math.Sqrt(float64(aMag))) * float32(math.Sqrt(float64(bMag))))
}

func normalize(v []float32) []float32 {
	var mag float32
	for _, x := range v {
		mag += x * x
	}
	if mag == 0 {
		return v
	}
	mag = float32(math.Sqrt(float64(mag)))
	for i := range v {
		v[i] /= mag
	}
	return v
}

func (e *Engine) HybridSearch(query string, limit int) ([]SearchResult, error) {
	keywordResults, err := e.Search(query, limit*2)
	if err != nil {
		return nil, err
	}

	semanticResults, err := e.SemanticSearch(query, limit*2)
	if err != nil {
		semanticResults = nil
	}

	combined := make(map[string]*SearchResult)

	for _, r := range keywordResults {
		r := r
		combined[r.Node.ID] = &r
	}

	for _, r := range semanticResults {
		if existing, ok := combined[r.Node.ID]; ok {
			existing.Score = (existing.Score + r.Score) / 2
		} else {
			r := r
			combined[r.Node.ID] = &r
		}
	}

	var results []SearchResult
	for _, r := range combined {
		results = append(results, *r)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}