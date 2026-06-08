package compression

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Enabled      bool   `yaml:"enabled"`
	MaxTokens    int    `yaml:"max_tokens"`
	TargetTokens int    `yaml:"target_tokens"`
	Mode         Mode   `yaml:"mode"`
	CacheSize    int    `yaml:"cache_size"`
	LogEnabled   bool   `yaml:"log_enabled"`
}

func DefaultConfig() Config {
	return Config{
		Enabled:      true,
		MaxTokens:    DefaultMaxTokens,
		TargetTokens: DefaultTargetTokens,
		Mode:         ModeBalanced,
		CacheSize:    100,
		LogEnabled:   true,
	}
}

func LoadConfig(path string) (Config, error) {
	cfg := DefaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}
	if !cfg.Mode.Valid() {
		cfg.Mode = ModeBalanced
	}
	if cfg.MaxTokens <= 0 {
		cfg.MaxTokens = DefaultMaxTokens
	}
	if cfg.TargetTokens <= 0 {
		cfg.TargetTokens = DefaultTargetTokens
	}
	if cfg.TargetTokens > cfg.MaxTokens {
		cfg.TargetTokens = cfg.MaxTokens
	}
	return cfg, nil
}

func (c Config) ShouldCompress(estimatedTokens int) bool {
	if !c.Enabled {
		return false
	}
	return estimatedTokens > c.TargetTokens
}
