package repoanalyzer

import (
	"os"
	"path/filepath"
	"strings"
)

type Inventory struct {
	Languages []string
	Frameworks []string
	HasDocker bool
	HasKubernetes bool
	HasTerraform bool
	HasHelm bool
	HasGitHubActions bool
	HasGitLabCI bool
	HasJenkins bool
}

func Analyze(root string) (Inventory, error) {
	inv := Inventory{}
	langs := map[string]bool{}
	frameworks := map[string]bool{}

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			name := strings.ToLower(d.Name())
			if name == ".git" || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		base := strings.ToLower(filepath.Base(path))
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".go":
			langs["go"] = true
		case ".py":
			langs["python"] = true
		case ".java":
			langs["java"] = true
		case ".js", ".ts":
			langs["nodejs"] = true
		case ".php":
			langs["php"] = true
		case ".tf":
			langs["terraform"] = true
			inv.HasTerraform = true
		case ".yaml", ".yml":
			if strings.Contains(path, ".github/workflows") {
				inv.HasGitHubActions = true
			}
			if strings.Contains(path, "k8s") || strings.Contains(path, "kubernetes") {
				inv.HasKubernetes = true
			}
			if strings.Contains(path, "helm") {
				inv.HasHelm = true
			}
		}

		if base == "dockerfile" || strings.HasPrefix(base, "dockerfile.") {
			inv.HasDocker = true
		}
		if base == ".gitlab-ci.yml" {
			inv.HasGitLabCI = true
		}
		if base == "jenkinsfile" {
			inv.HasJenkins = true
		}
		if base == "package.json" {
			frameworks["node"] = true
		}
		if base == "go.mod" {
			frameworks["go-modules"] = true
		}
		if base == "requirements.txt" || base == "pyproject.toml" {
			frameworks["python"] = true
		}
		return nil
	})
	if err != nil {
		return inv, err
	}
	for l := range langs {
		inv.Languages = append(inv.Languages, l)
	}
	for f := range frameworks {
		inv.Frameworks = append(inv.Frameworks, f)
	}
	return inv, nil
}
