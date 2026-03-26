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

// Model represents the active sessions focused view (lazygit-style split panels).
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
// actively working (not idle), needs-input, or recently failed.
// Idle sessions (running but stale >20min) are excluded — this view is
// "what are my agents doing right now."
func isActiveForView(s data.Session) bool {
	status := strings.ToLower(strings.TrimSpace(s.Status))
	if status == "needs-input" || status == "failed" {
		return true
	}
	// Only include active sessions that are actually working (not idle)
	return data.SessionIsActiveNotIdle(s)
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

// listItemHeight is the number of lines per session row (2 content lines).
const listItemHeight = 2

func (m *Model) visibleListItems(listHeight int) int {
	if listHeight <= 0 {
		return 1
	}
	return listHeight / listItemHeight
}

func (m *Model) ensureCursorVisible() {
	if len(m.sessions) == 0 {
		m.scrollOffset = 0
		return
	}
	// Panel chrome: 2 border lines + 1 title line
	listHeight := m.panelContentHeight()
	visible := m.visibleListItems(listHeight)
	if visible < 1 {
		visible = 1
	}
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

// panelContentHeight returns usable lines inside a bordered panel.
func (m *Model) panelContentHeight() int {
	// Total height minus: 1 hint bar line, 2 border lines per panel
	h := m.height - 3
	if h < 4 {
		h = 4
	}
	return h
}

// useHorizontalLayout returns true when terminal is wide enough for side-by-side.
func (m *Model) useHorizontalLayout() bool {
	return m.width >= 100
}

// ── Styles ──

var (
	borderColor    = compat.AdaptiveColor{Light: lipgloss.Color("249"), Dark: lipgloss.Color("238")}
	titleColor     = compat.AdaptiveColor{Light: lipgloss.Color("24"), Dark: lipgloss.Color("75")}
	dimColor       = compat.AdaptiveColor{Light: lipgloss.Color("244"), Dark: lipgloss.Color("241")}
	textColor      = compat.AdaptiveColor{Light: lipgloss.Color("236"), Dark: lipgloss.Color("252")}
	selectedBg     = compat.AdaptiveColor{Light: lipgloss.Color("254"), Dark: lipgloss.Color("236")}
	labelColor     = compat.AdaptiveColor{Light: lipgloss.Color("242"), Dark: lipgloss.Color("245")}
	urgentColor    = compat.AdaptiveColor{Light: lipgloss.Color("196"), Dark: lipgloss.Color("203")}
	activeColor    = compat.AdaptiveColor{Light: lipgloss.Color("28"), Dark: lipgloss.Color("42")}
)

// ── View ──

func (m *Model) View() string {
	if len(m.sessions) == 0 {
		return m.viewEmpty()
	}

	if m.useHorizontalLayout() {
		return m.viewHorizontal()
	}
	return m.viewVertical()
}

// statusBreakdown returns a panel title like " 3 running · 2 failed · 1 waiting "
func (m *Model) statusBreakdown() string {
	var running, failed, waiting, queued, other int
	for _, s := range m.sessions {
		switch strings.ToLower(strings.TrimSpace(s.Status)) {
		case "running", "active":
			running++
		case "failed":
			failed++
		case "needs-input":
			waiting++
		case "queued":
			queued++
		default:
			other++
		}
	}
	var parts []string
	if running > 0 {
		parts = append(parts, fmt.Sprintf("%d running", running))
	}
	if queued > 0 {
		parts = append(parts, fmt.Sprintf("%d queued", queued))
	}
	if waiting > 0 {
		parts = append(parts, fmt.Sprintf("%d waiting", waiting))
	}
	if failed > 0 {
		parts = append(parts, fmt.Sprintf("%d failed", failed))
	}
	if other > 0 {
		parts = append(parts, fmt.Sprintf("%d other", other))
	}
	if len(parts) == 0 {
		return " Sessions "
	}
	return " " + strings.Join(parts, " · ") + " "
}
func (m *Model) viewHorizontal() string {
	totalW := m.width - 1
	listW := totalW * 40 / 100
	if listW < 30 {
		listW = 30
	}
	detailW := totalW - listW

	contentH := m.panelContentHeight()
	listPanel := m.renderListPanel(listW, contentH)
	detailPanel := m.renderDetailPanel(detailW, contentH)

	main := lipgloss.JoinHorizontal(lipgloss.Top, listPanel, detailPanel)
	return main + "\n" + m.renderHintBar()
}

func (m *Model) viewVertical() string {
	contentH := m.panelContentHeight()
	listH := contentH * 40 / 100
	if listH < 4 {
		listH = 4
	}
	detailH := contentH - listH

	listPanel := m.renderListPanel(m.width, listH)
	detailPanel := m.renderDetailPanel(m.width, detailH)

	return listPanel + "\n" + detailPanel + "\n" + m.renderHintBar()
}

func (m *Model) renderListPanel(width, contentHeight int) string {
	panelTitle := lipgloss.NewStyle().Bold(true).Foreground(titleColor)
	dim := lipgloss.NewStyle().Foreground(dimColor)

	title := m.statusBreakdown()

	innerW := width - 4 // border + padding
	if innerW < 10 {
		innerW = 10
	}
	visible := m.visibleListItems(contentHeight)
	end := m.scrollOffset + visible
	if end > len(m.sessions) {
		end = len(m.sessions)
	}

	var rows []string
	for i := m.scrollOffset; i < end; i++ {
		s := m.sessions[i]
		selected := i == m.cursor
		rows = append(rows, m.renderListItem(s, selected, innerW))
	}

	// Pad remaining space with empty lines
	rendered := len(rows) * listItemHeight
	for rendered < contentHeight {
		rows = append(rows, "")
		rendered++
	}

	// Scroll indicator in title
	scrollInfo := ""
	if len(m.sessions) > visible {
		scrollInfo = dim.Render(fmt.Sprintf(" %d/%d", m.cursor+1, len(m.sessions)))
	}

	content := strings.Join(rows, "\n")
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width - 2).
		Height(contentHeight).
		Render(content)

	return panelTitle.Render(title) + scrollInfo + "\n" + box
}

func (m *Model) renderListItem(s data.Session, selected bool, innerW int) string {
	dim := lipgloss.NewStyle().Foreground(dimColor)
	text := lipgloss.NewStyle().Foreground(textColor)

	icon := m.statusIcon(s.Status)
	if m.animStatusIcon != nil && data.SessionIsActiveNotIdle(s) {
		icon = m.animStatusIcon(s.Status, m.animFrame)
	}

	// Line 1: icon + title
	title := s.Title
	maxTitle := innerW - 4
	if maxTitle < 10 {
		maxTitle = 10
	}
	if len(title) > maxTitle {
		title = title[:maxTitle-1] + "…"
	}

	line1Style := text
	if selected {
		line1Style = lipgloss.NewStyle().Bold(true).Foreground(textColor).Background(selectedBg)
	}
	line1 := fmt.Sprintf(" %s %s", icon, line1Style.Render(title))
	if selected {
		// Pad to full width for highlight bar
		pad := innerW - lipgloss.Width(line1)
		if pad > 0 {
			line1 += line1Style.Render(strings.Repeat(" ", pad))
		}
	}

	// Line 2: repo • branch
	var meta []string
	repo := shortRepo(s.Repository)
	if repo != "" {
		meta = append(meta, repo)
	}
	if s.Branch != "" && !data.IsDefaultBranch(s.Branch) {
		branch := s.Branch
		maxB := innerW / 2
		if maxB < 15 {
			maxB = 15
		}
		if len(branch) > maxB {
			branch = branch[:maxB-1] + "…"
		}
		meta = append(meta, branch)
	}
	line2 := "   " + dim.Render(strings.Join(meta, " • "))

	return line1 + "\n" + line2
}

func (m *Model) renderDetailPanel(width, contentHeight int) string {
	panelTitle := lipgloss.NewStyle().Bold(true).Foreground(titleColor)
	title := " Detail "

	s := m.SelectedSession()
	var content string
	if s == nil {
		content = lipgloss.NewStyle().Foreground(dimColor).Render(" No session selected")
	} else {
		content = m.renderDetail(*s, width-4, contentHeight)
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width - 2).
		Height(contentHeight).
		Render(content)

	return panelTitle.Render(title) + "\n" + box
}

func (m *Model) renderDetail(s data.Session, innerW, maxLines int) string {
	dim := lipgloss.NewStyle().Foreground(dimColor)
	label := lipgloss.NewStyle().Foreground(labelColor)
	text := lipgloss.NewStyle().Foreground(textColor)
	urgent := lipgloss.NewStyle().Bold(true).Foreground(urgentColor)
	active := lipgloss.NewStyle().Foreground(activeColor)

	var lines []string

	// Status line with icon
	icon := m.statusIcon(s.Status)
	if m.animStatusIcon != nil && data.SessionIsActiveNotIdle(s) {
		icon = m.animStatusIcon(s.Status, m.animFrame)
	}
	statusStyle := active
	st := strings.ToLower(strings.TrimSpace(s.Status))
	if st == "needs-input" || st == "failed" {
		statusStyle = urgent
	}
	lines = append(lines, fmt.Sprintf(" %s %s", icon, statusStyle.Render(s.Status)))
	lines = append(lines, "")

	// Metadata fields
	addField := func(lbl, val string) {
		if val != "" && len(lines) < maxLines-2 {
			lines = append(lines, fmt.Sprintf(" %s %s", label.Render(lbl), text.Render(val)))
		}
	}

	addField("repo:", shortRepo(s.Repository))
	if s.Branch != "" && !data.IsDefaultBranch(s.Branch) {
		addField("branch:", s.Branch)
	}
	if s.WorkDir != "" {
		addField("workdir:", s.WorkDir)
	}
	if s.PRNumber > 0 {
		addField("PR:", fmt.Sprintf("#%d", s.PRNumber))
	}
	if !s.CreatedAt.IsZero() {
		addField("elapsed:", formatDuration(time.Since(s.CreatedAt)))
	}
	if s.Telemetry != nil {
		if s.Telemetry.Model != "" {
			addField("model:", s.Telemetry.Model)
		}
		if s.Telemetry.InputTokens > 0 {
			addField("tokens:", data.FormatTokenCount(s.Telemetry.InputTokens)+" in / "+data.FormatTokenCount(s.Telemetry.OutputTokens)+" out")
		}
	}

	lines = append(lines, "")

	// Current activity
	action := mission.DeriveLastAction(s)
	lines = append(lines, " "+label.Render("activity:"))
	lines = append(lines, " "+text.Render(action))
	lines = append(lines, "")

	// Log tail — try to fill remaining space
	remaining := maxLines - len(lines)
	if remaining > 2 && s.Source == data.SourceLocalCopilot {
		lines = append(lines, " "+label.Render("recent log:"))
		logLines := m.fetchLogTail(s, remaining-1)
		if len(logLines) > 0 {
			for _, l := range logLines {
				if len(l) > innerW-2 {
					l = l[:innerW-3] + "…"
				}
				lines = append(lines, " "+dim.Render(l))
			}
		} else {
			lines = append(lines, " "+dim.Render("(no log data)"))
		}
	}

	return strings.Join(lines, "\n")
}

// fetchLogTail returns the last N meaningful events from the session log.
func (m *Model) fetchLogTail(s data.Session, maxLines int) []string {
	events, err := data.FetchSessionEvents(s.ID)
	if err != nil || len(events) == 0 {
		return nil
	}

	// Collect the last meaningful events
	var entries []string
	for _, ev := range events {
		var line string
		switch ev.Type {
		case "tool.execution_start":
			line = "🔧 " + ev.ToolName
		case "tool.execution_end":
			line = "✓ " + ev.ToolName + " done"
		case "assistant.message":
			msg := ev.Content
			if len(msg) > 80 {
				msg = msg[:77] + "..."
			}
			if msg != "" {
				line = "💬 " + msg
			}
		case "user.message":
			msg := ev.Content
			if len(msg) > 60 {
				msg = msg[:57] + "..."
			}
			if msg != "" {
				line = "👤 " + msg
			}
		}
		if line != "" {
			entries = append(entries, line)
		}
	}

	// Return the tail
	if len(entries) > maxLines {
		entries = entries[len(entries)-maxLines:]
	}
	return entries
}

func (m *Model) renderHintBar() string {
	hint := lipgloss.NewStyle().Foreground(dimColor)
	key := lipgloss.NewStyle().Bold(true).Foreground(textColor)

	hints := []string{
		key.Render("↑↓") + " navigate",
		key.Render("enter") + " details",
		key.Render("o") + " open PR",
		key.Render("l") + " logs",
		key.Render("c") + " copy ID",
		key.Render("x") + " dismiss",
		key.Render("esc") + " back",
	}

	joined := strings.Join(hints, hint.Render("  │  "))
	return " " + joined
}

func (m *Model) viewEmpty() string {
	dim := lipgloss.NewStyle().Foreground(dimColor)
	panelTitle := lipgloss.NewStyle().Bold(true).Foreground(titleColor)
	text := lipgloss.NewStyle().Foreground(textColor)

	var content []string
	content = append(content, "")
	content = append(content, dim.Render(" All quiet — no active sessions ✨"))
	content = append(content, "")

	recent := m.recentCompletions(3)
	if len(recent) > 0 {
		content = append(content, " "+lipgloss.NewStyle().Foreground(labelColor).Render("Just finished:"))
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
			content = append(content, fmt.Sprintf(" %s %s%s  %s", icon, text.Render(title), pr, dim.Render(ago)))
		}
	}

	contentH := m.panelContentHeight()
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(m.width - 2).
		Height(contentH).
		Render(strings.Join(content, "\n"))

	return panelTitle.Render(" Active Sessions ") + "\n" + box + "\n" + m.renderHintBar()
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

func shortRepo(repo string) string {
	repo = strings.TrimSpace(repo)
	if repo == "" {
		return "local"
	}
	repo = strings.TrimPrefix(repo, "https://github.com/")
	repo = strings.TrimPrefix(repo, "github.com/")
	return repo
}

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

func formatAge(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return formatDuration(time.Since(t)) + " ago"
}
