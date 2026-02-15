package tasklist

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/maxbeizer/gh-agent-viz/internal/data"
)

// Model represents the task list component state
type Model struct {
	titleStyle       lipgloss.Style
	tableHeaderStyle lipgloss.Style
	tableRowStyle    lipgloss.Style
	tableRowSelected lipgloss.Style
	sessions         []data.Session
	cursor           int
	loading          bool
	statusIcon       func(string) string
}

// New creates a new task list model
func New(titleStyle, headerStyle, rowStyle, rowSelectedStyle lipgloss.Style, statusIconFunc func(string) string) Model {
	return Model{
		titleStyle:       titleStyle,
		tableHeaderStyle: headerStyle,
		tableRowStyle:    rowStyle,
		tableRowSelected: rowSelectedStyle,
		sessions:         []data.Session{},
		cursor:           0,
		loading:          false,
		statusIcon:       statusIconFunc,
	}
}

// Init initializes the task list
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the task list
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

// View renders the task list as a table
func (m Model) View() string {
	if m.loading {
		return m.titleStyle.Render("Loading sessions...")
	}

	if len(m.sessions) == 0 {
		return m.titleStyle.Render("No sessions found")
	}

	var rows []string

	// Header
	header := m.tableHeaderStyle.Render("    Repository                       Task                                                     Updated")
	rows = append(rows, header)

	// Task rows
	for i, session := range m.sessions {
		selected := i == m.cursor
		row := m.renderRow(session, selected)
		rows = append(rows, row)
	}

	return strings.Join(rows, "\n")
}

// renderRow formats a single session as a table row
func (m Model) renderRow(session data.Session, selected bool) string {
	style := m.tableRowStyle
	if selected {
		style = m.tableRowSelected
	}

	icon := m.statusIcon(session.Status)
	repo := truncate(session.Repository, 30)
	title := truncate(session.Title, 50)
	updated := formatTime(session.UpdatedAt)

	row := fmt.Sprintf("%-3s %-32s %-52s %s", icon, repo, title, updated)
	return style.Render(row)
}

// SetTasks updates the session list
func (m *Model) SetTasks(sessions []data.Session) {
	m.sessions = sessions
	if m.cursor >= len(sessions) {
		m.cursor = len(sessions) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

// MoveCursor moves the cursor up or down
func (m *Model) MoveCursor(delta int) {
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.sessions) {
		m.cursor = len(m.sessions) - 1
	}
}

// SelectedTask returns the currently selected session
func (m Model) SelectedTask() *data.Session {
	if m.cursor >= 0 && m.cursor < len(m.sessions) {
		return &m.sessions[m.cursor]
	}
	return nil
}

// Helper functions

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	} else {
		return t.Format("Jan 2")
	}
}
