package agent

import (
	"context"

	"charm.land/fantasy"
	"github.com/SecDuckOps/duckops/internal/agent/tools"
)

// repairToolCall rewrites common model parameter aliases (e.g. path → file_path)
// so fantasy schema validation passes before the tool handler runs.
func repairToolCall(_ context.Context, opts fantasy.ToolCallRepairOptions) (*fantasy.ToolCallContent, error) {
	normalized, ok := tools.NormalizeToolCallInput(opts.OriginalToolCall.ToolName, opts.OriginalToolCall.Input)
	if !ok {
		return nil, opts.ValidationError
	}
	repaired := opts.OriginalToolCall
	repaired.Input = normalized
	return &repaired, nil
}
