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
	allSessions []data.Session
	statusIcon  func(string) string
	width       int
	height      int
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

	// Add PR info when available (any source)
	if m.session.PRNumber > 0 || m.session.PRURL != "" {
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

	if reason := attentionReason(m.session); reason != "" {
		details = append(details, "", reason)
	}

	if tl := RenderTimeline(m.session, timelineWidth(m.width)); tl != "" {
		details = append(details, sectionDivider(m.width-4))
		details = append(details, fmt.Sprintf("Timeline:   %s", tl))
	}

	// Show telemetry if available
	if m.session.Telemetry != nil {
		t := m.session.Telemetry
		details = append(details, sectionDivider(m.width-4))
		details = append(details, m.titleStyle.Render("Session Stats"))
		if t.Duration > 0 {
			details = append(details, fmt.Sprintf("⏱ Duration: %s", formatDuration(t.Duration)))
		}
		if t.ConversationTurns > 0 {
			details = append(details,
				fmt.Sprintf("💬 Turns: %d (%d user · %d assistant)",
					t.ConversationTurns, t.UserMessages, t.AssistantMessages),
			)
		}
		if t.ConversationTurns == 0 && t.Duration == 0 {
			details = append(details, "No usage data available for this session")
		}
		if t.InputTokens > 0 {
			details = append(details,
				fmt.Sprintf("🪙 Tokens: %s in, %s out, %s cached (%d calls)",
					data.FormatTokenCount(t.InputTokens),
					data.FormatTokenCount(t.OutputTokens),
					data.FormatTokenCount(t.CachedTokens),
					t.ModelCalls))
			if t.Model != "" {
				details = append(details, fmt.Sprintf("🤖 Model: %s", t.Model))
			}
		}
	}

	// Show dependency graph if relationships exist
	graph := ParseSessionDeps(m.session, m.allSessions)
	if rendered := RenderDepGraph(graph, m.width); rendered != "" {
		details = append(details, sectionDivider(m.width-4))
		details = append(details, m.titleStyle.Render("Related Sessions"), rendered)
	}

	return m.borderStyle.Render(joinVertical(details))
}

// SetTask updates the session being displayed
func (m *Model) SetTask(session *data.Session) {
	m.session = session
}

// SetAllSessions updates the full session list for dependency graph rendering.
func (m *Model) SetAllSessions(sessions []data.Session) {
	m.allSessions = sessions
}

// SetSize updates the available rendering size for responsive layout.
func (m *Model) SetSize(width, height int) {
	if width > 0 {
		m.width = width
	}
	if height > 0 {
		m.height = height
	}
}

// ViewSplit renders the detail pane for split-pane mode with a left border.
func (m Model) ViewSplit() string {
	if m.session == nil {
		content := m.titleStyle.Render("No session selected")
		style := lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderLeft(true).
			BorderTop(false).
			BorderBottom(false).
			BorderRight(false).
			BorderForeground(lipgloss.Color("63")).
			PaddingLeft(1)
		if m.width > 0 {
			style = style.Width(m.width - 2)
		}
		if m.height > 0 {
			style = style.Height(m.height)
		}
		return style.Render(content)
	}

	details := []string{
		m.titleStyle.Render(detailTitle(m.session.Title)),
		"",
		fmt.Sprintf("Status:     %s %s", m.statusIcon(m.session.Status), m.session.Status),
		fmt.Sprintf("Source:     %s", m.session.Source),
		fmt.Sprintf("Repository: %s", detailValue(m.session.Repository, "n/a")),
		fmt.Sprintf("Branch:     %s", detailValue(m.session.Branch, "n/a")),
	}

	if m.session.PRNumber > 0 || m.session.PRURL != "" {
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

	if reason := attentionReason(m.session); reason != "" {
		details = append(details, "", reason)
	}

	if tl := RenderTimeline(m.session, timelineWidth(m.width)); tl != "" {
		details = append(details, sectionDivider(m.width-4))
		details = append(details, fmt.Sprintf("Timeline:   %s", tl))
	}

	if m.session.Telemetry != nil {
		t := m.session.Telemetry
		details = append(details, sectionDivider(m.width-4))
		details = append(details, m.titleStyle.Render("Session Stats"))
		if t.Duration > 0 {
			details = append(details, fmt.Sprintf("⏱ Duration: %s", formatDuration(t.Duration)))
		}
		if t.ConversationTurns > 0 {
			details = append(details,
				fmt.Sprintf("💬 Turns: %d (%d user · %d assistant)",
					t.ConversationTurns, t.UserMessages, t.AssistantMessages),
			)
		}
		if t.ConversationTurns == 0 && t.Duration == 0 {
			details = append(details, "No usage data available")
		}
		if t.InputTokens > 0 {
			details = append(details,
				fmt.Sprintf("🪙 Tokens: %s in, %s out, %s cached (%d calls)",
					data.FormatTokenCount(t.InputTokens),
					data.FormatTokenCount(t.OutputTokens),
					data.FormatTokenCount(t.CachedTokens),
					t.ModelCalls))
			if t.Model != "" {
				details = append(details, fmt.Sprintf("🤖 Model: %s", t.Model))
			}
		}
	}

	// Show dependency graph if relationships exist
	graph := ParseSessionDeps(m.session, m.allSessions)
	if rendered := RenderDepGraph(graph, m.width); rendered != "" {
		details = append(details, sectionDivider(m.width-4))
		details = append(details, m.titleStyle.Render("Related Sessions"), rendered)
	}

	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderLeft(true).
		BorderTop(false).
		BorderBottom(false).
		BorderRight(false).
		BorderForeground(lipgloss.Color("63")).
		PaddingLeft(1)
	if m.width > 0 {
		style = style.Width(m.width - 2)
	}
	if m.height > 0 {
		style = style.Height(m.height)
	}

	return style.Render(joinVertical(details))
}

func joinVertical(lines []string) string {
	result := ""
	for _, line := range lines {
		result += line + "\n"
	}
	return result
}

func attentionReason(session *data.Session) string {
	if session == nil {
		return ""
	}
	status := strings.ToLower(strings.TrimSpace(session.Status))
	if status == "needs-input" {
		return "✋ This session is waiting for your input to continue."
	}
	if status == "failed" {
		return "🚨 This session has failed. Press 'l' to check logs."
	}
	return ""
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

// sectionDivider renders a horizontal rule for visual separation between sections.
func sectionDivider(width int) string {
	if width <= 0 {
		width = 40
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Render(strings.Repeat("─", width))
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

// timelineWidth picks a reasonable bar width based on available space.
func timelineWidth(availableWidth int) int {
	const defaultWidth = 24
	if availableWidth <= 0 {
		return defaultWidth
	}
	// Leave room for "Timeline:   " prefix (12) + "  Xh ago → now" suffix (~16)
	w := availableWidth - 28
	if w < 8 {
		return 8
	}
	if w > 48 {
		return 48
	}
	return w
}

