package compression_test

import (
	"context"
	"fmt"
	"time"

	"github.com/SecDuckOps/duckops/internal/compression"
)

func Example() {
	cfg := compression.DefaultConfig()
	cfg.Mode = compression.ModeBalanced
	cfg.TargetTokens = 500

	comp := compression.NewCompressor(cfg)

	ctx := context.Background()
	input := compression.Context{
		SystemPrompt: "You are a helpful assistant.",
		Messages: []compression.Message{
			{Role: "user", Content: "I need to build a Go web server with authentication.", Timestamp: time.Now().Add(-10 * time.Minute)},
			{Role: "assistant", Content: "Sure! Let's start with Go net/http. First, create a main.go file.", Timestamp: time.Now().Add(-9 * time.Minute)},
			{Role: "user", Content: "We need JWT auth middleware too.", Timestamp: time.Now().Add(-8 * time.Minute)},
			{Role: "assistant", Content: "Here's a JWT middleware implementation using golang-jwt.", Timestamp: time.Now().Add(-7 * time.Minute)},
			{Role: "user", Content: "Can we also add rate limiting?", Timestamp: time.Now().Add(-6 * time.Minute)},
			{Role: "assistant", Content: "Adding rate limiting with a token bucket approach.", Timestamp: time.Now().Add(-5 * time.Minute)},
			{Role: "user", Content: "Let me test the current implementation.", Timestamp: time.Now().Add(-4 * time.Minute)},
			{Role: "assistant", Content: "Running tests... All tests pass.", Timestamp: time.Now().Add(-3 * time.Minute)},
		},
		ToolOutputs: []compression.ToolOutput{
			{ToolName: "read", Input: "main.go", Output: "package main\n\nfunc main() {\n\t// server setup\n}", Truncated: false},
			{ToolName: "test", Input: "./...", Output: "ok 0.1s", Duration: time.Second},
		},
		CodeSnippets: []compression.CodeSnippet{
			{
				FilePath: "main.go",
				Language: "go",
				Content:  "package main\n\nimport \"net/http\"\n\nfunc main() {\n\thttp.HandleFunc(\"/\", handler)\n\thttp.ListenAndServe(\":8080\", nil)\n}\n\nfunc handler(w http.ResponseWriter, r *http.Request) {\n\tw.Write([]byte(\"Hello\"))\n}",
			},
		},
	}

	result, err := comp.Compress(ctx, input)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Original tokens: %d\n", result.Metrics.OriginalTokens)
	fmt.Printf("Compressed tokens: %d\n", result.Metrics.CompressedTokens)
	fmt.Printf("Ratio: %.1f%%\n", result.Metrics.Ratio*100)
	fmt.Printf("Mode: %s\n", result.Metrics.Mode)
	fmt.Printf("Duration: %v\n", result.Metrics.Duration)
}
