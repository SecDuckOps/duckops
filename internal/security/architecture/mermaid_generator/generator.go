package mermaid_generator

import (
	"fmt"
	"strings"

	"github.com/SecDuckOps/duckops/internal/security/architecture"
)

type Set struct {
	C4Context       string `json:"c4_context"`
	C4Container     string `json:"c4_container"`
	C4Component     string `json:"c4_component"`
	Sequence         string `json:"sequence"`
	Flowchart        string `json:"flowchart"`
	ThreatGraph      string `json:"threat_graph"`
	AttackPaths      string `json:"attack_paths"`
	TrustBoundaries  string `json:"trust_boundaries"`
}

func Generate(ir architecture.IR, attackPaths [][]string) Set {
	return Set{
		C4Context:      c4Context(ir),
		C4Container:    c4Container(ir),
		C4Component:    c4Component(ir),
		Sequence:       sequence(ir),
		Flowchart:      flowchart(ir),
		ThreatGraph:    threatGraph(ir),
		AttackPaths:    attackPathGraph(attackPaths),
		TrustBoundaries: trustBoundaries(ir),
	}
}

func c4Context(ir architecture.IR) string {
	var b strings.Builder
	b.WriteString("flowchart LR\n")
	b.WriteString("User[User] --> System[Platform]\n")
	for _, s := range ir.Services {
		b.WriteString(fmt.Sprintf("System --> %s[%s]\n", sanitizeID(s.ID), sanitizeLabel(s.Name)))
	}
	return b.String()
}

func c4Container(ir architecture.IR) string {
	var b strings.Builder
	b.WriteString("flowchart TD\n")
	for _, s := range ir.Services {
		b.WriteString(fmt.Sprintf("%s[%s (%s)]\n", sanitizeID(s.ID), sanitizeLabel(s.Name), sanitizeLabel(s.Language)))
	}
	for _, d := range ir.Databases {
		b.WriteString(fmt.Sprintf("%s[%s]\n", sanitizeID(d.ID), sanitizeLabel(d.Name)))
	}
	for _, f := range ir.DataFlows {
		b.WriteString(fmt.Sprintf("%s --> %s\n", sanitizeID(f.From), sanitizeID(f.To)))
	}
	return b.String()
}

func c4Component(ir architecture.IR) string {
	var b strings.Builder
	b.WriteString("flowchart LR\n")
	for _, a := range ir.APIs {
		b.WriteString(fmt.Sprintf("%s[%s %s] --> %s\n", sanitizeID(a.ID), a.Method, sanitizeLabel(a.Route), sanitizeID(a.ServiceID)))
	}
	return b.String()
}

func sequence(ir architecture.IR) string {
	var b strings.Builder
	b.WriteString("sequenceDiagram\n")
	b.WriteString("participant User\n")
	if len(ir.APIs) == 0 {
		b.WriteString("User->>System: Request\n")
		b.WriteString("System-->>User: Response\n")
		return b.String()
	}
	for _, s := range ir.Services {
		b.WriteString("participant " + sanitizeID(s.ID) + "\n")
	}
	for _, a := range ir.APIs {
		b.WriteString(fmt.Sprintf("User->>%s: %s %s\n", sanitizeID(a.ServiceID), a.Method, sanitizeLabel(a.Route)))
		b.WriteString(fmt.Sprintf("%s-->>User: response\n", sanitizeID(a.ServiceID)))
	}
	return b.String()
}

func flowchart(ir architecture.IR) string {
	var b strings.Builder
	b.WriteString("flowchart TD\n")
	for _, f := range ir.DataFlows {
		b.WriteString(fmt.Sprintf("%s -->|\"%s\"| %s\n", sanitizeID(f.From), sanitizeLabel(f.Protocol), sanitizeID(f.To)))
	}
	return b.String()
}

func threatGraph(ir architecture.IR) string {
	var b strings.Builder
	b.WriteString("flowchart LR\n")
	for _, f := range ir.Findings {
		nodeID := sanitizeID("finding_" + f.ID)
		comp := sanitizeID(f.Component)
		b.WriteString(fmt.Sprintf("%s[%s]\n", nodeID, sanitizeLabel(f.Title)))
		if comp != "" {
			b.WriteString(fmt.Sprintf("%s --> %s\n", comp, nodeID))
		}
	}
	return b.String()
}

func attackPathGraph(paths [][]string) string {
	var b strings.Builder
	b.WriteString("flowchart LR\n")
	for idx, p := range paths {
		for i := 0; i+1 < len(p); i++ {
			from := sanitizeID(fmt.Sprintf("p%d_%s", idx, p[i]))
			to := sanitizeID(fmt.Sprintf("p%d_%s", idx, p[i+1]))
			b.WriteString(fmt.Sprintf("%s[%s] --> %s[%s]\n", from, sanitizeLabel(p[i]), to, sanitizeLabel(p[i+1])))
		}
	}
	return b.String()
}

func trustBoundaries(ir architecture.IR) string {
	var b strings.Builder
	b.WriteString("flowchart TD\n")
	b.WriteString("subgraph internetZone [internet]\n")
	for _, a := range ir.APIs {
		if a.Public {
			b.WriteString(fmt.Sprintf("%s[%s %s]\n", sanitizeID(a.ID), a.Method, sanitizeLabel(a.Route)))
		}
	}
	b.WriteString("end\n")
	b.WriteString("subgraph internalZone [internal services]\n")
	for _, s := range ir.Services {
		b.WriteString(fmt.Sprintf("%s[%s]\n", sanitizeID(s.ID), sanitizeLabel(s.Name)))
	}
	b.WriteString("end\n")
	return b.String()
}

func sanitizeID(in string) string {
	return strings.NewReplacer(":", "_", "-", "_", "/", "_", ".", "_", " ", "_").Replace(in)
}

func sanitizeLabel(in string) string {
	if in == "" {
		return "unknown"
	}
	return strings.ReplaceAll(in, "\"", "'")
}
