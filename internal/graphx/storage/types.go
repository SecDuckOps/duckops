package storage

import (
	"encoding/json"
	"math/rand"
	"time"
)

type NodeType string

const (
	NodeTypeFile     NodeType = "file"
	NodeTypeFunction NodeType = "function"
	NodeTypeClass    NodeType = "class"
	NodeTypeAPI      NodeType = "api"
)

type EdgeType string

type Node struct {
	ID                  string                 `json:"id"`
	RepoID              string                 `json:"repo_id"`
	Type                NodeType               `json:"type"`
	Name                string                 `json:"name"`
	FullyQualifiedName string                 `json:"fully_qualified_name,omitempty"`
	FilePath            string                 `json:"file_path,omitempty"`
	Language            string                 `json:"language,omitempty"`
	Framework           string                 `json:"framework,omitempty"`
	SecuritySensitivity float64                `json:"security_sensitivity"`
	InternetExposed     bool                   `json:"internet_exposed"`
	AuthSensitive       bool                   `json:"auth_sensitive"`
	PIIHandling         bool                   `json:"pii_handling"`
	Ownership           string                 `json:"ownership,omitempty"`
	RiskScore           float64                `json:"risk_score"`
	BlastRadiusScore    float64                `json:"blast_radius_score"`
	ContentHash         string                 `json:"content_hash,omitempty"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
	IsExported          bool                   `json:"is_exported"`
	CreatedAt           int64                  `json:"created_at"`
	UpdatedAt           int64                  `json:"updated_at"`
	IndexedAt           int64                  `json:"indexed_at,omitempty"`
}

type Edge struct {
	ID                string                 `json:"id"`
	RepoID            string                 `json:"repo_id"`
	SourceID          string                 `json:"source_id"`
	TargetID          string                 `json:"target_id"`
	EdgeType          EdgeType               `json:"edge_type"`
	Confidence        float64                `json:"confidence"`
	ConfidenceLevel   string                 `json:"confidence_level"`
	ExtractionSource  string                 `json:"extraction_source"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt         int64                  `json:"created_at"`
	UpdatedAt         int64                  `json:"updated_at"`
}

func NewNode(repoID, name string, nodeType NodeType) *Node {
	now := time.Now().UnixMilli()
	return &Node{
		ID:                  GenerateID(),
		RepoID:              repoID,
		Type:                nodeType,
		Name:                name,
		SecuritySensitivity: 0.0,
		RiskScore:           0.0,
		BlastRadiusScore:    0.0,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

func NewEdge(repoID, sourceID, targetID string, edgeType EdgeType) *Edge {
	now := time.Now().UnixMilli()
	return &Edge{
		ID:                GenerateID(),
		RepoID:            repoID,
		SourceID:          sourceID,
		TargetID:          targetID,
		EdgeType:          edgeType,
		Confidence:        1.0,
		ConfidenceLevel:  "EXTRACTED",
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

func (n *Node) MetadataJSON() string {
	if n.Metadata == nil {
		return "{}"
	}
	b, _ := json.Marshal(n.Metadata)
	return string(b)
}

func (e *Edge) MetadataJSON() string {
	if e.Metadata == nil {
		return "{}"
	}
	b, _ := json.Marshal(e.Metadata)
	return string(b)
}

func GenerateID() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 16)
	rand.Seed(time.Now().UnixNano())
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}