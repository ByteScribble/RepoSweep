package watcher

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/ByteScribble/RepoSweep/internal/events"
	"github.com/ByteScribble/RepoSweep/internal/insight"
)

const debounceWindow = 140 * time.Millisecond

type Watcher struct {
	root    string
	hub     *events.Hub
	native  *fsnotify.Watcher
	mu      sync.Mutex
	pending map[string]*time.Timer
}

func New(root string, hub *events.Hub) (*Watcher, error) {
	native, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create native watcher: %w", err)
	}
	w := &Watcher{root: root, hub: hub, native: native, pending: make(map[string]*time.Timer)}
	if err := w.addTree(root); err != nil {
		native.Close()
		return nil, err
	}
	return w, nil
}

func (w *Watcher) Close() error {
	w.mu.Lock()
	for _, timer := range w.pending {
		timer.Stop()
	}
	w.pending = make(map[string]*time.Timer)
	w.mu.Unlock()
	return w.native.Close()
}

func (w *Watcher) Run(ctx context.Context) {
	w.hub.Publish(events.Event{
		Time: time.Now(), Kind: "system", Actor: "reposweep", Status: "active",
		Title: "Repository watcher online", Summary: "Native file notifications are active.",
		Risk: "none", Confidence: "observed", Evidence: []string{w.root},
	})
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-w.native.Events:
			if !ok {
				return
			}
			w.handle(event)
		case err, ok := <-w.native.Errors:
			if !ok {
				return
			}
			w.hub.Publish(events.Event{
				Time: time.Now(), Kind: "system", Actor: "reposweep", Status: "failed",
				Title: "Watcher error", Summary: err.Error(), Risk: "medium", Confidence: "observed",
				Evidence: []string{"fsnotify"},
			})
		}
	}
}

func (w *Watcher) addTree(root string) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			if path == root {
				return fmt.Errorf("scan repository: %w", err)
			}
			return nil
		}
		if !entry.IsDir() {
			return nil
		}
		if path != root && ignoredDir(entry.Name()) {
			return filepath.SkipDir
		}
		if err := w.native.Add(path); err != nil {
			return fmt.Errorf("watch %s: %w", path, err)
		}
		return nil
	})
}

func (w *Watcher) handle(event fsnotify.Event) {
	if w.ignored(event.Name) {
		return
	}
	if event.Has(fsnotify.Create) {
		if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
			_ = w.addTree(event.Name)
			return
		}
	}
	operation := operationName(event.Op)
	if operation == "" {
		return
	}
	relative, err := filepath.Rel(w.root, event.Name)
	if err != nil || strings.HasPrefix(relative, "..") {
		return
	}
	w.debounce(relative, operation)
}

func (w *Watcher) debounce(path, operation string) {
	w.mu.Lock()
	if timer, ok := w.pending[path]; ok {
		timer.Stop()
	}
	w.pending[path] = time.AfterFunc(debounceWindow, func() {
		event := insight.Classify(path, operation)
		event.Time = time.Now()
		w.hub.Publish(event)
		w.mu.Lock()
		delete(w.pending, path)
		w.mu.Unlock()
	})
	w.mu.Unlock()
}

func (w *Watcher) ignored(path string) bool {
	relative, err := filepath.Rel(w.root, path)
	if err != nil || strings.HasPrefix(relative, "..") {
		return true
	}
	for _, part := range strings.Split(filepath.ToSlash(relative), "/") {
		if ignoredDir(part) {
			return true
		}
	}
	return strings.HasSuffix(relative, ".swp") || strings.HasSuffix(relative, "~") || strings.HasSuffix(relative, ".tmp")
}

func ignoredDir(name string) bool {
	switch name {
	case ".git", ".hg", ".svn", "node_modules", "target", "dist", "build", ".next", ".venv", "vendor":
		return true
	default:
		return false
	}
}

func operationName(operation fsnotify.Op) string {
	switch {
	case operation.Has(fsnotify.Remove):
		return "removed"
	case operation.Has(fsnotify.Rename):
		return "renamed"
	case operation.Has(fsnotify.Create):
		return "created"
	case operation.Has(fsnotify.Write):
		return "modified"
	default:
		return ""
	}
}
