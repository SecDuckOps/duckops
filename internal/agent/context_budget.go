package agent

import (
	"errors"
	"strings"

	"charm.land/fantasy"
	"github.com/SecDuckOps/duckops/internal/message"
	"github.com/SecDuckOps/duckops/internal/skills"
)

const (
	contextRequestBuffer = 4096
	minMaxOutputTokens    = 1024
)

func resolveMaxOutputTokens(call SessionAgentCall, model Model) int64 {
	if call.MaxOutputTokens > 0 {
		return call.MaxOutputTokens
	}
	return model.CatwalkCfg.DefaultMaxTokens
}

func capMaxOutputTokens(contextWindow, estimatedInput, requested int64) int64 {
	if requested <= 0 {
		return requested
	}
	if contextWindow <= 0 {
		return requested
	}
	available := contextWindow - estimatedInput - contextRequestBuffer
	if available < minMaxOutputTokens {
		return minMaxOutputTokens
	}
	if requested > available {
		return available
	}
	return requested
}

func estimateRequestTokens(systemPrompt string, history []fantasy.Message, prompt string, files []fantasy.FilePart, attachments ...message.Attachment) int64 {
	total := int64(skills.ApproxTokenCount(systemPrompt))
	total += int64(skills.ApproxTokenCount(prompt))
	total += estimateFantasyMessagesTokens(history)
	for _, file := range files {
		total += int64(skills.ApproxTokenCount(file.Filename))
		total += int64(len(file.Data) / 4)
	}
	for _, att := range attachments {
		if att.IsText() {
			total += int64(skills.ApproxTokenCount(string(att.Content)))
		}
	}
	return total
}

func estimateFantasyMessagesTokens(msgs []fantasy.Message) int64 {
	var total int64
	for _, msg := range msgs {
		for _, part := range msg.Content {
			switch p := part.(type) {
			case fantasy.TextPart:
				total += int64(skills.ApproxTokenCount(p.Text))
			case fantasy.ToolResultPart:
				total += int64(estimateToolResultOutputTokens(p.Output))
			case fantasy.ToolCallPart:
				total += int64(skills.ApproxTokenCount(p.Input))
				total += int64(skills.ApproxTokenCount(p.ToolName))
			case fantasy.FilePart:
				total += int64(skills.ApproxTokenCount(p.Filename))
				total += int64(len(p.Data) / 4)
			case fantasy.ReasoningPart:
				total += int64(skills.ApproxTokenCount(p.Text))
			}
		}
	}
	return total
}

func truncateFantasyMessagesToTokenBudget(msgs []fantasy.Message, maxTokens int) []fantasy.Message {
	if maxTokens <= 0 || len(msgs) == 0 {
		return msgs
	}
	if int(estimateFantasyMessagesTokens(msgs)) <= maxTokens {
		return msgs
	}

	start := 0
	for start < len(msgs) && int(estimateFantasyMessagesTokens(msgs[start:])) > maxTokens {
		start++
		// Prefer cutting at a user-message boundary so we do not start mid-turn.
		for start < len(msgs) && msgs[start].Role != fantasy.MessageRoleUser {
			start++
		}
	}
	if start >= len(msgs) {
		return msgs[len(msgs)-1:]
	}
	return msgs[start:]
}

func isContextLengthError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "context length") ||
		strings.Contains(msg, "maximum context") ||
		strings.Contains(msg, "too many tokens") ||
		strings.Contains(msg, "context window") {
		return true
	}
	var providerErr *fantasy.ProviderError
	if errors.As(err, &providerErr) && providerErr.StatusCode == 400 {
		body := strings.ToLower(providerErr.Message)
		return strings.Contains(body, "context") && strings.Contains(body, "token")
	}
	return false
}

func estimateToolResultOutputTokens(output fantasy.ToolResultOutputContent) int {
	if text, ok := fantasy.AsToolResultOutputType[fantasy.ToolResultOutputContentText](output); ok {
		return skills.ApproxTokenCount(text.Text)
	}
	if errOut, ok := fantasy.AsToolResultOutputType[fantasy.ToolResultOutputContentError](output); ok {
		return skills.ApproxTokenCount(errOut.Error.Error())
	}
	if media, ok := fantasy.AsToolResultOutputType[fantasy.ToolResultOutputContentMedia](output); ok {
		return len(media.Data) / 4
	}
	return 0
}

// autoSummarizeReserveTokens is how much context headroom we keep before
// automatically summarizing. Matches the StopWhen threshold in Run().
func autoSummarizeReserveTokens(contextWindow int64) int64 {
	if contextWindow <= 0 {
		return contextRequestBuffer
	}
	if contextWindow > largeContextWindowThreshold {
		return largeContextWindowBuffer
	}
	return int64(float64(contextWindow) * smallContextWindowRatio)
}

// needsAutoSummarizeBeforeRequest reports whether the upcoming request should
// trigger automatic summarization before contacting the provider.
func needsAutoSummarizeBeforeRequest(contextWindow, estimatedInput, maxOutput int64) bool {
	if contextWindow <= 0 {
		return false
	}
	reserve := autoSummarizeReserveTokens(contextWindow)
	return estimatedInput+maxOutput > contextWindow-reserve
}
