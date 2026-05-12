package parser

import (
	"regexp"
	"strings"

	"github.com/SecDuckOps/duckops/internal/graphx/ir"
)

type PythonParser struct {
	*RegexParser
}

func NewPythonParser() *PythonParser {
	patterns := map[string]string{
		"func":  `^\s*(?:async\s+)?def\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`,
		"class": `^\s*class\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*[:\(]`,
		"import": `^\s*(?:import\s+([a-zA-Z_][a-zA-Z0-9_.]*)|from\s+([a-zA-Z_][a-zA-Z0-9_.]*)\s+import)`,
		"call":  `\b([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`,
		"api":   `@app\.(get|post|put|delete|patch)\s*\(['"]\/`,
	}
	p := NewRegexParser("python", patterns)
	p.BaseParser.language = "python"
	return &PythonParser{RegexParser: p}
}

func (p *PythonParser) Name() string  { return "python" }
func (p *PythonParser) Language() string { return "python" }

func (p *PythonParser) Parse(content []byte) (*ir.Program, error) {
	lines := strings.Split(string(content), "\n")
	program := &ir.Program{
		Language: "python",
		Elements: make([]ir.Element, 0),
		Errors:   make([]ir.ParseError, 0),
	}

	inClass := ""

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "class ") {
			parts := strings.Split(strings.TrimPrefix(trimmed, "class "), "(")
			className := strings.TrimSpace(parts[0])
			class := &ir.Class{
				ID:                irNodeID("class", className, lineNum),
				Name:              className,
				FullyQualifiedName: className,
				StartLine:         lineNum,
			}
			class.SecurityTags = ir.DetectSecurityTags(class)
			program.Elements = append(program.Elements, class)
			inClass = className
			continue
		}

		funcMatch := regexp.MustCompile(`^\s*(?:async\s+)?def\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`).FindStringSubmatch(line)
		if funcMatch != nil {
			funcName := funcMatch[1]
			fqn := funcName
			if inClass != "" {
				fqn = inClass + "." + funcName
			}

			isExported := strings.HasPrefix(trimmed, "def ") && len(trimmed) > 4 && strings.ToUpper(trimmed[4:5]) == trimmed[4:5]
			isPrivate := strings.HasPrefix(funcName, "_")

			fn := &ir.Function{
				ID:                 irNodeID("func", funcName, lineNum),
				Name:               funcName,
				FullyQualifiedName: fqn,
				StartLine:          lineNum,
				IsExported:         isExported,
				IsPrivate:          isPrivate,
			}

			if strings.HasPrefix(funcName, "__") {
				fn.SecurityTags = append(fn.SecurityTags, "dunder")
			}
			fn.SecurityTags = append(fn.SecurityTags, ir.DetectSecurityTags(fn)...)

			if fqn == "__init__" || fqn == inClass+"__init__" {
				fn.SecurityTags = append(fn.SecurityTags, "constructor")
			}

			program.Elements = append(program.Elements, fn)
			continue
		}

		importMatch := regexp.MustCompile(`^\s*(?:import\s+([a-zA-Z_][a-zA-Z0-9_.]*)|from\s+([a-zA-Z_][a-zA-Z0-9_.]*)\s+import)`).FindStringSubmatch(trimmed)
		if importMatch != nil {
			source := importMatch[1]
			if source == "" {
				source = importMatch[2]
			}
			if source != "" {
				imp := &ir.Import{
					ID:     irNodeID("imp", source, lineNum),
					Source: source,
				}
				imp.Kind = determineImportKind(source)
				program.Elements = append(program.Elements, imp)
			}
		}

		if trimmed == "" || !strings.HasPrefix(trimmed, " ") {
			inClass = ""
		}
	}

	program.Elements = p.detectAPIRoutes(program.Elements)

	return program, nil
}

func (p *PythonParser) detectAPIRoutes(elements []ir.Element) []ir.Element {
	var result []ir.Element

	for i, el := range elements {
		if fn, ok := elements[i].(*ir.Function); ok {
			for j := i - 1; j >= 0 && j >= i-5; j-- {
				if j >= 0 {
					if imp, ok := elements[j].(*ir.Import); ok && imp.Source == "flask" || imp.Source == "fastapi" || imp.Source == "django" {
						if strings.HasPrefix(fn.Name, "on_") || strings.HasPrefix(fn.Name, "handle_") {
							continue
						}
					}
				}
			}
		}
		result = append(result, el)
	}

	return result
}

func determineImportKind(source string) string {
	lower := strings.ToLower(source)
	if strings.HasPrefix(lower, "django.") || strings.HasPrefix(lower, "flask") || strings.HasPrefix(lower, "fastapi") {
		return "framework"
	}
	if strings.HasPrefix(lower, "test") || strings.HasPrefix(lower, "pytest") {
		return "test"
	}
	if strings.Contains(lower, "auth") || strings.Contains(lower, "security") {
		return "security"
	}
	return "library"
}

type GoParser struct {
	*RegexParser
}

func NewGoParser() *GoParser {
	patterns := map[string]string{
		"func":  `^func\s+(?:\([^\)]+\)\s+)?([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`,
		"class": ``,
		"import": `^\s*import\s+(?:\(\s*)?["']?([a-zA-Z0-9./]+)["']?`,
		"call":  `\b([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`,
		"api":   `@(Get|Post|Put|Delete|Patch)\(['"]\/`,
	}
	p := NewRegexParser("go", patterns)
	p.BaseParser.language = "go"
	return &GoParser{RegexParser: p}
}

func (p *GoParser) Name() string { return "go" }
func (p *GoParser) Language() string { return "go" }

func (p *GoParser) Parse(content []byte) (*ir.Program, error) {
	lines := strings.Split(string(content), "\n")
	program := &ir.Program{
		Language: "go",
		Elements: make([]ir.Element, 0),
		Errors:   make([]ir.ParseError, 0),
	}

	currentPackage := ""
	currentType := ""

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "package ") {
			parts := strings.Split(trimmed, " ")
			if len(parts) > 1 {
				currentPackage = parts[1]
			}
			continue
		}

		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			continue
		}

		typeMatch := regexp.MustCompile(`^\s*type\s+([a-zA-Z_][a-zA-Z0-9_]*)\s+(?:struct|interface)`).FindStringSubmatch(trimmed)
		if typeMatch != nil {
			currentType = typeMatch[1]
			class := &ir.Class{
				ID:                irNodeID("type", currentType, lineNum),
				Name:              currentType,
				FullyQualifiedName: currentPackage + "." + currentType,
				StartLine:         lineNum,
			}
			class.SecurityTags = ir.DetectSecurityTags(class)
			program.Elements = append(program.Elements, class)
			continue
		}

		if strings.HasPrefix(trimmed, "func ") {
			if !strings.HasPrefix(trimmed, "func (") {
				funcMatch := regexp.MustCompile(`^func\s+(?:(?:\([^)]+\)\s+))?([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`).FindStringSubmatch(trimmed)
				if funcMatch != nil {
					funcName := funcMatch[1]
					fqn := currentPackage + "." + funcName

					fn := &ir.Function{
						ID:                irNodeID("func", funcName, lineNum),
						Name:              funcName,
						FullyQualifiedName: fqn,
						StartLine:         lineNum,
						IsExported:        len(funcName) > 0 && funcName[0] >= 'A' && funcName[0] <= 'Z',
					}
					fn.SecurityTags = ir.DetectSecurityTags(fn)
					program.Elements = append(program.Elements, fn)
				}
			} else {
				methodMatch := regexp.MustCompile(`^func\s+\((?:(\w+)\s+\*?(\w+))\)\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`).FindStringSubmatch(trimmed)
				if methodMatch != nil {
					typeName := methodMatch[2]
					methodName := methodMatch[3]
					fqn := currentPackage + "." + typeName + "." + methodName

					fn := &ir.Function{
						ID:                 irNodeID("method", methodName, lineNum),
						Name:               methodName,
						FullyQualifiedName: fqn,
						StartLine:          lineNum,
						IsExported:         len(methodName) > 0 && methodName[0] >= 'A' && methodName[0] <= 'Z',
					}
					fn.SecurityTags = ir.DetectSecurityTags(fn)
					program.Elements = append(program.Elements, fn)
				}
			}
		}

		importMatch := regexp.MustCompile(`^\s*import\s+["']?([a-zA-Z0-9./]+)["']?`).FindStringSubmatch(trimmed)
		if importMatch != nil {
			source := importMatch[1]
			if !strings.HasPrefix(source, "_") && !strings.HasPrefix(source, ".") {
				imp := &ir.Import{
					ID:     irNodeID("imp", source, lineNum),
					Source: source,
					Kind:   determineGoImportKind(source),
				}
				program.Elements = append(program.Elements, imp)
			}
		}

		if strings.TrimSpace(trimmed) == "}" {
		}
	}

	return program, nil
}

func determineGoImportKind(source string) string {
	lower := strings.ToLower(source)
	if strings.HasPrefix(lower, "github.com/gin-gonic/gin") ||
		strings.HasPrefix(lower, "github.com/labstack/echo") ||
		strings.HasPrefix(lower, "github.com/gorilla/mux") {
		return "framework"
	}
	if strings.Contains(lower, "auth") || strings.Contains(lower, "jwt") {
		return "security"
	}
	if strings.HasPrefix(lower, "github.com/") {
		return "external"
	}
	return "stdlib"
}

type TypeScriptParser struct {
	*RegexParser
}

func NewTypeScriptParser() *TypeScriptParser {
	patterns := map[string]string{
		"func":  `(?:function\s+([a-zA-Z_$][a-zA-Z0-9_$]*)|const\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*(?:async\s*)?(?:\([^)]*\)|[^=])\s*[=>>]|([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\([^)]*\)\s*(?::\s*[^{]+)?\s*\{)`,
		"class": `class\s+([a-zA-Z_$][a-zA-Z0-9_$]*)`,
		"import": `import\s+(?:\{[^}]+\}|\*\s+as\s+\w+|\w+)\s+from\s+['"]([^'"]+)['"]`,
		"call":  `\b([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\(`,
		"api":   `@(Get|Post|Put|Delete|Patch|Use)\s*\(['"]\/`,
	}
	p := NewRegexParser("typescript", patterns)
	p.BaseParser.language = "typescript"
	return &TypeScriptParser{RegexParser: p}
}

func (p *TypeScriptParser) Name() string { return "typescript" }
func (p *TypeScriptParser) Language() string { return "typescript" }

func (p *TypeScriptParser) Parse(content []byte) (*ir.Program, error) {
	lines := strings.Split(string(content), "\n")
	program := &ir.Program{
		Language: "typescript",
		Elements: make([]ir.Element, 0),
		Errors:   make([]ir.ParseError, 0),
	}

	currentClass := ""

	reFunc := regexp.MustCompile(`(?:function\s+([a-zA-Z_$][a-zA-Z0-9_$]*)|(?:const|let|var)\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*(?:async\s+)?|([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\([^)]*\)\s*(?::\s*[^{]+)?\s*\{)`)

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "class ") {
			classMatch := regexp.MustCompile(`class\s+([a-zA-Z_$][a-zA-Z0-9_$]*)`).FindStringSubmatch(trimmed)
			if classMatch != nil {
				currentClass = classMatch[1]
				class := &ir.Class{
					ID:                irNodeID("class", currentClass, lineNum),
					Name:              currentClass,
					FullyQualifiedName: currentClass,
					StartLine:         lineNum,
				}
				class.SecurityTags = ir.DetectSecurityTags(class)
				program.Elements = append(program.Elements, class)
			}
			continue
		}

		if strings.HasPrefix(trimmed, "export ") || strings.HasPrefix(trimmed, "import ") {
			importMatch := regexp.MustCompile(`import\s+(?:\{[^}]+\}|\*\s+as\s+\w+|\w+)\s+from\s+['"]([^'"]+)['"]`).FindStringSubmatch(trimmed)
			if importMatch != nil {
				imp := &ir.Import{
					ID:     irNodeID("imp", importMatch[1], lineNum),
					Source: importMatch[1],
					Kind:   determineTSImportKind(importMatch[1]),
				}
				program.Elements = append(program.Elements, imp)
				continue
			}
		}

		if !strings.HasPrefix(trimmed, "class ") && !strings.HasPrefix(trimmed, "interface ") && !strings.HasPrefix(trimmed, "type ") {
			matches := reFunc.FindStringSubmatch(trimmed)
			if matches != nil {
				funcName := ""
				for _, m := range matches[1:] {
					if m != "" {
						funcName = m
						break
					}
				}

				if funcName != "" {
					fqn := funcName
					if currentClass != "" {
						fqn = currentClass + "." + funcName
					}

					isExported := strings.HasPrefix(trimmed, "export ")
					isAsync := strings.Contains(trimmed, "async")

					fn := &ir.Function{
						ID:                 irNodeID("func", funcName, lineNum),
						Name:               funcName,
						FullyQualifiedName: fqn,
						StartLine:          lineNum,
						IsExported:         isExported,
						IsAsync:            isAsync,
					}
					fn.SecurityTags = ir.DetectSecurityTags(fn)
					program.Elements = append(program.Elements, fn)
				}
			}
		}

		if trimmed == "" || (!strings.HasPrefix(trimmed, " ") && !strings.HasPrefix(trimmed, "\t")) {
			if !strings.HasPrefix(trimmed, "export ") && !strings.HasPrefix(trimmed, "import ") {
				currentClass = ""
			}
		}
	}

	return program, nil
}

func determineTSImportKind(source string) string {
	lower := strings.ToLower(source)
	if strings.HasPrefix(lower, "@types/") {
		return "types"
	}
	if strings.HasPrefix(lower, "@") || strings.HasPrefix(lower, "/") {
		return "module"
	}
	if strings.Contains(lower, "express") || strings.Contains(lower, "fastify") || strings.Contains(lower, "nest") {
		return "framework"
	}
	if strings.Contains(lower, "auth") || strings.Contains(lower, "jwt") {
		return "security"
	}
	return "library"
}

type RustParser struct {
	*RegexParser
}

func NewRustParser() *RustParser {
	patterns := map[string]string{
		"func":  `^\s*(?:pub\s+)?(?:async\s+)?fn\s+([a-zA-Z_][a-zA-Z0-9_]*)`,
		"class": `^\s*struct\s+([a-zA-Z_][a-zA-Z0-9_]*)`,
		"import": `^\s*use\s+([a-zA-Z_][a-zA-Z0-9_:]*)`,
		"call":  `\b([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`,
		"api":   ``,
	}
	p := NewRegexParser("rust", patterns)
	p.BaseParser.language = "rust"
	return &RustParser{RegexParser: p}
}

func (p *RustParser) Name() string { return "rust" }
func (p *RustParser) Language() string { return "rust" }

func (p *RustParser) Parse(content []byte) (*ir.Program, error) {
	lines := strings.Split(string(content), "\n")
	program := &ir.Program{
		Language: "rust",
		Elements: make([]ir.Element, 0),
		Errors:   make([]ir.ParseError, 0),
	}

	currentMod := ""
	reFunc := regexp.MustCompile(`^\s*(pub\s+)?(?:async\s+)?fn\s+([a-zA-Z_][a-zA-Z0-9_]*)`)

	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "mod ") {
			modMatch := regexp.MustCompile(`^mod\s+([a-zA-Z_][a-zA-Z0-9_]*)`).FindStringSubmatch(trimmed)
			if modMatch != nil {
				currentMod = modMatch[1]
			}
			continue
		}

		if strings.HasPrefix(trimmed, "use ") {
			useMatch := regexp.MustCompile(`^use\s+([a-zA-Z_][a-zA-Z0-9_:]*)`).FindStringSubmatch(trimmed)
			if useMatch != nil {
				imp := &ir.Import{
					ID:     irNodeID("use", useMatch[1], lineNum),
					Source: useMatch[1],
				}
				program.Elements = append(program.Elements, imp)
			}
			continue
		}

		matches := reFunc.FindStringSubmatch(trimmed)
		if matches != nil {
			isPub := matches[1] == "pub "
			funcName := matches[2]

			fqn := funcName
			if currentMod != "" {
				fqn = currentMod + "::" + funcName
			}

			fn := &ir.Function{
				ID:                irNodeID("fn", funcName, lineNum),
				Name:              funcName,
				FullyQualifiedName: fqn,
				StartLine:         lineNum,
				IsExported:        isPub,
			}
			fn.SecurityTags = ir.DetectSecurityTags(fn)
			program.Elements = append(program.Elements, fn)
		}
	}

	return program, nil
}

func init() {
}

type ParserOption func(p Parser)

func WithSecurityDetection() ParserOption {
	return func(p Parser) {
	}
}