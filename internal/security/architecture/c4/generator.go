package c4

import (
	"fmt"
	"strings"

	"github.com/SecDuckOps/duckops/internal/security/architecture"
)

type Models struct {
	Context     string
	Container   string
	Component   string
	PlantUML    string
	Structurizr string
}

func Generate(ir architecture.IR) Models {
	return Models{
		Context:     contextMermaid(ir),
		Container:   containerMermaid(ir),
		Component:   componentMermaid(ir),
		PlantUML:    plantUML(ir),
		Structurizr: structurizr(ir),
	}
}

func contextMermaid(ir architecture.IR) string {
	var b strings.Builder
	b.WriteString("flowchart LR\n")
	b.WriteString("User[User] --> System[System]\n")
	for _, s := range ir.Services {
		b.WriteString(fmt.Sprintf("System --> %s[%s]\n", sanitizeID(s.ID), s.Name))
	}
	return b.String()
}

func containerMermaid(ir architecture.IR) string {
	var b strings.Builder
	b.WriteString("flowchart TD\n")
	for _, s := range ir.Services {
		b.WriteString(fmt.Sprintf("%s[%s (%s)]\n", sanitizeID(s.ID), s.Name, s.Language))
	}
	for _, d := range ir.Databases {
		b.WriteString(fmt.Sprintf("%s[%s]\n", sanitizeID(d.ID), d.Name))
	}
	for _, f := range ir.DataFlows {
		b.WriteString(fmt.Sprintf("%s --> %s\n", sanitizeID(f.From), sanitizeID(f.To)))
	}
	return b.String()
}

func componentMermaid(ir architecture.IR) string {
	var b strings.Builder
	b.WriteString("flowchart LR\n")
	for _, a := range ir.APIs {
		b.WriteString(fmt.Sprintf("%s[%s %s] --> %s\n", sanitizeID(a.ID), a.Method, a.Route, sanitizeID(a.ServiceID)))
	}
	return b.String()
}

func plantUML(ir architecture.IR) string {
	var b strings.Builder
	b.WriteString("@startuml\n")
	b.WriteString("actor User\n")
	b.WriteString("rectangle System {\n")
	for _, s := range ir.Services {
		b.WriteString(fmt.Sprintf("  component \"%s\" as %s\n", s.Name, sanitizeID(s.ID)))
	}
	b.WriteString("}\n@enduml\n")
	return b.String()
}

func structurizr(ir architecture.IR) string {
	var b strings.Builder
	b.WriteString("workspace {\n")
	b.WriteString("  model {\n")
	b.WriteString("    user = person \"User\"\n")
	b.WriteString("    system = softwareSystem \"System\"\n")
	for _, s := range ir.Services {
		b.WriteString(fmt.Sprintf("    %s = container \"%s\" \"%s\"\n", sanitizeID(s.ID), s.Name, s.Language))
		b.WriteString(fmt.Sprintf("    system -> %s \"uses\"\n", sanitizeID(s.ID)))
	}
	b.WriteString("  }\n}\n")
	return b.String()
}

func sanitizeID(in string) string {
	return strings.NewReplacer(":", "_", "-", "_", "/", "_", ".", "_").Replace(in)
}
