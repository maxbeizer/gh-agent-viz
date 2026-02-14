package tui

// ProgramContext holds shared state and configuration for the TUI
type ProgramContext struct {
	Config       interface{} // Placeholder for config
	Width        int
	Height       int
	Error        error
	StatusFilter string // "all", "active", "completed", "failed"
}

// NewProgramContext initializes a new program context
func NewProgramContext() *ProgramContext {
	return &ProgramContext{
		Width:        80,
		Height:       24,
		StatusFilter: "all",
	}
}
