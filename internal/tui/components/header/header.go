package header

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// FilterCounts holds per-filter session counts for the tab bar
type FilterCounts struct {
	All       int
	Attention int
	Active    int
	Completed int
	Failed    int
}

// Model represents the header component state
type Model struct {
	titleStyle  lipgloss.Style
	tabActive   lipgloss.Style
	tabInactive lipgloss.Style
	tabCount    lipgloss.Style
	title       string
	filter      *string
	counts      FilterCounts
}

// New creates a new header model
func New(titleStyle, tabActive, tabInactive, tabCount lipgloss.Style, title string, filter *string) Model {
	return Model{
		titleStyle:  titleStyle,
		tabActive:   tabActive,
		tabInactive: tabInactive,
		tabCount:    tabCount,
		title:       title,
		filter:      filter,
	}
}

// SetCounts updates the filter counts displayed in tab badges
func (m *Model) SetCounts(counts FilterCounts) {
	m.counts = counts
}

// View renders the header as a tab bar
func (m Model) View() string {
	title := m.titleStyle.Render(m.title)

	activeFilter := "all"
	if m.filter != nil && *m.filter != "" {
		activeFilter = *m.filter
	}

	tabs := []struct {
		key   string
		label string
		count int
	}{
		{"all", "ALL", m.counts.All},
		{"attention", "ACTION", m.counts.Attention},
		{"active", "RUNNING", m.counts.Active},
		{"completed", "DONE", m.counts.Completed},
		{"failed", "FAILED", m.counts.Failed},
	}

	renderedTabs := make([]string, 0, len(tabs))
	for _, tab := range tabs {
		label := fmt.Sprintf("%s %s", tab.label, m.tabCount.Render(fmt.Sprintf("(%d)", tab.count)))
		if tab.key == activeFilter {
			renderedTabs = append(renderedTabs, m.tabActive.Render(label))
		} else {
			renderedTabs = append(renderedTabs, m.tabInactive.Render(label))
		}
	}

	tabBar := strings.Join(renderedTabs, "")
	header := lipgloss.JoinHorizontal(lipgloss.Center, title, "  ", tabBar)
	return header + "\n"
}
