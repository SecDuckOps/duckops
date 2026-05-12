package graphx

import (
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/SecDuckOps/duckops/internal/graphx/ir"
	"github.com/SecDuckOps/duckops/internal/graphx/parser"
)

type MultiLanguageParser struct {
	registry *parser.Registry
}

func NewMultiLanguageParser() *MultiLanguageParser {
	p := &MultiLanguageParser{
		registry: parser.NewRegistry(),
	}

	p.registry.Register(parser.NewPythonParser())
	p.registry.Register(parser.NewGoParser())
	p.registry.Register(parser.NewTypeScriptParser())
	p.registry.Register(parser.NewRustParser())

	return p
}

func (p *MultiLanguageParser) ParseFile(filePath string) (*ir.Program, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	langParser := p.registry.ForFile(filePath)
	if langParser == nil {
		return p.parseGeneric(content)
	}

	program, err := langParser.Parse(content)
	if err != nil {
		return nil, err
	}

	p.detectCallRelationships(program)
	p.detectSecurityPatterns(program)
	p.enrichWithAnnotations(program)

	return program, nil
}

func (p *MultiLanguageParser) parseGeneric(content []byte) (*ir.Program, error) {
	program := &ir.Program{
		Elements: make([]ir.Element, 0),
		Errors:   make([]ir.ParseError, 0),
	}

	lines := strings.Split(string(content), "\n")

	funcRe := regexp.MustCompile(`(?:func|function|def|fn)\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`)
	classRe := regexp.MustCompile(`(?:class|struct|type)\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*[:{\[]`)

	for i, line := range lines {
		lineNum := i + 1

		if matches := funcRe.FindStringSubmatch(line); len(matches) > 1 {
			fn := &ir.Function{
				ID:                generateNodeID("func", matches[1], lineNum),
				Name:              matches[1],
				FullyQualifiedName: matches[1],
				StartLine:         lineNum,
			}
			fn.SecurityTags = ir.DetectSecurityTags(fn)
			program.Elements = append(program.Elements, fn)
		}

		if matches := classRe.FindStringSubmatch(line); len(matches) > 1 {
			class := &ir.Class{
				ID:                generateNodeID("class", matches[1], lineNum),
				Name:              matches[1],
				FullyQualifiedName: matches[1],
				StartLine:         lineNum,
			}
			class.SecurityTags = ir.DetectSecurityTags(class)
			program.Elements = append(program.Elements, class)
		}
	}

	return program, nil
}

func (p *MultiLanguageParser) detectCallRelationships(program *ir.Program) {
	callRe := regexp.MustCompile(`\b([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`)

	for _, el := range program.Elements {
		if fn, ok := el.(*ir.Function); ok {
			calls := callRe.FindAllStringSubmatch(fn.FullyQualifiedName, -1)
			for _, call := range calls {
				if len(call) > 1 && call[1] != fn.Name {
					fn.Calls = append(fn.Calls, call[1])
				}
			}
		}
	}
}

func (p *MultiLanguageParser) detectSecurityPatterns(program *ir.Program) {
	securityPatterns := map[string][]string{
		"sql_query":     {`SELECT`, `INSERT`, `UPDATE`, `DELETE`, `EXECUTE`},
		"exec":           {`exec\(`, `system\(`, `popen`, `subprocess`},
		"crypto":         {`encrypt`, `decrypt`, `hash`, `sign`, `verify`},
		"auth":           {`authenticate`, `authorize`, `login`, `password`, `credential`},
		"file_access":    {`open\(`, `read`, `write`, `chmod`, `chown`},
		"network":       {`http\.`, `fetch\(`, `axios`, `requests\.`, `curl`},
		"env_var":        {`getenv`, `os\.environ`, `process\.env`, `System\.getProperty`},
		"serialize":      {`json\.marshal`, `json\.encode`, `pickle`, `yaml\.load`, `eval\(`},
	}

	for _, el := range program.Elements {
		if fn, ok := el.(*ir.Function); ok {
			for patternType, keywords := range securityPatterns {
				for _, keyword := range keywords {
					if strings.Contains(strings.ToLower(fn.Name), strings.ToLower(keyword)) {
						fn.SecurityTags = append(fn.SecurityTags, patternType)
					}
				}
			}
		}
	}
}

func (p *MultiLanguageParser) enrichWithAnnotations(program *ir.Program) {
	for _, el := range program.Elements {
		if fn, ok := el.(*ir.Function); ok {
			if strings.HasPrefix(fn.Name, "__") {
				fn.SecurityTags = append(fn.SecurityTags, "dunder")
			}
			if strings.HasSuffix(fn.Name, "_unsafe") {
				fn.SecurityTags = append(fn.SecurityTags, "unsafe")
			}
			if strings.Contains(fn.Name, "Handler") {
				fn.SecurityTags = append(fn.SecurityTags, "handler")
			}
			if strings.Contains(fn.Name, "Middleware") {
				fn.SecurityTags = append(fn.SecurityTags, "middleware")
			}
		}
	}
}

func generateNodeID(prefix, name string, line int) string {
	safeName := strings.ReplaceAll(name, "/", "_")
	safeName = strings.ReplaceAll(safeName, ".", "_")
	safeName = strings.ReplaceAll(safeName, "$", "_")
	return prefix + "_" + safeName + "_" + string(rune('a'+line%26))
}