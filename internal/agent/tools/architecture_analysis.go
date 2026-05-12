package tools

import (
	"context"
	_ "embed"
	"fmt"
	"path/filepath"
	"strings"

	"charm.land/fantasy"
	"github.com/SecDuckOps/duckops/internal/permission"
	"github.com/SecDuckOps/duckops/internal/security/architecture/pipeline"
)

//go:embed architecture_analysis.md
var architectureAnalysisDescription []byte

type ArchitectureAnalysisParams struct {
	RepositoryPath string `json:"repository_path" description:"Repository path to analyze. Defaults to current working directory."`
	OutputDir      string `json:"output_dir" description:"Output directory for reports and generated diagrams. Defaults to .duckops/architecture-report."`
}

const (
	AnalyzeArchitectureToolName  = "analyze_architecture"
	GenerateC4ModelToolName      = "generate_c4_model"
	GenerateThreatModelToolName  = "generate_threat_model"
	GenerateAttackPathsToolName  = "generate_attack_paths"
	ExportSecurityReportToolName = "export_security_report"
)

func NewAnalyzeArchitectureTool(permissions permission.Service, workingDir string) fantasy.AgentTool {
	return newArchitectureTool(AnalyzeArchitectureToolName, permissions, workingDir)
}
func NewGenerateC4ModelTool(permissions permission.Service, workingDir string) fantasy.AgentTool {
	return newArchitectureTool(GenerateC4ModelToolName, permissions, workingDir)
}
func NewGenerateThreatModelTool(permissions permission.Service, workingDir string) fantasy.AgentTool {
	return newArchitectureTool(GenerateThreatModelToolName, permissions, workingDir)
}
func NewGenerateAttackPathsTool(permissions permission.Service, workingDir string) fantasy.AgentTool {
	return newArchitectureTool(GenerateAttackPathsToolName, permissions, workingDir)
}
func NewExportSecurityReportTool(permissions permission.Service, workingDir string) fantasy.AgentTool {
	return newArchitectureTool(ExportSecurityReportToolName, permissions, workingDir)
}

func newArchitectureTool(name string, permissions permission.Service, workingDir string) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		name,
		FirstLineDescription(architectureAnalysisDescription),
		func(ctx context.Context, params ArchitectureAnalysisParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			repoPath := params.RepositoryPath
			if repoPath == "" {
				repoPath = workingDir
			}
			outDir := params.OutputDir
			if outDir == "" {
				outDir = filepath.Join(workingDir, ".duckops", "architecture-report")
			}
			sessionID := GetSessionFromContext(ctx)
			allowed, err := permissions.Request(ctx, permission.CreatePermissionRequest{
				SessionID:   sessionID,
				Path:        outDir,
				ToolCallID:  call.ID,
				ToolName:    name,
				Action:      "write",
				Description: "Generate threat model and architecture outputs",
				Params:      params,
			})
			if err != nil {
				return fantasy.ToolResponse{}, err
			}
			if !allowed {
				return NewPermissionDeniedResponse(), nil
			}

			result, err := pipeline.AnalyzeRepository(repoPath, outDir)
			if err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}
			var b strings.Builder
			b.WriteString("# Threat Modeling + C4 Analysis\n\n")
			b.WriteString(fmt.Sprintf("- Repository: `%s`\n", repoPath))
			b.WriteString(fmt.Sprintf("- Findings: `%d`\n", len(result.IR.Findings)))
			b.WriteString(fmt.Sprintf("- Attack paths: `%d`\n", len(result.AttackPaths)))
			b.WriteString(fmt.Sprintf("- IR snapshot: `%s`\n", result.IRSnapshotPath))
			b.WriteString(fmt.Sprintf("- Output directory: `%s`\n\n", outDir))
			b.WriteString("## Generated Artifacts\n\n")
			b.WriteString("- `report.md`\n")
			b.WriteString("- `report.json`\n")
			b.WriteString("- `report.sarif`\n")
			b.WriteString("- `dashboard.html`\n")
			b.WriteString("- `c4_context.mmd`\n")
			b.WriteString("- `c4_container.mmd`\n")
			b.WriteString("- `c4_component.mmd`\n")
			b.WriteString("- `c4.puml`\n")
			b.WriteString("- `workspace.dsl`\n\n")
			b.WriteString("## Top Findings\n\n")
			limit := min(5, len(result.IR.Findings))
			if limit == 0 {
				b.WriteString("- No findings generated.\n")
			} else {
				for i := 0; i < limit; i++ {
					f := result.IR.Findings[i]
					b.WriteString(fmt.Sprintf("- [%s] %s (`%s`)\n", strings.ToUpper(f.Severity), f.Title, f.Component))
				}
			}
			msg := b.String()
			return fantasy.NewTextResponse(msg), nil
		},
	)
}
