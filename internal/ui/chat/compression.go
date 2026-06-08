package chat

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/SecDuckOps/duckops/internal/ui/styles"
)

type CompressionIndicatorItem struct {
	*cachedMessageItem
	id              string
	sty             *styles.Styles
	savedTokens     int64
	originalTokens  int64
}

func NewCompressionIndicatorItem(sty *styles.Styles, id string, savedTokens, originalTokens int64) MessageItem {
	return &CompressionIndicatorItem{
		cachedMessageItem: &cachedMessageItem{},
		id:                fmt.Sprintf("%s:compressed", id),
		sty:               sty,
		savedTokens:       savedTokens,
		originalTokens:    originalTokens,
	}
}

func (c *CompressionIndicatorItem) ID() string {
	return c.id
}

func (c *CompressionIndicatorItem) RawRender(width int) string {
	innerWidth := max(0, width-MessageLeftPaddingTotal)
	content, _, ok := c.getCachedRender(innerWidth)
	if !ok {
		content = c.renderContent(innerWidth)
		height := lipgloss.Height(content)
		c.setCachedRender(content, innerWidth, height)
	}
	return content
}

func (c *CompressionIndicatorItem) Render(width int) string {
	prefix := c.sty.Messages.SectionHeader.Render()
	lines := strings.Split(c.RawRender(width), "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}

func (c *CompressionIndicatorItem) renderContent(width int) string {
	lineStyle := c.sty.Messages.CompressionLine
	labelStyle := c.sty.Messages.CompressionLabel
	line := c.centeredRule(width, labelStyle)
	return lineStyle.Render(line)
}

func (c *CompressionIndicatorItem) centeredRule(width int, labelStyle lipgloss.Style) string {
	label := labelStyle.Render(c.formatLabel())
	labelWidth := lipgloss.Width(label)

	if labelWidth >= width {
		return label
	}

	sideLen := (width - labelWidth) / 2
	left := strings.Repeat("─", sideLen)
	right := strings.Repeat("─", width-sideLen-labelWidth)

	return left + label + right
}

func (c *CompressionIndicatorItem) formatLabel() string {
	if c.originalTokens <= 0 {
		return "  compressed "
	}

	ratio := float64(c.savedTokens) / float64(c.originalTokens) * 100
	saved := formatTokenCount(c.savedTokens)
	cost := estimateCost(c.savedTokens)
	return fmt.Sprintf("  %.0f%% (%s, -$%.4f) ", ratio, saved, cost)
}

func formatTokenCount(n int64) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

func estimateCost(tokens int64) float64 {
	return float64(tokens) / 1000 * 0.002
}
