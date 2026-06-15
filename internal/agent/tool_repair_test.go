package agent

import (
	"context"
	"testing"

	"charm.land/fantasy"
	"github.com/stretchr/testify/require"
)

func TestRepairToolCallPathAlias(t *testing.T) {
	t.Parallel()

	repaired, err := repairToolCall(context.Background(), fantasy.ToolCallRepairOptions{
		OriginalToolCall: fantasy.ToolCallContent{
			ToolName: "view",
			Input:    `{"path":"main.go"}`,
		},
		ValidationError: context.Canceled, // unused when repair succeeds
	})
	require.NoError(t, err)
	require.NotNil(t, repaired)
	require.JSONEq(t, `{"file_path":"main.go"}`, repaired.Input)
}
