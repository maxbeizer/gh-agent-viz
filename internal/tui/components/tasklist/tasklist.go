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
const minNarrowColumnWidth = 20

// New creates a new task list model
func New(titleStyle, headerStyle, rowStyle, rowSelectedStyle lipgloss.Style, statusIconFunc func(string) string) Model {
	return Model{
		titleStyle:       titleStyle,
		tableHeaderStyle: headerStyle,
		tableRowStyle:    rowStyle,
		tableRowSelected: rowSelectedStyle,
		sessions:         []data.Session{},
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
		return m.titleStyle.Render("Warming up the radar...\n\nCollecting fresh session telemetry.")
	}

	if len(m.sessions) == 0 {
		return m.titleStyle.Render("The sky is clear — no active flights yet.\n\nPress 'r' to scan again, or Tab/Shift+Tab to check other lanes.")
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
		if narrowWidth < minNarrowColumnWidth {
			narrowWidth = minNarrowColumnWidth
		}
		hint := m.tableRowStyle.Render(fmt.Sprintf("NARROW MODE • showing %s lane only (use ←/→)", columnTitle(m.activeColumn)))
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
	agentCount := 0
	localCount := 0
	attentionCount := 0

	for _, session := range m.sessions {
		if session.Source == data.SourceAgentTask {
			agentCount++
		}
		if session.Source == data.SourceLocalCopilot {
			localCount++
		}
		if data.SessionNeedsAttention(session) {
			attentionCount++
		}
	}

	chips := []string{
		fmt.Sprintf("🛰 total %d", len(m.sessions)),
		fmt.Sprintf("🛫 active %d", activeCount),
		fmt.Sprintf("🛬 done %d", doneCount),
		fmt.Sprintf("🚨 failed %d", failedCount),
		fmt.Sprintf("🤖 agent %d", agentCount),
		fmt.Sprintf("💻 local %d", localCount),
	}
	if attentionCount > 0 {
		chips = append(chips, fmt.Sprintf("🚦 attention %d", attentionCount))
	}

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1).
		Render("ATC OVERVIEW  " + strings.Join(chips, "  •  "))
}

func (m Model) renderFlightDeck() string {
	selected := m.SelectedTask()
	if selected == nil {
		return ""
	}

	repository := selected.Repository
	if repository == "" {
		repository = "no-repo"
	}
	branch := selected.Branch
	if branch == "" {
		branch = "unknown"
	}

	actions := []string{"enter details"}
	if selected.Source == data.SourceAgentTask {
		actions = append(actions, "l logs", "o open PR")
	}
	if selected.Source == data.SourceLocalCopilot && isActiveStatus(selected.Status) && selected.ID != "" {
		actions = append(actions, "s resume")
	}

	lines := []string{
		"FLIGHT DECK",
		fmt.Sprintf("%s %s", m.statusIcon(selected.Status), selected.Title),
		fmt.Sprintf("status: %s", selected.Status),
		fmt.Sprintf("repo:   %s", repository),
		fmt.Sprintf("branch: %s", branch),
		fmt.Sprintf("source: %s", sourceLabel(selected.Source)),
		fmt.Sprintf("seen:   %s", formatTime(selected.UpdatedAt)),
		fmt.Sprintf("actions: %s", strings.Join(actions, " • ")),
	}
	if selected.Source == data.SourceAgentTask && selected.PRNumber > 0 {
		lines = append(lines, fmt.Sprintf("pr: #%d", selected.PRNumber))
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
	line := fmt.Sprintf("FLIGHT DECK • %s %s • %s • %s", m.statusIcon(selected.Status), truncate(selected.Title, 40), sourceLabel(selected.Source), formatTime(selected.UpdatedAt))
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
		rows = append(rows, m.renderRow(session, selected, width))
	}
	if end < len(indices) {
		rows = append(rows, m.tableRowStyle.Render(fmt.Sprintf("  ↓ %d more", len(indices)-end)))
	}

	return lipgloss.NewStyle().Width(width).PaddingRight(1).Render(strings.Join(rows, "\n"))
}

func (m Model) renderRow(session data.Session, selected bool, width int) string {
	style := m.tableRowStyle
	if selected {
		style = m.tableRowSelected
	}

	icon := m.statusIcon(session.Status)
	title := truncate(session.Title, maxInt(16, width-10))
	repoWithBranch := session.Repository
	if repoWithBranch == "" {
		repoWithBranch = "no-repo"
	}
	if session.Branch != "" {
		repoWithBranch = fmt.Sprintf("%s@%s", repoWithBranch, session.Branch)
	}
	updated := formatTime(session.UpdatedAt)
	repo := truncate(repoWithBranch, maxInt(18, width-len(updated)-8))
	if title == "" {
		title = "Untitled Session"
	}
	badge := sessionBadge(session)

	titleLine := fmt.Sprintf("%s %s", icon, title)
	if badge != "" {
		titleLine += " " + badge
	}

	row := fmt.Sprintf("%s\n  %s • %s", titleLine, repo, updated)
	if selected {
		row += fmt.Sprintf("\n  ↳ %s", truncate(rowContext(session), maxInt(18, width-6)))
	}
	return style.Render(row)
}

// SetTasks updates sessions and recategorizes columns
func (m *Model) SetTasks(sessions []data.Session) {
	if selected := m.SelectedTask(); selected != nil {
		m.selectedSessionID = selected.ID
	}

	m.sessions = append([]data.Session(nil), sessions...)
	sort.SliceStable(m.sessions, func(i, j int) bool {
		return m.sessions[i].UpdatedAt.After(m.sessions[j].UpdatedAt)
	})

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
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "unknown"
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
	case "running", "queued", "in progress", "active", "open", "needs-input":
		return 0
	default:
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
	normalized := strings.ToLower(strings.TrimSpace(status))
	return normalized == "running" || normalized == "queued" || normalized == "active" || normalized == "open" || normalized == "in progress" || normalized == "needs-input"
}

func sessionBadge(session data.Session) string {
	if strings.EqualFold(strings.TrimSpace(session.Status), "needs-input") {
		return "🧑 input needed"
	}
	if strings.EqualFold(strings.TrimSpace(session.Status), "failed") {
		return "🚨 failed"
	}
	if data.SessionNeedsAttention(session) {
		return "⚠ attention"
	}
	if !isActiveStatus(session.Status) || session.UpdatedAt.IsZero() {
		return ""
	}
	return "• live"
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

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
