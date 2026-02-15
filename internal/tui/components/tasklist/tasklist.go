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
	columnSessionIdx  [3][]int
	activeColumn      int
	rowCursor         [3]int
	loading           bool
	statusIcon        func(string) string
	selectedSessionID string
	width             int
	height            int
}

const defaultColumnWidth = 42
const minColumnWidth = 30

// New creates a new task list model
func New(titleStyle, headerStyle, rowStyle, rowSelectedStyle lipgloss.Style, statusIconFunc func(string) string) Model {
	return Model{
		titleStyle:       titleStyle,
		tableHeaderStyle: headerStyle,
		tableRowStyle:    rowStyle,
		tableRowSelected: rowSelectedStyle,
		sessions:         []data.Session{},
		deEmphasizedIdx:  map[int]struct{}{},
		activeColumn:     0,
		loading:          false,
		statusIcon:       statusIconFunc,
		width:            defaultColumnWidth * 3,
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

// View renders the sessions as a kanban board
func (m Model) View() string {
	if m.loading {
		return m.titleStyle.Render("Loading sessions...\n\nGathering the latest Copilot session updates.")
	}

	if len(m.sessions) == 0 {
		return m.titleStyle.Render("No sessions to show yet.\n\nPress 'r' to refresh, or Tab/Shift+Tab to switch filters.")
	}

	overview := m.renderOverview()
	board := m.renderBoard()
	flightDeck := m.renderFlightDeck()
	compactFlightDeck := m.renderCompactFlightDeck()

	switch {
	case m.height <= 24:
		return lipgloss.JoinVertical(lipgloss.Left, overview, board)
	case m.height <= 30:
		return lipgloss.JoinVertical(lipgloss.Left, overview, board, compactFlightDeck)
	default:
		return lipgloss.JoinVertical(lipgloss.Left, overview, "", board, "", flightDeck)
	}
}

func (m Model) renderBoard() string {
	if m.width < minColumnWidth*3+4 {
		narrowWidth := m.width - 2
		if narrowWidth < 3 {
			narrowWidth = 3
		}
		hint := m.tableRowStyle.Render(fmt.Sprintf("COMPACT VIEW • showing %s lane only (use ←/→)", columnTitle(m.activeColumn)))
		column := m.renderColumn(m.activeColumn, narrowWidth)
		return lipgloss.JoinVertical(lipgloss.Left, hint, column)
	}

	columns := make([]string, 0, 3)
	columnWidth := m.columnWidth()
	for col := 0; col < 3; col++ {
		columns = append(columns, m.renderColumn(col, columnWidth))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, columns...)
}

func (m Model) renderOverview() string {
	activeCount := len(m.columnSessionIdx[0])
	doneCount := len(m.columnSessionIdx[1])
	failedCount := len(m.columnSessionIdx[2])
	attentionCount := 0

	for _, session := range m.sessions {
		if data.SessionNeedsAttention(session) {
			attentionCount++
		}
	}

	chips := []string{
		fmt.Sprintf("total %d", len(m.sessions)),
		fmt.Sprintf("running %d", activeCount),
		fmt.Sprintf("done %d", doneCount),
		fmt.Sprintf("failed %d", failedCount),
		fmt.Sprintf("needs action %d", attentionCount),
	}

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1).
		Render("SESSIONS AT A GLANCE  " + strings.Join(chips, "  •  "))
}

func (m Model) renderFlightDeck() string {
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

	lines := []string{
		"SESSION SUMMARY",
		fmt.Sprintf("%s %s", m.statusIcon(selected.Status), sessionTitle(*selected)),
		fmt.Sprintf("Status: %s", selected.Status),
		fmt.Sprintf("Needs your action: %s", attentionReason(*selected)),
		fmt.Sprintf("Repository: %s", panelRepository(*selected)),
		fmt.Sprintf("Branch: %s", panelBranch(*selected)),
		fmt.Sprintf("Source: %s", sourceLabel(selected.Source)),
		fmt.Sprintf("Last update: %s", formatTime(selected.UpdatedAt)),
		fmt.Sprintf("Available actions: %s", strings.Join(actions, " • ")),
	}
	if selected.Source == data.SourceAgentTask && selected.PRNumber > 0 {
		lines = append(lines, fmt.Sprintf("Pull Request: #%d", selected.PRNumber))
	}

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("241")).
		Padding(0, 1).
		Render(strings.Join(lines, "\n"))
}

func (m Model) renderCompactFlightDeck() string {
	selected := m.SelectedTask()
	if selected == nil {
		return ""
	}
	line := fmt.Sprintf("Session Summary • %s %s • Needs action: %s • Last update: %s", m.statusIcon(selected.Status), truncate(sessionTitle(*selected), 32), attentionReason(*selected), formatTime(selected.UpdatedAt))
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("241")).
		Padding(0, 1).
		Render(line)
}

func (m Model) renderColumn(column int, width int) string {
	headerStyle := m.tableHeaderStyle
	if column == m.activeColumn {
		headerStyle = m.tableRowSelected.Bold(true)
	}

	indices := m.columnSessionIdx[column]
	rows := []string{headerStyle.Render(fmt.Sprintf("%s (%d)", columnTitle(column), len(indices)))}
	if len(indices) == 0 {
		rows = append(rows, m.tableRowStyle.Render("  —"))
		return lipgloss.NewStyle().Width(width).PaddingRight(1).Render(strings.Join(rows, "\n"))
	}

	cursor := m.rowCursor[column]
	if cursor >= len(indices) {
		cursor = len(indices) - 1
	}
	if cursor < 0 {
		cursor = 0
	}

	start, end := visibleRange(len(indices), cursor, m.pageSize())
	if start > 0 {
		rows = append(rows, m.tableRowStyle.Render(fmt.Sprintf("  ↑ %d above", start)))
	}
	for i, idx := range indices[start:end] {
		session := m.sessions[idx]
		selected := column == m.activeColumn && (start+i) == cursor
		rows = append(rows, m.renderRow(idx, session, selected, width))
	}
	if end < len(indices) {
		rows = append(rows, m.tableRowStyle.Render(fmt.Sprintf("  ↓ %d more", len(indices)-end)))
	}

	return lipgloss.NewStyle().Width(width).PaddingRight(1).Render(strings.Join(rows, "\n"))
}

func (m Model) renderRow(sessionIdx int, session data.Session, selected bool, width int) string {
	style := m.tableRowStyle
	if selected {
		style = m.tableRowSelected
	} else if m.isDeEmphasized(sessionIdx) {
		style = style.Faint(true)
	}

	icon := m.statusIcon(session.Status)
	titleMax := width - 4
	if titleMax < 3 {
		titleMax = 3
	}
	title := truncate(sessionTitle(session), titleMax)
	badge := sessionBadge(session, m.isDeEmphasized(sessionIdx))
	attention := fmt.Sprintf("Needs your action: %s", attentionReason(session))

	titleLine := fmt.Sprintf("%s %s", icon, title)
	if badge != "" {
		titleLine += " " + badge
	}

	if width < 20 {
		row := fmt.Sprintf("%s\n  %s", titleLine, truncate(attention, titleMax))
		if selected {
			row += fmt.Sprintf("\n  ↳ %s", truncate(rowContext(session), titleMax))
		}
		return style.Render(row)
	}

	metaMax := width - 4
	if metaMax < 3 {
		metaMax = 3
	}
	rowLines := []string{
		titleLine,
		fmt.Sprintf("  %s", truncate(fmt.Sprintf("Repository: %s", rowRepository(session)), metaMax)),
		fmt.Sprintf("  %s", truncate(fmt.Sprintf("%s • Last update: %s", attention, formatTime(session.UpdatedAt)), metaMax)),
	}
	if selected {
		contextMax := width - 6
		if contextMax < 3 {
			contextMax = 3
		}
		rowLines = append(rowLines, fmt.Sprintf("  ↳ %s", truncate(rowContext(session), contextMax)))
	}
	return style.Render(strings.Join(rowLines, "\n"))
}

// SetTasks updates sessions and recategorizes columns
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

	m.columnSessionIdx = [3][]int{}
	for i, session := range m.sessions {
		column := statusColumn(session.Status)
		m.columnSessionIdx[column] = append(m.columnSessionIdx[column], i)
	}

	for col := 0; col < 3; col++ {
		if len(m.columnSessionIdx[col]) == 0 {
			m.rowCursor[col] = 0
			continue
		}
		if m.rowCursor[col] >= len(m.columnSessionIdx[col]) {
			m.rowCursor[col] = len(m.columnSessionIdx[col]) - 1
		}
		if m.rowCursor[col] < 0 {
			m.rowCursor[col] = 0
		}
	}

	if m.selectedSessionID != "" {
		for idx, session := range m.sessions {
			if session.ID != m.selectedSessionID {
				continue
			}
			for col := 0; col < 3; col++ {
				for row, sessionIdx := range m.columnSessionIdx[col] {
					if sessionIdx == idx {
						m.activeColumn = col
						m.rowCursor[col] = row
						return
					}
				}
			}
		}
	}

	if len(m.columnSessionIdx[m.activeColumn]) == 0 {
		for col := 0; col < 3; col++ {
			if len(m.columnSessionIdx[col]) > 0 {
				m.activeColumn = col
				break
			}
		}
	}
}

// MoveCursor moves the active row cursor
func (m *Model) MoveCursor(delta int) {
	columnSessions := m.columnSessionIdx[m.activeColumn]
	if len(columnSessions) == 0 {
		m.rowCursor[m.activeColumn] = 0
		return
	}

	m.rowCursor[m.activeColumn] += delta
	if m.rowCursor[m.activeColumn] < 0 {
		m.rowCursor[m.activeColumn] = 0
	}
	if m.rowCursor[m.activeColumn] >= len(columnSessions) {
		m.rowCursor[m.activeColumn] = len(columnSessions) - 1
	}
}

// MoveColumn moves the active column left/right
func (m *Model) MoveColumn(delta int) {
	m.activeColumn += delta
	if m.activeColumn < 0 {
		m.activeColumn = 0
	}
	if m.activeColumn > 2 {
		m.activeColumn = 2
	}
}

// SelectedTask returns the selected session
func (m Model) SelectedTask() *data.Session {
	columnSessions := m.columnSessionIdx[m.activeColumn]
	if len(columnSessions) == 0 {
		return nil
	}

	cursor := m.rowCursor[m.activeColumn]
	if cursor >= len(columnSessions) {
		cursor = len(columnSessions) - 1
	}
	if cursor < 0 {
		cursor = 0
	}

	sessionIdx := columnSessions[cursor]
	return &m.sessions[sessionIdx]
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

func statusColumn(status string) int {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "failed", "cancelled", "canceled":
		return 2
	case "needs-input":
		return 0
	default:
		if data.StatusIsActive(status) {
			return 0
		}
		return 1
	}
}

func columnTitle(column int) string {
	switch column {
	case 0:
		return "🛫 Running"
	case 1:
		return "🛬 Done"
	default:
		return "🚨 Failed"
	}
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

func (m Model) columnWidth() int {
	usable := m.width - 4
	if usable <= 0 {
		return defaultColumnWidth
	}
	width := usable / 3
	if width < minColumnWidth {
		return minColumnWidth
	}
	return width
}

func (m Model) pageSize() int {
	available := m.height - 14
	if m.height <= 24 {
		available = m.height - 8
	} else if m.height <= 30 {
		available = m.height - 9
	}
	size := available / 2
	if size < 2 {
		return 2
	}
	if size > 12 {
		return 12
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
