package diagram_validation

import (
	"errors"
	"strings"
)

type Result struct {
	Name              string   `json:"name"`
	ValidSyntax       bool     `json:"valid_syntax"`
	DisconnectedNodes []string `json:"disconnected_nodes,omitempty"`
	MalformedEdges    []string `json:"malformed_edges,omitempty"`
	Errors            []string `json:"errors,omitempty"`
}

func Validate(name, diagram string) Result {
	r := Result{Name: name, ValidSyntax: true}
	lines := strings.Split(diagram, "\n")
	nodes := map[string]bool{}
	connected := map[string]bool{}

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "flowchart") || strings.HasPrefix(line, "sequenceDiagram") || strings.HasPrefix(line, "subgraph") || line == "end" {
			continue
		}
		if strings.Contains(line, "-->") {
			parts := strings.Split(line, "-->")
			if len(parts) != 2 {
				r.MalformedEdges = append(r.MalformedEdges, line)
				r.ValidSyntax = false
				continue
			}
			left := extractNodeID(parts[0])
			right := extractNodeID(parts[1])
			if left == "" || right == "" {
				r.MalformedEdges = append(r.MalformedEdges, line)
				r.ValidSyntax = false
				continue
			}
			nodes[left] = true
			nodes[right] = true
			connected[left] = true
			connected[right] = true
			continue
		}
		nodeID := extractNodeID(line)
		if nodeID != "" {
			nodes[nodeID] = true
		}
	}

	for n := range nodes {
		if !connected[n] {
			r.DisconnectedNodes = append(r.DisconnectedNodes, n)
		}
	}
	if !r.ValidSyntax {
		r.Errors = append(r.Errors, "diagram syntax validation failed")
	}
	return r
}

func ValidateSet(diagrams map[string]string) ([]Result, error) {
	results := make([]Result, 0, len(diagrams))
	anyInvalid := false
	for name, d := range diagrams {
		res := Validate(name, d)
		if !res.ValidSyntax {
			anyInvalid = true
		}
		results = append(results, res)
	}
	if anyInvalid {
		return results, errors.New("one or more diagrams are invalid")
	}
	return results, nil
}

func extractNodeID(part string) string {
	p := strings.TrimSpace(part)
	if p == "" {
		return ""
	}
	if idx := strings.Index(p, "|"); idx >= 0 {
		p = strings.TrimSpace(p[idx+1:])
	}
	p = strings.TrimSpace(p)
	for _, sep := range []string{"[", "(", "{", "\""} {
		if idx := strings.Index(p, sep); idx >= 0 {
			p = p[:idx]
			break
		}
	}
	return strings.TrimSpace(strings.Trim(p, " "))
}
