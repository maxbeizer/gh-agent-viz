package data

import (
	"encoding/json"
	"os"
	"time"
)

// Snapshot captures the TUI state as a machine-readable artifact.
type Snapshot struct {
	ViewMode       string            `json:"view_mode"`
	TerminalSize   SnapshotSize      `json:"terminal_size"`
	RenderedOutput string            `json:"rendered_output"`
	SessionCount   int               `json:"session_count"`
	FilterCounts   SnapshotCounts    `json:"filter_counts"`
	Sessions       []SnapshotSession `json:"sessions"`
	FocusedPanel   string            `json:"focused_panel"`
	Timestamp      string            `json:"timestamp"`
}

// SnapshotSize holds terminal dimensions.
type SnapshotSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// SnapshotCounts mirrors filter tab counts.
type SnapshotCounts struct {
	All       int `json:"all"`
	Attention int `json:"attention"`
	Warning   int `json:"warning"`
	Active    int `json:"active"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
}

// SnapshotSession is a minimal session summary for the snapshot.
type SnapshotSession struct {
	ID             string        `json:"id"`
	Status         string        `json:"status"`
	Title          string        `json:"title"`
	Repository     string        `json:"repository"`
	Source         SessionSource `json:"source"`
	AttentionLevel string        `json:"attention_level"`
}

// WriteSnapshot serialises a Snapshot to path as indented JSON.
func WriteSnapshot(path string, snap *Snapshot) error {
	if snap.Timestamp == "" {
		snap.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	b, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}
