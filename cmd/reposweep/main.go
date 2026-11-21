package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/ByteScribble/RepoSweep/internal/events"
	"github.com/ByteScribble/RepoSweep/internal/server"
	"github.com/ByteScribble/RepoSweep/internal/watcher"
)

const version = "0.1.0-dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "reposweep:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return watchCommand(nil)
	}
	switch args[0] {
	case "watch":
		return watchCommand(args[1:])
	case "emit":
		return emitCommand(args[1:])
	case "version", "--version", "-v":
		fmt.Println("reposweep", version)
		return nil
	case "help", "--help", "-h":
		printHelp()
		return nil
	default:
		if strings.HasPrefix(args[0], "-") {
			return watchCommand(args)
		}
		return watchCommand(args)
	}
}

func printHelp() {
	fmt.Print(`RepoSweep — live, local repository intelligence

Usage:
  reposweep watch [path] [--addr 127.0.0.1:7331] [--open=false]
  reposweep emit --agent codex --kind intent --message "Refactor auth flow"
  reposweep version

Run "reposweep watch ." and open http://127.0.0.1:7331.
`)
}

func watchCommand(args []string) error {
	fs := flag.NewFlagSet("watch", flag.ContinueOnError)
	addr := fs.String("addr", "127.0.0.1:7331", "dashboard listen address")
	quiet := fs.Bool("quiet", false, "suppress startup output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root := "."
	if fs.NArg() > 0 {
		root = fs.Arg(0)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve repository path: %w", err)
	}
	info, err := os.Stat(absRoot)
	if err != nil {
		return fmt.Errorf("inspect repository path: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", absRoot)
	}

	hub := events.NewHub(500)
	w, err := watcher.New(absRoot, hub)
	if err != nil {
		return err
	}
	defer w.Close()

	srv := server.New(absRoot, hub)
	httpServer := &http.Server{
		Addr:              *addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go w.Run(ctx)
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()

	if !*quiet {
		fmt.Printf("RepoSweep is watching %s\n", absRoot)
		fmt.Printf("Dashboard: http://%s\n", *addr)
	}
	err = httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func emitCommand(args []string) error {
	fs := flag.NewFlagSet("emit", flag.ContinueOnError)
	addr := fs.String("addr", "http://127.0.0.1:7331", "RepoSweep server URL")
	agent := fs.String("agent", "unknown-agent", "agent name")
	kind := fs.String("kind", "activity", "activity kind")
	message := fs.String("message", "", "human-readable activity")
	path := fs.String("file", "", "related repository path")
	status := fs.String("status", "active", "active, passed, failed, or done")
	if err := fs.Parse(args); err != nil {
		return err
	}
	payload := map[string]string{
		"agent": *agent, "kind": *kind, "message": *message, "path": *path, "status": *status,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := http.Post(strings.TrimRight(*addr, "/")+"/api/events", "application/json", strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("send event: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("server returned %s", resp.Status)
	}
	return nil
}
