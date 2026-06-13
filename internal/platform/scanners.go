package platform

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type ScannerResult struct {
	Tool        string `json:"tool"`
	Severity    string `json:"severity"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	FilePath    string `json:"file_path,omitempty"`
	LineStart   int    `json:"line_start,omitempty"`
	LineEnd     int    `json:"line_end,omitempty"`
	RuleID      string `json:"rule_id,omitempty"`
	CWE         string `json:"cwe,omitempty"`
	CVE         string `json:"cve,omitempty"`
	CVSS        float64 `json:"cvss,omitempty"`
	PackageName string `json:"package_name,omitempty"`
	PackageVer  string `json:"package_version,omitempty"`
	FixVersion  string `json:"fixed_version,omitempty"`
	Raw         any    `json:"-"`
}

type NormalizedFinding struct {
	FindingPayload
	Scanner  string
	Raw      any
}

func NormalizeSemgrepResult(raw json.RawMessage) ([]ScannerResult, error) {
	var report struct {
		Results []struct {
			CheckID string `json:"check_id"`
			Path    string `json:"path"`
			Start   struct {
				Line int `json:"line"`
			} `json:"start"`
			End struct {
				Line int `json:"line"`
			} `json:"end"`
			Extra struct {
				Severity   string `json:"severity"`
				Message    string `json:"message"`
				Metadata   map[string]any `json:"metadata"`
				Lines      string `json:"lines"`
			} `json:"extra"`
		} `json:"results"`
	}

	if err := json.Unmarshal(raw, &report); err != nil {
		return nil, fmt.Errorf("semgrep parse error: %w", err)
	}

	var results []ScannerResult
	for _, r := range report.Results {
		cwe := ""
		if r.Extra.Metadata != nil {
			if c, ok := r.Extra.Metadata["cwe"].(string); ok {
				cwe = c
			}
		}

		results = append(results, ScannerResult{
			Tool:        "semgrep",
			Severity:    mapSeverity(r.Extra.Severity),
			Title:       r.Extra.Message,
			Description: r.Extra.Lines,
			FilePath:    r.Path,
			LineStart:   r.Start.Line,
			LineEnd:     r.End.Line,
			RuleID:      r.CheckID,
			CWE:         cwe,
		})
	}
	return results, nil
}

func NormalizeTrivyResult(raw json.RawMessage) ([]ScannerResult, error) {
	var report struct {
		Results []struct {
			Target        string `json:"target"`
			Vulnerabilities []struct {
				VulnerabilityID  string  `json:"vulnerabilityID"`
				PkgName          string  `json:"pkgName"`
				InstalledVersion string  `json:"installedVersion"`
				FixedVersion     string  `json:"fixedVersion"`
				Severity         string  `json:"severity"`
				Title            string  `json:"title"`
				Description      string  `json:"description"`
				CVSS             map[string]struct {
					V3Score float64 `json:"v3Score"`
				} `json:"cvss"`
				CweIDs []string `json:"cweIDs"`
			} `json:"vulnerabilities"`
		} `json:"Results"`
	}

	if err := json.Unmarshal(raw, &report); err != nil {
		return nil, fmt.Errorf("trivy parse error: %w", err)
	}

	var results []ScannerResult
	for _, res := range report.Results {
		for _, v := range res.Vulnerabilities {
			cvss := 0.0
			if v.CVSS != nil {
				for _, s := range v.CVSS {
					if s.V3Score > cvss {
						cvss = s.V3Score
					}
				}
			}

			cwe := ""
			if len(v.CweIDs) > 0 {
				cwe = v.CweIDs[0]
			}

			results = append(results, ScannerResult{
				Tool:        "trivy",
				Severity:    mapSeverity(v.Severity),
				Title:       v.Title,
				Description: v.Description,
				RuleID:      v.VulnerabilityID,
				CVE:         v.VulnerabilityID,
				CWE:         cwe,
				CVSS:        cvss,
				PackageName: v.PkgName,
				PackageVer:  v.InstalledVersion,
				FixVersion:  v.FixedVersion,
			})
		}
	}
	return results, nil
}

func NormalizeGosecResult(raw json.RawMessage) ([]ScannerResult, error) {
	var report struct {
		Issues []struct {
			Severity   string `json:"severity"`
			Confidence string `json:"confidence"`
			CWE        struct {
				ID  string `json:"id"`
			} `json:"cwe"`
			RuleID string `json:"rule_id"`
			Details string `json:"details"`
			File    string `json:"file"`
			Line    string `json:"line"`
		} `json:"Issues"`
	}

	if err := json.Unmarshal(raw, &report); err != nil {
		return nil, fmt.Errorf("gosec parse error: %w", err)
	}

	var results []ScannerResult
	for _, issue := range report.Issues {
		line := 0
		fmt.Sscanf(issue.Line, "%d", &line)

		results = append(results, ScannerResult{
			Tool:        "gosec",
			Severity:    mapSeverity(issue.Severity),
			Title:       issue.Details,
			Description: issue.Details,
			FilePath:    issue.File,
			LineStart:   line,
			LineEnd:     line,
			RuleID:      issue.RuleID,
			CWE:         issue.CWE.ID,
		})
	}
	return results, nil
}

func NormalizeBanditResult(raw json.RawMessage) ([]ScannerResult, error) {
	var report struct {
		Results []struct {
			IssueText string `json:"issue_text"`
			Severity  string `json:"issue_severity"`
			Confidence string `json:"issue_confidence"`
			TestID    string `json:"test_id"`
			Filename  string `json:"filename"`
			LineNum   int    `json:"line_number"`
			Col       int    `json:"col_offset"`
			EndLine   int    `json:"end_line"`
			MoreInfo  string `json:"more_info"`
		} `json:"results"`
	}

	if err := json.Unmarshal(raw, &report); err != nil {
		return nil, fmt.Errorf("bandit parse error: %w", err)
	}

	var results []ScannerResult
	for _, r := range report.Results {
		results = append(results, ScannerResult{
			Tool:        "bandit",
			Severity:    mapSeverity(r.Severity),
			Title:       r.IssueText,
			Description: fmt.Sprintf("%s (confidence: %s)", r.IssueText, r.Confidence),
			FilePath:    r.Filename,
			LineStart:   r.LineNum,
			LineEnd:     r.EndLine,
			RuleID:      r.TestID,
		})
	}
	return results, nil
}

func NormalizeGitleaksResult(raw json.RawMessage) ([]ScannerResult, error) {
	var findings []struct {
		Description string `json:"description"`
		StartLine   int    `json:"startLine"`
		EndLine     int    `json:"endLine"`
		File        string `json:"file"`
		RuleID      string `json:"rule"`
		Severity    string `json:"severity"`
		Fingerprint string `json:"fingerprint"`
	}

	if err := json.Unmarshal(raw, &findings); err != nil {
		return nil, fmt.Errorf("gitleaks parse error: %w", err)
	}

	var results []ScannerResult
	for _, f := range findings {
		results = append(results, ScannerResult{
			Tool:        "gitleaks",
			Severity:    mapSeverity(f.Severity),
			Title:       fmt.Sprintf("Secret leaked: %s", f.Description),
			Description: f.Description,
			FilePath:    f.File,
			LineStart:   f.StartLine,
			LineEnd:     f.EndLine,
			RuleID:      f.RuleID,
		})
	}
	return results, nil
}

func NormalizeTrufflehogResult(raw json.RawMessage) ([]ScannerResult, error) {
	var results []ScannerResult

	var rawResults []json.RawMessage
	if err := json.Unmarshal(raw, &rawResults); err != nil {
		return nil, fmt.Errorf("trufflehog parse error: %w", err)
	}

	for _, r := range rawResults {
		var entry struct {
			SourceMetadata struct {
				Data struct {
					File string `json:"file"`
					Line int    `json:"line"`
				} `json:"data"`
			} `json:"SourceMetadata"`
			DetectorName string `json:"DetectorName"`
			Severity     int    `json:"Severity"`
			RawV2        string `json:"RawV2"`
		}

		if err := json.Unmarshal(r, &entry); err != nil {
			continue
		}

		sev := "medium"
		if entry.Severity >= 3 {
			sev = "critical"
		} else if entry.Severity == 2 {
			sev = "high"
		} else if entry.Severity <= 0 {
			sev = "low"
		}

		results = append(results, ScannerResult{
			Tool:        "trufflehog",
			Severity:    sev,
			Title:       fmt.Sprintf("Secret detected: %s", entry.DetectorName),
			Description: entry.RawV2,
			FilePath:    entry.SourceMetadata.Data.File,
			LineStart:   entry.SourceMetadata.Data.Line,
			RuleID:      entry.DetectorName,
		})
	}
	return results, nil
}

func NormalizeAll(tool string, raw json.RawMessage) ([]ScannerResult, error) {
	switch tool {
	case "sast__semgrep", "semgrep":
		return NormalizeSemgrepResult(raw)
	case "sca__trivy_fs", "trivy":
		return NormalizeTrivyResult(raw)
	case "sast__gosec", "gosec":
		return NormalizeGosecResult(raw)
	case "sast__bandit", "bandit":
		return NormalizeBanditResult(raw)
	case "secrets__gitleaks", "gitleaks":
		return NormalizeGitleaksResult(raw)
	case "secrets__trufflehog", "trufflehog":
		return NormalizeTrufflehogResult(raw)
	default:
		return nil, fmt.Errorf("unknown scanner tool: %s", tool)
	}
}

func (sr ScannerResult) ToFindingPayload() FindingPayload {
	return FindingPayload{
		Tool:         sr.Tool,
		Severity:     sr.Severity,
		Title:        sr.Title,
		Description:  sr.Description,
		FilePath:     sr.FilePath,
		LineStart:    sr.LineStart,
		LineEnd:      sr.LineEnd,
		RuleID:       sr.RuleID,
		CWEID:        sr.CWE,
		CVEID:        sr.CVE,
		CVSSScore:    sr.CVSS,
		PackageName:  sr.PackageName,
		PackageVer:   sr.PackageVer,
		FixedVersion: sr.FixVersion,
		SourceTool:   sr.Tool,
		RawDetails:   sr.Raw,
	}
}

func (sr ScannerResult) ToVulnerabilityPayload(scanID string) VulnerabilityPayload {
	return VulnerabilityPayload{
		ScanID:       scanID,
		CVEID:        sr.CVE,
		Title:        sr.Title,
		Severity:     sr.Severity,
		CVSSScore:    sr.CVSS,
		PackageName:  sr.PackageName,
		PackageVer:   sr.PackageVer,
		FixedVersion: sr.FixVersion,
		SourceTool:   sr.Tool,
		Status:       "open",
		DetectedAt:   time.Now().UTC().Format(time.RFC3339),
	}
}

func mapSeverity(s string) string {
	switch strings.ToLower(s) {
	case "critical", "crítica", "crítico", "3":
		return "critical"
	case "high", "alta", "alto", "2":
		return "high"
	case "medium", "media", "medio", "1":
		return "medium"
	case "low", "baja", "baixo", "0":
		return "low"
	case "info", "informational", "informative":
		return "info"
	default:
		if s == "" {
			return "low"
		}
		return "medium"
	}
}
