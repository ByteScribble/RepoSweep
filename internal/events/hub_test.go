package events

import (
	"testing"
	"time"
)

func TestHubBoundsHistoryAndAttributesActiveAgent(t *testing.T) {
	hub := NewHub(2)
	hub.Publish(Event{Kind: "agent", Actor: "codex", Status: "active", Time: time.Now()})
	event := hub.Publish(Event{Kind: "change", Time: time.Now()})
	if event.Actor != "codex" {
		t.Fatalf("expected active actor attribution, got %q", event.Actor)
	}
	hub.Publish(Event{Kind: "change", Time: time.Now()})
	if got := len(hub.Snapshot()); got != 2 {
		t.Fatalf("expected bounded history of 2, got %d", got)
	}
}

func TestHubFailsClosedWhenMultipleAgentsAreActive(t *testing.T) {
	hub := NewHub(10)
	now := time.Now()
	hub.Publish(Event{Kind: "agent", Actor: "codex", Status: "active", Time: now})
	hub.Publish(Event{Kind: "agent", Actor: "claude", Status: "active", Time: now})
	event := hub.Publish(Event{Kind: "change", Time: now})
	if event.Actor != "unattributed" {
		t.Fatalf("expected ambiguous activity to be unattributed, got %q", event.Actor)
	}
	if event.Metadata["candidate_agents"] != "claude,codex" {
		t.Fatalf("expected stable candidate evidence, got %#v", event.Metadata)
	}
}
