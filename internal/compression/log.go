package compression

import (
	"strings"
	"sync"
)

type LogCompressor struct {
	mu sync.Mutex
}

func NewLogCompressor() *LogCompressor {
	return &LogCompressor{}
}

func (l *LogCompressor) Compress(logs []string, mode Mode) []string {
	if len(logs) == 0 || mode == ModeNone {
		return logs
	}

	ratio := mode.Ratio()
	targetCount := int(float64(len(logs)) * ratio)
	if targetCount < 1 {
		targetCount = 1
	}

	var important []string
	var regular []string

	for _, log := range logs {
		if l.IsImportant(log) {
			important = append(important, l.CompressLine(log))
		} else {
			regular = append(regular, log)
		}
	}

	if len(regular) > targetCount {
		skipped := len(regular) - targetCount
		regular = regular[:targetCount]
		regular = append(regular, "["+itoa(skipped)+" regular log lines suppressed]")
	}

	result := make([]string, 0, len(important)+len(regular))
	result = append(result, important...)
	result = append(result, regular...)
	return result
}

func (l *LogCompressor) CompressLine(log string) string {
	if len(log) <= 200 {
		return log
	}

	trimmed := strings.TrimSpace(log)
	if len(trimmed) <= 200 {
		return trimmed
	}

	parts := strings.Fields(trimmed)
	if len(parts) > 20 {
		parts = append(parts[:20], "...")
		return strings.Join(parts, " ")
	}

	return trimmed[:200] + "..."
}

func (l *LogCompressor) IsImportant(log string) bool {
	lower := toLower(log)
	indicators := []string{
		"error", "fatal", "panic", "fail", "exception",
		"warning", "warn", "critical", "alert",
		"traceback", "stack trace", "segfault",
		"denied", "rejected", "timeout",
		"exit code", "exit status", "signal",
		"oom", "out of memory", "disk full",
		"summary", "statistics", "results:",
		"passed", "failed", "ok", "not ok",
	}
	for _, ind := range indicators {
		if strings.Contains(lower, ind) {
			return true
		}
	}
	return false
}
