package agent

import (
	"errors"
	"testing"

	"charm.land/fantasy"
	"github.com/stretchr/testify/require"
)

func TestCapMaxOutputTokens(t *testing.T) {
	t.Parallel()

	require.Equal(t, int64(1024), capMaxOutputTokens(131072, 130000, 65536))
	require.Equal(t, int64(65536), capMaxOutputTokens(131072, 50000, 65536))
	require.Equal(t, int64(8976), capMaxOutputTokens(131072, 118000, 65536))
}

func TestTruncateFantasyMessagesToTokenBudget(t *testing.T) {
	t.Parallel()

	msgs := []fantasy.Message{
		fantasy.NewUserMessage(stringsRepeat("hello ", 1000)),
		{Role: fantasy.MessageRoleAssistant, Content: []fantasy.MessagePart{fantasy.TextPart{Text: stringsRepeat("world ", 1000)}}},
		fantasy.NewUserMessage("recent question"),
	}
	truncated := truncateFantasyMessagesToTokenBudget(msgs, 50)
	require.NotEmpty(t, truncated)
	require.LessOrEqual(t, estimateFantasyMessagesTokens(truncated), estimateFantasyMessagesTokens(msgs))
}

func TestNeedsAutoSummarizeBeforeRequest(t *testing.T) {
	t.Parallel()

	const cw int64 = 131072
	// 20% reserve => summarize when input+output > 104858
	require.False(t, needsAutoSummarizeBeforeRequest(cw, 80000, 13000))
	require.True(t, needsAutoSummarizeBeforeRequest(cw, 105219, 65536))
	require.True(t, needsAutoSummarizeBeforeRequest(cw, 100000, 10000))
}

func TestIsContextLengthError(t *testing.T) {
	t.Parallel()

	require.True(t, isContextLengthError(errors.New("maximum context length is 131072 tokens")))
	require.False(t, isContextLengthError(errors.New("unrelated bad request")))
}

func stringsRepeat(s string, n int) string {
	out := ""
	for range n {
		out += s
	}
	return out
}
