package taskdetail

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/maxbeizer/gh-agent-viz/internal/data"
)

// Model represents the task detail component state
type Model struct {
	titleStyle  lipgloss.Style
	borderStyle lipgloss.Style
	task        *data.AgentTask
	statusIcon  func(string) string
}

// New creates a new task detail model
func New(titleStyle, borderStyle lipgloss.Style, statusIconFunc func(string) string) Model {
	return Model{
		titleStyle:  titleStyle,
		borderStyle: borderStyle,
		statusIcon:  statusIconFunc,
	}
}

// View renders the task detail pane
func (m Model) View() string {
	if m.task == nil {
		return m.titleStyle.Render("No task selected")
	}

	details := []string{
		m.titleStyle.Render(m.task.Title),
		"",
		fmt.Sprintf("Status:     %s %s", m.statusIcon(m.task.Status), m.task.Status),
		fmt.Sprintf("Repository: %s", m.task.Repository),
		fmt.Sprintf("Branch:     %s", m.task.Branch),
		fmt.Sprintf("PR:         #%d", m.task.PRNumber),
		fmt.Sprintf("PR URL:     %s", m.task.PRURL),
		fmt.Sprintf("Created:    %s", m.task.CreatedAt.Format("2006-01-02 15:04:05")),
		fmt.Sprintf("Updated:    %s", m.task.UpdatedAt.Format("2006-01-02 15:04:05")),
		fmt.Sprintf("Task ID:    %s", m.task.ID),
	}

	return m.borderStyle.Render(joinVertical(details))
}

// SetTask updates the task being displayed
func (m *Model) SetTask(task *data.AgentTask) {
	m.task = task
}

func joinVertical(lines []string) string {
	result := ""
	for _, line := range lines {
		result += line + "\n"
	}
	return result
}
