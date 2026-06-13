package platform

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	queueDirName = "queue"
	maxQueueSize = 10000
)

type Queue struct {
	dir   string
	mu    sync.Mutex
	count int
}

func NewQueue(baseDir string) (*Queue, error) {
	dir := filepath.Join(baseDir, queueDirName)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create queue dir %s: %w", dir, err)
	}

	q := &Queue{dir: dir}
	q.count = q.scanCount()
	return q, nil
}

func (q *Queue) Enqueue(itemType string, payload []byte) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.count >= maxQueueSize {
		return &PlatformError{Code: ErrQueueFull, Message: fmt.Sprintf("queue full (%d items)", maxQueueSize)}
	}

	id, err := newID()
	if err != nil {
		return err
	}

	item := QueueItem{
		ID:        id,
		Type:      itemType,
		Payload:   payload,
		CreatedAt: time.Now(),
		Retries:   0,
	}

	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal queue item: %w", err)
	}

	filename := fmt.Sprintf("%d_%s.json", time.Now().UnixNano(), id)
	path := filepath.Join(q.dir, filename)

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write queue file %s: %w", path, err)
	}

	q.count++
	return nil
}

func (q *Queue) DequeueAll() ([]QueueItem, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	entries, err := os.ReadDir(q.dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read queue dir: %w", err)
	}

	var items []QueueItem
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		path := filepath.Join(q.dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var item QueueItem
		if err := json.Unmarshal(data, &item); err != nil {
			continue
		}
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})

	return items, nil
}

func (q *Queue) Remove(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	entries, err := os.ReadDir(q.dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if strings.Contains(entry.Name(), id) {
			path := filepath.Join(q.dir, entry.Name())
			if err := os.Remove(path); err == nil {
				q.count--
			}
			return nil
		}
	}
	return nil
}

func (q *Queue) Clear() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	entries, err := os.ReadDir(q.dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			os.Remove(filepath.Join(q.dir, entry.Name()))
		}
	}
	q.count = 0
	return nil
}

func (q *Queue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.count
}

func (q *Queue) Dir() string {
	return q.dir
}

func (q *Queue) scanCount() int {
	entries, err := os.ReadDir(q.dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			count++
		}
	}
	return count
}

func newID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate ID: %w", err)
	}
	return hex.EncodeToString(b), nil
}
