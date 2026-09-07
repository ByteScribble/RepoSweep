package events

import (
	"sort"
	"strings"
	"sync"
	"time"
)

const activeActorTTL = 10 * time.Minute

type Hub struct {
	mu           sync.RWMutex
	capacity     int
	sequence     uint64
	events       []Event
	subscribers  map[chan Event]struct{}
	activeActors map[string]time.Time
}

func NewHub(capacity int) *Hub {
	if capacity < 1 {
		capacity = 1
	}
	return &Hub{
		capacity: capacity, subscribers: make(map[chan Event]struct{}), activeActors: make(map[string]time.Time),
	}
}

func (h *Hub) Publish(event Event) Event {
	h.mu.Lock()
	h.sequence++
	event.Sequence = h.sequence
	h.expireActors(event.Time)
	if event.Actor == "" {
		event.Actor = h.attributeActor(&event)
	}
	if event.Kind == "agent" && event.Status == "active" {
		h.activeActors[event.Actor] = event.Time
	}
	if event.Kind == "agent" && (event.Status == "done" || event.Status == "failed") {
		delete(h.activeActors, event.Actor)
	}
	h.events = append(h.events, event)
	if len(h.events) > h.capacity {
		h.events = append([]Event(nil), h.events[len(h.events)-h.capacity:]...)
	}
	for subscriber := range h.subscribers {
		select {
		case subscriber <- event:
		default:
		}
	}
	h.mu.Unlock()
	return event
}

func (h *Hub) expireActors(now time.Time) {
	for actor, lastSeen := range h.activeActors {
		if now.Sub(lastSeen) > activeActorTTL {
			delete(h.activeActors, actor)
		}
	}
}

func (h *Hub) attributeActor(event *Event) string {
	if len(h.activeActors) == 1 {
		for actor := range h.activeActors {
			return actor
		}
	}
	if len(h.activeActors) > 1 {
		actors := make([]string, 0, len(h.activeActors))
		for actor := range h.activeActors {
			actors = append(actors, actor)
		}
		sort.Strings(actors)
		if event.Metadata == nil {
			event.Metadata = make(map[string]string)
		}
		event.Metadata["candidate_agents"] = strings.Join(actors, ",")
		return "unattributed"
	}
	return "filesystem"
}

func (h *Hub) Snapshot() []Event {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]Event, len(h.events))
	copy(out, h.events)
	return out
}

func (h *Hub) Subscribe() (<-chan Event, func()) {
	ch := make(chan Event, 32)
	h.mu.Lock()
	h.subscribers[ch] = struct{}{}
	h.mu.Unlock()
	return ch, func() {
		h.mu.Lock()
		if _, ok := h.subscribers[ch]; ok {
			delete(h.subscribers, ch)
			close(ch)
		}
		h.mu.Unlock()
	}
}
