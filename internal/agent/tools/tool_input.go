package tools

import (
	"encoding/json"
)

// filePathParamTools are tools that require a file_path parameter. Models
// sometimes send path or location instead; NormalizeToolCallInput rewrites
// those aliases before fantasy schema validation runs.
var filePathParamTools = map[string]struct{}{
	ViewToolName:                {},
	WriteToolName:               {},
	EditToolName:                {},
	MultiEditToolName:           {},
	DownloadToolName:            {},
	VulnerabilityReportToolName: {},
}

var filePathParamAliases = []string{"path", "location"}

// NormalizeToolCallInput rewrites common parameter aliases in tool call JSON
// so schema validation accepts them. Returns the (possibly updated) input and
// whether any rewrite was applied.
func NormalizeToolCallInput(toolName, input string) (string, bool) {
	if _, ok := filePathParamTools[toolName]; !ok {
		return input, false
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(input), &raw); err != nil {
		return input, false
	}

	if _, hasFilePath := raw["file_path"]; hasFilePath {
		return input, false
	}

	for _, alias := range filePathParamAliases {
		if v, ok := raw[alias]; ok {
			raw["file_path"] = v
			delete(raw, alias)
			out, err := json.Marshal(raw)
			if err != nil {
				return input, false
			}
			return string(out), true
		}
	}

	return input, false
}
