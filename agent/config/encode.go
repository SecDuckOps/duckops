package config

import (
	"encoding/json"
	"os"
	"strconv"
)

func (c *Config) decode(data []byte) error {
	return json.Unmarshal(data, c)
}

func (c *Config) applyEnvOverrides() {
	if v := os.Getenv("DUCKOPS_SERVER_URL"); v != "" {
		c.ServerURL = v
	}
	if v := os.Getenv("DUCKOPS_DATA_DIR"); v != "" {
		c.DataDir = v
	}
	if v := os.Getenv("DUCKOPS_HEARTBEAT_INTERVAL"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			c.HeartbeatIntervalSec = i
		}
	}
	if v := os.Getenv("DUCKOPS_REQUEST_TIMEOUT"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			c.RequestTimeoutSec = i
		}
	}
	if v := os.Getenv("DUCKOPS_MAX_RETRIES"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			c.MaxRetries = i
		}
	}
	if v := os.Getenv("DUCKOPS_LOG_LEVEL"); v != "" {
		c.LogLevel = v
	}
}

func Save(path string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
