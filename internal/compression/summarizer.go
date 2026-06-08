package compression

import (
	"fmt"
	"strings"
	"sync"
)

type Summarizer struct {
	mu sync.Mutex
}

func NewSummarizer() *Summarizer {
	return &Summarizer{}
}

func (s *Summarizer) SummarizeMessages(messages []Message) *Summary {
	if len(messages) == 0 {
		return nil
	}

	summary := &Summary{}
	objective := s.extractObjective(messages)

	if objective != "" {
		summary.UserObjective = objective
	}

	summary.CurrentProgress = s.extractProgress(messages)
	summary.DecisionsMade = s.extractDecisions(messages)
	summary.Findings = s.extractFindings(messages)
	summary.OpenIssues = s.extractIssues(messages)
	summary.PendingTasks = s.extractPendingTasks(messages)

	return summary
}

func (s *Summarizer) extractObjective(messages []Message) string {
	for _, msg := range messages {
		if msg.Role == "user" && len(msg.Content) > 10 {
			text := strings.TrimSpace(msg.Content)
			if idx := strings.Index(text, "\n"); idx > 0 && idx < 300 {
				text = text[:idx]
			}
			if len(text) > 300 {
				text = text[:300]
			}
			return text
		}
	}
	return ""
}

func (s *Summarizer) extractProgress(messages []Message) string {
	if len(messages) == 0 {
		return ""
	}

	last := messages[len(messages)-1]
	if !last.ToolCall && len(last.Content) > 0 {
		content := last.Content
		if len(content) > 500 {
			content = content[:500]
		}
		lines := strings.SplitN(content, "\n", 4)
		var result []string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				result = append(result, line)
			}
		}
		if len(result) > 3 {
			result = result[:3]
		}
		return strings.Join(result, "; ")
	}
	return ""
}

func (s *Summarizer) extractDecisions(messages []Message) []string {
	var decisions []string
	seen := make(map[string]bool)

	for _, msg := range messages {
		lower := toLower(msg.Content)
		if !strings.Contains(lower, "decid") &&
			!strings.Contains(lower, "chose") &&
			!strings.Contains(lower, "pick") &&
			!strings.Contains(lower, "select") &&
			!strings.Contains(lower, "going with") {
			continue
		}

		for _, line := range splitLines(msg.Content) {
			trimmed := trimSpace(line)
			if trimmed == "" {
				continue
			}
			lowerLine := toLower(trimmed)
			if !strings.Contains(lowerLine, "decid") &&
				!strings.Contains(lowerLine, "chose") &&
				!strings.Contains(lowerLine, "pick") &&
				!strings.Contains(lowerLine, "select") &&
				!strings.Contains(lowerLine, "going with") {
				continue
			}
			if len(trimmed) > 200 {
				trimmed = trimmed[:200] + "..."
			}
			if !seen[trimmed] {
				decisions = append(decisions, trimmed)
				seen[trimmed] = true
			}
		}
	}

	if len(decisions) > 10 {
		decisions = decisions[:10]
	}
	return decisions
}

func (s *Summarizer) extractFindings(messages []Message) []string {
	var findings []string
	seen := make(map[string]bool)

	for _, msg := range messages {
		lower := toLower(msg.Content)
		if !strings.Contains(lower, "found") &&
			!strings.Contains(lower, "discover") &&
			!strings.Contains(lower, "result") &&
			!strings.Contains(lower, "detect") &&
			!strings.Contains(lower, "identified") {
			continue
		}

		for _, line := range splitLines(msg.Content) {
			trimmed := trimSpace(line)
			if trimmed == "" {
				continue
			}
			lowerLine := toLower(trimmed)
			if !strings.Contains(lowerLine, "found") &&
				!strings.Contains(lowerLine, "discover") &&
				!strings.Contains(lowerLine, "result") &&
				!strings.Contains(lowerLine, "detect") &&
				!strings.Contains(lowerLine, "identified") {
				continue
			}
			if len(trimmed) > 200 {
				trimmed = trimmed[:200] + "..."
			}
			if !seen[trimmed] {
				findings = append(findings, trimmed)
				seen[trimmed] = true
			}
		}
	}

	if len(findings) > 10 {
		findings = findings[:10]
	}
	return findings
}

func (s *Summarizer) extractIssues(messages []Message) []string {
	return extractIssues(messages)
}

func (s *Summarizer) extractPendingTasks(messages []Message) []string {
	return extractTasks(messages)
}

func (s *Summarizer) FormatSummary(summary *Summary) string {
	if summary == nil {
		return ""
	}

	var b strings.Builder
	b.WriteString("[Context Summary]\n")

	if summary.UserObjective != "" {
		b.WriteString(fmt.Sprintf("Objective: %s\n", summary.UserObjective))
	}
	if summary.CurrentProgress != "" {
		b.WriteString(fmt.Sprintf("Progress: %s\n", summary.CurrentProgress))
	}
	if len(summary.DecisionsMade) > 0 {
		b.WriteString("Decisions:\n")
		for _, d := range summary.DecisionsMade {
			b.WriteString(fmt.Sprintf("  - %s\n", d))
		}
	}
	if len(summary.Findings) > 0 {
		b.WriteString("Findings:\n")
		for _, f := range summary.Findings {
			b.WriteString(fmt.Sprintf("  - %s\n", f))
		}
	}
	if len(summary.OpenIssues) > 0 {
		b.WriteString("Open Issues:\n")
		for _, o := range summary.OpenIssues {
			b.WriteString(fmt.Sprintf("  - %s\n", o))
		}
	}
	if len(summary.PendingTasks) > 0 {
		b.WriteString("Pending Tasks:\n")
		for _, p := range summary.PendingTasks {
			b.WriteString(fmt.Sprintf("  - %s\n", p))
		}
	}

	return b.String()
}
