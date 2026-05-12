package chat

import (
	"encoding/json"

	"github.com/SecDuckOps/duckops/internal/agent/tools"
	"github.com/SecDuckOps/duckops/internal/message"
	"github.com/SecDuckOps/duckops/internal/ui/styles"
)

// ArchitectureToolMessageItem renders threat-model/C4 architecture tool output.
type ArchitectureToolMessageItem struct {
	*baseToolMessageItem
}

var _ ToolMessageItem = (*ArchitectureToolMessageItem)(nil)

func NewArchitectureToolMessageItem(
	sty *styles.Styles,
	toolCall message.ToolCall,
	result *message.ToolResult,
	canceled bool,
) ToolMessageItem {
	return newBaseToolMessageItem(sty, toolCall, result, &ArchitectureToolRenderContext{}, canceled)
}

type ArchitectureToolRenderContext struct{}

func (a *ArchitectureToolRenderContext) RenderTool(sty *styles.Styles, width int, opts *ToolRenderOpts) string {
	cappedWidth := cappedMessageWidth(width)
	name := humanizedToolName(opts.ToolCall.Name)

	if opts.IsPending() {
		return pendingTool(sty, name, opts.Anim, opts.Compact)
	}

	var params map[string]any
	var toolParams []string
	if err := json.Unmarshal([]byte(opts.ToolCall.Input), &params); err == nil {
		if repoPath, ok := params["repository_path"].(string); ok && repoPath != "" {
			toolParams = append(toolParams, repoPath)
		}
	}

	header := toolHeader(sty, opts.Status, name, cappedWidth, opts.Compact, toolParams...)
	if opts.Compact {
		return header
	}
	if earlyState, ok := toolEarlyStateContent(sty, opts, cappedWidth); ok {
		return joinToolParts(header, earlyState)
	}
	if !opts.HasResult() || opts.Result.Content == "" {
		return header
	}

	bodyWidth := cappedWidth - toolBodyLeftPaddingTotal
	body := sty.Tool.Body.Render(toolOutputMarkdownContent(sty, opts.Result.Content, bodyWidth, opts.ExpandedContent))
	return joinToolParts(header, body)
}

func IsArchitectureTool(name string) bool {
	return name == tools.AnalyzeArchitectureToolName ||
		name == tools.GenerateC4ModelToolName ||
		name == tools.GenerateThreatModelToolName ||
		name == tools.GenerateAttackPathsToolName ||
		name == tools.ExportSecurityReportToolName
}
