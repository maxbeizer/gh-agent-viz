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
	titleStyle     lipgloss.Style
	tabActive      lipgloss.Style
	tabInactive    lipgloss.Style
	tabCount       lipgloss.Style
	title          string
	filter         *string
	counts         FilterCounts
	useAsciiHeader bool
	width          int
	height         int
}

// Banner is the compact ASCII art header
const Banner = "┌─────────────────────────┐\n│  A G E N T   V I Z  ⚡  │\n└─────────────────────────┘"

// bannerWidth is the visual width of the banner (excluding ANSI codes)
const bannerWidth = 27

// minHeightForBanner is the minimum terminal height to show the banner
const minHeightForBanner = 15

// New creates a new header model
func New(titleStyle, tabActive, tabInactive, tabCount lipgloss.Style, title string, filter *string, useAsciiHeader bool) Model {
	return Model{
		titleStyle:     titleStyle,
		tabActive:      tabActive,
		tabInactive:    tabInactive,
		tabCount:       tabCount,
		title:          title,
		filter:         filter,
		useAsciiHeader: useAsciiHeader,
		width:          80,
		height:         24,
	}
}

// SetSize updates the terminal dimensions for responsive layout
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetCounts updates the filter counts displayed in tab badges
func (m *Model) SetCounts(counts FilterCounts) {
	m.counts = counts
}

// showBanner returns true when the ASCII banner should be displayed
func (m Model) showBanner() bool {
	return m.useAsciiHeader && m.height >= minHeightForBanner && m.width >= bannerWidth
}

// View renders the header as a tab bar
func (m Model) View() string {
	title := m.titleStyle.Render(m.title)

	activeFilter := "attention"
	if m.filter != nil && *m.filter != "" {
		activeFilter = *m.filter
	}

	tabs := []struct {
		key   string
		label string
		count int
	}{
		{"attention", "ATTENTION", m.counts.Attention},
		{"active", "RUNNING", m.counts.Active},
		{"completed", "DONE", m.counts.Completed},
		{"failed", "FAILED", m.counts.Failed},
		{"all", "ALL", m.counts.All},
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
	tabLine := lipgloss.JoinHorizontal(lipgloss.Center, title, "  ", tabBar)

	separator := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "249", Dark: "240"}).
		Render(strings.Repeat("━", m.width))

	if m.showBanner() {
		styledBanner := m.titleStyle.Render(Banner)
		return styledBanner + "\n" + tabLine + "\n" + separator + "\n"
	}

	return tabLine + "\n" + separator + "\n"
}
