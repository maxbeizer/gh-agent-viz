package statsbar

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
	"github.com/maxbeizer/gh-agent-viz/internal/data"
)

// Counts holds the stats to display.
type Counts struct {
	Active      int
	Idle        int
	Attention   int // urgent
	Warning     int
	Completed   int
	TotalTokens int64
	TotalCost   float64
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

	activeStyle := lipgloss.NewStyle().Foreground(compat.AdaptiveColor{Light: lipgloss.Color("28"), Dark: lipgloss.Color("42")})
	urgentStyle := lipgloss.NewStyle().Foreground(compat.AdaptiveColor{Light: lipgloss.Color("160"), Dark: lipgloss.Color("203")})
	doneStyle := lipgloss.NewStyle().Foreground(compat.AdaptiveColor{Light: lipgloss.Color("30"), Dark: lipgloss.Color("72")})
	tokenStyle := lipgloss.NewStyle().Foreground(compat.AdaptiveColor{Light: lipgloss.Color("244"), Dark: lipgloss.Color("245")})
	dimStyle := lipgloss.NewStyle().Foreground(compat.AdaptiveColor{Light: lipgloss.Color("249"), Dark: lipgloss.Color("240")})

	if m.counts.Active > 0 {
		parts = append(parts, activeStyle.Render(fmt.Sprintf("● %d active", m.counts.Active)))
	}
	if m.counts.Attention > 0 {
		parts = append(parts, urgentStyle.Render(fmt.Sprintf("✋ %d attention", m.counts.Attention)))
	}
	parts = append(parts, doneStyle.Render(fmt.Sprintf("✅ %d done", m.counts.Completed)))
	if m.counts.TotalTokens > 0 {
		parts = append(parts, tokenStyle.Render(fmt.Sprintf("🪙 %s tokens", data.FormatTokenCount(m.counts.TotalTokens))))
	}
	if m.counts.TotalCost > 0 {
		parts = append(parts, tokenStyle.Render(fmt.Sprintf("💰 %s", data.FormatCost(m.counts.TotalCost))))
	}

	bar := strings.Join(parts, dimStyle.Render("  │  "))

	return dimStyle.Render("  ") + bar
}
