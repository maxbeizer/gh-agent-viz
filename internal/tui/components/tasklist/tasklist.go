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
	titleStyle        lipgloss.Style
	tableHeaderStyle  lipgloss.Style
	tableRowStyle     lipgloss.Style
	tableRowSelected  lipgloss.Style
	sessions          []data.Session
	deEmphasizedIdx   map[int]struct{}
	rowCursor         int
	loading           bool
	statusIcon        func(string) string
	selectedSessionID string
	width             int
	height            int
}

// New creates a new task list model
func New(titleStyle, headerStyle, rowStyle, rowSelectedStyle lipgloss.Style, statusIconFunc func(string) string) Model {
	return Model{
		titleStyle:       titleStyle,
		tableHeaderStyle: headerStyle,
		tableRowStyle:    rowStyle,
		tableRowSelected: rowSelectedStyle,
		sessions:         []data.Session{},
		deEmphasizedIdx:  map[int]struct{}{},
		rowCursor:        0,
		loading:          false,
		statusIcon:       statusIconFunc,
		width:            80,
		height:           24,
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

// SetLoading sets the loading state, clearing the list for visual feedback
func (m *Model) SetLoading(loading bool) {
	m.loading = loading
}

// View renders the sessions as a focused single-column list
func (m Model) View() string {
	if m.loading {
		return m.titleStyle.Render("🔄 Switching gears...\n\nFetching sessions, one moment.")
	}

	if len(m.sessions) == 0 {
		return m.titleStyle.Render("✨ All quiet on the agent front.\n\nNo sessions match this filter — your agents are either napping or haven't checked in yet.\nPress 'r' to refresh, or tab to try another filter.")
	}

	list := m.renderFocusedList()
	detail := m.renderInlineDetail()

	if m.height <= 20 {
		return list
	}

	return lipgloss.JoinVertical(lipgloss.Left, list, "", detail)
}

func (m Model) renderFocusedList() string {
	usableWidth := m.width - 4
	if usableWidth < 20 {
		usableWidth = 20
	}

	rows := []string{}

	cursor := m.rowCursor
	if cursor >= len(m.sessions) {
		cursor = len(m.sessions) - 1
	}
	if cursor < 0 {
		cursor = 0
	}

	pageSize := m.pageSize()
	start, end := visibleRange(len(m.sessions), cursor, pageSize)

	if start > 0 {
		rows = append(rows, m.tableRowStyle.Render(
			fmt.Sprintf("  ↑ %d more above", start)))
	}

	for i := start; i < end; i++ {
		session := m.sessions[i]
		selected := i == cursor
		rows = append(rows, m.renderRow(i, session, selected, usableWidth))
	}

	if end < len(m.sessions) {
		rows = append(rows, m.tableRowStyle.Render(
			fmt.Sprintf("  ↓ %d more below", len(m.sessions)-end)))
	}

	return strings.Join(rows, "\n")
}

func (m Model) renderInlineDetail() string {
	selected := m.SelectedTask()
	if selected == nil {
		return ""
	}

	actions := []string{"enter details"}
	if selected.Source == data.SourceAgentTask {
		actions = append(actions, "l logs")
		if sessionHasLinkedPR(*selected) {
			actions = append(actions, "o open PR")
		}
	}
	if selected.Source == data.SourceLocalCopilot && isActiveStatus(selected.Status) && selected.ID != "" {
		actions = append(actions, "s resume")
	}

	maxW := m.width - 6
	if maxW < 20 {
		maxW = 20
	}

	lines := []string{
		fmt.Sprintf("  %s %s", m.statusIcon(selected.Status), truncate(sessionTitle(*selected), maxW)),
		fmt.Sprintf("  Status: %s  •  Needs action: %s", selected.Status, attentionReason(*selected)),
		fmt.Sprintf("  Repository: %s  •  Branch: %s", panelRepository(*selected), panelBranch(*selected)),
		fmt.Sprintf("  Source: %s  •  Last update: %s", sourceLabel(selected.Source), formatTime(selected.UpdatedAt)),
		fmt.Sprintf("  Actions: %s", strings.Join(actions, " • ")),
	}
	if selected.Source == data.SourceAgentTask && selected.PRNumber > 0 {
		lines = append(lines, fmt.Sprintf("  Pull Request: #%d", selected.PRNumber))
	}

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1).
		Render(strings.Join(lines, "\n"))
}

func (m Model) renderRow(sessionIdx int, session data.Session, selected bool, width int) string {
	style := m.tableRowStyle
	if selected {
		style = m.tableRowSelected
	} else if m.isDeEmphasized(sessionIdx) {
		style = style.Faint(true)
	}

	icon := m.statusIcon(session.Status)
	titleMax := width - 8
	if titleMax < 3 {
		titleMax = 3
	}
	title := truncate(sessionTitle(session), titleMax)
	badge := sessionBadge(session, m.isDeEmphasized(sessionIdx))

	// Gutter indicator: selected row gets a bar, others get a space
	gutter := "  "
	if selected {
		gutter = "▎ "
	}

	titleLine := fmt.Sprintf("%s%s %s", gutter, icon, title)
	if badge != "" {
		titleLine += " " + badge
	}

	metaMax := width - 8
	if metaMax < 3 {
		metaMax = 3
	}

	if width < 40 {
		return style.Render(titleLine)
	}

	repo := truncate(rowRepository(session), metaMax)
	attention := attentionReason(session)
	meta := fmt.Sprintf("    %s • %s • %s", repo, attention, formatTime(session.UpdatedAt))

	return style.Render(titleLine + "\n" + meta)
}

// SetTasks updates sessions with sorting and de-emphasis
func (m *Model) SetTasks(sessions []data.Session) {
	if selected := m.SelectedTask(); selected != nil {
		m.selectedSessionID = selected.ID
	}

	type rankedSession struct {
		session      data.Session
		deEmphasized bool
		sortPriority int
	}

	deEmphasizedInputIdx := quietDuplicateIndices(sessions)
	ranked := make([]rankedSession, 0, len(sessions))
	for i, session := range sessions {
		_, deEmphasized := deEmphasizedInputIdx[i]
		ranked = append(ranked, rankedSession{
			session:      session,
			deEmphasized: deEmphasized,
			sortPriority: sessionSortPriority(session, deEmphasized),
		})
	}

	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].sortPriority != ranked[j].sortPriority {
			return ranked[i].sortPriority < ranked[j].sortPriority
		}
		return ranked[i].session.UpdatedAt.After(ranked[j].session.UpdatedAt)
	})

	m.sessions = make([]data.Session, len(ranked))
	m.deEmphasizedIdx = map[int]struct{}{}
	for i, candidate := range ranked {
		m.sessions[i] = candidate.session
		if candidate.deEmphasized {
			m.deEmphasizedIdx[i] = struct{}{}
		}
	}

	// Clamp cursor
	if m.rowCursor >= len(m.sessions) {
		m.rowCursor = len(m.sessions) - 1
	}
	if m.rowCursor < 0 {
		m.rowCursor = 0
	}

	// Restore selection by ID
	if m.selectedSessionID != "" {
		for i, session := range m.sessions {
			if session.ID == m.selectedSessionID {
				m.rowCursor = i
				return
			}
		}
	}
}

// MoveCursor moves the cursor up/down in the focused list
func (m *Model) MoveCursor(delta int) {
	if len(m.sessions) == 0 {
		m.rowCursor = 0
		return
	}

	m.rowCursor += delta
	if m.rowCursor < 0 {
		m.rowCursor = 0
	}
	if m.rowCursor >= len(m.sessions) {
		m.rowCursor = len(m.sessions) - 1
	}
}

// MoveColumn is a no-op in focused list mode (kept for interface compat)
func (m *Model) MoveColumn(delta int) {
	// No columns in focused list mode
}

// SelectedTask returns the currently selected session
func (m Model) SelectedTask() *data.Session {
	if len(m.sessions) == 0 {
		return nil
	}

	cursor := m.rowCursor
	if cursor >= len(m.sessions) {
		cursor = len(m.sessions) - 1
	}
	if cursor < 0 {
		cursor = 0
	}

	return &m.sessions[cursor]
}

func truncate(s string, maxLen int) string {
	if maxLen < 3 {
		maxLen = 3
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "not recorded"
	}

	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	}
	return t.Format("Jan 2")
}

func sourceLabel(source data.SessionSource) string {
	switch source {
	case data.SourceLocalCopilot:
		return "local"
	case data.SourceAgentTask:
		return "agent"
	default:
		return "other"
	}
}

func isActiveStatus(status string) bool {
	return data.StatusIsActive(status) || strings.EqualFold(strings.TrimSpace(status), "needs-input")
}

func sessionBadge(session data.Session, deEmphasized bool) string {
	if strings.EqualFold(strings.TrimSpace(session.Status), "needs-input") {
		return "🧑 waiting on you"
	}
	if strings.EqualFold(strings.TrimSpace(session.Status), "failed") {
		return "🚨 failed"
	}
	if deEmphasized {
		return "↺ quiet duplicate"
	}
	if data.SessionNeedsAttention(session) {
		return "⚠ check progress"
	}
	if !isActiveStatus(session.Status) || session.UpdatedAt.IsZero() {
		return ""
	}
	return "• in progress"
}

func (m Model) isDeEmphasized(sessionIdx int) bool {
	_, ok := m.deEmphasizedIdx[sessionIdx]
	return ok
}

func sessionSortPriority(session data.Session, deEmphasized bool) int {
	status := strings.ToLower(strings.TrimSpace(session.Status))
	switch {
	case status == "needs-input" || status == "failed":
		return 0
	case deEmphasized:
		return 4
	case data.SessionNeedsAttention(session):
		return 1
	case isActiveStatus(session.Status):
		return 2
	default:
		return 3
	}
}

func quietDuplicateIndices(sessions []data.Session) map[int]struct{} {
	grouped := map[string][]int{}
	for i, session := range sessions {
		if !isQuietDuplicateSession(session) {
			continue
		}
		key := quietDuplicateKey(session)
		grouped[key] = append(grouped[key], i)
	}

	result := map[int]struct{}{}
	for _, indexes := range grouped {
		if len(indexes) < 2 {
			continue
		}

		sort.SliceStable(indexes, func(i, j int) bool {
			return sessions[indexes[i]].UpdatedAt.After(sessions[indexes[j]].UpdatedAt)
		})

		for _, idx := range indexes[1:] {
			result[idx] = struct{}{}
		}
	}

	return result
}

func isQuietDuplicateSession(session data.Session) bool {
	status := strings.ToLower(strings.TrimSpace(session.Status))
	if status == "needs-input" || status == "failed" {
		return false
	}
	return data.SessionNeedsAttention(session)
}

func quietDuplicateKey(session data.Session) string {
	return fmt.Sprintf(
		"%s|%s|%s|%s",
		strings.ToLower(strings.TrimSpace(sessionTitle(session))),
		strings.ToLower(strings.TrimSpace(session.Repository)),
		strings.ToLower(strings.TrimSpace(session.Branch)),
		strings.ToLower(strings.TrimSpace(string(session.Source))),
	)
}

func sessionTitle(session data.Session) string {
	if strings.TrimSpace(session.Title) == "" {
		return "Untitled Session"
	}
	return session.Title
}

func rowRepository(session data.Session) string {
	repository := strings.TrimSpace(session.Repository)
	branch := strings.TrimSpace(session.Branch)
	if repository == "" {
		repository = "not available"
	}
	if branch == "" {
		return repository
	}
	return fmt.Sprintf("%s @ %s", repository, branch)
}

func panelRepository(session data.Session) string {
	repository := strings.TrimSpace(session.Repository)
	if repository == "" {
		return "not available"
	}
	return repository
}

func panelBranch(session data.Session) string {
	branch := strings.TrimSpace(session.Branch)
	if branch == "" {
		return "not available"
	}
	return branch
}

func sessionHasLinkedPR(session data.Session) bool {
	if session.Source != data.SourceAgentTask {
		return false
	}
	if strings.TrimSpace(session.PRURL) != "" {
		return true
	}
	return session.PRNumber > 0 && strings.TrimSpace(session.Repository) != ""
}

func attentionReason(session data.Session) string {
	status := strings.ToLower(strings.TrimSpace(session.Status))
	switch {
	case status == "needs-input":
		return "waiting on your input"
	case status == "failed":
		return "run failed"
	case data.SessionNeedsAttention(session):
		return "running but quiet"
	case isActiveStatus(session.Status):
		return "in progress"
	default:
		return "no action needed"
	}
}

func rowContext(session data.Session) string {
	switch session.Source {
	case data.SourceLocalCopilot:
		if strings.EqualFold(strings.TrimSpace(session.Status), "needs-input") {
			return "waiting for your response • press 's' to resume"
		}
		if strings.EqualFold(strings.TrimSpace(session.Status), "failed") {
			return "failed session • inspect logs or retry"
		}
		if data.SessionNeedsAttention(session) {
			return "active session is quiet • check if it needs direction"
		}
		if isActiveStatus(session.Status) {
			return "local session can be resumed with 's'"
		}
		return "local session"
	case data.SourceAgentTask:
		if strings.EqualFold(strings.TrimSpace(session.Status), "failed") {
			return "remote task failed • open details/logs"
		}
		if data.SessionNeedsAttention(session) {
			return "remote task is quiet • verify progress"
		}
		if session.PRNumber > 0 {
			return fmt.Sprintf("remote task • PR #%d", session.PRNumber)
		}
		return "remote task"
	default:
		return "session"
	}
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

func (m Model) pageSize() int {
	// Reserve space for detail panel + padding
	available := m.height - 12
	if m.height <= 20 {
		available = m.height - 4
	}
	// Each row is 2 lines
	size := available / 2
	if size < 3 {
		return 3
	}
	if size > 20 {
		return 20
	}
	return size
}

func visibleRange(total, cursor, size int) (int, int) {
	if total <= 0 || size <= 0 {
		return 0, 0
	}
	if total <= size {
		return 0, total
	}

	start := cursor - (size / 2)
	if start < 0 {
		start = 0
	}
	end := start + size
	if end > total {
		end = total
		start = end - size
	}
	return start, end
}
