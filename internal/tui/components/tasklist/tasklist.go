package tasklist

import (
	"fmt"
	"sort"
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
	tasks            []data.AgentTask
	columnTaskIdx    [3][]int
	activeColumn     int
	rowCursor        [3]int
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
		activeColumn:     0,
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

	if len(m.tasks) == 0 {
		return m.titleStyle.Render("No sessions found")
	}

	columns := make([]string, 0, 3)
	for col := 0; col < 3; col++ {
		columns = append(columns, m.renderColumn(col))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, columns...)
}

func (m Model) renderColumn(column int) string {
	headerStyle := m.tableHeaderStyle
	if column == m.activeColumn {
		headerStyle = m.tableRowSelected.Bold(true)
	}

	indices := m.columnTaskIdx[column]
	rows := []string{headerStyle.Render(fmt.Sprintf("%s (%d)", columnTitle(column), len(indices)))}
	if len(indices) == 0 {
		rows = append(rows, m.tableRowStyle.Render("  —"))
		return lipgloss.NewStyle().Width(42).PaddingRight(1).Render(strings.Join(rows, "\n"))
	}

	cursor := m.rowCursor[column]
	if cursor >= len(indices) {
		cursor = len(indices) - 1
	}
	if cursor < 0 {
		cursor = 0
	}

	for i, taskIdx := range indices {
		task := m.tasks[taskIdx]
		rows = append(rows, m.renderRow(task, column == m.activeColumn && i == cursor))
	}

	return lipgloss.NewStyle().Width(42).PaddingRight(1).Render(strings.Join(rows, "\n"))
}

// renderRow formats a single task as a card row
func (m Model) renderRow(task data.AgentTask, selected bool) string {
	style := m.tableRowStyle
	if selected {
		style = m.tableRowSelected
	}

	icon := m.statusIcon(task.Status)
	title := truncate(task.Title, 34)
	repo := truncate(task.Repository, 18)
	updated := formatTime(task.UpdatedAt)

	row := fmt.Sprintf("%s %s\n  %s • %s", icon, title, repo, updated)
	return style.Render(row)
}

// SetTasks updates the task list
func (m *Model) SetTasks(tasks []data.AgentTask) {
	m.tasks = append([]data.AgentTask(nil), tasks...)
	sort.SliceStable(m.tasks, func(i, j int) bool {
		return m.tasks[i].UpdatedAt.After(m.tasks[j].UpdatedAt)
	})

	m.columnTaskIdx = [3][]int{}
	for i, task := range m.tasks {
		column := statusColumn(task.Status)
		m.columnTaskIdx[column] = append(m.columnTaskIdx[column], i)
	}

	for col := 0; col < 3; col++ {
		if len(m.columnTaskIdx[col]) == 0 {
			m.rowCursor[col] = 0
			continue
		}
		if m.rowCursor[col] >= len(m.columnTaskIdx[col]) {
			m.rowCursor[col] = len(m.columnTaskIdx[col]) - 1
		}
		if m.rowCursor[col] < 0 {
			m.rowCursor[col] = 0
		}
	}

	if len(m.columnTaskIdx[m.activeColumn]) == 0 {
		for col := 0; col < 3; col++ {
			if len(m.columnTaskIdx[col]) > 0 {
				m.activeColumn = col
				break
			}
		}
	}
}

// MoveCursor moves the cursor up or down
func (m *Model) MoveCursor(delta int) {
	columnTasks := m.columnTaskIdx[m.activeColumn]
	if len(columnTasks) == 0 {
		m.rowCursor[m.activeColumn] = 0
		return
	}

	m.rowCursor[m.activeColumn] += delta
	if m.rowCursor[m.activeColumn] < 0 {
		m.rowCursor[m.activeColumn] = 0
	}
	if m.rowCursor[m.activeColumn] >= len(columnTasks) {
		m.rowCursor[m.activeColumn] = len(columnTasks) - 1
	}
}

// MoveColumn moves the active column left or right.
func (m *Model) MoveColumn(delta int) {
	m.activeColumn += delta
	if m.activeColumn < 0 {
		m.activeColumn = 0
	}
	if m.activeColumn > 2 {
		m.activeColumn = 2
	}
}

// SelectedTask returns the currently selected task
func (m Model) SelectedTask() *data.AgentTask {
	columnTasks := m.columnTaskIdx[m.activeColumn]
	if len(columnTasks) == 0 {
		return nil
	}

	cursor := m.rowCursor[m.activeColumn]
	if cursor >= len(columnTasks) {
		cursor = len(columnTasks) - 1
	}
	if cursor < 0 {
		cursor = 0
	}

	taskIdx := columnTasks[cursor]
	return &m.tasks[taskIdx]
}

// Helper functions

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}

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

func statusColumn(status string) int {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "failed", "cancelled", "canceled":
		return 2
	case "running", "queued", "in progress", "active", "open":
		return 0
	default:
		return 1
	}
}

func columnTitle(column int) string {
	switch column {
	case 0:
		return "Running"
	case 1:
		return "Done"
	default:
		return "Failed"
	}
}
