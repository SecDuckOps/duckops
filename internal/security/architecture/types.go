package architecture

import "time"

const IRVersion = "1.0.0"

type Confidence string

const (
	ConfidenceHigh   Confidence = "high"
	ConfidenceMedium Confidence = "medium"
	ConfidenceLow    Confidence = "low"
)

type Service struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Language   string   `json:"language"`
	Frameworks []string `json:"frameworks,omitempty"`
	Exposure   string   `json:"exposure,omitempty"`
}

type API struct {
	ID         string `json:"id"`
	ServiceID  string `json:"service_id"`
	Method     string `json:"method"`
	Route      string `json:"route"`
	AuthMethod string `json:"auth_method,omitempty"`
	Public     bool   `json:"public"`
}

type DataStore struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Sensitivity string `json:"sensitivity,omitempty"`
}

type Queue struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
}

type ExternalIntegration struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category,omitempty"`
}

type TrustBoundary struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Zone  string `json:"zone"`
	Scope string `json:"scope,omitempty"`
}

type AuthenticationControl struct {
	ID          string `json:"id"`
	ServiceID   string `json:"service_id,omitempty"`
	Method      string `json:"method"`
	HasJWT      bool   `json:"has_jwt,omitempty"`
	HasSession  bool   `json:"has_session,omitempty"`
	HasRateLimit bool  `json:"has_rate_limit,omitempty"`
}

type DataFlow struct {
	ID                 string `json:"id"`
	From               string `json:"from"`
	To                 string `json:"to"`
	Protocol           string `json:"protocol,omitempty"`
	AuthMethod         string `json:"auth_method,omitempty"`
	TrustZoneSrc       string `json:"trust_zone_src,omitempty"`
	TrustZoneDst       string `json:"trust_zone_dst,omitempty"`
	DataClassification string `json:"data_classification,omitempty"`
}

type Dependency struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Type    string `json:"type,omitempty"`
}

type Evidence struct {
	Path       string     `json:"path"`
	Extractor  string     `json:"extractor"`
	Confidence Confidence `json:"confidence"`
	Detail     string     `json:"detail,omitempty"`
}

type Finding struct {
	ID             string     `json:"id"`
	Title          string     `json:"title"`
	Severity       string     `json:"severity"`
	Component      string     `json:"component"`
	CWE            string     `json:"cwe,omitempty"`
	AttackScenario string     `json:"attack_scenario,omitempty"`
	Impact         string     `json:"impact,omitempty"`
	Likelihood     string     `json:"likelihood,omitempty"`
	Mitigation     string     `json:"mitigation,omitempty"`
	References     []string   `json:"references,omitempty"`
	ATTACK         string     `json:"attack,omitempty"`
	Category       string     `json:"category,omitempty"`
	Confidence     Confidence `json:"confidence"`
	Hypothesis     bool       `json:"hypothesis,omitempty"`
}

type IR struct {
	IRVersion            string                `json:"ir_version"`
	GeneratedAt          time.Time             `json:"generated_at"`
	RepositoryPath       string                `json:"repository_path"`
	Services             []Service             `json:"services"`
	APIs                 []API                 `json:"apis"`
	Databases            []DataStore           `json:"databases"`
	Queues               []Queue               `json:"queues"`
	ExternalIntegrations []ExternalIntegration `json:"external_integrations"`
	TrustBoundaries      []TrustBoundary       `json:"trust_boundaries"`
	Authentication       []AuthenticationControl `json:"authentication"`
	DataFlows            []DataFlow            `json:"data_flows"`
	Dependencies         []Dependency          `json:"dependencies"`
	Findings             []Finding             `json:"findings"`
	Evidence             []Evidence            `json:"evidence"`
}

func NewIR(repoPath string) IR {
	return IR{
		IRVersion:      IRVersion,
		GeneratedAt:    time.Now().UTC(),
		RepositoryPath: repoPath,
	}
}
