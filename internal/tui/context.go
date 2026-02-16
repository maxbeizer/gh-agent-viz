package tui

import "github.com/maxbeizer/gh-agent-viz/internal/config"

// FilterCounts holds the session count for each filter state
type FilterCounts struct {
	All       int
	Attention int
	Active    int
	Completed int
	Failed    int
}

// ProgramContext holds shared state and configuration for the TUI
type ProgramContext struct {
	Config       *config.Config
	Width        int
	Height       int
	Error        error
	Debug        bool
	StatusFilter string // "all", "attention", "active", "completed", "failed"
	Counts       FilterCounts
}

// NewProgramContext initializes a new program context
func NewProgramContext() *ProgramContext {
	return &ProgramContext{
		Config:       config.DefaultConfig(),
		Width:        80,
		Height:       24,
		StatusFilter: "all",
	}
}
