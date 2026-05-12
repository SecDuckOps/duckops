package parser

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/SecDuckOps/duckops/internal/security/architecture"
)

var routeExpr = regexp.MustCompile("(?i)(get|post|put|patch|delete)\\s*\\(\\s*[\"'`](.+?)[\"'`]")

func ParseRepository(root string) (architecture.IR, error) {
	ir := architecture.NewIR(root)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() {
			return nil
		}
		if strings.Contains(path, "/.git/") || strings.Contains(path, "/node_modules/") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		text := string(content)

		if strings.HasSuffix(path, "Dockerfile") || strings.Contains(strings.ToLower(filepath.Base(path)), "dockerfile") {
			ir.Evidence = append(ir.Evidence, architecture.Evidence{
				Path:      path,
				Extractor: "regex",
				Confidence: architecture.ConfidenceMedium,
				Detail:    "dockerfile detected",
			})
		}

		if strings.HasSuffix(path, ".go") || strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".py") {
			matches := routeExpr.FindAllStringSubmatch(text, -1)
			for _, m := range matches {
				route := m[2]
				method := strings.ToUpper(m[1])
				serviceID := "service:monolith"
				if len(ir.Services) == 0 {
					ir.Services = append(ir.Services, architecture.Service{
						ID:       serviceID,
						Name:     "monolith",
						Language: inferLang(path),
						Exposure: "internal",
					})
				}
				id := "api:" + serviceID + ":" + method + ":" + route
				ir.APIs = append(ir.APIs, architecture.API{
					ID:        id,
					ServiceID: serviceID,
					Method:    method,
					Route:     route,
					Public:    strings.HasPrefix(route, "/api"),
				})
				ir.Evidence = append(ir.Evidence, architecture.Evidence{
					Path:      path,
					Extractor: "regex",
					Confidence: architecture.ConfidenceLow,
					Detail:    "http route pattern",
				})
			}
		}

		if strings.Contains(path, ".github/workflows") {
			ir.ExternalIntegrations = append(ir.ExternalIntegrations, architecture.ExternalIntegration{
				ID:       "ci:github-actions",
				Name:     "GitHub Actions",
				Category: "ci_cd",
			})
		}
		return nil
	})
	return ir, err
}

func inferLang(path string) string {
	switch filepath.Ext(path) {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".java":
		return "java"
	case ".js", ".ts":
		return "nodejs"
	case ".php":
		return "php"
	default:
		return "unknown"
	}
}
