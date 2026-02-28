package statsbar

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/maxbeizer/gh-agent-viz/internal/data"
)

// Counts holds the stats to display.
type Counts struct {
	Active    int
	Attention int // urgent
	Warning   int
	Completed int
	TotalTokens int64
}

// Model represents the always-visible stats bar.
type Model struct {
	counts Counts
	width  int
}

// New creates a new stats bar model.
func New() Model {
	return Model{width: 80}
}

// SetCounts updates the displayed stats.
func (m *Model) SetCounts(counts Counts) {
	m.counts = counts
}

// SetWidth updates the available rendering width.
func (m *Model) SetWidth(width int) {
	m.width = width
}

// View renders the stats bar as a single compact line.
func (m Model) View() string {
	var parts []string

	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "28", Dark: "42"})
	urgentStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "160", Dark: "203"})
	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "172", Dark: "214"})
	doneStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "30", Dark: "72"})
	tokenStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "244", Dark: "245"})
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "249", Dark: "240"})

	if m.counts.Active > 0 {
		parts = append(parts, activeStyle.Render(fmt.Sprintf("● %d active", m.counts.Active)))
	}
	if m.counts.Attention > 0 {
		parts = append(parts, urgentStyle.Render(fmt.Sprintf("🔴 %d urgent", m.counts.Attention)))
	}
	if m.counts.Warning > 0 {
		parts = append(parts, warningStyle.Render(fmt.Sprintf("🟡 %d warning", m.counts.Warning)))
	}
	parts = append(parts, doneStyle.Render(fmt.Sprintf("✅ %d done", m.counts.Completed)))
	if m.counts.TotalTokens > 0 {
		parts = append(parts, tokenStyle.Render(fmt.Sprintf("🪙 %s tokens", data.FormatTokenCount(m.counts.TotalTokens))))
	}

	bar := strings.Join(parts, dimStyle.Render("  │  "))

	return dimStyle.Render("  ") + bar
}
