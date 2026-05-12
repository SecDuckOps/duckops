package incremental

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/SecDuckOps/duckops/internal/graphx/storage"
)

type Cache struct {
	mu       sync.RWMutex
	Files    map[string]FileInfo `json:"files"`
	RepoID   string              `json:"repo_id"`
	Path     string               `json:"path"`
	Modified int64                `json:"modified"`
}

type FileInfo struct {
	Path        string `json:"path"`
	ContentHash string `json:"content_hash"`
	Size        int64  `json:"size"`
	ModTime     int64  `json:"mod_time"`
	IndexedAt   int64  `json:"indexed_at"`
}

func NewCache(repoID, path string) *Cache {
	return &Cache{
		Files:  make(map[string]FileInfo),
		RepoID: repoID,
		Path:   path,
	}
}

func (c *Cache) Load() error {
	data, err := os.ReadFile(c.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(data, c)
}

func (c *Cache) Save() error {
	c.Modified = time.Now().Unix()
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.Path, data, 0644)
}

func (c *Cache) GetFileHash(path string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Files[path].ContentHash
}

func (c *Cache) HasChanged(path, currentHash string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	info, exists := c.Files[path]
	if !exists {
		return true
	}
	return info.ContentHash != currentHash
}

func (c *Cache) UpdateFile(path, hash string, size int64, modTime int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Files[path] = FileInfo{
		Path:        path,
		ContentHash: hash,
		Size:        size,
		ModTime:     modTime,
		IndexedAt:   time.Now().Unix(),
	}
}

func (c *Cache) RemoveFile(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.Files, path)
}

func (c *Cache) GetChangedFiles(currentFiles map[string]int64) ([]string, []string, []string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var added []string
	var modified []string
	var deleted []string

	currentSet := make(map[string]bool)
	for path := range currentFiles {
		currentSet[path] = true
		if _, exists := c.Files[path]; !exists {
			added = append(added, path)
		}
	}

	for path := range c.Files {
		if !currentSet[path] {
			deleted = append(deleted, path)
		}
	}

	return added, modified, deleted
}

type Engine struct {
	cache *Cache
	store *storage.JSONStore
}

func NewIncrementalEngine(store *storage.JSONStore) *Engine {
	return &Engine{
		cache: nil,
		store: store,
	}
}

func hashFile(path string) (string, int64, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", 0, err
	}
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:]), int64(len(content)), nil
}

func (e *Engine) ProcessFile(path string) error {
	hash, size, err := hashFile(path)
	if err != nil {
		return err
	}

	if !e.cache.HasChanged(path, hash) {
		return nil
	}

	stat, err := os.Stat(path)
	if err != nil {
		return err
	}

	e.cache.UpdateFile(path, hash, size, stat.ModTime().Unix())
	return nil
}

func (e *Engine) MarkDeleted(path string) error {
	e.cache.RemoveFile(path)
	return nil
}

func (e *Engine) Save() error {
	return e.cache.Save()
}

func (e *Engine) GetFilesToReprocess(rootPath string) ([]string, error) {
	var files []string

	extensions := map[string]bool{
		".py": true, ".go": true, ".ts": true, ".tsx": true,
		".js": true, ".jsx": true, ".rs": true,
	}

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		ext := filepath.Ext(path)
		if !extensions[ext] {
			return nil
		}

		hash, _, err := hashFile(path)
		if err != nil {
			return nil
		}

		if e.cache.HasChanged(path, hash) {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}