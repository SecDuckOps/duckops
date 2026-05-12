package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Repository RepositoryConfig `yaml:"repository"`
	Graph      GraphConfig      `yaml:"graph"`
	Parser     ParserConfig     `yaml:"parser"`
	Analysis   AnalysisConfig  `yaml:"analysis"`
	Watch      WatchConfig      `yaml:"watch"`
	Export     ExportConfig    `yaml:"export"`
}

type RepositoryConfig struct {
	Path      string   `yaml:"path"`
	Ignore    []string `yaml:"ignore"`
	RootPath  string   `yaml:"root_path"`
}

type GraphConfig struct {
	Storage  string `yaml:"storage"`
	Path     string `yaml:"path"`
	WALMode  bool   `yaml:"wal_mode"`
}

type ParserConfig struct {
	Languages []string `yaml:"languages"`
	Parallel  int      `yaml:"parallel"`
}

type AnalysisConfig struct {
	Security    SecurityAnalysisConfig    `yaml:"security"`
	BlastRadius BlastRadiusConfig          `yaml:"blast_radius"`
}

type SecurityAnalysisConfig struct {
	Enabled             bool    `yaml:"enabled"`
	SensitivityThreshold float64 `yaml:"sensitivity_threshold"`
}

type BlastRadiusConfig struct {
	MaxDepth   int     `yaml:"max_depth"`
	WeightAuth float64 `yaml:"weight_auth"`
	WeightAPI  float64 `yaml:"weight_api"`
}

type WatchConfig struct {
	Enabled    bool `yaml:"enabled"`
	DebounceMs int  `yaml:"debounce_ms"`
}

type ExportConfig struct {
	Formats []string `yaml:"formats"`
}

func Default() *Config {
	return &Config{
		Repository: RepositoryConfig{
			Path:   ".",
			Ignore: []string{"**/node_modules/**", "**/vendor/**", "**/.git/**"},
		},
		Graph: GraphConfig{
			Storage: "sqlite",
			Path:    ".graphx/graph.json",
			WALMode: true,
		},
		Parser: ParserConfig{
			Languages: []string{"python", "typescript", "go"},
			Parallel:  4,
		},
		Analysis: AnalysisConfig{
			Security: SecurityAnalysisConfig{
				Enabled:             true,
				SensitivityThreshold: 0.5,
			},
			BlastRadius: BlastRadiusConfig{
				MaxDepth:   10,
				WeightAuth: 2.0,
				WeightAPI:  1.5,
			},
		},
		Watch: WatchConfig{
			Enabled:    true,
			DebounceMs: 500,
		},
		Export: ExportConfig{
			Formats: []string{"mermaid", "graphml", "json"},
		},
	}
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Default(), nil
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func Save(cfg *Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}