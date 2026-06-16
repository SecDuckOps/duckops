package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type ScannerTool string

const (
	ScannerSemgrep        ScannerTool = "semgrep"
	ScannerTrivy          ScannerTool = "trivy"
	ScannerGosec          ScannerTool = "gosec"
	ScannerBandit         ScannerTool = "bandit"
	ScannerGitleaks       ScannerTool = "gitleaks"
	ScannerTrufflehog     ScannerTool = "trufflehog"
	ScannerSemgrepMapped  ScannerTool = "sast__semgrep"
	ScannerTrivyMapped    ScannerTool = "sca__trivy_fs"
	ScannerGosecMapped    ScannerTool = "sast__gosec"
	ScannerBanditMapped   ScannerTool = "sast__bandit"
	ScannerGitleaksMapped ScannerTool = "secrets__gitleaks"
)

type ScanRunConfig struct {
	Type        string
	Environment string
	WorkspaceID string
	ProjectID   string
	TargetPath  string
	Tools       []ScannerTool
	Timeout     time.Duration
}

type ScannerRunner struct {
	client *Client
	config ScanRunConfig
	logf   func(string, ...any)
}

func NewScannerRunner(client *Client, config ScanRunConfig) *ScannerRunner {
	return &ScannerRunner{
		client: client,
		config: config,
		logf:   func(format string, args ...any) { client.log(format, args...) },
	}
}

func (r *ScannerRunner) RunAll(ctx context.Context) error {
	for _, tool := range r.config.Tools {
		if err := r.RunTool(ctx, tool); err != nil {
			r.logf("scanner %s failed: %v", tool, err)
		}
	}
	return nil
}

func (r *ScannerRunner) RunTool(ctx context.Context, tool ScannerTool) error {
	r.logf("starting scan with %s", tool)

	triggerResp, err := r.client.TriggerScan(ctx, TriggerScanPayload{
		Type:        r.config.Type,
		Environment: r.config.Environment,
		WorkspaceID: r.config.WorkspaceID,
		ProjectID:   r.config.ProjectID,
	})
	if err != nil {
		return fmt.Errorf("trigger scan failed: %w", err)
	}

	scanID := triggerResp.Data.Scan.UUID
	scanUUID := triggerResp.Data.Scan.UUID
	r.logf("scan triggered: %s (uuid=%s)", triggerResp.Data.Scan.ScanID, scanUUID)

	raw, scanErr := r.runLocalScanner(ctx, tool)

	if scanErr != nil {
		r.logf("scanner %s execution failed: %v", tool, scanErr)
		r.client.UpdateScan(ctx, scanID, UpdateScanPayload{
			Status:      "failed",
			CompletedAt: time.Now().UTC().Format(time.RFC3339),
		})
		return scanErr
	}

	results, normErr := NormalizeAll(string(tool), raw)
	if normErr != nil {
		r.logf("normalization failed for %s: %v", tool, normErr)
		r.client.UpdateScan(ctx, scanID, UpdateScanPayload{
			Status:      "failed",
			CompletedAt: time.Now().UTC().Format(time.RFC3339),
		})
		return normErr
	}

	r.logf("normalized %d findings from %s", len(results), tool)

	var findings []FindingPayload
	var vulns []VulnerabilityPayload
	for _, res := range results {
		findings = append(findings, res.ToFindingPayload())
		vulns = append(vulns, res.ToVulnerabilityPayload(scanID))
	}

	if err := r.client.UploadScanResults(ctx, scanID, UploadResultsPayload{
		Findings:       findings,
		Vulnerabilities: vulns,
	}); err != nil {
		r.logf("failed to upload results for %s: %v", tool, err)
	}

	r.client.UpdateScan(ctx, scanID, UpdateScanPayload{
		Status:      "completed",
		CompletedAt: time.Now().UTC().Format(time.RFC3339),
	})

	r.logf("scan %s completed: %d findings, %d vulns", tool, len(findings), len(vulns))
	return nil
}

func (r *ScannerRunner) runLocalScanner(ctx context.Context, tool ScannerTool) (json.RawMessage, error) {
	cmd, err := r.buildCommand(tool)
	if err != nil {
		return nil, err
	}

	r.logf("executing: %s", strings.Join(cmd.Args, " "))
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("scanner exited with code %d: %s", exitErr.ExitCode(), string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to run %s: %w", tool, err)
	}

	return json.RawMessage(output), nil
}

func (r *ScannerRunner) buildCommand(tool ScannerTool) (*exec.Cmd, error) {
	switch tool {
	case ScannerSemgrep, ScannerSemgrepMapped:
		return exec.Command("semgrep", "--json", r.config.TargetPath), nil
	case ScannerTrivy, ScannerTrivyMapped:
		return exec.Command("trivy", "fs", "--format", "json", r.config.TargetPath), nil
	case ScannerGosec, ScannerGosecMapped:
		return exec.Command("gosec", "-fmt=json", r.config.TargetPath), nil
	case ScannerBandit, ScannerBanditMapped:
		return exec.Command("bandit", "-f", "json", r.config.TargetPath), nil
	case ScannerGitleaks, ScannerGitleaksMapped:
		return exec.Command("gitleaks", "detect", "--report-format", "json", "--report-path", "/dev/stdout", "--source", r.config.TargetPath), nil
	default:
		return nil, fmt.Errorf("unsupported scanner tool: %s", tool)
	}
}
