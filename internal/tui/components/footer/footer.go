package footer

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the footer component state
type Model struct {
	style lipgloss.Style
	hints []string
	width int
}

// New creates a new footer model with key binding hints
func New(style lipgloss.Style, keys []key.Binding) Model {
	hints := make([]string, 0, len(keys))
	for _, k := range keys {
		hints = append(hints, k.Help().Key+" "+k.Help().Desc)
	}

	return Model{
		style: style,
		hints: hints,
	}
}

// View renders the footer with key binding hints, truncating to fit width
func (m Model) View() string {
	sep := " • "
	joined := strings.Join(m.hints, sep)

	// Account for horizontal padding in the style
	hPad := m.style.GetHorizontalPadding()
	available := m.width - hPad
	if available > 0 && lipgloss.Width(joined) > available {
		ellipsis := " …"
		joined = ""
		for i, h := range m.hints {
			candidate := joined
			if i > 0 {
				candidate += sep
			}
			candidate += h
			if lipgloss.Width(candidate+ellipsis) > available {
				joined += ellipsis
				break
			}
			joined = candidate
		}
	}

	footer := m.style.Render(joined)
	return "\n" + footer
}

// SetHints updates the key binding hints
func (m *Model) SetHints(keys []key.Binding) {
	m.hints = make([]string, 0, len(keys))
	for _, k := range keys {
		m.hints = append(m.hints, k.Help().Key+" "+k.Help().Desc)
	}
}

// SetWidth updates the available terminal width for truncation
func (m *Model) SetWidth(width int) {
	m.width = width
}
