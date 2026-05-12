package graph_visualizer

import (
	"encoding/json"
	"strings"

	"github.com/SecDuckOps/duckops/internal/security/architecture"
	"github.com/SecDuckOps/duckops/internal/security/architecture/graph"
)

type CytoscapeElement struct {
	Group string         `json:"group"`
	Data  map[string]any `json:"data"`
}

func ToCytoscapeJSON(ir architecture.IR, g graph.Model, attackPaths [][]string) (string, error) {
	var elems []CytoscapeElement
	for _, n := range g.Nodes {
		severity := nodeSeverity(ir, n.ID)
		elems = append(elems, CytoscapeElement{
			Group: "nodes",
			Data: map[string]any{
				"id":       n.ID,
				"label":    firstNonEmpty(n.Attrs["name"], n.ID),
				"type":     n.Type,
				"severity": severity,
			},
		})
	}
	for i, e := range g.Edges {
		elems = append(elems, CytoscapeElement{
			Group: "edges",
			Data: map[string]any{
				"id":        "edge_" + strings.ReplaceAll(strings.ReplaceAll(e.From+"_"+e.To, ":", "_"), "/", "_") + "_" + firstNonEmpty(e.Attrs["protocol"], "na") + "_" + string(rune(i+'0')),
				"source":    e.From,
				"target":    e.To,
				"protocol":  e.Attrs["protocol"],
				"auth":      e.Attrs["auth_method"],
				"pathScore": inAttackPathCount(attackPaths, e.From, e.To),
			},
		})
	}
	raw, err := json.MarshalIndent(elems, "", "  ")
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func nodeSeverity(ir architecture.IR, component string) string {
	sev := "safe"
	for _, f := range ir.Findings {
		if f.Component != component {
			continue
		}
		switch strings.ToLower(f.Severity) {
		case "critical":
			return "critical"
		case "high":
			if sev != "critical" {
				sev = "high"
			}
		case "medium":
			if sev == "safe" {
				sev = "medium"
			}
		}
	}
	return sev
}

func inAttackPathCount(paths [][]string, from, to string) int {
	count := 0
	for _, p := range paths {
		for i := 0; i+1 < len(p); i++ {
			if p[i] == from && p[i+1] == to {
				count++
			}
		}
	}
	return count
}

func firstNonEmpty(v, fallback string) string {
	if v != "" {
		return v
	}
	return fallback
}
