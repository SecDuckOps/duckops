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

func main() {
  wd := filepath.Join(string(os.PathSeparator), "tmp", "duckops-test", "TestCoderAgent", "glm-5.1", "simple_test")
  os.MkdirAll(wd, 0755)
  cfg, err := config.Init(wd, "", false)
  if err != nil { panic(err) }
  cfg.Config().Options.Attribution = &config.Attribution{TrailerStyle: "co-authored-by", GeneratedWith: true}
  cfg.Config().Options.SkillsPaths = nil
  cfg.Config().Options.DisabledSkills = []string{"duckops-config"}
  cfg.Config().Options.ContextPaths = nil
  cfg.Config().LSP = nil
  tplData, err := os.ReadFile(filepath.Join("internal","agent","templates","coder.md.tpl"))
  if err != nil { panic(err) }
  promptObj, err := prompt.NewPrompt("coder", string(tplData), prompt.WithTimeFunc(func() time.Time { return time.Time{} }), prompt.WithPlatform("linux"), prompt.WithWorkingDir(filepath.ToSlash(wd)))
  if err != nil { panic(err) }
  systemPrompt, err := promptObj.Build(context.Background(), "hyper", "glm-5.1", cfg)
  if err != nil { panic(err) }
  b, _ := json.Marshal(systemPrompt)
  fmt.Println(string(b))
}
