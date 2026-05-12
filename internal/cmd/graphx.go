package cmd

import (
	"github.com/SecDuckOps/duckops/internal/graphx/cli"
	"github.com/spf13/cobra"
)

var graphxCmd = &cobra.Command{
	Use:   "graphx",
	Short: "Persistent Security Code Graph Engine",
	Long: `GraphX is a persistent AI-native code intelligence and security graph engine.

It provides:
- Repository understanding and code intelligence
- Blast radius analysis for changes
- Threat modeling with attack path generation
- Architectural analysis and pattern detection
- Token-efficient AI context generation`,
	SilenceUsage: true,
}

func init() {
	graphxCLI := cli.NewCommands()
	graphxCmd.AddCommand(graphxCLI.GraphCommand())
	graphxCmd.AddCommand(graphxCLI.AnalysisCommand())
	graphxCmd.AddCommand(graphxCLI.ThreatCommand())
	graphxCmd.AddCommand(graphxCLI.ExportCommand())
	graphxCmd.AddCommand(graphxCLI.SearchCommand())

	rootCmd.AddCommand(graphxCmd)
}