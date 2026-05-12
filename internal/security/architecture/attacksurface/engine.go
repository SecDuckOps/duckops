package attacksurface

import (
	"strings"

	"github.com/SecDuckOps/duckops/internal/security/architecture"
)

func Analyze(ir architecture.IR) []architecture.Finding {
	var out []architecture.Finding
	for _, api := range ir.APIs {
		route := strings.ToLower(api.Route)
		if api.Public {
			out = append(out, architecture.Finding{
				ID:         "surface:public:" + api.ID,
				Title:      "Public API exposed",
				Severity:   "medium",
				Component:  api.ID,
				Category:   "attack_surface",
				Confidence: architecture.ConfidenceHigh,
				Mitigation: "Apply authz, rate limits, and strict input validation for internet-facing endpoints.",
			})
		}
		if strings.Contains(route, "fetch") || strings.Contains(route, "proxy") {
			out = append(out, architecture.Finding{
				ID:             "surface:ssrf:" + api.ID,
				Title:          "Potential SSRF vector",
				Severity:       "high",
				Component:      api.ID,
				CWE:            "CWE-918",
				Category:       "attack_surface",
				Confidence:     architecture.ConfidenceMedium,
				AttackScenario: "User-supplied URL reaches internal metadata or private network targets.",
				Mitigation:     "Block link-local/private CIDRs and enforce hostname allowlist.",
			})
		}
	}
	return out
}
