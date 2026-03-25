package activeview

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
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

// cardLineCount returns the number of lines a card will consume (including trailing spacer).
func (m *Model) cardLineCount(s data.Session, focused bool) int {
	lines := 3 // title + meta + action
	if focused {
		// Extra detail line (time/tokens/model) if any available
		hasTime := !s.CreatedAt.IsZero()
		hasTokens := s.Telemetry != nil && s.Telemetry.InputTokens > 0
		hasModel := s.Telemetry != nil && s.Telemetry.Model != ""
		if hasTime || hasTokens || hasModel {
			lines++
		}
		lines++ // action hints line
	}
	lines++ // spacer
	return lines
}

// headerLines is the number of lines consumed by the view header (title + blank).
const headerLines = 2

// footerReserve is lines reserved for scroll indicator + breathing room.
const footerReserve = 2

func (m *Model) ensureCursorVisible() {
	if len(m.sessions) == 0 {
		m.scrollOffset = 0
		return
	}

	// Scroll up if cursor is above viewport
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	}

	// Scroll down if cursor is below viewport — walk cards from scrollOffset
	// and see if the cursor card fits.
	for {
		budget := m.height - headerLines - footerReserve
		used := 0
		cursorVisible := false
		for i := m.scrollOffset; i < len(m.sessions); i++ {
			h := m.cardLineCount(m.sessions[i], i == m.cursor)
			if used+h > budget && used > 0 {
				break
			}
			used += h
			if i == m.cursor {
				cursorVisible = true
				break
			}
		}
		if cursorVisible {
			break
		}
		m.scrollOffset++
		if m.scrollOffset >= len(m.sessions) {
			m.scrollOffset = len(m.sessions) - 1
			break
		}
	}

	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

// View renders the active sessions view, constrained to the terminal height.
func (m *Model) View() string {
	if len(m.sessions) == 0 {
		return m.viewEmpty()
	}

	dim := lipgloss.NewStyle().Faint(true)
	titleStyle := lipgloss.NewStyle().Bold(true).
		Foreground(compat.AdaptiveColor{Light: lipgloss.Color("24"), Dark: lipgloss.Color("75")})

	budget := m.height - headerLines - footerReserve
	if budget < 4 {
		budget = 4
	}

	// Header
	countLabel := fmt.Sprintf("%d active", len(m.sessions))
	headerLine := titleStyle.Render("  ⚡ Active Sessions") + "  " + dim.Render(countLabel)

	var lines []string
	lines = append(lines, headerLine, "")

	// Render cards until we run out of vertical budget
	used := 0
	lastRendered := m.scrollOffset - 1
	for i := m.scrollOffset; i < len(m.sessions); i++ {
		focused := i == m.cursor
		h := m.cardLineCount(m.sessions[i], focused)
		if used+h > budget && used > 0 {
			break
		}
		card := m.renderCard(m.sessions[i], focused)
		lines = append(lines, card)
		lines = append(lines, "") // spacer
		used += h
		lastRendered = i
	}

	// Scroll position indicator
	above := m.scrollOffset
	below := len(m.sessions) - lastRendered - 1
	if above > 0 || below > 0 {
		parts := []string{fmt.Sprintf("  %d/%d", m.cursor+1, len(m.sessions))}
		if above > 0 {
			parts = append(parts, fmt.Sprintf("↑%d above", above))
		}
		if below > 0 {
			parts = append(parts, fmt.Sprintf("↓%d below", below))
		}
		lines = append(lines, dim.Render(strings.Join(parts, "  ")))
	}

	return strings.Join(lines, "\n")
}

func (m *Model) renderCard(s data.Session, focused bool) string {
	dim := lipgloss.NewStyle().Faint(true)
	sessionStyle := lipgloss.NewStyle().
		Foreground(compat.AdaptiveColor{Light: lipgloss.Color("236"), Dark: lipgloss.Color("252")})
	boldStyle := lipgloss.NewStyle().Bold(true).
		Foreground(compat.AdaptiveColor{Light: lipgloss.Color("236"), Dark: lipgloss.Color("252")})
	hintStyle := lipgloss.NewStyle().Faint(true).
		Foreground(compat.AdaptiveColor{Light: lipgloss.Color("244"), Dark: lipgloss.Color("241")})

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

	// Line 2: repo • branch • PR
	var meta []string
	repo := shortRepo(s.Repository)
	if repo != "" {
		meta = append(meta, repo)
	}
	if s.Branch != "" && !data.IsDefaultBranch(s.Branch) {
		branch := s.Branch
		maxBranch := m.width/3
		if maxBranch < 20 { maxBranch = 20 }
		if len(branch) > maxBranch {
			branch = branch[:maxBranch-1] + "…"
		}
		meta = append(meta, branch)
	}
	if s.WorkDir != "" && s.Branch == "" {
		meta = append(meta, s.WorkDir)
	}
	if s.PRNumber > 0 {
		meta = append(meta, fmt.Sprintf("PR #%d", s.PRNumber))
	}
	line2 := gutter + "   " + dim.Render(strings.Join(meta, "  •  "))

	// Line 3: last action
	action := mission.DeriveLastAction(s)
	line3 := gutter + "   " + dim.Render("▸ "+action)

	lines := []string{line1, line2, line3}

	// Expanded lines for focused card
	if focused {
		var extras []string
		if !s.CreatedAt.IsZero() {
			extras = append(extras, formatDuration(time.Since(s.CreatedAt)))
		}
		if s.Telemetry != nil && s.Telemetry.InputTokens > 0 {
			extras = append(extras, data.FormatTokenCount(s.Telemetry.InputTokens)+" tokens")
		}
		if s.Telemetry != nil && s.Telemetry.Model != "" {
			extras = append(extras, s.Telemetry.Model)
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
		Foreground(compat.AdaptiveColor{Light: lipgloss.Color("24"), Dark: lipgloss.Color("75")})
	sessionStyle := lipgloss.NewStyle().
		Foreground(compat.AdaptiveColor{Light: lipgloss.Color("236"), Dark: lipgloss.Color("252")})

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
