package watcher

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Hrtkk/reposweep/internal/events"
)

func TestWatcherPublishesCreatedFile(t *testing.T) {
	root := t.TempDir()
	hub := events.NewHub(20)
	w, err := New(root, hub)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go w.Run(ctx)

	path := filepath.Join(root, "handler.go")
	if err := os.WriteFile(path, []byte("package example\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		for _, event := range hub.Snapshot() {
			if event.Path == "handler.go" && event.Kind == "change" {
				return
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("created file event was not published: %#v", hub.Snapshot())
}
