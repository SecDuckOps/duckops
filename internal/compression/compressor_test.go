package compression

import (
	"context"
	"testing"
	"time"
)

func TestCompressorLight(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.TargetTokens = 100
	cfg.Mode = ModeLight

	comp := NewCompressor(cfg)
	ctx := context.Background()
	input := testContext(50)

	result, err := comp.Compress(ctx, input)
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}
	if result.Metrics.CompressedTokens > result.Metrics.OriginalTokens {
		t.Errorf("compressed > original: %d > %d", result.Metrics.CompressedTokens, result.Metrics.OriginalTokens)
	}
}

func TestCompressorBalanced(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.TargetTokens = 100
	cfg.Mode = ModeBalanced

	comp := NewCompressor(cfg)
	ctx := context.Background()
	input := testContext(50)

	result, err := comp.Compress(ctx, input)
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}
	if result.Metrics.CompressedTokens > result.Metrics.OriginalTokens {
		t.Errorf("compressed > original: %d > %d", result.Metrics.CompressedTokens, result.Metrics.OriginalTokens)
	}
}

func TestCompressorAggressive(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.TargetTokens = 100
	cfg.Mode = ModeAggressive

	comp := NewCompressor(cfg)
	ctx := context.Background()
	input := testContext(50)

	result, err := comp.Compress(ctx, input)
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}
	if result.Metrics.CompressedTokens > result.Metrics.OriginalTokens {
		t.Errorf("compressed > original: %d > %d", result.Metrics.CompressedTokens, result.Metrics.OriginalTokens)
	}
	if result.Metrics.Ratio < 0.5 {
		t.Errorf("expected aggressive compression ratio >= 0.5, got %.2f", result.Metrics.Ratio)
	}
}

func TestCompressorDisabled(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = false
	cfg.Mode = ModeAggressive

	comp := NewCompressor(cfg)
	ctx := context.Background()
	input := testContext(50)

	result, err := comp.Compress(ctx, input)
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}
	if result.Metrics.Ratio != 0 {
		t.Errorf("disabled should have ratio=0, got %.2f", result.Metrics.Ratio)
	}
}

func TestCompressorUnderTarget(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.TargetTokens = 999999
	cfg.Mode = ModeBalanced

	comp := NewCompressor(cfg)
	ctx := context.Background()
	input := testContext(5)

	result, err := comp.Compress(ctx, input)
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}
	if result.Metrics.OriginalTokens != result.Metrics.CompressedTokens {
		t.Errorf("under-target should not compress: %d != %d", result.Metrics.CompressedTokens, result.Metrics.OriginalTokens)
	}
}

func TestEstimateTokens(t *testing.T) {
	comp := NewCompressor(DefaultConfig())
	ctx := context.Background()
	input := testContext(10)

	tokens := comp.EstimateTokens(ctx, input)
	if tokens <= 0 {
		t.Errorf("expected positive token count, got %d", tokens)
	}
}

func TestSummarize(t *testing.T) {
	comp := NewCompressor(DefaultConfig())
	ctx := context.Background()

	messages := []Message{
		{Role: "user", Content: "I need to build a web server with authentication."},
		{Role: "assistant", Content: "Let me help you set up JWT auth."},
		{Role: "user", Content: "We decided to use golang-jwt library."},
	}

	summary := comp.Summarize(ctx, messages)
	if summary == nil {
		t.Fatal("expected non-nil summary")
	}
}

func TestCompressLargeContext(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.TargetTokens = 1000
	cfg.Mode = ModeBalanced

	comp := NewCompressor(cfg)
	ctx := context.Background()

	input := Context{
		SystemPrompt: "You are a helpful assistant with security expertise.",
		Messages:     make([]Message, 200),
		Logs:         make([]string, 100),
	}

	for i := 0; i < 200; i++ {
		role := "user"
		if i%2 == 0 {
			role = "assistant"
		}
		input.Messages[i] = Message{
			Role:    role,
			Content: repeatString("This is test message number "+itoa(i)+" with some content to make it realistic. ", 5),
		}
	}
	for i := 0; i < 100; i++ {
		input.Logs[i] = "INFO: processing step " + itoa(i) + " completed"
	}
	input.Logs[10] = "ERROR: step 10 failed with timeout"
	input.Logs[50] = "WARN: step 50 slow response"

	result, err := comp.Compress(ctx, input)
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}

	if result.Metrics.OriginalTokens <= 0 {
		t.Error("expected positive original tokens")
	}
	if result.Metrics.Duration <= 0 {
		t.Error("expected positive duration")
	}
}

func TestDedupMessages(t *testing.T) {
	dp := NewDedupProcessor()

	msgs := []Message{
		{Role: "user", Content: "hello", Timestamp: time.Now()},
		{Role: "user", Content: "hello", Timestamp: time.Now()},
		{Role: "assistant", Content: "hi", Timestamp: time.Now()},
	}

	result, removed := dp.DedupMessages(msgs)
	if removed != 1 {
		t.Errorf("expected 1 duplicate, got %d", removed)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 messages, got %d", len(result))
	}
}

func TestDedupLogs(t *testing.T) {
	dp := NewDedupProcessor()

	logs := []string{
		"INFO: step 1",
		"INFO: step 2",
		"INFO: step 1",
		"INFO: step 3",
	}

	result, removed := dp.DedupLogs(logs)
	if removed != 1 {
		t.Errorf("expected 1 duplicate, got %d", removed)
	}
	if len(result) != 3 {
		t.Errorf("expected 3 logs, got %d", len(result))
	}
}

func TestCodeCompressBalanced(t *testing.T) {
	cc := NewCodeCompressor()

	code := `package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func notUsedFunction() string {
	return "this function is dead code"
}
`

	snippets := []CodeSnippet{
		{FilePath: "main.go", Language: "go", Content: code},
	}

	result := cc.Compress(snippets, ModeBalanced)
	if len(result) != 1 {
		t.Fatalf("expected 1 snippet, got %d", len(result))
	}

	if result[0].Content == "" {
		t.Fatal("compressed content should not be empty")
	}
}

func TestLogCompress(t *testing.T) {
	lc := NewLogCompressor()

	logs := []string{
		"INFO: starting server",
		"ERROR: connection refused on port 8080",
		"INFO: retrying connection",
		"WARN: timeout exceeded, retrying",
		"INFO: connected successfully",
		"DEBUG: request details: GET /api/v1/users",
		"ERROR: failed to authenticate user",
		"INFO: shutting down",
		"INFO: cleanup complete",
		"CRITICAL: out of memory",
	}

	result := lc.Compress(logs, ModeAggressive)
	if len(result) >= len(logs) {
		t.Errorf("expected fewer logs after aggressive compression, got %d", len(result))
	}

	hasError := false
	hasCritical := false
	for _, l := range result {
		if lc.IsImportant(l) {
			if containsFold(l, "error") {
				hasError = true
			}
			if containsFold(l, "critical") {
				hasCritical = true
			}
		}
	}
	if !hasError {
		t.Error("expected errors to be preserved")
	}
	if !hasCritical {
		t.Error("expected critical messages to be preserved")
	}
}

func TestSummarizer(t *testing.T) {
	s := NewSummarizer()

	messages := []Message{
		{Role: "user", Content: "I need a Go web server with JWT auth and rate limiting."},
		{Role: "assistant", Content: "We decided to use gin-gonic/gin as the framework."},
		{Role: "assistant", Content: "Found that golang-jwt v5 is the best option."},
		{Role: "user", Content: "We need rate limiting middleware too."},
		{Role: "assistant", Content: "Added rate limiting with token bucket. Issue: need to tune the limits."},
	}

	summary := s.SummarizeMessages(messages)
	if summary == nil {
		t.Fatal("expected non-nil summary")
	}
	if summary.UserObjective == "" {
		t.Error("expected user objective to be extracted")
	}
	if len(summary.PendingTasks) == 0 {
		t.Error("expected pending tasks to be extracted")
	}
}

func TestContextCache(t *testing.T) {
	cache := NewContextCache(10)

	ctx := Context{
		SystemPrompt: "test",
		Messages:     []Message{{Role: "user", Content: "hello"}},
	}

	key := cache.KeyForContext(ctx)
	entry := &CacheEntry{Key: key, CreatedAt: time.Now()}

	cache.Set(key, entry)

	got, ok := cache.Get(key)
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got.HitCount != 1 {
		t.Errorf("expected hit count 1, got %d", got.HitCount)
	}
}

func TestContextCacheEviction(t *testing.T) {
	cache := NewContextCache(2)

	for i := 0; i < 5; i++ {
		key := "key-" + itoa(i)
		entry := &CacheEntry{
			Key:       key,
			CreatedAt: time.Now(),
		}
		cache.Set(key, entry)
	}

	hits, misses, size := cache.Stats()
	if size > 2 {
		t.Errorf("expected max 2 entries, got %d", size)
	}
	if misses > 0 {
		t.Errorf("expected 0 misses for this test, got %d", misses)
	}
	_ = hits
}

func TestMetricsTracker(t *testing.T) {
	m := NewMetricsTracker()

	m.Record(10000, 5000, ModeBalanced, 100*time.Millisecond)
	m.Record(20000, 8000, ModeAggressive, 200*time.Millisecond)

	snap := m.Snapshot()
	if snap.OriginalTokens != 30000 {
		t.Errorf("expected 30000 original tokens, got %d", snap.OriginalTokens)
	}
	if snap.CompressedTokens != 13000 {
		t.Errorf("expected 13000 compressed tokens, got %d", snap.CompressedTokens)
	}
	if snap.CostSaved <= 0 {
		t.Error("expected positive cost saved")
	}
}

func TestTokenizer(t *testing.T) {
	tok := NewTokenizer()

	text := "Hello, world! This is a test."
	tokens := tok.EstimateText(text)
	if tokens <= 0 {
		t.Errorf("expected positive tokens, got %d", tokens)
	}

	messages := []Message{
		{Content: "Hello"},
		{Content: "World"},
	}
	tokens = tok.EstimateMessages(messages)
	if tokens <= 0 {
		t.Errorf("expected positive tokens for messages, got %d", tokens)
	}
}

func TestModeRatio(t *testing.T) {
	tests := []struct {
		mode Mode
		want float64
	}{
		{ModeNone, 1.0},
		{ModeLight, 0.8},
		{ModeBalanced, 0.5},
		{ModeAggressive, 0.2},
	}
	for _, tt := range tests {
		got := tt.mode.Ratio()
		if got != tt.want {
			t.Errorf("Mode(%s).Ratio() = %.1f, want %.1f", tt.mode, got, tt.want)
		}
	}
}

func TestConfig(t *testing.T) {
	cfg := DefaultConfig()
	if !cfg.Enabled {
		t.Error("default config should be enabled")
	}
	if cfg.Mode != ModeBalanced {
		t.Errorf("default mode should be balanced, got %s", cfg.Mode)
	}

	cfg.Mode = ModeAggressive
	if !cfg.Mode.Valid() {
		t.Error("ModeAggressive should be valid")
	}

	cfg.Mode = Mode("invalid")
	if cfg.Mode.Valid() {
		t.Error("invalid mode should not be valid")
	}
}

func TestLogCompressorIsImportant(t *testing.T) {
	lc := NewLogCompressor()

	important := []string{
		"ERROR: connection failed",
		"FATAL: out of memory",
		"WARN: timeout",
		"exit code 1",
		"Stack trace:",
		"Results: 5 passed, 2 failed",
	}
	notImportant := []string{
		"INFO: server started",
		"DEBUG: request received",
		"processing item 42",
	}

	for _, log := range important {
		if !lc.IsImportant(log) {
			t.Errorf("expected '%s' to be important", log)
		}
	}
	for _, log := range notImportant {
		if lc.IsImportant(log) {
			t.Errorf("expected '%s' to not be important", log)
		}
	}
}

func TestCodeExtractSignaturesLight(t *testing.T) {
	cc := NewCodeCompressor()

	smallCode := "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}"
	snippets := []CodeSnippet{
		{FilePath: "main.go", Language: "go", Content: smallCode},
	}

	result := cc.Compress(snippets, ModeLight)
	if len(result) != 1 {
		t.Fatalf("expected 1 snippet, got %d", len(result))
	}
	if !containsFold(result[0].Content, "func main") {
		t.Error("expected function signature to be preserved")
	}
}

func TestCompressorConcurrent(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.TargetTokens = 500
	cfg.Mode = ModeBalanced

	comp := NewCompressor(cfg)
	ctx := context.Background()
	input := testContext(30)

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := comp.Compress(ctx, input)
			if err != nil {
				t.Errorf("concurrent Compress failed: %v", err)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestCompressorPreservesSystemPrompt(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.TargetTokens = 100
	cfg.Mode = ModeAggressive

	comp := NewCompressor(cfg)
	ctx := context.Background()
	input := testContext(50)
	input.SystemPrompt = "You are a security-focused agent. Never ignore safety rules."

	result, err := comp.Compress(ctx, input)
	if err != nil {
		t.Fatalf("Compress failed: %v", err)
	}
	if result.SystemPrompt != input.SystemPrompt {
		t.Error("system prompt must be preserved unchanged")
	}
}

func testContext(msgCount int) Context {
	msgs := make([]Message, msgCount)
	for i := 0; i < msgCount; i++ {
		role := "user"
		if i%2 == 0 {
			role = "assistant"
		}
		msgs[i] = Message{
			Role:    role,
			Content: "Test message " + itoa(i) + " with enough content to generate tokens.",
		}
	}

	logs := make([]string, 20)
	for i := 0; i < 20; i++ {
		logs[i] = "INFO: step " + itoa(i) + " completed"
	}
	logs[3] = "ERROR: step 3 failed"
	logs[15] = "WARN: slow operation detected"

	return Context{
		SystemPrompt: "You are a helpful assistant.",
		Messages:     msgs,
		Logs:         logs,
	}
}

func repeatString(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
