package tooltimeline

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ToolEvent represents a single tool execution in the timeline
type ToolEvent struct {
	Timestamp string
	ToolName  string
	Icon      string
}

// Model holds the state of the tool timeline view
type Model struct {
	events   []ToolEvent
	viewport viewport.Model
	width    int
	height   int
	ready    bool
}

// New creates a new tool timeline model
func New(width, height int) Model {
	vp := viewport.New(width, height)
	return Model{
		viewport: vp,
		width:    width,
		height:   height,
	}
}

// SetEvents replaces the current events and re-renders
func (m *Model) SetEvents(events []ToolEvent) {
	m.events = events
	m.viewport.SetContent(m.renderTimeline())
	m.ready = true
}

// SetSize updates the viewport dimensions
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.Width = width
	m.viewport.Height = height
	if m.ready {
		m.viewport.SetContent(m.renderTimeline())
	}
}

// Update handles messages for viewport scrolling
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the timeline inside a bordered box
func (m Model) View() string {
	if !m.ready || len(m.events) == 0 {
		boxStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(1, 2).
			Width(m.width - 4)
		title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99")).Render("Tool Timeline")
		return boxStyle.Render(title + "\n\nNo tool executions recorded.")
	}

	return m.viewport.View()
}

// LineDown scrolls down one line
func (m *Model) LineDown() { m.viewport.ScrollDown(1) }

// LineUp scrolls up one line
func (m *Model) LineUp() { m.viewport.ScrollUp(1) }

// HalfPageDown scrolls down half a page
func (m *Model) HalfPageDown() { m.viewport.HalfPageDown() }

// HalfPageUp scrolls up half a page
func (m *Model) HalfPageUp() { m.viewport.HalfPageUp() }

func (m *Model) renderTimeline() string {
	if len(m.events) == 0 {
		return ""
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	iconStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("75"))

	innerWidth := m.width - 8
	if innerWidth < 20 {
		innerWidth = 20
	}

	var lines []string
	var prevTime string
	for _, ev := range m.events {
		ts := formatTimelineTimestamp(ev.Timestamp)

		// Group rapid sequences: dim the timestamp if same as previous
		tsDisplay := dimStyle.Render(ts)
		if ts == prevTime {
			tsDisplay = dimStyle.Render("  ·  ")
		}
		prevTime = ts

		line := fmt.Sprintf("  %s  %s %s",
			tsDisplay,
			iconStyle.Render(ev.Icon),
			ev.ToolName,
		)
		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")

	boxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("238")).
		Padding(1, 2).
		Width(m.width - 4)

	header := titleStyle.Render("Tool Timeline") +
		dimStyle.Render(fmt.Sprintf("  (%d executions)", len(m.events)))

	return boxStyle.Render(header + "\n\n" + content)
}

func formatTimelineTimestamp(ts string) string {
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		t, err = time.Parse(time.RFC3339, ts)
		if err != nil {
			return ts
		}
	}
	return t.Format("15:04")
}

// ToolIcon returns an emoji icon for a tool name
func ToolIcon(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "bash"), strings.Contains(lower, "terminal"):
		return "🔧"
	case strings.Contains(lower, "edit"), strings.Contains(lower, "write"), strings.Contains(lower, "create"):
		return "✏️"
	case strings.Contains(lower, "read"), strings.Contains(lower, "view"), strings.Contains(lower, "file"):
		return "📄"
	case strings.Contains(lower, "grep"), strings.Contains(lower, "search"), strings.Contains(lower, "glob"):
		return "🔍"
	case strings.Contains(lower, "git"), strings.Contains(lower, "commit"):
		return "📤"
	case strings.Contains(lower, "test"):
		return "🧪"
	default:
		return "⚙️"
	}
}
