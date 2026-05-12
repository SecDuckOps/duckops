package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseRepositoryRoutes(t *testing.T) {
	dir := t.TempDir()
	code := `app.get("/api/health", handler)`
	if err := os.WriteFile(filepath.Join(dir, "app.js"), []byte(code), 0o644); err != nil {
		t.Fatal(err)
	}
	ir, err := ParseRepository(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ir.APIs) == 0 {
		t.Fatalf("expected APIs to be detected")
	}
}
