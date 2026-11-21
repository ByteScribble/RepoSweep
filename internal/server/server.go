package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/ByteScribble/RepoSweep/internal/events"
	webassets "github.com/ByteScribble/RepoSweep/web"
)

type Server struct {
	root string
	hub  *events.Hub
}

func New(root string, hub *events.Hub) *Server {
	return &Server{root: root, hub: hub}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/state", s.state)
	mux.HandleFunc("GET /api/stream", s.stream)
	mux.HandleFunc("POST /api/events", s.ingest)
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.Handle("GET /", http.FileServer(http.FS(webassets.Files)))
	return securityHeaders(mux)
}

func (s *Server) state(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"repository": filepath.Base(s.root),
		"root":       s.root,
		"events":     s.hub.Snapshot(),
	})
}

func (s *Server) stream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	stream, unsubscribe := s.hub.Subscribe()
	defer unsubscribe()
	fmt.Fprint(w, ": connected\n\n")
	flusher.Flush()
	for {
		select {
		case <-r.Context().Done():
			return
		case event, ok := <-stream:
			if !ok {
				return
			}
			payload, err := json.Marshal(event)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "id: %d\ndata: %s\n\n", event.Sequence, payload)
			flusher.Flush()
		}
	}
}

func (s *Server) ingest(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	defer r.Body.Close()
	var incoming events.Incoming
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&incoming); err != nil {
		http.Error(w, "invalid event payload", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(incoming.Agent) == "" || strings.TrimSpace(incoming.Message) == "" {
		http.Error(w, "agent and message are required", http.StatusBadRequest)
		return
	}
	status := incoming.Status
	if status == "" {
		status = "active"
	}
	event := s.hub.Publish(events.Event{
		Time: time.Now(), Kind: "agent", Actor: incoming.Agent, Status: status,
		Path: incoming.Path, Title: titleForAgentEvent(incoming), Summary: incoming.Message,
		Risk: "none", Confidence: "reported", Evidence: []string{"agent-event-api"},
	})
	writeJSON(w, http.StatusAccepted, event)
}

func titleForAgentEvent(incoming events.Incoming) string {
	if incoming.Kind == "" || incoming.Kind == "activity" {
		return incoming.Agent + " reported activity"
	}
	return incoming.Agent + " · " + strings.ReplaceAll(incoming.Kind, "_", " ")
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; connect-src 'self'")
		next.ServeHTTP(w, r)
	})
}
