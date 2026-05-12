package parser

import (
	"regexp"
	"strings"

	"github.com/SecDuckOps/duckops/internal/graphx/ir"
)

type Parser interface {
	Name() string
	Language() string
	Parse(content []byte) (*ir.Program, error)
	LanguageSupported(filename string) bool
}

type Registry struct {
	parsers map[string]Parser
}

func NewRegistry() *Registry {
	return &Registry{
		parsers: make(map[string]Parser),
	}
}

func (r *Registry) Register(p Parser) {
	r.parsers[p.Language()] = p
}

func (r *Registry) Get(language string) Parser {
	return r.parsers[language]
}

func (r *Registry) ForFile(filename string) Parser {
	ext := strings.TrimPrefix(strings.ToLower(filename), ".")
	langMap := map[string]string{
		"py":     "python",
		"ts":     "typescript",
		"tsx":    "typescript",
		"js":     "javascript",
		"jsx":    "javascript",
		"go":     "go",
		"rs":     "rust",
		"java":   "java",
		"cs":     "csharp",
		"rb":     "ruby",
		"php":    "php",
		"kt":     "kotlin",
		"swift":  "swift",
		"tf":     "terraform",
		"yaml":   "yaml",
		"yml":    "yaml",
		"Dockerfile": "dockerfile",
	}

	if lang, ok := langMap[ext]; ok {
		if p, ok := r.parsers[lang]; ok {
			return p
		}
	}

	for _, p := range r.parsers {
		if p.LanguageSupported(filename) {
			return p
		}
	}

	return nil
}

func (r *Registry) Languages() []string {
	var langs []string
	for lang := range r.parsers {
		langs = append(langs, lang)
	}
	return langs
}

type BaseParser struct {
	language string
}

func (p *BaseParser) Language() string {
	return p.language
}

func (p *BaseParser) LanguageSupported(filename string) bool {
	return strings.HasSuffix(filename, "."+p.language)
}

type RegexParser struct {
	BaseParser
	funcPattern    *regexp.Regexp
	classPattern   *regexp.Regexp
	importPattern  *regexp.Regexp
	callPattern    *regexp.Regexp
	apiPattern     *regexp.Regexp
}

func NewRegexParser(language string, patterns map[string]string) *RegexParser {
	return &RegexParser{
		BaseParser:    BaseParser{language: language},
		funcPattern:   compilePattern(patterns["func"]),
		classPattern:  compilePattern(patterns["class"]),
		importPattern: compilePattern(patterns["import"]),
		callPattern:   compilePattern(patterns["call"]),
		apiPattern:    compilePattern(patterns["api"]),
	}
}

func compilePattern(s string) *regexp.Regexp {
	if s == "" {
		return nil
	}
	return regexp.MustCompile(s)
}

func (p *RegexParser) Parse(content []byte) (*ir.Program, error) {
	lines := strings.Split(string(content), "\n")
	program := &ir.Program{
		Elements: make([]ir.Element, 0),
		Errors:   make([]ir.ParseError, 0),
	}

	for i, line := range lines {
		lineNum := i + 1

		if p.classPattern != nil {
			if match := p.classPattern.FindStringSubmatch(line); len(match) > 1 {
				class := &ir.Class{
					ID:                irNodeID("class", match[1], lineNum),
					Name:              match[1],
					FullyQualifiedName: match[1],
					StartLine:         lineNum,
				}
				program.Elements = append(program.Elements, class)
			}
		}

		if p.funcPattern != nil {
			if match := p.funcPattern.FindStringSubmatch(line); len(match) > 1 {
				isExported := len(match) > 2 && match[2] != ""
				fn := &ir.Function{
					ID:                irNodeID("func", match[1], lineNum),
					Name:              match[1],
					FullyQualifiedName: match[1],
					StartLine:         lineNum,
					IsExported:        isExported,
				}
				fn.SecurityTags = ir.DetectSecurityTags(fn)
				program.Elements = append(program.Elements, fn)
			}
		}

		if p.importPattern != nil {
			if match := p.importPattern.FindStringSubmatch(line); len(match) > 1 {
				imp := &ir.Import{
					ID:     irNodeID("imp", match[1], lineNum),
					Source: match[1],
				}
				program.Elements = append(program.Elements, imp)
			}
		}
	}

	return program, nil
}

func irNodeID(prefix, name string, line int) string {
	safeName := strings.ReplaceAll(name, "/", "_")
	safeName = strings.ReplaceAll(safeName, ".", "_")
	return prefix + "_" + safeName + "_" + string(rune('a'+line%26))
}