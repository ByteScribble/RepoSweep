package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ByteScribble/RepoSweep/internal/events"
)

func TestIngestRejectsUnknownFields(t *testing.T) {
	srv := New("/tmp/example", events.NewHub(10))
	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewBufferString(`{"agent":"codex","message":"work","extra":true}`))
	recorder := httptest.NewRecorder()
	srv.Handler().ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestIngestPublishesAgentEvent(t *testing.T) {
	hub := events.NewHub(10)
	srv := New("/tmp/example", hub)
	req := httptest.NewRequest(http.MethodPost, "/api/events", bytes.NewBufferString(`{"agent":"codex","kind":"intent","message":"Refactor auth"}`))
	recorder := httptest.NewRecorder()
	srv.Handler().ServeHTTP(recorder, req)
	if recorder.Code != http.StatusAccepted || len(hub.Snapshot()) != 1 {
		t.Fatalf("unexpected response %d or event count %d", recorder.Code, len(hub.Snapshot()))
	}
}
