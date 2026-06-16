package config

import (
	"log/slog"
	"os"
	"path/filepath"
)

type Config struct {
	ServerURL string `json:"server_url"`
	DataDir   string `json:"data_dir"`

	HeartbeatIntervalSec int `json:"heartbeat_interval_sec"`
	RequestTimeoutSec    int `json:"request_timeout_sec"`
	MaxRetries           int `json:"max_retries"`

	LogLevel string `json:"log_level"`
	LogFile  string `json:"log_file"`
}

func Defaults() *Config {
	return &Config{
		ServerURL:           "http://192.168.1.14:8000/api/v1",
		DataDir:             defaultDataDir(),
		HeartbeatIntervalSec: 30,
		RequestTimeoutSec:    30,
		MaxRetries:           5,
		LogLevel:             "info",
	}
}

func defaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "duckops-agent")
	}
	return filepath.Join(home, ".duckops", "agent")
}

func (c *Config) EnsureDataDir() error {
	if err := os.MkdirAll(c.DataDir, 0o700); err != nil {
		return err
	}
	return nil
}

type Manager struct {
	cfg  *Config
	path string
}

func NewManager(path string) *Manager {
	return &Manager{path: path}
}

func (m *Manager) Load() (*Config, error) {
	cfg := Defaults()

	data, err := os.ReadFile(m.path)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Warn("config file not found, using defaults", "path", m.path)
			return cfg, nil
		}
		return nil, err
	}

	if err := cfg.decode(data); err != nil {
		return nil, err
	}

	cfg.applyEnvOverrides()

	return cfg, nil
}

func (m *Manager) Path() string { return m.path }
