package mission

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/maxbeizer/gh-agent-viz/internal/data"
)

// SessionCard pairs a session with its derived last-action text.
type SessionCard struct {
	Session    data.Session
	LastAction string
}

// Model represents the mission control dashboard state.
type Model struct {
	cards          []SessionCard
	cursor         int
	statusIcon     func(string) string
	animStatusIcon func(string, int) string
	animFrame      int
	width          int
	height         int
	titleStyle     lipgloss.Style
	cardStyle      lipgloss.Style
	cardSelStyle   lipgloss.Style
}

// New creates a new mission control model.
func New(
	titleStyle lipgloss.Style,
	cardStyle lipgloss.Style,
	cardSelStyle lipgloss.Style,
	statusIconFunc func(string) string,
	animStatusIconFunc func(string, int) string,
) Model {
	return Model{
		statusIcon:     statusIconFunc,
		animStatusIcon: animStatusIconFunc,
		titleStyle:     titleStyle,
		cardStyle:      cardStyle,
		cardSelStyle:   cardSelStyle,
		width:          80,
		height:         24,
	}
}

// SetSessions populates the dashboard cards from sessions.
func (m *Model) SetSessions(sessions []data.Session) {
	m.cards = make([]SessionCard, len(sessions))
	for i, s := range sessions {
		m.cards[i] = SessionCard{
			Session:    s,
			LastAction: DeriveLastAction(s),
		}
	}
	m.clampCursor()
}

// SetSize sets the available rendering dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetAnimFrame updates the animation frame counter.
func (m *Model) SetAnimFrame(frame int) {
	m.animFrame = frame
}

// MoveCursor moves the cursor by delta.
func (m *Model) MoveCursor(delta int) {
	if len(m.cards) == 0 {
		return
	}
	m.cursor += delta
	m.clampCursor()
}

// Cursor returns the current cursor position (for testing).
func (m *Model) Cursor() int {
	return m.cursor
}

// Cards returns the current cards (for testing).
func (m *Model) Cards() []SessionCard {
	return m.cards
}

// SelectedSession returns a pointer to the currently selected session, or nil.
func (m *Model) SelectedSession() *data.Session {
	if len(m.cards) == 0 || m.cursor < 0 || m.cursor >= len(m.cards) {
		return nil
	}
	s := m.cards[m.cursor].Session
	return &s
}

// View renders the mission control dashboard with pagination.
func (m *Model) View() string {
	if len(m.cards) == 0 {
		return m.titleStyle.Render("  No sessions to display")
	}

	header := m.titleStyle.Render("⚡ Mission Control") +
		lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("  %d sessions", len(m.cards)))

	// Each card is 3 lines (line1 + line2 + separator)
	cardHeight := 3
	availableHeight := m.height - 6 // header + footer chrome
	pageSize := availableHeight / cardHeight
	if pageSize < 2 {
		pageSize = 2
	}

	// Compute visible window around cursor
	start := m.cursor - pageSize/2
	if start < 0 {
		start = 0
	}
	end := start + pageSize
	if end > len(m.cards) {
		end = len(m.cards)
		start = end - pageSize
		if start < 0 {
			start = 0
		}
	}

	var lines []string
	lines = append(lines, header)
	lines = append(lines, "")

	if start > 0 {
		lines = append(lines, lipgloss.NewStyle().Faint(true).Render(
			fmt.Sprintf("  ↑ %d more above", start)))
	}

	for i := start; i < end; i++ {
		isSelected := i == m.cursor
		lines = append(lines, m.renderCard(m.cards[i], isSelected))
		if i < end-1 {
			sep := strings.Repeat("─", m.cardWidth())
			lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("238")).Render("  "+sep))
		}
	}

	if end < len(m.cards) {
		lines = append(lines, lipgloss.NewStyle().Faint(true).Render(
			fmt.Sprintf("  ↓ %d more below", len(m.cards)-end)))
	}

	return strings.Join(lines, "\n")
}

func (m *Model) cardWidth() int {
	w := m.width - 4
	if w < 40 {
		w = 40
	}
	return w
}

func (m *Model) renderCard(card SessionCard, selected bool) string {
	s := card.Session
	w := m.cardWidth()

	icon := m.statusIcon(s.Status)
	if m.animStatusIcon != nil {
		icon = m.animStatusIcon(s.Status, m.animFrame)
	}

	gutter := "  "
	if selected {
		gutter = "▎ "
	}

	// Line 1: gutter + icon + title + repo + duration (right-aligned)
	title := s.Title
	repo := s.Repository
	if repo == "" {
		repo = "local"
	}
	rightInfo := formatDuration(s)

	fixedLen := len(gutter) + runeWidth(icon) + 1 + 2 + len(repo) + 2 + len(rightInfo)
	maxTitle := w - fixedLen
	if maxTitle < 5 {
		maxTitle = 5
	}
	if len(title) > maxTitle {
		title = title[:maxTitle-1] + "…"
	}

	left := fmt.Sprintf("%s%s %s", gutter, icon, title)
	right := fmt.Sprintf("%s  %s", repo, rightInfo)
	padding := w - runeWidth(left) - len(right)
	if padding < 2 {
		padding = 2
	}
	line1 := left + strings.Repeat(" ", padding) + right

	// Line 2: action (indented, dimmed)
	action := card.LastAction
	maxAction := w - 4
	if maxAction < 10 {
		maxAction = 10
	}
	if len(action) > maxAction {
		action = action[:maxAction-1] + "…"
	}
	actionStyle := lipgloss.NewStyle().Faint(true)
	line2 := "    " + actionStyle.Render(action)

	cardText := line1 + "\n" + line2

	if selected {
		return m.cardSelStyle.Width(w + 4).Render(cardText)
	}
	return m.cardStyle.Width(w + 4).Render(cardText)
}

func (m *Model) clampCursor() {
	if len(m.cards) == 0 {
		m.cursor = 0
		return
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.cards) {
		m.cursor = len(m.cards) - 1
	}
}

// runeWidth is a simple approximation for display width.
func runeWidth(s string) int {
	return len([]rune(s))
}

// formatDuration returns a concise duration or status label for the session.
func formatDuration(s data.Session) string {
	status := strings.ToLower(strings.TrimSpace(s.Status))
	switch status {
	case "completed":
		return "done"
	case "failed":
		return "failed"
	case "queued":
		return "queued"
	}
	if s.CreatedAt.IsZero() {
		return ""
	}
	d := time.Since(s.CreatedAt)
	switch {
	case d < time.Minute:
		return "⏱ <1m"
	case d < time.Hour:
		return fmt.Sprintf("⏱ %dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("⏱ %dh", int(d.Hours()))
	default:
		return fmt.Sprintf("⏱ %dd", int(d.Hours()/24))
	}
}

// DeriveLastAction returns a brief description of what the session is currently doing.
func DeriveLastAction(s data.Session) string {
	status := strings.ToLower(strings.TrimSpace(s.Status))

	switch status {
	case "queued":
		return "⏳ Waiting to start"
	case "failed":
		return "❌ Session failed"
	case "completed":
		if s.PRNumber > 0 {
			return fmt.Sprintf("📤 PR #%d ready for review", s.PRNumber)
		}
		return "✅ Completed"
	case "needs-input":
		// For local sessions, try to get the last assistant message
		if s.Source == data.SourceLocalCopilot {
			if msg := data.FetchLastAssistantMessage(s.ID); msg != "" {
				truncated := msg
				if len(truncated) > 80 {
					truncated = truncated[:77] + "..."
				}
				return "❓ \"" + truncated + "\""
			}
		}
		return "🧑 Waiting for input"
	case "running":
		if s.Source == data.SourceLocalCopilot {
			if action := data.FetchLastSessionAction(s); action != "" {
				return action
			}
		}
		return "● Working..."
	default:
		return "● Working..."
	}
}
