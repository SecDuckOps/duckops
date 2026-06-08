package compression

import (
	"strings"
	"sync"
)

type CodeCompressor struct {
	mu sync.Mutex
}

func NewCodeCompressor() *CodeCompressor {
	return &CodeCompressor{}
}

func (c *CodeCompressor) Compress(snippets []CodeSnippet, mode Mode) []CodeSnippet {
	if len(snippets) == 0 || mode == ModeNone {
		return snippets
	}

	result := make([]CodeSnippet, 0, len(snippets))
	for _, snippet := range snippets {
		compressed := c.compressSnippet(snippet, mode)
		result = append(result, compressed)
	}
	return result
}

func (c *CodeCompressor) compressSnippet(snippet CodeSnippet, mode Mode) CodeSnippet {
	lines := strings.Split(snippet.Content, "\n")
	if len(lines) == 0 {
		return snippet
	}

	threshold := c.lineThreshold(mode)
	if len(lines) <= threshold {
		return snippet
	}

	var kept []string
	kept = append(kept, c.extractImports(lines)...)
	kept = append(kept, c.extractSignatures(lines)...)
	kept = append(kept, c.extractImportantComments(lines)...)

	if mode == ModeAggressive {
		kept = append(kept, c.summarizeBody(lines, threshold)...)
	} else if mode == ModeBalanced {
		kept = append(kept, c.keepFirstLast(lines, threshold)...)
	} else if mode == ModeLight {
		kept = append(kept, c.keepFirstLast(lines, threshold*2)...)
	}

	snippet.Content = strings.Join(kept, "\n")
	return snippet
}

func (c *CodeCompressor) lineThreshold(mode Mode) int {
	switch mode {
	case ModeLight:
		return 100
	case ModeBalanced:
		return 50
	case ModeAggressive:
		return 20
	default:
		return 200
	}
}

func (c *CodeCompressor) extractImports(lines []string) []string {
	var imports []string
	inImport := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "import") || strings.HasPrefix(trimmed, "#include") ||
			strings.HasPrefix(trimmed, "use ") || strings.HasPrefix(trimmed, "require(") ||
			strings.HasPrefix(trimmed, "from ") {
			inImport = true
		}
		if inImport {
			imports = append(imports, line)
			if trimmed == "" || strings.HasSuffix(trimmed, ";") || trimmed == ")" || trimmed == "}" {
				inImport = false
			}
			continue
		}
		if strings.HasPrefix(trimmed, "package ") || strings.HasPrefix(trimmed, "// ") ||
			strings.HasPrefix(trimmed, "# ") || strings.HasPrefix(trimmed, "/// ") {
			imports = append(imports, line)
		}
		if !inImport && len(imports) > 0 && trimmed != "" {
			break
		}
	}
	return imports
}

func (c *CodeCompressor) extractSignatures(lines []string) []string {
	var sigs []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if c.isSignature(trimmed) {
			sigs = append(sigs, line)
		}
	}
	return sigs
}

func (c *CodeCompressor) isSignature(line string) bool {
	keywords := []string{
		"func ", "def ", "fn ", "function ",
		"class ", "struct ", "interface ", "trait ",
		"type ", "enum ", "impl ",
		"public ", "private ", "protected ",
		"async ", "const ", "let ", "var ",
		"void ", "int ", "string ", "bool ",
		"export ", "default ",
	}
	for _, kw := range keywords {
		if strings.HasPrefix(line, kw) {
			return true
		}
	}
	return false
}

func (c *CodeCompressor) extractImportantComments(lines []string) []string {
	var comments []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "// TODO") ||
			strings.HasPrefix(trimmed, "// FIXME") ||
			strings.HasPrefix(trimmed, "// HACK") ||
			strings.HasPrefix(trimmed, "// BUG") ||
			strings.HasPrefix(trimmed, "// OPTIMIZE") ||
			strings.HasPrefix(trimmed, "# TODO") ||
			strings.HasPrefix(trimmed, "# FIXME") ||
			strings.HasPrefix(trimmed, "/* TODO") ||
			strings.HasPrefix(trimmed, "/* FIXME") {
			comments = append(comments, line)
		}
	}
	return comments
}

func (c *CodeCompressor) summarizeBody(lines []string, threshold int) []string {
	if len(lines) <= threshold {
		return lines
	}

	bodyLines := lines[threshold:]
	nonEmpty := 0
	for _, line := range bodyLines {
		if strings.TrimSpace(line) != "" {
			nonEmpty++
		}
	}

	return []string{
		"// ... [" + itoa(len(bodyLines)) + " lines, " + itoa(nonEmpty) + " non-empty lines compressed] ...",
	}
}

func (c *CodeCompressor) keepFirstLast(lines []string, keepCount int) []string {
	if len(lines) <= keepCount {
		return lines
	}

	first := keepCount / 2
	last := keepCount / 2
	if first+last > len(lines) {
		first = len(lines) / 2
		last = len(lines) - first
	}

	result := make([]string, 0, first+last+1)
	result = append(result, lines[:first]...)
	result = append(result, "// ... ["+itoa(len(lines)-first-last)+" lines compressed] ...")
	result = append(result, lines[len(lines)-last:]...)
	return result
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
