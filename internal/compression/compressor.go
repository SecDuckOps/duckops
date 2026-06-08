package compression

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type ContextCompressor interface {
	Compress(ctx context.Context, input Context) (*CompressedContext, error)
	EstimateTokens(ctx context.Context, input Context) int
	Summarize(ctx context.Context, messages []Message) *Summary
}

type Compressor struct {
	config    Config
	tokenizer *Tokenizer
	dedup     *DedupProcessor
	codec     *CodeCompressor
	logc      *LogCompressor
	cache     *ContextCache
	metrics   *MetricsTracker
	mu        sync.Mutex
}

func NewCompressor(cfg Config) *Compressor {
	return &Compressor{
		config:    cfg,
		tokenizer: NewTokenizer(),
		dedup:     NewDedupProcessor(),
		codec:     NewCodeCompressor(),
		logc:      NewLogCompressor(),
		cache:     NewContextCache(cfg.CacheSize),
		metrics:   NewMetricsTracker(),
	}
}

func (c *Compressor) Compress(ctx context.Context, input Context) (*CompressedContext, error) {
	start := time.Now()
	originalTokens := c.tokenizer.EstimateContext(input)

	if !c.config.ShouldCompress(originalTokens) {
		return &CompressedContext{
			SystemPrompt: input.SystemPrompt,
			Messages:     input.Messages,
			ToolOutputs:  input.ToolOutputs,
			CodeSnippets: input.CodeSnippets,
			Metrics: CompressionMetrics{
				OriginalTokens:   originalTokens,
				CompressedTokens: originalTokens,
				Ratio:            0,
				Mode:             c.config.Mode,
				Duration:         0,
			},
		}, nil
	}

	ratio := c.config.Mode.Ratio()

	c.mu.Lock()
	defer c.mu.Unlock()

	result := &CompressedContext{
		SystemPrompt: input.SystemPrompt,
	}

	result.Summary = c.buildCompressionSummary(input, ratio)

	result.Messages = c.compressMessages(ctx, input.Messages, c.config.TargetTokens, ratio)

	result.ToolOutputs = c.compressToolOutputs(input.ToolOutputs, ratio)

	result.CodeSnippets = c.codec.Compress(input.CodeSnippets, c.config.Mode)

	c.compressLogsInline(&input, ratio)

	compressedTokens := c.tokenizer.EstimateCompressed(*result)

	metrics := c.metrics.Record(originalTokens, compressedTokens, c.config.Mode, time.Since(start))
	result.Metrics = metrics

	if c.config.LogEnabled {
		slog.Info("[compression]",
			"original_tokens", metrics.OriginalTokens,
			"compressed_tokens", metrics.CompressedTokens,
			"ratio", fmt.Sprintf("%.0f%%", metrics.Ratio*100),
			"mode", string(metrics.Mode),
			"duration_ms", metrics.Duration.Milliseconds(),
		)
	}

	return result, nil
}

func (c *Compressor) EstimateTokens(_ context.Context, input Context) int {
	return c.tokenizer.EstimateContext(input)
}

func (c *Compressor) Summarize(_ context.Context, messages []Message) *Summary {
	return c.buildSummary(messages)
}

func (c *Compressor) buildCompressionSummary(input Context, ratio float64) string {
	summary := c.buildSummary(input.Messages)
	if summary == nil {
		return ""
	}

	text := fmt.Sprintf(
		"[Session Summary]\n"+
			"Objective: %s\n"+
			"Progress: %s\n"+
			"Compression: %.0f%% of original\n",
		summary.UserObjective,
		summary.CurrentProgress,
		ratio*100,
	)

	if len(summary.DecisionsMade) > 0 {
		text += "\nKey Decisions:\n"
		for _, d := range summary.DecisionsMade {
			text += fmt.Sprintf("  - %s\n", d)
		}
	}
	if len(summary.Findings) > 0 {
		text += "\nFindings:\n"
		for _, f := range summary.Findings {
			text += fmt.Sprintf("  - %s\n", f)
		}
	}
	if len(summary.OpenIssues) > 0 {
		text += "\nOpen Issues:\n"
		for _, o := range summary.OpenIssues {
			text += fmt.Sprintf("  - %s\n", o)
		}
	}
	if len(summary.PendingTasks) > 0 {
		text += "\nPending Tasks:\n"
		for _, p := range summary.PendingTasks {
			text += fmt.Sprintf("  - %s\n", p)
		}
	}

	return text
}

func (c *Compressor) buildSummary(messages []Message) *Summary {
	if len(messages) == 0 {
		return nil
	}

	last := messages[len(messages)-1]

	summary := &Summary{
		UserObjective: extractObjective(messages),
		OpenIssues:    extractIssues(messages),
		PendingTasks:  extractTasks(messages),
	}

	if !last.ToolCall {
		summary.CurrentProgress = truncate(last.Content, 200)
	}

	return summary
}

func (c *Compressor) compressMessages(ctx context.Context, messages []Message, targetTokens int, ratio float64) []Message {
	if len(messages) <= 3 {
		return messages
	}

	var system, recent []Message
	var compressible []Message
	cutoff := len(messages) - 5
	if cutoff < 0 {
		cutoff = 0
	}

	for i, msg := range messages {
		switch {
		case i < cutoff && msg.Role == "system":
			system = append(system, msg)
		case i >= cutoff:
			recent = append(recent, msg)
		default:
			compressible = append(compressible, msg)
		}
	}

	compressed := c.compressMessageBatch(ctx, compressible, ratio)

	result := make([]Message, 0, len(system)+len(compressed)+len(recent))
	result = append(result, system...)
	result = append(result, compressed...)
	result = append(result, recent...)
	return result
}

func (c *Compressor) compressMessageBatch(_ context.Context, messages []Message, ratio float64) []Message {
	if len(messages) == 0 || ratio >= 1.0 {
		return messages
	}

	targetCount := int(float64(len(messages)) * ratio)
	if targetCount < 1 {
		targetCount = 1
	}

	summary := c.buildSummary(messages)
	if summary == nil {
		return messages[:targetCount]
	}

	compressed := Message{
		Role:      "system",
		Content:   c.formatSummary(summary),
		Timestamp: messages[0].Timestamp,
	}
	return []Message{compressed}
}

func (c *Compressor) formatSummary(s *Summary) string {
	text := "[Compressed History]\n"
	if s.UserObjective != "" {
		text += "Goal: " + s.UserObjective + "\n"
	}
	if s.CurrentProgress != "" {
		text += "Status: " + s.CurrentProgress + "\n"
	}
	if len(s.DecisionsMade) > 0 {
		text += "Decisions:\n"
		for _, d := range s.DecisionsMade {
			text += "  - " + d + "\n"
		}
	}
	if len(s.Findings) > 0 {
		text += "Findings:\n"
		for _, f := range s.Findings {
			text += "  - " + f + "\n"
		}
	}
	if len(s.OpenIssues) > 0 {
		text += "Issues:\n"
		for _, o := range s.OpenIssues {
			text += "  - " + o + "\n"
		}
	}
	if len(s.PendingTasks) > 0 {
		text += "Pending:\n"
		for _, p := range s.PendingTasks {
			text += "  - " + p + "\n"
		}
	}
	return text
}

func (c *Compressor) compressToolOutputs(outputs []ToolOutput, ratio float64) []ToolOutput {
	if len(outputs) == 0 {
		return nil
	}

	var important []ToolOutput
	var compressible []ToolOutput

	for _, o := range outputs {
		if !o.Truncated && len(o.Output) < 1000 {
			important = append(important, o)
		} else {
			compressible = append(compressible, o)
		}
	}

	keepCount := int(float64(len(compressible)) * ratio)
	if keepCount > len(compressible) {
		keepCount = len(compressible)
	}

	compressed := make([]ToolOutput, 0, len(important)+keepCount)
	compressed = append(compressed, important...)

	for i := 0; i < keepCount && i < len(compressible); i++ {
		out := compressible[i]
		if len(out.Output) > 500 {
			out.Output = truncate(out.Output, 500) + " [truncated]"
		}
		compressed = append(compressed, out)
	}

	return compressed
}

func (c *Compressor) compressLogsInline(input *Context, ratio float64) {
	if len(input.Logs) == 0 {
		return
	}

	keepCount := int(float64(len(input.Logs)) * ratio)
	if keepCount < 1 {
		keepCount = 1
	}

	var kept []string
	compressedCount := 0

	for i, log := range input.Logs {
		if c.logc.IsImportant(log) || i >= len(input.Logs)-keepCount {
			compressed := c.logc.CompressLine(log)
			kept = append(kept, compressed)
		} else {
			compressedCount++
		}
	}

	if compressedCount > 0 {
		kept = append(kept, fmt.Sprintf("[%d log lines compressed]", compressedCount))
	}

	input.Logs = kept
}

func extractObjective(messages []Message) string {
	for _, msg := range messages {
		if msg.Role == "user" && len(msg.Content) > 10 && len(msg.Content) < 500 {
			return truncate(msg.Content, 300)
		}
	}
	return ""
}

func extractIssues(messages []Message) []string {
	var issues []string
	seen := make(map[string]bool)
	for _, msg := range messages {
		lines := extractLines(msg.Content, "error", "issue", "bug", "fail")
		for _, line := range lines {
			if !seen[line] {
				issues = append(issues, line)
				seen[line] = true
			}
		}
	}
	if len(issues) > 5 {
		issues = issues[:5]
	}
	return issues
}

func extractTasks(messages []Message) []string {
	var tasks []string
	seen := make(map[string]bool)
	for _, msg := range messages {
		lines := extractLines(msg.Content, "todo", "task", "need", "pending")
		for _, line := range lines {
			if !seen[line] {
				tasks = append(tasks, line)
				seen[line] = true
			}
		}
	}
	if len(tasks) > 5 {
		tasks = tasks[:5]
	}
	return tasks
}

func extractLines(content string, keywords ...string) []string {
	if len(content) > 10000 {
		content = content[:10000]
	}
	var result []string
	lines := splitLines(content)
	for _, line := range lines {
		line = trimSpace(line)
		if line == "" {
			continue
		}
		for _, kw := range keywords {
			if containsFold(line, kw) {
				if len(line) > 200 {
					line = line[:200] + "..."
				}
				result = append(result, line)
				break
			}
		}
	}
	return result
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

var splitLines = func(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

var trimSpace = func(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

var containsFold = func(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	return len(s) >= len(substr) && contains(s, substr)
}

var toLower = func(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			b[i] = s[i] + 32
		} else {
			b[i] = s[i]
		}
	}
	return string(b)
}

var contains = func(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

var findSubstring = func(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
