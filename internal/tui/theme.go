package tui

import "github.com/charmbracelet/lipgloss"

// Theme contains all Lip Gloss styles for the UI
type Theme struct {
	StatusRunning    lipgloss.Style
	StatusQueued     lipgloss.Style
	StatusCompleted  lipgloss.Style
	StatusFailed     lipgloss.Style
	TableHeader      lipgloss.Style
	TableRow         lipgloss.Style
	TableRowSelected lipgloss.Style
	Border           lipgloss.Style
	Title            lipgloss.Style
	Footer           lipgloss.Style
	// Tab bar styles
	TabActive   lipgloss.Style
	TabInactive lipgloss.Style
	TabCount    lipgloss.Style
	// Focus area styles
	FocusBorder   lipgloss.Style
	RowGutter     lipgloss.Style
	RowGutterSel  lipgloss.Style
	SectionHeader lipgloss.Style
}

// NewTheme creates a default theme
func NewTheme() *Theme {
	return &Theme{
		StatusRunning:   lipgloss.NewStyle().Foreground(lipgloss.Color("42")),  // Green
		StatusQueued:    lipgloss.NewStyle().Foreground(lipgloss.Color("226")), // Yellow
		StatusCompleted: lipgloss.NewStyle().Foreground(lipgloss.Color("46")),  // Bright green
		StatusFailed:    lipgloss.NewStyle().Foreground(lipgloss.Color("196")), // Red
		TableHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true),
		TableRow: lipgloss.NewStyle().
			Padding(0, 1),
		TableRowSelected: lipgloss.NewStyle().
			Padding(0, 1).
			Background(lipgloss.Color("237")).
			Foreground(lipgloss.Color("15")),
		Border: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")),
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			Padding(0, 1),
		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(0, 1),
		// Active tab: bold with magenta accent background
		TabActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("99")).
			Padding(0, 2),
		// Inactive tab: subdued
		TabInactive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Padding(0, 2),
		// Count badge inside tabs
		TabCount: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		// Focus area border
		FocusBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")),
		// Left gutter indicator for selected row
		RowGutter: lipgloss.NewStyle().
			Foreground(lipgloss.Color("238")),
		RowGutterSel: lipgloss.NewStyle().
			Foreground(lipgloss.Color("99")).
			Bold(true),
		// Section headers within the focus area
		SectionHeader: lipgloss.NewStyle().
			Foreground(lipgloss.Color("63")).
			Bold(true).
			Padding(0, 1),
	}
}

// StatusIcon returns the appropriate emoji icon for a given status
func StatusIcon(status string) string {
	switch status {
	case "running":
		return "🟢"
	case "queued":
		return "🟡"
	case "needs-input":
		return "🧑"
	case "completed":
		return "✅"
	case "failed":
		return "❌"
	default:
		return "⚪"
	}
}
