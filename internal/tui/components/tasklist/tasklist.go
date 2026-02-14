package tasklist

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/maxbeizer/gh-agent-viz/internal/data"
)

// Model represents the task list component state
type Model struct {
	titleStyle       lipgloss.Style
	tableHeaderStyle lipgloss.Style
	tableRowStyle    lipgloss.Style
	tableRowSelected lipgloss.Style
	tasks            []data.AgentTask
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
		tasks:            []data.AgentTask{},
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
		return m.titleStyle.Render("Loading agent tasks...")
	}

	if len(m.tasks) == 0 {
		return m.titleStyle.Render("No agent tasks found")
	}

	var rows []string
	
	// Header
	header := m.tableHeaderStyle.Render("    Repository                       Task                                                     Updated")
	rows = append(rows, header)

	// Task rows
	for i, task := range m.tasks {
		selected := i == m.cursor
		row := m.renderRow(task, selected)
		rows = append(rows, row)
	}

	return strings.Join(rows, "\n")
}

// renderRow formats a single task as a table row
func (m Model) renderRow(task data.AgentTask, selected bool) string {
	style := m.tableRowStyle
	if selected {
		style = m.tableRowSelected
	}

	icon := m.statusIcon(task.Status)
	repo := truncate(task.Repository, 30)
	title := truncate(task.Title, 50)
	updated := formatTime(task.UpdatedAt)

	row := fmt.Sprintf("%-3s %-32s %-52s %s", icon, repo, title, updated)
	return style.Render(row)
}

// SetTasks updates the task list
func (m *Model) SetTasks(tasks []data.AgentTask) {
	m.tasks = tasks
	if m.cursor >= len(tasks) {
		m.cursor = len(tasks) - 1
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
	if m.cursor >= len(m.tasks) {
		m.cursor = len(m.tasks) - 1
	}
}

// SelectedTask returns the currently selected task
func (m Model) SelectedTask() *data.AgentTask {
	if m.cursor >= 0 && m.cursor < len(m.tasks) {
		return &m.tasks[m.cursor]
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

