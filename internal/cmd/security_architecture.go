package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SecDuckOps/duckops/internal/security/architecture/diagram_renderer"
	"github.com/SecDuckOps/duckops/internal/security/architecture/pipeline"
	"github.com/spf13/cobra"
)

var threatModelCmd = &cobra.Command{
	Use:   "threat-model [repo]",
	Short: "Generate threat model from repository",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := "."
		if len(args) == 1 {
			repo = args[0]
		}
		outDir, _ := cmd.Flags().GetString("output")
		if outDir == "" {
			outDir = filepath.Join(".duckops", "threat-model")
		}
		result, err := pipeline.AnalyzeRepository(repo, outDir)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Threat model generated: findings=%d output=%s\n", len(result.IR.Findings), outDir)
		return nil
	},
}

var architectureReviewCmd = &cobra.Command{
	Use:   "architecture",
	Short: "Architecture analysis commands",
}

var architectureReviewRunCmd = &cobra.Command{
	Use:   "review [repo]",
	Short: "Generate architecture review",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := "."
		if len(args) == 1 {
			repo = args[0]
		}
		outDir, _ := cmd.Flags().GetString("output")
		if outDir == "" {
			outDir = filepath.Join(".duckops", "architecture-review")
		}
		result, err := pipeline.AnalyzeRepository(repo, outDir)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Architecture review completed: services=%d apis=%d output=%s\n", len(result.IR.Services), len(result.IR.APIs), outDir)
		return nil
	},
}

var c4Cmd = &cobra.Command{
	Use:   "c4",
	Short: "C4 architecture commands",
}

var c4GenerateCmd = &cobra.Command{
	Use:   "generate [repo]",
	Short: "Generate C4 diagrams from repository",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := "."
		if len(args) == 1 {
			repo = args[0]
		}
		outDir, _ := cmd.Flags().GetString("output")
		if outDir == "" {
			outDir = filepath.Join(".duckops", "c4")
		}
		result, err := pipeline.AnalyzeRepository(repo, outDir)
		if err != nil {
			return err
		}
		if err := renderExports(cmd, outDir, result.C4.Context); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "C4 generated: context=%d chars output=%s\n", len(result.C4.Context), outDir)
		return nil
	},
}

func init() {
	threatModelCmd.Flags().StringP("output", "o", "", "Output directory")
	c4GenerateCmd.Flags().StringP("output", "o", "", "Output directory")
	c4GenerateCmd.Flags().Bool("export-svg", false, "Render C4 context to SVG via Mermaid CLI (mmdc)")
	c4GenerateCmd.Flags().Bool("export-png", false, "Render C4 context to PNG via Mermaid CLI (mmdc)")
	c4GenerateCmd.Flags().Bool("export-pdf", false, "Render C4 context to PDF via Mermaid CLI (mmdc)")
	architectureReviewRunCmd.Flags().StringP("output", "o", "", "Output directory")

	c4Cmd.AddCommand(c4GenerateCmd)
	architectureReviewCmd.AddCommand(architectureReviewRunCmd)
}

func renderExports(cmd *cobra.Command, outDir, source string) error {
	if v, _ := cmd.Flags().GetBool("export-svg"); v {
		if err := diagram_renderer.RenderWithMermaidCLI(source, filepath.Join(outDir, "c4_context.svg")); err != nil {
			return err
		}
	}
	if v, _ := cmd.Flags().GetBool("export-png"); v {
		if err := diagram_renderer.RenderWithMermaidCLI(source, filepath.Join(outDir, "c4_context.png")); err != nil {
			return err
		}
	}
	if v, _ := cmd.Flags().GetBool("export-pdf"); v {
		if err := diagram_renderer.RenderWithMermaidCLI(source, filepath.Join(outDir, "c4_context.pdf")); err != nil {
			return err
		}
	}
	if v, _ := cmd.Flags().GetBool("export-svg"); !v {
		if v2, _ := cmd.Flags().GetBool("export-png"); !v2 {
			if v3, _ := cmd.Flags().GetBool("export-pdf"); !v3 {
				return nil
			}
		}
	}
	if _, err := os.Stat(filepath.Join(outDir, "c4_context.mmd")); err != nil {
		return err
	}
	return nil
}
