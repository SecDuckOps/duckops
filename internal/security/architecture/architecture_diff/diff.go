package architecture_diff

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Snapshot struct {
	Hash      string    `json:"hash"`
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
}

type History struct {
	Snapshots []Snapshot `json:"snapshots"`
}

func Record(historyPath string, content []byte, snapshotPath string) (History, error) {
	sum := sha256.Sum256(content)
	hash := hex.EncodeToString(sum[:])
	history := History{}

	if data, err := os.ReadFile(historyPath); err == nil {
		_ = json.Unmarshal(data, &history)
	}
	if len(history.Snapshots) == 0 || history.Snapshots[len(history.Snapshots)-1].Hash != hash {
		history.Snapshots = append(history.Snapshots, Snapshot{
			Hash:      hash,
			Timestamp: time.Now().UTC(),
			Path:      snapshotPath,
		})
	}
	if err := os.MkdirAll(filepath.Dir(historyPath), 0o755); err != nil {
		return history, err
	}
	raw, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return history, err
	}
	return history, os.WriteFile(historyPath, raw, 0o644)
}
