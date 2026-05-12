package diagram_renderer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// RenderWithMermaidCLI renders Mermaid source to output format using mmdc when available.
// outputPath extension controls format: .svg, .png, .pdf.
func RenderWithMermaidCLI(mermaidSource, outputPath string) error {
	if _, err := exec.LookPath("mmdc"); err != nil {
		return fmt.Errorf("mermaid CLI (mmdc) not found in PATH")
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp("", "duckops-diagram-*.mmd")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(mermaidSource); err != nil {
		_ = tmp.Close()
		return err
	}
	_ = tmp.Close()

	cmd := exec.Command("mmdc", "-i", tmp.Name(), "-o", outputPath, "-b", "transparent")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mmdc failed: %w (%s)", err, string(out))
	}
	return nil
}
