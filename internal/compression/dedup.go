package compression

import (
	"crypto/sha256"
	"fmt"
	"sync"
)

type DedupProcessor struct {
	mu     sync.Mutex
	hashes map[string]bool
}

func NewDedupProcessor() *DedupProcessor {
	return &DedupProcessor{
		hashes: make(map[string]bool),
	}
}

func (d *DedupProcessor) DedupMessages(messages []Message) ([]Message, int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	seen := make(map[string]bool)
	var result []Message
	removed := 0

	for _, msg := range messages {
		key := d.messageKey(msg)
		if seen[key] {
			removed++
			continue
		}
		seen[key] = true
		result = append(result, msg)
	}

	return result, removed
}

func (d *DedupProcessor) DedupToolOutputs(outputs []ToolOutput) ([]ToolOutput, int) {
	seen := make(map[string]bool)
	var result []ToolOutput
	removed := 0

	for _, o := range outputs {
		key := d.toolOutputKey(o)
		if seen[key] {
			removed++
			continue
		}
		seen[key] = true
		result = append(result, o)
	}

	return result, removed
}

func (d *DedupProcessor) DedupLogs(logs []string) ([]string, int) {
	seen := make(map[string]bool)
	var result []string
	removed := 0

	for _, log := range logs {
		key := d.hashString(log)
		if seen[key] {
			removed++
			continue
		}
		seen[key] = true
		result = append(result, log)
	}

	return result, removed
}

func (d *DedupProcessor) DedupCodeSnippets(snippets []CodeSnippet) ([]CodeSnippet, int) {
	seen := make(map[string]bool)
	var result []CodeSnippet
	removed := 0

	for _, s := range snippets {
		key := d.hashString(s.FilePath + s.Content)
		if seen[key] {
			removed++
			continue
		}
		seen[key] = true
		result = append(result, s)
	}

	return result, removed
}

func (d *DedupProcessor) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.hashes = make(map[string]bool)
}

func (d *DedupProcessor) messageKey(msg Message) string {
	return fmt.Sprintf("%s|%s|%d", msg.Role, d.hashString(msg.Content), msg.Timestamp.Unix())
}

func (d *DedupProcessor) toolOutputKey(o ToolOutput) string {
	return fmt.Sprintf("%s|%s|%s", o.ToolName, d.hashString(o.Input), d.hashString(o.Output))
}

func (d *DedupProcessor) hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return string(h[:16])
}
