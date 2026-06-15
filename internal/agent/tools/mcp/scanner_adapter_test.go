package mcp

import (
	"context"
	"encoding/json"
	"testing"

	agentmcp "github.com/SecDuckOps/agent/internal/domain/mcp"
	"github.com/SecDuckOps/shared/scanner/domain"
)

type stubMCPClient struct {
	call      agentmcp.ToolCall
	tools     []agentmcp.ToolInfo
	callError error
}

func (s *stubMCPClient) CallTool(_ context.Context, call agentmcp.ToolCall) (agentmcp.ToolResult, error) {
	s.call = call
	payload, _ := json.Marshal(domain.ScanResult{
		ScannerName: "semgrep",
		Target:      "target",
		Findings:    []domain.Finding{{ID: "1"}},
	})
	return agentmcp.ToolResult{Content: string(payload)}, s.callError
}

func (s *stubMCPClient) ListTools(context.Context) ([]agentmcp.ToolInfo, error) {
	return s.tools, nil
}

func (s *stubMCPClient) ListServerTools(context.Context, string) ([]agentmcp.ToolInfo, error) {
	return s.tools, nil
}

func (s *stubMCPClient) IsConnected(string) bool                      { return true }
func (s *stubMCPClient) ConnectedServers() []string                   { return []string{"scanner"} }
func (s *stubMCPClient) ConnectionError(string) error                 { return nil }
func (s *stubMCPClient) HealthCheck(context.Context) map[string]error { return map[string]error{} }
func (s *stubMCPClient) Close() error                                 { return nil }

func TestScannerAdapterRunScanMapsCanonicalScannerToMCPTool(t *testing.T) {
	client := &stubMCPClient{}
	adapter := NewScannerAdapter(client, "scanner")

	result, err := adapter.RunScan(context.Background(), "/workspace", "trivy")
	if err != nil {
		t.Fatalf("RunScan() error = %v", err)
	}

	if client.call.ToolName != "sca__trivy_fs" {
		t.Fatalf("expected tool sca__trivy_fs, got %q", client.call.ToolName)
	}

	if got := client.call.Arguments["path"]; got != "/workspace" {
		t.Fatalf("expected path argument to be /workspace, got %#v", got)
	}

	if len(result.Findings) != 1 {
		t.Fatalf("expected parsed findings, got %#v", result.Findings)
	}
}

func TestScannerAdapterAvailableScannersMapsServerTools(t *testing.T) {
	client := &stubMCPClient{
		tools: []agentmcp.ToolInfo{
			{Name: "sast__semgrep"},
			{Name: "sca__trivy_fs"},
			{Name: "secrets__gitleaks"},
			{Name: "ignored__tool"},
		},
	}
	adapter := NewScannerAdapter(client, "scanner")

	got := adapter.AvailableScanners()
	want := []string{"gitleaks", "semgrep", "trivy"}
	if len(got) != len(want) {
		t.Fatalf("expected %d scanners, got %d (%v)", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected scanner %q at index %d, got %q", want[i], i, got[i])
		}
	}
}
