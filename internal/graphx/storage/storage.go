package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

type Database struct {
	path string
}

func NewDatabase(path string) (*Database, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	db := &Database{
		path: path,
	}

	if err := db.init(); err != nil {
		return nil, fmt.Errorf("failed to init database: %w", err)
	}

	return db, nil
}

func (d *Database) init() error {
	return nil
}

func (d *Database) Close() error {
	return nil
}

type Stmt struct{}

func (s *Stmt) Scan(dest ...interface{}) error { return nil }
func (s *Stmt) Next() bool { return false }
func (s *Stmt) Finalize() error { return nil }

func (d *Database) Query(sql string, args ...interface{}) (*Stmt, error) {
	return &Stmt{}, nil
}

func (d *Database) Exec(sql string, args ...interface{}) (int64, int64, error) {
	return 0, 0, nil
}

func RunInTransaction(ctx interface{}, db *Database, fn func(tx *Database) error) error {
	return nil
}