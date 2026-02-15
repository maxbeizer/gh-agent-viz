package tasklist

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/maxbeizer/gh-agent-viz/internal/data"
)

// Column widths for table alignment
const (
	statusWidth = 3   // Status icon (emoji)
	sourceWidth = 6   // Source badge (emoji + padding)
	repoWidth   = 30  // Repository name
	titleWidth  = 52  // Task title
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
	selectedTaskID   string // Store selected task ID for persistence
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
		return m.titleStyle.Render("No agent tasks found\n\nTry:\n  • Refresh with 'r'\n  • Toggle filter with Tab\n  • Check your repository settings")
	}

	var rows []string

	// Header with count - using format string with column widths
	headerFormat := fmt.Sprintf("%%-%ds%%-%ds%%-%ds%%-%ds Updated  (%%d tasks)",
		statusWidth, sourceWidth, repoWidth, titleWidth)
	header := m.tableHeaderStyle.Render(fmt.Sprintf(headerFormat,
		"", "Source", "Repository", "Task", len(m.tasks)))
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
	source := sourceIcon(task.Source)
	repo := truncate(task.Repository, repoWidth-2)  // -2 for padding
	title := truncate(task.Title, titleWidth-2)     // -2 for padding
	updated := formatTime(task.UpdatedAt)

	// Use constants for column widths
	row := fmt.Sprintf("%-*s %-*s %-*s %-*s %s",
		statusWidth, icon,
		sourceWidth, source,
		repoWidth, repo,
		titleWidth, title,
		updated)
	return style.Render(row)
}

// SetTasks updates the task list and preserves selection
func (m *Model) SetTasks(tasks []data.AgentTask) {
	// Store selected task ID before updating
	if m.cursor >= 0 && m.cursor < len(m.tasks) {
		m.selectedTaskID = m.tasks[m.cursor].ID
	}

	m.tasks = tasks

	// Try to restore cursor to the same task ID
	if m.selectedTaskID != "" {
		for i, task := range tasks {
			if task.ID == m.selectedTaskID {
				m.cursor = i
				return
			}
		}
	}

	// If task ID not found or no previous selection, clamp cursor
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

func sourceIcon(source string) string {
	switch source {
	case "agent-task":
		return "🤖"
	case "local":
		return "💻"
	default:
		// Two spaces to maintain alignment with emoji characters above
		return "  "
	}
}
