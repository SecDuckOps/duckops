package ir

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/SecDuckOps/duckops/internal/security/architecture"
)

func SaveSnapshot(cacheDir string, model architecture.IR) (string, error) {
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", err
	}
	raw, err := json.MarshalIndent(model, "", "  ")
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	hash := hex.EncodeToString(sum[:])
	target := filepath.Join(cacheDir, "ir_"+hash[:12]+".json")
	if err := os.WriteFile(target, raw, 0o644); err != nil {
		return "", err
	}
	return target, nil
}

func Merge(base architecture.IR, additional architecture.IR) architecture.IR {
	base.Services = append(base.Services, additional.Services...)
	base.APIs = append(base.APIs, additional.APIs...)
	base.Databases = append(base.Databases, additional.Databases...)
	base.Queues = append(base.Queues, additional.Queues...)
	base.ExternalIntegrations = append(base.ExternalIntegrations, additional.ExternalIntegrations...)
	base.TrustBoundaries = append(base.TrustBoundaries, additional.TrustBoundaries...)
	base.Authentication = append(base.Authentication, additional.Authentication...)
	base.DataFlows = append(base.DataFlows, additional.DataFlows...)
	base.Dependencies = append(base.Dependencies, additional.Dependencies...)
	base.Findings = append(base.Findings, additional.Findings...)
	base.Evidence = append(base.Evidence, additional.Evidence...)
	return base
}
