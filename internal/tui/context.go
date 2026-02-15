package tui

import "github.com/maxbeizer/gh-agent-viz/internal/config"

// ProgramContext holds shared state and configuration for the TUI
type ProgramContext struct {
	Config       *config.Config
	Width        int
	Height       int
	Error        error
	StatusFilter string // "all", "active", "completed", "failed"
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
