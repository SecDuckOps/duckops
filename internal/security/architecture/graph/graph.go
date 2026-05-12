package graph

import (
	"slices"

	"github.com/SecDuckOps/duckops/internal/security/architecture"
)

type Node struct {
	ID    string
	Type  string
	Attrs map[string]string
}

type Edge struct {
	From  string
	To    string
	Attrs map[string]string
}

type Model struct {
	Nodes map[string]Node
	Edges []Edge
}

func Build(ir architecture.IR) Model {
	m := Model{Nodes: map[string]Node{}}
	for _, s := range ir.Services {
		m.Nodes[s.ID] = Node{ID: s.ID, Type: "service", Attrs: map[string]string{"name": s.Name, "language": s.Language}}
	}
	for _, a := range ir.APIs {
		m.Nodes[a.ID] = Node{ID: a.ID, Type: "api", Attrs: map[string]string{"method": a.Method, "route": a.Route}}
		m.Edges = append(m.Edges, Edge{From: a.ID, To: a.ServiceID, Attrs: map[string]string{"relation": "served_by"}})
	}
	for _, d := range ir.Databases {
		m.Nodes[d.ID] = Node{ID: d.ID, Type: "database", Attrs: map[string]string{"name": d.Name, "type": d.Type}}
	}
	for _, f := range ir.DataFlows {
		m.Edges = append(m.Edges, Edge{
			From: f.From,
			To:   f.To,
			Attrs: map[string]string{
				"protocol":      f.Protocol,
				"auth_method":   f.AuthMethod,
				"trust_zone_src": f.TrustZoneSrc,
				"trust_zone_dst": f.TrustZoneDst,
			},
		})
	}
	return m
}

func AttackPaths(m Model, starts []string, maxDepth int) [][]string {
	var paths [][]string
	for _, s := range starts {
		dfs(m, s, nil, maxDepth, &paths)
	}
	return paths
}

func dfs(m Model, current string, prefix []string, depth int, out *[][]string) {
	if depth < 0 {
		return
	}
	path := append(slices.Clone(prefix), current)
	*out = append(*out, path)
	for _, e := range m.Edges {
		if e.From == current && !slices.Contains(path, e.To) {
			dfs(m, e.To, path, depth-1, out)
		}
	}
}

func Neo4jCypher(m Model) string {
	out := ""
	for _, n := range m.Nodes {
		out += "MERGE (n:" + n.Type + " {id: '" + n.ID + "'})\n"
	}
	for _, e := range m.Edges {
		out += "MATCH (a {id: '" + e.From + "'}), (b {id: '" + e.To + "'}) MERGE (a)-[:CONNECTS_TO]->(b)\n"
	}
	return out
}
