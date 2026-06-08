package tools

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeToolCallInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		tool      string
		input     string
		wantInput string
		wantOK    bool
	}{
		{
			name:      "view path alias",
			tool:      ViewToolName,
			input:     `{"path":"main.go"}`,
			wantInput: `{"file_path":"main.go"}`,
			wantOK:    true,
		},
		{
			name:      "view location alias",
			tool:      ViewToolName,
			input:     `{"location":"duckops://skills/foo/SKILL.md"}`,
			wantInput: `{"file_path":"duckops://skills/foo/SKILL.md"}`,
			wantOK:    true,
		},
		{
			name:      "already has file_path",
			tool:      ViewToolName,
			input:     `{"file_path":"main.go","offset":1}`,
			wantInput: `{"file_path":"main.go","offset":1}`,
			wantOK:    false,
		},
		{
			name:      "unrelated tool",
			tool:      BashToolName,
			input:     `{"command":"ls"}`,
			wantInput: `{"command":"ls"}`,
			wantOK:    false,
		},
		{
			name:      "invalid json",
			tool:      ViewToolName,
			input:     `{bad`,
			wantInput: `{bad`,
			wantOK:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotInput, gotOK := NormalizeToolCallInput(tt.tool, tt.input)
			require.Equal(t, tt.wantOK, gotOK)

			if tt.wantOK {
				var got, want map[string]any
				require.NoError(t, json.Unmarshal([]byte(gotInput), &got))
				require.NoError(t, json.Unmarshal([]byte(tt.wantInput), &want))
				require.Equal(t, want, got)
			} else {
				require.Equal(t, tt.wantInput, gotInput)
			}
		})
	}
}
