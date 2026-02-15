package taskdetail

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/maxbeizer/gh-agent-viz/internal/data"
)

// Model represents the task detail component state
type Model struct {
	titleStyle  lipgloss.Style
	borderStyle lipgloss.Style
	session     *data.Session
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

// View renders the session detail pane
func (m Model) View() string {
	if m.session == nil {
		return m.titleStyle.Render("No session selected")
	}

	details := []string{
		m.titleStyle.Render(detailTitle(m.session.Title)),
		"",
		fmt.Sprintf("Status:     %s %s", m.statusIcon(m.session.Status), m.session.Status),
		fmt.Sprintf("Source:     %s", m.session.Source),
		fmt.Sprintf("Repository: %s", detailValue(m.session.Repository, "not available")),
		fmt.Sprintf("Branch:     %s", detailValue(m.session.Branch, "not available")),
	}

	// Add PR info for agent-task sessions
	if m.session.Source == data.SourceAgentTask {
		details = append(details,
			fmt.Sprintf("PR:         #%d", m.session.PRNumber),
			fmt.Sprintf("PR URL:     %s", m.session.PRURL),
		)
	}

	details = append(details,
		fmt.Sprintf("Created:    %s", detailTimestamp(m.session.CreatedAt)),
		fmt.Sprintf("Updated:    %s", detailTimestamp(m.session.UpdatedAt)),
		fmt.Sprintf("Session ID: %s", m.session.ID),
	)

	return m.borderStyle.Render(joinVertical(details))
}

// SetTask updates the session being displayed
func (m *Model) SetTask(session *data.Session) {
	m.session = session
}

func joinVertical(lines []string) string {
	result := ""
	for _, line := range lines {
		result += line + "\n"
	}
	return result
}

func detailValue(value string, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func detailTimestamp(ts time.Time) string {
	if ts.IsZero() {
		return "not recorded"
	}
	return ts.Format("2006-01-02 15:04:05")
}

func detailTitle(title string) string {
	return detailValue(title, "Untitled Session")
}
