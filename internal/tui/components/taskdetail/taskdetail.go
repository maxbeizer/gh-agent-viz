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

	// Show telemetry if available
	if m.session.Telemetry != nil {
		t := m.session.Telemetry
		details = append(details, "", m.titleStyle.Render("Usage"))
		if t.Duration > 0 {
			details = append(details, fmt.Sprintf("Duration:   %s", formatDuration(t.Duration)))
		}
		if t.ConversationTurns > 0 {
			details = append(details,
				fmt.Sprintf("Messages:   %d total (%d user, %d assistant)",
					t.ConversationTurns, t.UserMessages, t.AssistantMessages),
			)
		}
		if t.ConversationTurns == 0 && t.Duration == 0 {
			details = append(details, "No usage data available for this session")
		}
	}

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

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if hours >= 24 {
		days := hours / 24
		hours = hours % 24
		if hours > 0 {
			return fmt.Sprintf("%dd %dh", days, hours)
		}
		return fmt.Sprintf("%dd", days)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dh", hours)
}
