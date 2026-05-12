package ir

import (
	"path/filepath"
	"strings"
)

type Program struct {
	RepoID     string
	Language   string
	FilePath   string
	SourceHash string
	Elements   []Element
	Errors     []ParseError
}

type Element interface {
	GetID() string
	GetType() string
}

type Function struct {
	ID                string
	Name              string
	FullyQualifiedName string
	StartLine         int
	EndLine           int
	Parameters        []Parameter
	ReturnType        string
	IsExported        bool
	IsPrivate         bool
	IsAsync           bool
	IsConstructor     bool
	SecurityTags      []string
	Calls             []string
	Reads             []string
	Writes            []string
	Imports           []string
}

func (f *Function) GetID() string   { return f.ID }
func (f *Function) GetType() string { return "function" }

type Parameter struct {
	Name       string
	Type       string
	IsOptional bool
	IsVariadic bool
}

type Class struct {
	ID                string
	Name              string
	FullyQualifiedName string
	StartLine         int
	EndLine           int
	Methods           []string
	Fields            []Field
	Parent            string
	Implements        []string
	SecurityTags      []string
}

func (c *Class) GetID() string   { return c.ID }
func (c *Class) GetType() string { return "class" }

type Field struct {
	Name       string
	Type       string
	IsPrivate  bool
	IsReadOnly bool
}

type Import struct {
	ID       string
	Source   string
	Alias    string
	Kind     string
	IsDynamic bool
}

func (i *Import) GetID() string   { return i.ID }
func (i *Import) GetType() string { return "import" }

type APIEndpoint struct {
	ID          string
	Path        string
	Method      string
	Handler     string
	AuthRequired bool
	Middleware  []string
	RequestBody string
	ResponseType string
	SecurityTags []string
}

func (a *APIEndpoint) GetID() string   { return a.ID }
func (a *APIEndpoint) GetType() string { return "api" }

type DatabaseAccess struct {
	ID          string
	Query       string
	Table       string
	Operation   string
	IsSensitive bool
}

func (d *DatabaseAccess) GetID() string   { return d.ID }
func (d *DatabaseAccess) GetType() string { return "database" }

type QueueProducer struct {
	ID        string
	QueueName string
	MessageType string
}

func (q *QueueProducer) GetID() string   { return q.ID }
func (q *QueueProducer) GetType() string { return "queue_producer" }

type QueueConsumer struct {
	ID        string
	QueueName string
	Handler   string
}

func (q *QueueConsumer) GetID() string   { return q.ID }
func (q *QueueConsumer) GetType() string { return "queue_consumer" }

type InfraResource struct {
	ID          string
	Type        string
	Name        string
	Provider    string
	Config      map[string]interface{}
	DependsOn   []string
}

func (i *InfraResource) GetID() string   { return i.ID }
func (i *InfraResource) GetType() string { return "infra" }

type ParseError struct {
	Message string
	Line    int
	Column  int
}

type SecurityMetadata struct {
	InternetExposed     bool
	AuthSensitive       bool
	HandlesTokens       bool
	HandlesPII         bool
	RequiresAuth       bool
	TrustBoundary       string
	SensitivityLevel    float64
}

func DetectSecurityTags(el Element) []string {
	var tags []string

	switch e := el.(type) {
	case *Function:
		name := strings.ToLower(e.Name)
		if strings.Contains(name, "auth") || strings.Contains(name, "login") || strings.Contains(name, "password") {
			tags = append(tags, "auth")
		}
		if strings.Contains(name, "admin") {
			tags = append(tags, "admin")
		}
		if strings.Contains(name, "api") || strings.Contains(name, "endpoint") {
			tags = append(tags, "api")
		}
		if e.IsExported && isPublicAPI(e) {
			tags = append(tags, "public_api")
		}

	case *APIEndpoint:
		tags = append(tags, "api")
		if e.AuthRequired {
			tags = append(tags, "auth_required")
		}
		if e.Method == "GET" {
			tags = append(tags, "read")
		} else {
			tags = append(tags, "write")
		}

	case *DatabaseAccess:
		tags = append(tags, "database")
		if e.IsSensitive {
			tags = append(tags, "sensitive_data")
		}
	}

	return tags
}

func isPublicAPI(f *Function) bool {
	pathLower := strings.ToLower(filepath.Ext(f.FullyQualifiedName))
	return pathLower == ".go" && f.Name[0] >= 'A' && f.Name[0] <= 'Z' ||
		strings.Contains(f.FullyQualifiedName, "/api/") ||
		strings.Contains(f.FullyQualifiedName, "/public/")
}

func (p *Program) GetFunctions() []*Function {
	var funcs []*Function
	for _, el := range p.Elements {
		if f, ok := el.(*Function); ok {
			funcs = append(funcs, f)
		}
	}
	return funcs
}

func (p *Program) GetClasses() []*Class {
	var classes []*Class
	for _, el := range p.Elements {
		if c, ok := el.(*Class); ok {
			classes = append(classes, c)
		}
	}
	return classes
}

func (p *Program) GetImports() []*Import {
	var imports []*Import
	for _, el := range p.Elements {
		if i, ok := el.(*Import); ok {
			imports = append(imports, i)
		}
	}
	return imports
}

func (p *Program) GetAPIs() []*APIEndpoint {
	var apis []*APIEndpoint
	for _, el := range p.Elements {
		if a, ok := el.(*APIEndpoint); ok {
			apis = append(apis, a)
		}
	}
	return apis
}

func DetectLanguage(filePath string) string {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".ts", ".tsx":
		return "typescript"
	case ".js", ".jsx":
		return "javascript"
	case ".rs":
		return "rust"
	case ".java":
		return "java"
	case ".rb":
		return "ruby"
	case ".php":
		return "php"
	case ".kt":
		return "kotlin"
	case ".swift":
		return "swift"
	default:
		return "unknown"
	}
}

func IsSecuritySensitive(source string) bool {
	lower := strings.ToLower(source)
	securityKeywords := []string{
		"auth", "jwt", "oauth", "password", "credential", "secret",
		"token", "session", "crypt", "ssl", "tls", "bcrypt",
		"scrypt", "hmac", "sign", "verify", "rsa", "aes",
	}
	for _, keyword := range securityKeywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}