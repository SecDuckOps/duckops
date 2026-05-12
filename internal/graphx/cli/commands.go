package cli

import (
	"fmt"
	"os"

	"github.com/SecDuckOps/duckops/internal/graphx/config"
	"github.com/SecDuckOps/duckops/internal/graphx/engine"
	"github.com/SecDuckOps/duckops/internal/graphx/storage"
	"github.com/SecDuckOps/duckops/internal/graphx/analysis/blast"
	"github.com/SecDuckOps/duckops/internal/graphx/analysis/threat"
	"github.com/spf13/cobra"
)

type Commands struct {
	root *cobra.Command
	cfg  *config.Config
	eng  *engine.Engine
}

func NewCommands() *Commands {
	return &Commands{
		cfg: config.Default(),
	}
}

func (c *Commands) Root() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "graphx",
		Short: "GraphX - Persistent Security Code Graph Engine",
		Long: `GraphX is a persistent AI-native code intelligence and security graph engine.
It provides repository understanding, architectural reasoning, security analysis,
threat modeling, blast radius analysis, and attack path simulation.`,
		SilenceUsage: true,
	}

	rootCmd.AddCommand(c.GraphCommand())
	rootCmd.AddCommand(c.AnalysisCommand())
	rootCmd.AddCommand(c.ThreatCommand())
	rootCmd.AddCommand(c.ExportCommand())
	rootCmd.AddCommand(c.SearchCommand())

	return rootCmd
}

func (c *Commands) GraphCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "graph",
		Short: "Graph management commands",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "init [path]",
		Short: "Initialize a repository for graph analysis",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				c.cfg.Repository.Path = args[0]
			}
			return c.initGraph()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "build",
		Short: "Build the knowledge graph",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.buildGraph()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Show graph status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.graphStatus()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "watch",
		Short: "Watch for file changes and rebuild",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.watchGraph()
		},
	})

	return cmd
}

func (c *Commands) AnalysisCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "analyze",
		Short: "Security analysis commands",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "blast-radius [path]",
		Short: "Analyze blast radius of changes",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.blastRadius(args)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "risks",
		Short: "List high-risk components",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.listRisks()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "auth-flows",
		Short: "Analyze authentication flows",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.analyzeAuthFlows()
		},
	})

	return cmd
}

func (c *Commands) ThreatCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "threat-model",
		Short: "Threat modeling commands",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "generate",
		Short: "Generate threat model",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.generateThreatModel()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "attack-paths",
		Short: "Find potential attack paths",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.findAttackPaths()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "boundaries",
		Short: "List trust boundaries",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.listTrustBoundaries()
		},
	})

	return cmd
}

func (c *Commands) ExportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "export",
		Short: "Export graph in various formats",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "mermaid",
		Short: "Export as Mermaid diagram",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.exportMermaid()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "graphml",
		Short: "Export as GraphML",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.exportGraphML()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "json",
		Short: "Export as JSON",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.exportJSON()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "cypher",
		Short: "Export as Neo4j Cypher",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.exportCypher()
		},
	})

	return cmd
}

func (c *Commands) SearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "search",
		Short: "Search the knowledge graph",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.search(args)
		},
	}

	cmd.Flags().IntP("limit", "l", 10, "Maximum results")
	cmd.Flags().BoolP("security", "s", false, "Focus on security-relevant results")

	return cmd
}

func (c *Commands) initGraph() error {
	fmt.Println("Initializing GraphX graph...")

	cfgPath := "graphx.yaml"

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		if err := config.Save(c.cfg, cfgPath); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Printf("Created default config: %s\n", cfgPath)
	} else {
		loadedCfg, err := config.Load(cfgPath)
		if err == nil {
			c.cfg = loadedCfg
		}
	}

	if err := config.Save(c.cfg, cfgPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	eng, err := engine.NewEngine(c.cfg)
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}
	defer eng.Close()

	if err := config.Save(c.cfg, cfgPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("Graph initialized successfully!")
	fmt.Printf("Repository: %s\n", c.cfg.Repository.Path)
	fmt.Printf("Database: %s\n", c.cfg.Graph.Path)

	return nil
}

func (c *Commands) buildGraph() error {
	cfgPath := "graphx.yaml"
	if loadedCfg, err := config.Load(cfgPath); err == nil {
		c.cfg = loadedCfg
	}

	fmt.Println("Building knowledge graph...")

	eng, err := engine.NewEngine(c.cfg)
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}
	defer eng.Close()

	if err := eng.Build(nil); err != nil {
		return fmt.Errorf("failed to build graph: %w", err)
	}

	store := eng.GetStore()
	nodeCount, _ := store.CountNodes(eng.GetRepoID())
	edgeCount, _ := store.CountEdges(eng.GetRepoID())

	fmt.Printf("Graph built successfully!\n")
	fmt.Printf("Nodes: %d\n", nodeCount)
	fmt.Printf("Edges: %d\n", edgeCount)

	return nil
}

func (c *Commands) graphStatus() error {
	eng, err := engine.NewEngine(c.cfg)
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}
	defer eng.Close()

	store := eng.GetStore()
	nodeCount, _ := store.CountNodes(eng.GetRepoID())
	edgeCount, _ := store.CountEdges(eng.GetRepoID())
	risks, _ := store.GetHighRiskNodes(eng.GetRepoID(), 0.7)

	fmt.Printf("Graph Status\n")
	fmt.Printf("============\n")
	fmt.Printf("Repository: %s\n", c.cfg.Repository.Path)
	fmt.Printf("Nodes: %d\n", nodeCount)
	fmt.Printf("Edges: %d\n", edgeCount)
	fmt.Printf("High-risk components: %d\n", len(risks))

	return nil
}

func (c *Commands) watchGraph() error {
	fmt.Println("Watch mode not yet implemented")
	return nil
}

func (c *Commands) blastRadius(args []string) error {
	cfgPath := "graphx.yaml"
	if loadedCfg, err := config.Load(cfgPath); err == nil {
		c.cfg = loadedCfg
	}

	eng, err := engine.NewEngine(c.cfg)
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}
	defer eng.Close()

	store := eng.GetStore()
	blastEngine := blast.NewBlastRadiusEngine(store)

	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}

	nodes, _, err := store.Query(&storage.NodeQuery{
		RepoID:     eng.GetRepoID(),
		NamePattern: "",
		Limit:      100,
	})
	if err != nil {
		return err
	}

	var targetNodeID string
	for _, n := range nodes {
		if n.FilePath == targetPath {
			targetNodeID = n.ID
			break
		}
	}

	if targetNodeID == "" && len(nodes) > 0 {
		targetNodeID = nodes[0].ID
	}

	if targetNodeID == "" {
		return fmt.Errorf("no target node found")
	}

	result, err := blastEngine.Analyze(targetNodeID)
	if err != nil {
		return err
	}

	fmt.Printf("Blast Radius Analysis for %s\n", targetPath)
	fmt.Printf("================================\n\n")

	fmt.Printf("Affected Nodes: %d\n", len(result.AffectedNodes))
	fmt.Printf("Affected Edges: %d\n", len(result.AffectedEdges))
	fmt.Printf("Risk Score: %.2f\n\n", result.RiskScore)

	if len(result.AuthFlows) > 0 {
		fmt.Printf("Authentication Flows Affected: %d\n", len(result.AuthFlows))
	}

	if len(result.ImpactedAPIs) > 0 {
		fmt.Printf("APIs Impacted: %d\n", len(result.ImpactedAPIs))
	}

	if len(result.CriticalPath) > 0 {
		fmt.Printf("\nCritical Path (Top 5):\n")
		for i, id := range result.CriticalPath[:min(5, len(result.CriticalPath))] {
			fmt.Printf("  %d. %s\n", i+1, id)
		}
	}

	return nil
}

func (c *Commands) listRisks() error {
	eng, err := engine.NewEngine(c.cfg)
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}
	defer eng.Close()

	store := eng.GetStore()
	risks, err := store.GetHighRiskNodes(eng.GetRepoID(), 0.7)
	if err != nil {
		return err
	}

	fmt.Printf("High-Risk Components (score >= 0.7)\n")
	fmt.Printf("=====================================\n")
	for _, n := range risks {
		fmt.Printf("%s [%s] - Risk: %.2f\n", n.Name, n.Type, n.RiskScore)
	}

	return nil
}

func (c *Commands) analyzeAuthFlows() error {
	eng, err := engine.NewEngine(c.cfg)
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}
	defer eng.Close()

	store := eng.GetStore()
	authNodes, err := store.GetAuthSensitiveNodes(eng.GetRepoID())
	if err != nil {
		return err
	}

	fmt.Printf("Authentication-Sensitive Components\n")
	fmt.Printf("=====================================\n")
	for _, n := range authNodes {
		fmt.Printf("%s [%s] - %s\n", n.Name, n.Type, n.FilePath)
	}

	return nil
}

func (c *Commands) generateThreatModel() error {
	cfgPath := "graphx.yaml"
	if loadedCfg, err := config.Load(cfgPath); err == nil {
		c.cfg = loadedCfg
	}

	eng, err := engine.NewEngine(c.cfg)
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}
	defer eng.Close()

	store := eng.GetStore()
	threatEngine := threat.NewThreatModelEngine(store)

	model, err := threatEngine.GenerateModel(eng.GetRepoID())
	if err != nil {
		return err
	}

	fmt.Printf("Threat Model for Repository\n")
	fmt.Printf("============================\n\n")

	fmt.Printf("Entry Points: %d\n", len(model.EntryPoints))
	fmt.Printf("Trust Boundaries: %d\n", len(model.TrustBoundaries))
	fmt.Printf("Total Threats: %d\n\n", len(model.Threats))

	if model.Summary != nil {
		fmt.Printf("STRIDE Summary:\n")
		fmt.Printf("  Spoofing: %d\n", model.Summary.Spoofing)
		fmt.Printf("  Tampering: %d\n", model.Summary.Tampering)
		fmt.Printf("  Repudiation: %d\n", model.Summary.Repudiation)
		fmt.Printf("  Info Disclosure: %d\n", model.Summary.InformationDisclosure)
		fmt.Printf("  DoS: %d\n", model.Summary.DenialOfService)
		fmt.Printf("  Elevation: %d\n\n", model.Summary.ElevationOfPrivilege)
	}

	if len(model.Threats) > 0 {
		fmt.Printf("Top Threats:\n")
		for i, t := range model.Threats[:min(10, len(model.Threats))] {
			fmt.Printf("  %d. [%s] %s - %s\n", i+1, t.Severity, t.Title, t.Category)
		}
	}

	fmt.Printf("\nAttack Paths Found: %d\n", len(model.AttackPaths))

	return nil
}

func (c *Commands) findAttackPaths() error {
	fmt.Println("Attack path analysis not yet implemented")
	return nil
}

func (c *Commands) listTrustBoundaries() error {
	fmt.Println("Trust boundary analysis not yet implemented")
	return nil
}

func (c *Commands) exportMermaid() error {
	eng, err := engine.NewEngine(c.cfg)
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}
	defer eng.Close()

	store := eng.GetStore()
	nodes, _, err := store.Query(&storage.NodeQuery{
		RepoID: eng.GetRepoID(),
		Limit:  100,
	})
	if err != nil {
		return err
	}

	fmt.Println("```mermaid")
	fmt.Println("graph TD")
	for _, n := range nodes {
		fmt.Printf("  %s[%s]\n", n.ID, n.Name)
	}
	fmt.Println("```")

	return nil
}

func (c *Commands) exportGraphML() error {
	fmt.Println("GraphML export not yet implemented")
	return nil
}

func (c *Commands) exportJSON() error {
	eng, err := engine.NewEngine(c.cfg)
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}
	defer eng.Close()

	store := eng.GetStore()
	nodes, edges, err := store.Query(&storage.NodeQuery{
		RepoID:    eng.GetRepoID(),
		WithEdges: true,
		Limit:     100,
	})
	if err != nil {
		return err
	}

	fmt.Printf("{\n")
	fmt.Printf("  \"nodes\": %d,\n", len(nodes))
	fmt.Printf("  \"edges\": %d\n", len(edges))
	fmt.Printf("}\n")

	return nil
}

func (c *Commands) exportCypher() error {
	fmt.Println("Cypher export not yet implemented")
	return nil
}

func (c *Commands) search(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("search query required")
	}

	eng, err := engine.NewEngine(c.cfg)
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}
	defer eng.Close()

	store := eng.GetStore()
	nodes, _, err := store.Query(&storage.NodeQuery{
		RepoID:      eng.GetRepoID(),
		NamePattern: args[0],
		Limit:       10,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Search results for: %s\n", args[0])
	fmt.Printf("====================\n")
	for _, n := range nodes {
		fmt.Printf("%s [%s] - %s\n", n.Name, n.Type, n.FilePath)
	}

	return nil
}