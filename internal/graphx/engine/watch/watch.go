package watch

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/SecDuckOps/duckops/internal/graphx/config"
	"github.com/SecDuckOps/duckops/internal/graphx/engine"
)

type Watcher struct {
	cfg      *config.Config
	engine   *engine.Engine
	mu       sync.RWMutex
	 running bool
	stopCh   chan struct{}
	events   chan Event
}

type Event struct {
	Type    EventType
	Path    string
	Time    time.Time
}

type EventType int

const (
	EventCreate EventType = iota
	EventModify
	EventDelete
)

func NewWatcher(cfg *config.Config, eng *engine.Engine) *Watcher {
	return &Watcher{
		cfg:    cfg,
		engine: eng,
		stopCh: make(chan struct{}),
		events: make(chan Event, 100),
	}
}

func (w *Watcher) Start() error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return fmt.Errorf("watcher already running")
	}
	w.running = true
	w.mu.Unlock()

	go w.runLoop()
	return nil
}

func (w *Watcher) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return fmt.Errorf("watcher not running")
	}

	close(w.stopCh)
	w.running = false
	return nil
}

func (w *Watcher) runLoop() {
	debounce := time.Duration(w.cfg.Watch.DebounceMs) * time.Millisecond
	var pending map[string]EventType
	var timer *time.Timer

	extensions := map[string]bool{
		".py": true, ".go": true, ".ts": true, ".tsx": true,
		".js": true, ".jsx": true, ".rs": true,
	}

	processEvents := func() {
		for path, et := range pending {
			ext := filepath.Ext(path)
			if !extensions[ext] {
				continue
			}

			switch et {
			case EventCreate, EventModify:
				fmt.Printf("GraphX: Detected change in %s, re-processing...\n", path)
				if err := w.processFile(path); err != nil {
					fmt.Printf("GraphX: Error processing %s: %v\n", path, err)
				}
			case EventDelete:
				fmt.Printf("GraphX: Detected deletion of %s\n", path)
			}
		}
		pending = nil
	}

	for {
		select {
		case <-w.stopCh:
			if timer != nil {
				timer.Stop()
			}
			return
		case ev := <-w.events:
			if pending == nil {
				pending = make(map[string]EventType)
			}
			pending[ev.Path] = ev.Type

			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(debounce, processEvents)
		}
	}
}

func (w *Watcher) processFile(path string) error {
	return nil
}

func (w *Watcher) WatchDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			if err := w.WatchDir(path); err != nil {
				continue
			}
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		w.events <- Event{
			Type: EventCreate,
			Path: path,
			Time: info.ModTime(),
		}
	}

	return nil
}

type Poller struct {
	cfg    *config.Config
	engine *engine.Engine
	stopCh chan struct{}
}

func NewPoller(cfg *config.Config, eng *engine.Engine) *Poller {
	return &Poller{
		cfg:    cfg,
		engine: eng,
		stopCh: make(chan struct{}),
	}
}

func (p *Poller) Start(interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopCh:
			return nil
		case <-ticker.C:
			if err := p.poll(); err != nil {
				fmt.Printf("GraphX: Poll error: %v\n", err)
			}
		}
	}
}

func (p *Poller) Stop() error {
	close(p.stopCh)
	return nil
}

func (p *Poller) poll() error {
	_ = p.cfg
	return nil
}