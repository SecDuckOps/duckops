package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var autoScanCmd = &cobra.Command{
	Use:   "auto-scan",
	Short: "Run all bundled security scanners and aggregate results",
	RunE: func(cmd *cobra.Command, _ []string) error {
		port, token, ok := EnsureToolServer()
		if !ok {
			return fmt.Errorf("failed to start tool server")
		}

		client := &http.Client{Timeout: 30 * time.Second}
		payload := map[string]interface{}{
			"agent_id": "auto-scan",
			"path":     ".",
			"target":   ".",
		}
		data, _ := json.Marshal(payload)
		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:%d/scan", port), bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("failed to create scan request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := client.Do(req)
		if err != nil {
			slog.Error("scan request failed", "err", err)
			return err
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(autoScanCmd)
}
