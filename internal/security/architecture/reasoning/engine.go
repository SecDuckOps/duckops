package reasoning

import (
	"fmt"
	"strings"

	"github.com/SecDuckOps/duckops/internal/security/architecture"
)

type PromptContext struct {
	ArchitectureSummary string
	TopRisks            []architecture.Finding
	AttackPaths         [][]string
}

func BuildPrompt(ir architecture.IR, findings []architecture.Finding, paths [][]string) string {
	top := findings
	if len(top) > 15 {
		top = top[:15]
	}
	ctx := PromptContext{
		ArchitectureSummary: fmt.Sprintf("services=%d apis=%d flows=%d boundaries=%d", len(ir.Services), len(ir.APIs), len(ir.DataFlows), len(ir.TrustBoundaries)),
		TopRisks:            top,
		AttackPaths:         paths,
	}
	return render(ctx)
}

func render(ctx PromptContext) string {
	var b strings.Builder
	b.WriteString("You are a security architect reasoning over deterministic evidence.\n")
	b.WriteString("Architecture Summary: " + ctx.ArchitectureSummary + "\n")
	b.WriteString("Top Deterministic Findings:\n")
	for _, f := range ctx.TopRisks {
		b.WriteString(fmt.Sprintf("- [%s] %s (%s, %s)\n", strings.ToUpper(f.Severity), f.Title, f.Component, f.CWE))
	}
	b.WriteString("Attack Paths:\n")
	for _, p := range ctx.AttackPaths {
		b.WriteString("- " + strings.Join(p, " -> ") + "\n")
	}
	b.WriteString("Task: correlate, prioritize, and suggest mitigations. Do not invent components not in provided context.\n")
	return b.String()
}
