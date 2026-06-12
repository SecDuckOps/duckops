//go:build ignore
// +build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/SecDuckOps/duckops/internal/agent/prompt"
	"github.com/SecDuckOps/duckops/internal/config"
)

func fixedTime() time.Time {
	t, _ := time.Parse("1/2/2006", "1/1/2025")
	return t
}

func main() {
	tests := []string{
		"simple_test",
		"read_a_file",
		"update_a_file",
		"bash_tool",
		"download_tool",
		"fetch_tool",
		"glob_tool",
		"grep_tool",
		"ls_tool",
		"multiedit_tool",
		"parallel_tool_calls",
		"sourcegraph_tool",
		"write_tool",
	}
	out := map[string]string{}
	tmplPath := filepath.Join("internal", "agent", "templates", "coder.md.tpl")
	tplData, err := os.ReadFile(tmplPath)
	if err != nil {
		panic(err)
	}
	for _, name := range tests {
		wd := filepath.Join(string(os.PathSeparator), "tmp", "duckops-test", "TestCoderAgent", "glm-5.1", name)
		if err := os.MkdirAll(wd, 0o755); err != nil {
			panic(err)
		}
		cfg, err := config.Init(wd, "", false)
		if err != nil {
			panic(err)
		}
		cfg.Config().Options.Attribution = &config.Attribution{
			TrailerStyle:  "co-authored-by",
			GeneratedWith: true,
		}
		cfg.Config().Options.SkillsPaths = nil
		cfg.Config().Options.DisabledSkills = []string{"duckops-config"}
		cfg.Config().Options.ContextPaths = nil
		cfg.Config().LSP = nil
		promptObj, err := prompt.NewPrompt("coder", string(tplData), prompt.WithTimeFunc(fixedTime), prompt.WithPlatform("linux"), prompt.WithWorkingDir(wd))
		if err != nil {
			panic(err)
		}
		systemPrompt, err := promptObj.Build(context.Background(), "hyper", "glm-5.1", cfg)
		if err != nil {
			panic(err)
		}
		out[name] = systemPrompt
	}
	enc, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(enc))
}
