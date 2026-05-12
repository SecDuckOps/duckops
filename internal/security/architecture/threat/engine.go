package threat

import (
	"fmt"
	"strings"

	"github.com/SecDuckOps/duckops/internal/security/architecture"
)

var stride = []string{"Spoofing", "Tampering", "Repudiation", "Information Disclosure", "Denial of Service", "Elevation of Privilege"}

func Generate(ir architecture.IR) []architecture.Finding {
	var out []architecture.Finding
	for _, api := range ir.APIs {
		for _, category := range stride {
			severity := "medium"
			cwe := ""
			if category == "Information Disclosure" && api.Public {
				severity = "high"
				cwe = "CWE-200"
			}
			if strings.Contains(strings.ToLower(api.Route), "upload") {
				severity = "high"
				cwe = "CWE-434"
			}
			out = append(out, architecture.Finding{
				ID:         fmt.Sprintf("threat:%s:%s", api.ID, strings.ReplaceAll(strings.ToLower(category), " ", "_")),
				Title:      category + " risk on " + api.Route,
				Severity:   severity,
				Component:  api.ID,
				CWE:        cwe,
				Category:   "STRIDE",
				Confidence: architecture.ConfidenceMedium,
				Mitigation: "Add explicit control mapping and validate enforcement in middleware/policies.",
			})
		}
	}
	return out
}

func AbuseCases(ir architecture.IR) []architecture.Finding {
	var out []architecture.Finding
	for _, api := range ir.APIs {
		route := strings.ToLower(api.Route)
		if strings.Contains(route, "webhook") {
			out = append(out, architecture.Finding{
				ID:         "abuse:webhook:" + api.ID,
				Title:      "Poisoned webhook abuse case",
				Severity:   "high",
				Component:  api.ID,
				CWE:        "CWE-345",
				Category:   "abuse_case",
				Confidence: architecture.ConfidenceHigh,
				AttackScenario: "Attacker sends forged webhook payload to trigger privileged workflow.",
				Mitigation: "Validate signatures, enforce replay protection, and constrain webhook actions.",
			})
		}
		if strings.Contains(route, "upload") {
			out = append(out, architecture.Finding{
				ID:         "abuse:upload:" + api.ID,
				Title:      "Malicious file upload abuse case",
				Severity:   "high",
				Component:  api.ID,
				CWE:        "CWE-434",
				Category:   "abuse_case",
				Confidence: architecture.ConfidenceHigh,
				AttackScenario: "Malicious SVG/script payload upload leading to stored XSS or parser exploit.",
				Mitigation: "Enforce MIME + magic checks, content sanitization, and execution isolation.",
			})
		}
	}
	return out
}
