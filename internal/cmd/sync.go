package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/SecDuckOps/duckops/internal/config"
	"github.com/SecDuckOps/duckops/internal/platform"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize local data with DuckOps Platform",
	Long: `Upload local scan results, pipeline events, and agent status 
to the DuckOps Platform server.

Requires authentication via 'duckops login' first.`,
	RunE: syncRun,
}

var (
	syncScanResults string
	syncTool        string
	syncWatch       bool
)

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().StringVarP(&syncScanResults, "results", "r", "", "Path to scan results JSON file")
	syncCmd.Flags().StringVarP(&syncTool, "tool", "t", "", "Scanner tool name (semgrep, trivy, gosec, bandit, gitleaks, trufflehog)")
	syncCmd.Flags().BoolVarP(&syncWatch, "watch", "w", false, "Watch queue directory and sync continuously")
}

func syncRun(cmd *cobra.Command, _ []string) error {
	client, err := platformClient()
	if err != nil {
		return fmt.Errorf("failed to initialize platform client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.VerifyPAT(ctx); err != nil {
		return fmt.Errorf("PAT verification failed: %w\nRun 'duckops login' to authenticate.", err)
	}

	fmt.Println("✓ PAT verified")

	if syncScanResults != "" {
		return syncFileResults(client, syncScanResults, syncTool)
	}

	if syncWatch {
		return watchQueue(client)
	}

	// Default: show sync status
	q := client.Queue()
	if q != nil {
		size := q.Size()
		if size > 0 {
			fmt.Printf("📦 %d items in offline queue\n", size)
		} else {
			fmt.Println("✓ Queue empty")
		}
	}

	return syncAll(client)
}

func syncAll(client *platform.Client) error {
	ctx := context.Background()

	info, err := client.GetServerInfo(ctx)
	if err != nil {
		return fmt.Errorf("cannot reach server: %w", err)
	}
	fmt.Printf("✓ Connected to server v%s\n", info.Data.Version)

	q := client.Queue()
	if q == nil {
		return nil
	}

	items, err := q.DequeueAll()
	if err != nil {
		return fmt.Errorf("failed to read queue: %w", err)
	}

	if len(items) == 0 {
		fmt.Println("✓ No queued items to sync")
		return nil
	}

	fmt.Printf("Syncing %d queued items...\n", len(items))
	for _, item := range items {
		if err := replayItem(client, item); err != nil {
			fmt.Printf("✗ %s: %v\n", item.Type, err)
			continue
		}
		q.Remove(item.ID)
		fmt.Printf("✓ %s synced\n", item.Type)
	}

	return nil
}

func syncFileResults(client *platform.Client, path, tool string) error {
	if tool == "" {
		return errors.New("--tool is required when uploading scan results")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	raw := json.RawMessage(data)
	results, err := platform.NormalizeAll(tool, raw)
	if err != nil {
		return fmt.Errorf("normalization failed: %w", err)
	}

	fmt.Printf("Normalized %d findings from %s\n", len(results), tool)

	ctx := context.Background()
	var findings []platform.FindingPayload
	for _, r := range results {
		findings = append(findings, r.ToFindingPayload())
	}

	if err := client.UploadFindingsBulk(ctx, findings); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	fmt.Printf("✓ Uploaded %d findings\n", len(findings))
	return nil
}

func watchQueue(client *platform.Client) error {
	fmt.Println("Watching queue for new items (Ctrl+C to stop)...")
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		q := client.Queue()
		if q == nil {
			continue
		}

		items, err := q.DequeueAll()
		if err != nil {
			fmt.Fprintf(os.Stderr, "queue read error: %v\n", err)
			continue
		}

		for _, item := range items {
			if err := replayItem(client, item); err != nil {
				fmt.Fprintf(os.Stderr, "✗ %s: %v\n", item.Type, err)
				continue
			}
			q.Remove(item.ID)
			fmt.Printf("✓ %s replayed\n", item.Type)
		}
	}
	return nil
}

func replayItem(client *platform.Client, item platform.QueueItem) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch item.Type {
	case "scan":
		var p platform.ScanPayload
		if err := json.Unmarshal(item.Payload, &p); err != nil {
			return err
		}
		return client.UploadScan(ctx, p)
	case "bulk_scans":
		var p []platform.ScanPayload
		if err := json.Unmarshal(item.Payload, &p); err != nil {
			return err
		}
		return client.UploadScansBulk(ctx, p)
	case "vulnerability":
		var p platform.VulnerabilityPayload
		if err := json.Unmarshal(item.Payload, &p); err != nil {
			return err
		}
		return client.UploadVulnerability(ctx, p)
	case "bulk_vulnerabilities":
		var p []platform.VulnerabilityPayload
		if err := json.Unmarshal(item.Payload, &p); err != nil {
			return err
		}
		return client.UploadVulnerabilitiesBulk(ctx, p)
	case "finding":
		var p platform.FindingPayload
		if err := json.Unmarshal(item.Payload, &p); err != nil {
			return err
		}
		return client.UploadFinding(ctx, p)
	case "bulk_findings":
		var p []platform.FindingPayload
		if err := json.Unmarshal(item.Payload, &p); err != nil {
			return err
		}
		return client.UploadFindingsBulk(ctx, p)
	case "pipeline_event":
		var p platform.PipelineEventPayload
		if err := json.Unmarshal(item.Payload, &p); err != nil {
			return err
		}
		return client.UploadPipelineEvent(ctx, p)
	case "session":
		var p platform.SessionPayload
		if err := json.Unmarshal(item.Payload, &p); err != nil {
			return err
		}
		return client.UploadSession(ctx, p)
	case "bulk_sessions":
		var p []platform.SessionPayload
		if err := json.Unmarshal(item.Payload, &p); err != nil {
			return err
		}
		return client.UploadSessionsBulk(ctx, p)
	default:
		return fmt.Errorf("unknown type: %s", item.Type)
	}
}

func platformClient() (*platform.Client, error) {
	apiKey, err := duckopsAPIKey()
	if err != nil {
		return nil, err
	}

	duckopsDir := filepath.Dir(config.GlobalConfig())

	client := platform.NewClient(apiKey,
		platform.WithLogger(func(format string, args ...any) {
			fmt.Fprintf(os.Stderr, format+"\n", args...)
		}),
		platform.WithQueueDir(duckopsDir),
	)

	return client, nil
}

func duckopsAPIKey() (string, error) {
	data, err := os.ReadFile(config.GlobalConfigData())
	if err != nil {
		return "", errors.New("not logged in to DuckOps Platform\nRun 'duckops login' first")
	}

	apiKey := gjson.Get(string(data), "duckops_api_key").String()
	if apiKey == "" {
		apiKey = gjson.Get(string(data), "providers.duckops.api_key").String()
	}
	if apiKey == "" {
		return "", errors.New("not logged in to DuckOps Platform\nRun 'duckops login' first")
	}

	return apiKey, nil
}


