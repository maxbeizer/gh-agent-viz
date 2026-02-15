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
}

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
		return m.titleStyle.Render("Loading sessions...")
	}

	if len(m.sessions) == 0 {
		return m.titleStyle.Render("No sessions found\n\nTry: refresh with 'r' or toggle filter with Tab")
	}

	columns := make([]string, 0, 3)
	for col := 0; col < 3; col++ {
		columns = append(columns, m.renderColumn(col))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, columns...)
}

func (m Model) renderColumn(column int) string {
	headerStyle := m.tableHeaderStyle
	if column == m.activeColumn {
		headerStyle = m.tableRowSelected.Bold(true)
	}

	indices := m.columnSessionIdx[column]
	rows := []string{headerStyle.Render(fmt.Sprintf("%s (%d)", columnTitle(column), len(indices)))}
	if len(indices) == 0 {
		rows = append(rows, m.tableRowStyle.Render("  —"))
		return lipgloss.NewStyle().Width(42).PaddingRight(1).Render(strings.Join(rows, "\n"))
	}

	cursor := m.rowCursor[column]
	if cursor >= len(indices) {
		cursor = len(indices) - 1
	}
	if cursor < 0 {
		cursor = 0
	}

	for i, idx := range indices {
		session := m.sessions[idx]
		rows = append(rows, m.renderRow(session, column == m.activeColumn && i == cursor))
	}

	return lipgloss.NewStyle().Width(42).PaddingRight(1).Render(strings.Join(rows, "\n"))
}

func (m Model) renderRow(session data.Session, selected bool) string {
	style := m.tableRowStyle
	if selected {
		style = m.tableRowSelected
	}

	icon := m.statusIcon(session.Status)
	title := truncate(session.Title, 34)
	repo := truncate(session.Repository, 16)
	source := sourceLabel(session.Source)
	updated := formatTime(session.UpdatedAt)
	if repo == "" {
		repo = "no-repo"
	}
	if title == "" {
		title = "Untitled Session"
	}

	row := fmt.Sprintf("%s %s\n  %s • %s • %s", icon, title, repo, source, updated)
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
	case "running", "queued", "in progress", "active", "open":
		return 0
	default:
		return 1
	}
}

func columnTitle(column int) string {
	switch column {
	case 0:
		return "Running"
	case 1:
		return "Done"
	default:
		return "Failed"
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
