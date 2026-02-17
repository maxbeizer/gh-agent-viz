package help

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Model represents the help overlay state.
type Model struct {
	visible bool
	width   int
	height  int
}

// New creates a new help overlay model.
func New() Model {
	return Model{}
}

// Toggle flips help overlay visibility.
func (m *Model) Toggle() {
	m.visible = !m.visible
}

// Visible returns whether the help overlay is shown.
func (m Model) Visible() bool {
	return m.visible
}

// SetSize updates the available dimensions for the overlay.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// View renders the help overlay panel. Returns empty string when hidden.
func (m Model) View() string {
	if !m.visible {
		return ""
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "55", Dark: "99"})
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "28", Dark: "42"}).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "240", Dark: "252"})
	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "27", Dark: "63"}).MarginBottom(1)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "244", Dark: "241"}).Italic(true)

	formatKey := func(k, desc string) string {
		return keyStyle.Render(k) + "  " + descStyle.Render(desc)
	}

	// Navigation section
	nav := sectionStyle.Render("Navigation") + "\n" +
		formatKey("↑/↓ j/k", "navigate") + "\n" +
		formatKey("enter", "details") + "\n" +
		formatKey("esc", "back") + "\n" +
		formatKey("tab", "cycle filter") + "\n" +
		formatKey("a", "attention tab")

	// Actions section
	actions := sectionStyle.Render("Actions") + "\n" +
		formatKey("o", "open PR") + "\n" +
		formatKey("s", "resume session") + "\n" +
		formatKey("x", "dismiss") + "\n" +
		formatKey("r", "refresh") + "\n" +
		formatKey("p", "toggle preview")

	// Views section
	views := sectionStyle.Render("Views") + "\n" +
		formatKey("K", "kanban board") + "\n" +
		formatKey("l", "logs") + "\n" +
		formatKey("p", "preview pane")

	// Groups section
	groups := sectionStyle.Render("Groups") + "\n" +
		formatKey("g", "cycle grouping") + "\n" +
		formatKey("⎵", "expand/collapse")

	// Log View section
	logView := sectionStyle.Render("Log View") + "\n" +
		formatKey("d/u", "page down/up") + "\n" +
		formatKey("g/G", "top/bottom") + "\n" +
		formatKey("f", "toggle follow")

	// Layout: two columns for Navigation+Actions, Views+Groups
	colWidth := 28
	col := func(content string) string {
		return lipgloss.NewStyle().Width(colWidth).Render(content)
	}

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, col(nav), col(actions))
	midRow := lipgloss.JoinHorizontal(lipgloss.Top, col(views), col(groups))

	body := strings.Join([]string{topRow, midRow, logView}, "\n\n")

	closeHint := dimStyle.Render("Press ? or esc to close")
	body += "\n\n" + lipgloss.NewStyle().Width(colWidth*2).Align(lipgloss.Center).Render(closeHint)

	boxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.AdaptiveColor{Light: "249", Dark: "238"}).
		Padding(1, 3).
		Width(colWidth*2 + 8)

	title := titleStyle.Render(" Keyboard Shortcuts ")
	box := boxStyle.BorderTop(true).Render(title + "\n\n" + body)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
