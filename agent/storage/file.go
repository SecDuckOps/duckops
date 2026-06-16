package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type FileStorage struct {
	path string
}

func NewFileStorage(dataDir string) *FileStorage {
	return &FileStorage{
		path: filepath.Join(dataDir, "identity.json"),
	}
}

func (s *FileStorage) LoadIdentity(_ context.Context) (*StoredIdentity, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read identity file: %w", err)
	}

	var id StoredIdentity
	if err := json.Unmarshal(data, &id); err != nil {
		return nil, fmt.Errorf("parse identity file: %w", err)
	}

	return &id, nil
}

func (s *FileStorage) SaveIdentity(_ context.Context, id *StoredIdentity) error {
	data, err := json.MarshalIndent(id, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal identity: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0o600); err != nil {
		return fmt.Errorf("write identity file: %w", err)
	}

	return nil
}

func (s *FileStorage) ClearIdentity(_ context.Context) error {
	if err := os.Remove(s.path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
