package activeview

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/maxbeizer/gh-agent-viz/internal/data"
	"github.com/maxbeizer/gh-agent-viz/internal/tui/components/mission"
)

// Model represents the active sessions focused view.
type Model struct {
	sessions     []data.Session // filtered to active/needs-input/failed only
	allSessions  []data.Session // unfiltered, for "just finished" fallback
	cursor       int
	scrollOffset int
	width        int
	height       int
	dismissedIDs map[string]struct{}
	statusIcon     func(string) string
	animStatusIcon func(string, int) string
	animFrame      int
}

// New creates a new active sessions view model.
func New(
	statusIconFunc func(string) string,
	animStatusIconFunc func(string, int) string,
) Model {
	return Model{
		width:          80,
		height:         24,
		dismissedIDs:   make(map[string]struct{}),
		statusIcon:     statusIconFunc,
		animStatusIcon: animStatusIconFunc,
	}
}

// SetSize updates the available dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetAnimFrame updates the animation frame counter.
func (m *Model) SetAnimFrame(frame int) {
	m.animFrame = frame
}

// isActiveForView returns true for sessions that belong in this view:
// running, queued, active, needs-input, or failed.
func isActiveForView(s data.Session) bool {
	status := strings.ToLower(strings.TrimSpace(s.Status))
	return data.StatusIsActive(s.Status) ||
		status == "needs-input" ||
		status == "failed"
}

// SetSessions filters and stores sessions for display.
func (m *Model) SetSessions(sessions []data.Session) {
	m.allSessions = sessions
	m.sessions = make([]data.Session, 0)
	for _, s := range sessions {
		if _, dismissed := m.dismissedIDs[s.ID]; dismissed {
			continue
		}
		if isActiveForView(s) {
			m.sessions = append(m.sessions, s)
		}
	}
	// Sort: needs-input first, then failed, then by most recently updated.
	sort.SliceStable(m.sessions, func(i, j int) bool {
		si := strings.ToLower(strings.TrimSpace(m.sessions[i].Status))
		sj := strings.ToLower(strings.TrimSpace(m.sessions[j].Status))
		pi := statusPriority(si)
		pj := statusPriority(sj)
		if pi != pj {
			return pi < pj
		}
		return m.sessions[i].UpdatedAt.After(m.sessions[j].UpdatedAt)
	})
	m.clampCursor()
}

// statusPriority returns sort priority (lower = higher priority).
func statusPriority(status string) int {
	switch status {
	case "needs-input":
		return 0
	case "failed":
		return 1
	default:
		return 2
	}
}

// SessionCount returns the number of active sessions.
func (m *Model) SessionCount() int {
	return len(m.sessions)
}

// MoveCursor moves the cursor by delta, clamping to bounds.
func (m *Model) MoveCursor(delta int) {
	if len(m.sessions) == 0 {
		return
	}
	m.cursor += delta
	m.clampCursor()
	m.ensureCursorVisible()
}

// SelectedSession returns the session at the cursor, or nil.
func (m *Model) SelectedSession() *data.Session {
	if m.cursor < 0 || m.cursor >= len(m.sessions) {
		return nil
	}
	s := m.sessions[m.cursor]
	return &s
}

// DismissSelected removes the focused session from the view.
func (m *Model) DismissSelected() {
	s := m.SelectedSession()
	if s == nil {
		return
	}
	m.dismissedIDs[s.ID] = struct{}{}
	// Remove from slice.
	updated := make([]data.Session, 0, len(m.sessions)-1)
	for _, session := range m.sessions {
		if session.ID != s.ID {
			updated = append(updated, session)
		}
	}
	m.sessions = updated
	m.clampCursor()
}

func (m *Model) clampCursor() {
	if m.cursor >= len(m.sessions) {
		m.cursor = len(m.sessions) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

// unfocusedCardHeight is the number of lines per unfocused card (3 content + 1 blank separator).
const unfocusedCardHeight = 4

// focusedCardHeight is the number of lines per focused card (5 content + 1 blank separator).
const focusedCardHeight = 6

func (m *Model) visibleCardCount() int {
	if m.height <= 0 {
		return 1
	}
	// Rough estimate: assume mostly unfocused cards + one focused.
	return (m.height + unfocusedCardHeight - 1) / unfocusedCardHeight
}

func (m *Model) ensureCursorVisible() {
	visible := m.visibleCardCount()
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	}
	if m.cursor >= m.scrollOffset+visible {
		m.scrollOffset = m.cursor - visible + 1
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

// View renders the active sessions view.
func (m *Model) View() string {
	if len(m.sessions) == 0 {
		return m.viewEmpty()
	}

	dim := lipgloss.NewStyle().Faint(true)
	titleStyle := lipgloss.NewStyle().Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "24", Dark: "75"})

	// Header line
	countLabel := fmt.Sprintf("%d active", len(m.sessions))
	headerLine := titleStyle.Render("  ⚡ Active Sessions") + "  " + dim.Render(countLabel)

	var lines []string
	lines = append(lines, headerLine, "")

	maxCards := m.visibleCardCount()
	end := m.scrollOffset + maxCards + 1 // render one extra for partial visibility
	if end > len(m.sessions) {
		end = len(m.sessions)
	}

	for i := m.scrollOffset; i < end; i++ {
		focused := i == m.cursor
		card := m.renderCard(m.sessions[i], focused)
		lines = append(lines, card)
		lines = append(lines, "") // spacer between cards
	}

	// Scroll indicator
	if m.scrollOffset > 0 || end < len(m.sessions) {
		indicator := dim.Render(fmt.Sprintf("  %d/%d", m.cursor+1, len(m.sessions)))
		lines = append(lines, indicator)
	}

	return strings.Join(lines, "\n")
}

func (m *Model) renderCard(s data.Session, focused bool) string {
	dim := lipgloss.NewStyle().Faint(true)
	sessionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "236", Dark: "252"})
	boldStyle := lipgloss.NewStyle().Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "236", Dark: "252"})
	hintStyle := lipgloss.NewStyle().Faint(true).
		Foreground(lipgloss.AdaptiveColor{Light: "244", Dark: "241"})

	icon := m.statusIcon(s.Status)
	if m.animStatusIcon != nil && data.SessionIsActiveNotIdle(s) {
		icon = m.animStatusIcon(s.Status, m.animFrame)
	}

	gutter := "  "
	if focused {
		gutter = "▎ "
	}

	// Line 1: icon + title
	title := s.Title
	maxTitle := m.width - 10
	if maxTitle < 20 {
		maxTitle = 20
	}
	if len(title) > maxTitle {
		title = title[:maxTitle-1] + "…"
	}
	if focused {
		title = boldStyle.Render(title)
	} else {
		title = sessionStyle.Render(title)
	}
	line1 := fmt.Sprintf("%s%s %s", gutter, icon, title)

	// Line 2: repo • PR • elapsed • tokens
	var meta []string
	repo := shortRepo(s.Repository)
	if repo != "" {
		meta = append(meta, repo)
	}
	if s.PRNumber > 0 {
		meta = append(meta, fmt.Sprintf("PR #%d", s.PRNumber))
	}
	if !s.CreatedAt.IsZero() {
		meta = append(meta, formatDuration(time.Since(s.CreatedAt)))
	}
	if s.Telemetry != nil && s.Telemetry.InputTokens > 0 {
		meta = append(meta, data.FormatTokenCount(s.Telemetry.InputTokens)+" tokens")
	}
	line2 := gutter + "   " + dim.Render(strings.Join(meta, "  •  "))

	// Line 3: last action
	action := mission.DeriveLastAction(s)
	line3 := gutter + "   " + dim.Render("▸ "+action)

	lines := []string{line1, line2, line3}

	// Expanded lines for focused card
	if focused {
		var extras []string
		if s.Branch != "" && !data.IsDefaultBranch(s.Branch) {
			extras = append(extras, "branch: "+s.Branch)
		}
		if s.Telemetry != nil && s.Telemetry.Model != "" {
			extras = append(extras, "model: "+s.Telemetry.Model)
		}
		if len(extras) > 0 {
			line4 := gutter + "   " + dim.Render(strings.Join(extras, "  •  "))
			lines = append(lines, line4)
		}

		hints := hintStyle.Render(gutter + "   [enter] details  [o] open PR  [l] logs  [c] copy ID  [x] dismiss")
		lines = append(lines, hints)
	}

	return strings.Join(lines, "\n")
}

func (m *Model) viewEmpty() string {
	dim := lipgloss.NewStyle().Faint(true)
	titleStyle := lipgloss.NewStyle().Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "24", Dark: "75"})
	sessionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "236", Dark: "252"})

	var lines []string
	lines = append(lines, titleStyle.Render("  ⚡ Active Sessions"))
	lines = append(lines, "")
	lines = append(lines, dim.Render("  All quiet — no active sessions ✨"))
	lines = append(lines, "")

	// Show last 2-3 completions as "just finished"
	recent := m.recentCompletions(3)
	if len(recent) > 0 {
		lines = append(lines, dim.Render("  Recently finished:"))
		for _, s := range recent {
			icon := "✅"
			if strings.EqualFold(s.Status, "failed") {
				icon = "❌"
			}
			title := s.Title
			if len(title) > 50 {
				title = title[:47] + "…"
			}
			ago := formatAge(s.UpdatedAt)
			pr := ""
			if s.PRNumber > 0 {
				pr = dim.Render(fmt.Sprintf(" PR #%d", s.PRNumber))
			}
			lines = append(lines, fmt.Sprintf("  %s %s%s  %s",
				icon, sessionStyle.Render(title), pr, dim.Render(ago)))
		}
	}

	return strings.Join(lines, "\n")
}

func (m *Model) recentCompletions(n int) []data.Session {
	var completed []data.Session
	for _, s := range m.allSessions {
		status := strings.ToLower(strings.TrimSpace(s.Status))
		if status == "completed" || status == "failed" {
			completed = append(completed, s)
		}
	}
	sort.SliceStable(completed, func(i, j int) bool {
		return completed[i].UpdatedAt.After(completed[j].UpdatedAt)
	})
	if len(completed) > n {
		completed = completed[:n]
	}
	return completed
}

// shortRepo trims "github.com/" prefix and returns just owner/repo.
func shortRepo(repo string) string {
	repo = strings.TrimSpace(repo)
	if repo == "" {
		return "local"
	}
	repo = strings.TrimPrefix(repo, "https://github.com/")
	repo = strings.TrimPrefix(repo, "github.com/")
	return repo
}

// formatDuration formats a duration as "2m", "1h23m", "3d", etc.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "<1m"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	h := int(d.Hours())
	mins := int(d.Minutes()) - h*60
	if h < 24 {
		if mins > 0 {
			return fmt.Sprintf("%dh%dm", h, mins)
		}
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dd", h/24)
}

// formatAge returns a human-readable "3m ago" style string.
func formatAge(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return formatDuration(time.Since(t)) + " ago"
}
