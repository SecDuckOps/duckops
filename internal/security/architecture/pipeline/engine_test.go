package pipeline

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeRepository(t *testing.T) {
	dir := t.TempDir()
	src := `package main
import "net/http"
func main() {
  _ = http.MethodGet
}
`
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "out")
	res, err := AnalyzeRepository(dir, out)
	if err != nil {
		t.Fatalf("analyze failed: %v", err)
	}
	if res.IR.IRVersion == "" {
		t.Fatalf("expected IR version")
	}
	if _, err := os.Stat(filepath.Join(out, "report.md")); err != nil {
		t.Fatalf("expected markdown report: %v", err)
	}
}
