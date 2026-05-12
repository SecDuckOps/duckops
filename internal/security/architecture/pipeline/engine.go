package pipeline

import (
	"encoding/json"
	"path/filepath"

	"github.com/SecDuckOps/duckops/internal/security/architecture"
	"github.com/SecDuckOps/duckops/internal/security/architecture/architecture_diff"
	"github.com/SecDuckOps/duckops/internal/security/architecture/attack_path_renderer"
	"github.com/SecDuckOps/duckops/internal/security/architecture/attacksurface"
	"github.com/SecDuckOps/duckops/internal/security/architecture/c4"
	"github.com/SecDuckOps/duckops/internal/security/architecture/diagram_validation"
	"github.com/SecDuckOps/duckops/internal/security/architecture/graph"
	"github.com/SecDuckOps/duckops/internal/security/architecture/graph_visualizer"
	irpkg "github.com/SecDuckOps/duckops/internal/security/architecture/ir"
	"github.com/SecDuckOps/duckops/internal/security/architecture/mermaid_generator"
	"github.com/SecDuckOps/duckops/internal/security/architecture/parser"
	"github.com/SecDuckOps/duckops/internal/security/architecture/reasoning"
	"github.com/SecDuckOps/duckops/internal/security/architecture/repoanalyzer"
	"github.com/SecDuckOps/duckops/internal/security/architecture/reporting"
	"github.com/SecDuckOps/duckops/internal/security/architecture/threat"
	"github.com/SecDuckOps/duckops/internal/security/architecture/validation"
)

type Result struct {
	IR              architecture.IR
	Graph           graph.Model
	AttackPaths     [][]string
	C4              c4.Models
	ReasoningPrompt string
	Reports         reporting.Bundle
	IRSnapshotPath  string
	Neo4jCypher     string
	DiagramValidation []diagram_validation.Result
}

func AnalyzeRepository(repoPath, outDir string) (Result, error) {
	var result Result
	_, err := repoanalyzer.Analyze(repoPath)
	if err != nil {
		return result, err
	}

	ir, err := parser.ParseRepository(repoPath)
	if err != nil {
		return result, err
	}
	ir.Findings = append(ir.Findings, threat.Generate(ir)...)
	ir.Findings = append(ir.Findings, threat.AbuseCases(ir)...)
	ir.Findings = append(ir.Findings, attacksurface.Analyze(ir)...)
	ir.Findings = append(ir.Findings, validation.Validate(ir)...)

	g := graph.Build(ir)
	starts := []string{}
	for _, a := range ir.APIs {
		if a.Public {
			starts = append(starts, a.ID)
		}
	}
	paths := graph.AttackPaths(g, starts, 4)
	diagrams := c4.Generate(ir)
	mermaidSet := mermaid_generator.Generate(ir, paths)
	validationInput := map[string]string{
		"c4_context":       mermaidSet.C4Context,
		"c4_container":     mermaidSet.C4Container,
		"c4_component":     mermaidSet.C4Component,
		"sequence":         mermaidSet.Sequence,
		"flowchart":        mermaidSet.Flowchart,
		"threat_graph":     mermaidSet.ThreatGraph,
		"attack_paths":     mermaidSet.AttackPaths,
		"trust_boundaries": mermaidSet.TrustBoundaries,
	}
	validations, _ := diagram_validation.ValidateSet(validationInput)
	cytoscapeJSON, err := graph_visualizer.ToCytoscapeJSON(ir, g, paths)
	if err != nil {
		return result, err
	}
	replays := make([][]attack_path_renderer.Frame, 0, len(paths))
	for _, p := range paths {
		replays = append(replays, attack_path_renderer.BuildReplay(p))
	}
	replayRaw, err := json.MarshalIndent(replays, "", "  ")
	if err != nil {
		return result, err
	}
	prompt := reasoning.BuildPrompt(ir, ir.Findings, paths)
	reports, err := reporting.Build(ir, diagrams, mermaidSet, validations, cytoscapeJSON, string(replayRaw), paths)
	if err != nil {
		return result, err
	}
	if err := reporting.WriteAll(outDir, reports); err != nil {
		return result, err
	}
	if _, err := architecture_diff.Record(filepath.Join(outDir, "cache", "history.json"), []byte(reports.IRJSON), filepath.Join(outDir, "report.json")); err != nil {
		return result, err
	}
	snap, err := irpkg.SaveSnapshot(filepath.Join(outDir, "cache"), ir)
	if err != nil {
		return result, err
	}

	result = Result{
		IR:              ir,
		Graph:           g,
		AttackPaths:     paths,
		C4:              diagrams,
		ReasoningPrompt: prompt,
		Reports:         reports,
		IRSnapshotPath:  snap,
		Neo4jCypher:     graph.Neo4jCypher(g),
		DiagramValidation: validations,
	}
	return result, nil
}
