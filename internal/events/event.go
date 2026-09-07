package events

import "time"

type Event struct {
	Sequence   uint64            `json:"sequence"`
	Time       time.Time         `json:"time"`
	Kind       string            `json:"kind"`
	Actor      string            `json:"actor"`
	Status     string            `json:"status"`
	Path       string            `json:"path,omitempty"`
	Title      string            `json:"title"`
	Summary    string            `json:"summary"`
	Risk       string            `json:"risk"`
	Confidence string            `json:"confidence"`
	Evidence   []string          `json:"evidence"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type Incoming struct {
	Agent   string `json:"agent"`
	Kind    string `json:"kind"`
	Message string `json:"message"`
	Path    string `json:"path"`
	Status  string `json:"status"`
}
