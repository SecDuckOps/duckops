package validation

import (
	"strings"

	"github.com/SecDuckOps/duckops/internal/security/architecture"
)

func Validate(ir architecture.IR) []architecture.Finding {
	var out []architecture.Finding
	hasRateLimit := false
	for _, a := range ir.Authentication {
		if a.HasRateLimit {
			hasRateLimit = true
			break
		}
	}
	if !hasRateLimit {
		out = append(out, architecture.Finding{
			ID:         "validation:missing_rate_limit",
			Title:      "Rate limiting evidence missing",
			Severity:   "medium",
			Component:  "global",
			Category:   "validation",
			Confidence: architecture.ConfidenceMedium,
			Hypothesis: true,
			Mitigation: "Add deterministic checks for gateway/app-level throttling and emit control evidence.",
		})
	}
	for _, dep := range ir.Dependencies {
		if strings.Contains(strings.ToLower(dep.Name), "jwt") && dep.Version == "" {
			out = append(out, architecture.Finding{
				ID:         "validation:dep_version_missing:" + dep.Name,
				Title:      "Dependency version missing for security-sensitive package",
				Severity:   "low",
				Component:  dep.Name,
				Category:   "validation",
				Confidence: architecture.ConfidenceHigh,
				Hypothesis: true,
				Mitigation: "Pin version and validate against vulnerability feeds in CI.",
			})
		}
	}
	return out
}
