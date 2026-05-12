package plugins

import "github.com/SecDuckOps/duckops/internal/security/architecture"

type DetectorPlugin interface {
	Name() string
	Detect(path string, ir *architecture.IR) error
}

type ThreatRulePlugin interface {
	Name() string
	Apply(ir *architecture.IR) ([]architecture.Finding, error)
}

type CloudPlugin interface {
	Name() string
	Enrich(ir *architecture.IR) error
}
