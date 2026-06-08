package compression

import (
	"time"
)

type Mode string

const (
	ModeNone      Mode = "none"
	ModeLight     Mode = "light"
	ModeBalanced  Mode = "balanced"
	ModeAggressive Mode = "aggressive"
)

func (m Mode) Valid() bool {
	switch m {
	case ModeNone, ModeLight, ModeBalanced, ModeAggressive:
		return true
	}
	return false
}

func (m Mode) Ratio() float64 {
	switch m {
	case ModeNone:
		return 1.0
	case ModeLight:
		return 0.8
	case ModeBalanced:
		return 0.5
	case ModeAggressive:
		return 0.2
	default:
		return 1.0
	}
}

type Context struct {
	SystemPrompt  string
	Messages      []Message
	ToolOutputs   []ToolOutput
	CodeSnippets  []CodeSnippet
	Logs          []string
	AgentMemory   string
	Config        string
	SecurityRules string
}

type Message struct {
	Role      string
	Content   string
	Timestamp time.Time
	ToolCall  bool
	TokenCost int
}

type ToolOutput struct {
	ToolName   string
	Input      string
	Output     string
	Duration   time.Duration
	Timestamp  time.Time
	TokenCost  int
	Truncated  bool
}

type CodeSnippet struct {
	FilePath string
	Language string
	Content  string
	Imports  []string
	Funcs    []FuncSignature
	Size     int
}

type FuncSignature struct {
	Name      string
	Params    string
	Returns   string
	Comment   string
}

type CompressedContext struct {
	SystemPrompt    string
	Messages        []Message
	ToolOutputs     []ToolOutput
	Summary         string
	CodeSnippets    []CodeSnippet
	Metrics         CompressionMetrics
}

type CompressionMetrics struct {
	OriginalTokens   int           `json:"original_tokens"`
	CompressedTokens int           `json:"compressed_tokens"`
	Ratio            float64       `json:"ratio"`
	Mode             Mode          `json:"mode"`
	Duration         time.Duration `json:"duration_ms"`
	CostSaved        float64       `json:"cost_saved"`
	MessagesRemoved  int           `json:"messages_removed"`
	DuplicatesFound  int           `json:"duplicates_found"`
	SummariesCreated int           `json:"summaries_created"`
}

type Summary struct {
	UserObjective   string   `json:"user_objective"`
	CurrentProgress string   `json:"current_progress"`
	DecisionsMade   []string `json:"decisions_made"`
	Findings        []string `json:"findings"`
	OpenIssues      []string `json:"open_issues"`
	PendingTasks    []string `json:"pending_tasks"`
}

type CacheEntry struct {
	Key        string
	Summary    *Summary
	Compressed *CompressedContext
	CreatedAt  time.Time
	HitCount   int
}

const (
	CostPerThousandTokens = 0.002
	DefaultMaxTokens      = 120000
	DefaultTargetTokens   = 60000
)
