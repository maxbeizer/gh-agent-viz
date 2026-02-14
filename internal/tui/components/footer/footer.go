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

// View renders the footer with key binding hints
func (m Model) View() string {
	footer := m.style.Render(strings.Join(m.hints, " • "))
	return "\n" + footer
}

// SetHints updates the key binding hints
func (m *Model) SetHints(keys []key.Binding) {
	m.hints = make([]string, 0, len(keys))
	for _, k := range keys {
		m.hints = append(m.hints, k.Help().Key+" "+k.Help().Desc)
	}
}
