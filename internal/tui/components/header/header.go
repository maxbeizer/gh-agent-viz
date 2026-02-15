package header

import (
	"github.com/charmbracelet/lipgloss"
)

// Model represents the header component state
type Model struct {
	titleStyle  lipgloss.Style
	filterStyle lipgloss.Style
	title       string
	filter      *string
}

// New creates a new header model
func New(titleStyle lipgloss.Style, title string, filter *string) Model {
	filterStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(0, 2)

	return Model{
		titleStyle:  titleStyle,
		filterStyle: filterStyle,
		title:       title,
		filter:      filter,
	}
}

// View renders the header
func (m Model) View() string {
	title := m.titleStyle.Render(m.title)
	filterText := ""
	if m.filter != nil && *m.filter != "" {
		filterText = m.filterStyle.Render("Filter: " + filterLabel(*m.filter))
	}

	header := lipgloss.JoinHorizontal(lipgloss.Top, title, filterText)
	return header + "\n"
}

func filterLabel(filter string) string {
	switch filter {
	case "attention":
		return "needs action"
	case "active":
		return "running"
	case "completed":
		return "done"
	default:
		return filter
	}
}
