package pipeline

import (
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkAnalyzeRepository(b *testing.B) {
	dir := b.TempDir()
	code := `package main
func get() {}
`
	for i := 0; i < 50; i++ {
		_ = os.WriteFile(filepath.Join(dir, "svc"+string(rune('a'+(i%26)))+".go"), []byte(code), 0o644)
	}
	out := filepath.Join(dir, "out")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := AnalyzeRepository(dir, out); err != nil {
			b.Fatal(err)
		}
	}
}
