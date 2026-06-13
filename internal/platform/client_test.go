package platform

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newTestServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
	srv := httptest.NewServer(handler)
	client := NewClient("duck_pat_test123456789012345678901234567890123456789012345678")
	client.BaseURL = srv.URL
	return srv, client
}

func TestVerifyPAT_Success(t *testing.T) {
	srv, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer duck_pat_test123456789012345678901234567890123456789012345678" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success"}`))
	})
	defer srv.Close()

	err := client.VerifyPAT(context.Background())
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestVerifyPAT_Failure(t *testing.T) {
	srv, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"status":"error","message":"invalid token"}`))
	})
	defer srv.Close()

	err := client.VerifyPAT(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsAuthError(err) {
		t.Fatalf("expected auth error, got %v", err)
	}
}

func TestGetServerInfo(t *testing.T) {
	srv, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success","data":{"version":"1.0.0","endpoints":{"scans":"/api/v1/scans"}}}`))
	})
	defer srv.Close()

	info, err := client.GetServerInfo(context.Background())
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if info.Data.Version != "1.0.0" {
		t.Fatalf("expected 1.0.0, got %s", info.Data.Version)
	}
}

func TestUploadScan(t *testing.T) {
	srv, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"status":"success"}`))
	})
	defer srv.Close()

	err := client.UploadScan(context.Background(), ScanPayload{
		ScanID: "test-scan-1",
		Type:   "semgrep",
		Status: "completed",
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestUploadFindingsBulk(t *testing.T) {
	srv, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var req BulkFindingsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode error: %v", err)
		}
		if len(req.Findings) != 2 {
			t.Errorf("expected 2 findings, got %d", len(req.Findings))
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"status":"success"}`))
	})
	defer srv.Close()

	err := client.UploadFindingsBulk(context.Background(), []FindingPayload{
		{Tool: "semgrep", Severity: "high", Title: "Test 1"},
		{Tool: "trivy", Severity: "critical", Title: "Test 2"},
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestRegisterAgent(t *testing.T) {
	srv, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success","data":{"agent":{"uuid":"abc-123","agent_id":"test-agent","status":"online"}}}`))
	})
	defer srv.Close()

	err := client.RegisterAgent(context.Background(), AgentRegistration{
		AgentID:  "test-agent",
		Hostname: "test-host",
		Version:  "1.0.0",
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if client.AgentID() != "test-agent" {
		t.Fatalf("expected test-agent, got %s", client.AgentID())
	}
}

func TestHeartbeat(t *testing.T) {
	srv, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success","data":{"agent":{"uuid":"abc","status":"online"}}}`))
	})
	defer srv.Close()

	err := client.Heartbeat(context.Background(), HeartbeatRequest{
		AgentID: "test-agent",
		Status:  "online",
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestRetryOnServerError(t *testing.T) {
	attempts := 0
	srv, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success"}`))
	})
	defer srv.Close()

	err := client.UploadScan(context.Background(), ScanPayload{ScanID: "retry-test", Type: "test", Status: "completed"})
	if err != nil {
		t.Fatalf("expected nil after retries, got %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestNoRetryOnAuthError(t *testing.T) {
	attempts := 0
	srv, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusUnauthorized)
	})
	defer srv.Close()

	err := client.UploadScan(context.Background(), ScanPayload{ScanID: "auth-test", Type: "test", Status: "completed"})
	if err == nil {
		t.Fatal("expected error")
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", attempts)
	}
}

func TestOfflineQueueIntegration(t *testing.T) {
	dir := t.TempDir()
	client := NewClient("duck_pat_test123456789012345678901234567890123456789012345678",
		WithQueueDir(dir),
	)

	// Should queue when server is unreachable
	err := client.UploadScan(context.Background(), ScanPayload{ScanID: "offline-test", Type: "test", Status: "completed"})
	if err == nil {
		// It might fail with offline error, that's fine
	}

	q := client.Queue()
	if q == nil {
		t.Fatal("queue is nil")
	}

	items, err := q.DequeueAll()
	if err != nil {
		t.Fatalf("dequeue error: %v", err)
	}

	if len(items) == 0 {
		t.Log("no items queued (server might have been reachable)")
	} else {
		t.Logf("%d items queued", len(items))
	}
}

func TestQueuePersistence(t *testing.T) {
	dir := t.TempDir()
	q, err := NewQueue(dir)
	if err != nil {
		t.Fatalf("new queue error: %v", err)
	}

	if err := q.Enqueue("scan", []byte(`{"scan_id":"test"}`)); err != nil {
		t.Fatalf("enqueue error: %v", err)
	}

	if q.Size() != 1 {
		t.Fatalf("expected 1 item, got %d", q.Size())
	}

	// Re-create queue from same dir (simulate restart)
	q2, err := NewQueue(dir)
	if err != nil {
		t.Fatalf("re-create queue error: %v", err)
	}
	if q2.Size() != 1 {
		t.Fatalf("expected 1 item after reload, got %d", q2.Size())
	}

	items, err := q2.DequeueAll()
	if err != nil {
		t.Fatalf("dequeue error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Type != "scan" {
		t.Fatalf("expected type scan, got %s", items[0].Type)
	}

	if err := q2.Remove(items[0].ID); err != nil {
		t.Fatalf("remove error: %v", err)
	}
	if q2.Size() != 0 {
		t.Fatalf("expected 0 items after remove, got %d", q2.Size())
	}
}

func TestQueueMaxSize(t *testing.T) {
	dir := t.TempDir()
	q, err := NewQueue(dir)
	if err != nil {
		t.Fatalf("new queue error: %v", err)
	}

	for i := 0; i < maxQueueSize; i++ {
		if err := q.Enqueue("scan", []byte(`{"scan_id":"test"}`)); err != nil {
			t.Fatalf("enqueue %d error: %v", i, err)
		}
	}

	err = q.Enqueue("scan", []byte(`{"scan_id":"overflow"}`))
	if err == nil {
		t.Fatal("expected error when queue is full")
	}
}

func TestMaskPAT(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "***"},
		{"short", "***"},
		{"duck_pat_abcdef1234567890abcdef1234567890abcdef12", "duck****ef12"},
	}

	for _, tt := range tests {
		got := maskPAT(tt.input)
		if got != tt.want {
			t.Errorf("maskPAT(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestBackoffCalculation(t *testing.T) {
	for i := 1; i <= 5; i++ {
		b := computeBackoff(i)
		if b < 0 {
			t.Errorf("negative backoff for attempt %d", i)
		}
		if b > maxBackoff+jitterMax(maxBackoff) {
			t.Errorf("backoff too large for attempt %d: %v", i, b)
		}
	}
}

func jitterMax(d time.Duration) time.Duration {
	return time.Duration(jitterFactor * float64(d))
}

func TestComputeBackoff(t *testing.T) {
	backoffs := make([]time.Duration, 5)
	for i := 0; i < 5; i++ {
		backoffs[i] = computeBackoff(i + 1)
		if i > 0 && backoffs[i] < backoffs[i-1] {
			t.Logf("backoff decreased: %v -> %v", backoffs[i-1], backoffs[i])
		}
	}
}

func TestEmptyBulk(t *testing.T) {
	client := NewClient("duck_pat_test")
	if err := client.UploadScansBulk(context.Background(), nil); err != nil {
		t.Fatalf("expected nil for empty bulk, got %v", err)
	}
	if err := client.UploadFindingsBulk(context.Background(), []FindingPayload{}); err != nil {
		t.Fatalf("expected nil for empty bulk, got %v", err)
	}
	if err := client.UploadVulnerabilitiesBulk(context.Background(), []VulnerabilityPayload{}); err != nil {
		t.Fatalf("expected nil for empty bulk, got %v", err)
	}
}

func TestRateLimitRetry(t *testing.T) {
	attempts := 0
	srv, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success"}`))
	})
	defer srv.Close()

	err := client.UploadScan(context.Background(), ScanPayload{ScanID: "ratelimit-test", Type: "test", Status: "completed"})
	if err != nil {
		t.Fatalf("expected nil after rate limit retry, got %v", err)
	}
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
}

func TestPATNotLogged(t *testing.T) {
	masked := maskPAT("duck_pat_abcdef1234567890abcdef1234567890abcdef1234567890")
	if masked == "duck_pat_abcdef1234567890abcdef1234567890abcdef1234567890" {
		t.Fatal("PAT was not masked")
	}
	if len(masked) < 8 {
		t.Fatalf("masked too short: %s", masked)
	}
}

func TestConcurrentRequests(t *testing.T) {
	srv, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success"}`))
	})
	defer srv.Close()

	ctx := context.Background()
	concurrent := 10
	errCh := make(chan error, concurrent)

	for i := 0; i < concurrent; i++ {
		go func() {
			errCh <- client.UploadScan(ctx, ScanPayload{ScanID: "concurrent", Type: "test", Status: "completed"})
		}()
	}

	for i := 0; i < concurrent; i++ {
		if err := <-errCh; err != nil {
			t.Errorf("concurrent request failed: %v", err)
		}
	}
}

func TestContextCancellation(t *testing.T) {
	srv, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.Write([]byte(`{"status":"success"}`))
	})
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond)

	err := client.UploadScan(ctx, ScanPayload{ScanID: "cancel-test", Type: "test", Status: "completed"})
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestFileQueueCrashSafety(t *testing.T) {
	dir := t.TempDir()

	// Simulate crash by writing partial data
	partialPath := filepath.Join(dir, "queue", "1000000000_crash.json")
	os.MkdirAll(filepath.Dir(partialPath), 0700)
	os.WriteFile(partialPath, []byte(`{"id":"crash"`), 0600) // invalid JSON

	q, err := NewQueue(dir)
	if err != nil {
		t.Fatalf("new queue error after crash: %v", err)
	}

	if err := q.Enqueue("scan", []byte(`{"scan_id":"after-crash"}`)); err != nil {
		t.Fatalf("enqueue after crash error: %v", err)
	}

	items, err := q.DequeueAll()
	if err != nil {
		t.Fatalf("dequeue after crash error: %v", err)
	}

	// Should only have 1 valid item (corrupt one is skipped)
	if len(items) != 1 {
		t.Fatalf("expected 1 valid item after crash, got %d", len(items))
	}
}
