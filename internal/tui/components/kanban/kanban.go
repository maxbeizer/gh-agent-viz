package kanban

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/maxbeizer/gh-agent-viz/internal/data"
)

// Column represents a status column in the kanban board
type Column struct {
	Title    string
	Status   string // status key used in column header
	Sessions []data.Session
}

// defaultColumns returns the fixed set of kanban columns.
func defaultColumns() []Column {
	return []Column{
		{Title: "RUNNING", Status: "running"},
		{Title: "NEEDS INPUT", Status: "needs-input"},
		{Title: "COMPLETED", Status: "completed"},
		{Title: "FAILED", Status: "failed"},
	}
}

// statusBelongsToColumn returns true when a session status belongs in the given column.
func statusBelongsToColumn(sessionStatus, colStatus string) bool {
	s := strings.ToLower(strings.TrimSpace(sessionStatus))
	switch colStatus {
	case "running":
		return s == "running" || s == "active" || s == "queued"
	case "needs-input":
		return s == "needs-input"
	case "completed":
		return s == "completed"
	case "failed":
		return s == "failed"
	}
	return false
}

// Model represents the kanban board state
type Model struct {
	columns           []Column
	colCursor         int
	rowCursor         int
	statusIcon        func(string) string
	animStatusIcon    func(string, int) string
	animFrame         int
	width             int
	height            int
	titleStyle        lipgloss.Style
	columnStyle       lipgloss.Style
	cardStyle         lipgloss.Style
	cardSelectedStyle lipgloss.Style
}

// New creates a new kanban board model.
func New(
	titleStyle lipgloss.Style,
	borderStyle lipgloss.Style,
	rowStyle lipgloss.Style,
	rowSelectedStyle lipgloss.Style,
	statusIconFunc func(string) string,
	animStatusIconFunc func(string, int) string,
) Model {
	return Model{
		columns:           defaultColumns(),
		statusIcon:        statusIconFunc,
		animStatusIcon:    animStatusIconFunc,
		titleStyle:        titleStyle,
		columnStyle:       borderStyle,
		cardStyle:         rowStyle,
		cardSelectedStyle: rowSelectedStyle,
		width:             80,
		height:            24,
	}
}

// SetSessions distributes sessions into columns by status.
func (m *Model) SetSessions(sessions []data.Session) {
	cols := defaultColumns()
	for i := range cols {
		cols[i].Sessions = nil
	}
	for _, s := range sessions {
		for i := range cols {
			if statusBelongsToColumn(s.Status, cols[i].Status) {
				cols[i].Sessions = append(cols[i].Sessions, s)
				break
			}
		}
	}
	m.columns = cols
	// Clamp cursors
	m.clampCursors()
}

// MoveColumn moves the column cursor by delta (negative = left, positive = right).
func (m *Model) MoveColumn(delta int) {
	if len(m.columns) == 0 {
		return
	}
	m.colCursor += delta
	if m.colCursor < 0 {
		m.colCursor = 0
	}
	if m.colCursor >= len(m.columns) {
		m.colCursor = len(m.columns) - 1
	}
	// Clamp row cursor to new column
	m.clampRowCursor()
}

// MoveRow moves the row cursor within the current column by delta.
func (m *Model) MoveRow(delta int) {
	col := m.currentColumn()
	if col == nil || len(col.Sessions) == 0 {
		return
	}
	m.rowCursor += delta
	if m.rowCursor < 0 {
		m.rowCursor = 0
	}
	if m.rowCursor >= len(col.Sessions) {
		m.rowCursor = len(col.Sessions) - 1
	}
}

// SelectedSession returns a pointer to the currently focused session, or nil.
func (m *Model) SelectedSession() *data.Session {
	col := m.currentColumn()
	if col == nil || len(col.Sessions) == 0 {
		return nil
	}
	if m.rowCursor < 0 || m.rowCursor >= len(col.Sessions) {
		return nil
	}
	s := col.Sessions[m.rowCursor]
	return &s
}

// SetSize sets the available dimensions for rendering.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetAnimFrame updates the animation frame counter.
func (m *Model) SetAnimFrame(frame int) {
	m.animFrame = frame
}

// Columns returns the current columns (for testing).
func (m *Model) Columns() []Column {
	return m.columns
}

// ColCursor returns the current column cursor position (for testing).
func (m *Model) ColCursor() int {
	return m.colCursor
}

// RowCursor returns the current row cursor position (for testing).
func (m *Model) RowCursor() int {
	return m.rowCursor
}

// ColumnWidth computes the width for each column based on non-empty column count.
func (m *Model) ColumnWidth() int {
	nonEmpty := 0
	for _, col := range m.columns {
		if len(col.Sessions) > 0 {
			nonEmpty++
		}
	}
	count := len(m.columns)
	if nonEmpty > 0 {
		count = nonEmpty
	}
	// Always show all columns, but size based on visible count
	count = len(m.columns)
	if count == 0 {
		return m.width
	}
	// Account for gaps between columns (1 space each)
	gaps := count - 1
	available := m.width - gaps
	w := available / count
	if w < 20 {
		w = 20
	}
	return w
}

// View renders the full kanban board.
func (m *Model) View() string {
	colWidth := m.ColumnWidth()
	// Available height for cards inside a column (subtract header/border chrome)
	// Border top + title line + border bottom = 3 lines of chrome
	cardAreaHeight := m.height - 6
	if cardAreaHeight < 3 {
		cardAreaHeight = 3
	}

	columnViews := make([]string, len(m.columns))
	for i, col := range m.columns {
		isFocusedCol := (i == m.colCursor)
		columnViews[i] = m.renderColumn(col, i, colWidth, cardAreaHeight, isFocusedCol)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, columnViews...)
}

// renderColumn renders a single kanban column.
func (m *Model) renderColumn(col Column, colIdx, width, cardAreaHeight int, focused bool) string {
	innerWidth := width - 4 // account for border left/right + padding
	if innerWidth < 10 {
		innerWidth = 10
	}

	// Build card lines
	var cardLines []string
	if len(col.Sessions) == 0 {
		placeholder := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true).
			Render("(no sessions)")
		cardLines = append(cardLines, placeholder)
	} else {
		for rowIdx, session := range col.Sessions {
			isSelected := focused && rowIdx == m.rowCursor
			card := m.renderCard(session, innerWidth, isSelected)
			cardLines = append(cardLines, card)
		}
	}

	content := strings.Join(cardLines, "\n")

	// Pad or truncate to fill card area height
	lines := strings.Split(content, "\n")
	for len(lines) < cardAreaHeight {
		lines = append(lines, strings.Repeat(" ", innerWidth))
	}
	if len(lines) > cardAreaHeight {
		lines = lines[:cardAreaHeight]
	}
	content = strings.Join(lines, "\n")

	// Build title with column header
	title := fmt.Sprintf(" %s ", col.Title)

	borderColor := lipgloss.Color("238")
	if focused {
		borderColor = lipgloss.Color("63")
	}

	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width - 2). // -2 for border characters
		Padding(0, 1)

	titleRendered := lipgloss.NewStyle().
		Bold(true).
		Foreground(borderColor).
		Render(title)

	return titleRendered + "\n" + style.Render(content)
}

// renderCard renders a single session card.
func (m *Model) renderCard(session data.Session, width int, selected bool) string {
	icon := m.statusIcon(session.Status)
	if m.animStatusIcon != nil {
		icon = m.animStatusIcon(session.Status, m.animFrame)
	}

	// First line: icon + title (truncated)
	titleMaxLen := width - 4 // icon + space + padding
	if titleMaxLen < 5 {
		titleMaxLen = 5
	}
	title := session.Title
	if len(title) > titleMaxLen {
		title = title[:titleMaxLen-1] + "…"
	}
	line1 := fmt.Sprintf("%s %s", icon, title)

	// Second line: repo + age
	repo := session.Repository
	if repo == "" {
		repo = "local"
	}
	// Shorten repo: "owner/repo" → "repo"
	if parts := strings.SplitN(repo, "/", 2); len(parts) == 2 {
		repo = parts[1]
	}
	age := formatAge(session.CreatedAt)
	line2 := fmt.Sprintf("  %s • %s", repo, age)
	if len(line2) > width {
		line2 = line2[:width]
	}

	cardText := line1 + "\n" + line2

	if selected {
		return m.cardSelectedStyle.Width(width).Render(cardText)
	}
	return m.cardStyle.Width(width).Render(cardText)
}

// formatAge formats a time as a human-readable duration (e.g., "12m", "1h", "2d").
func formatAge(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "<1m"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

// currentColumn returns a pointer to the column at colCursor, or nil.
func (m *Model) currentColumn() *Column {
	if len(m.columns) == 0 || m.colCursor < 0 || m.colCursor >= len(m.columns) {
		return nil
	}
	return &m.columns[m.colCursor]
}

// clampCursors clamps both column and row cursors to valid ranges.
func (m *Model) clampCursors() {
	if len(m.columns) == 0 {
		m.colCursor = 0
		m.rowCursor = 0
		return
	}
	if m.colCursor >= len(m.columns) {
		m.colCursor = len(m.columns) - 1
	}
	if m.colCursor < 0 {
		m.colCursor = 0
	}
	m.clampRowCursor()
}

// clampRowCursor clamps the row cursor for the current column.
func (m *Model) clampRowCursor() {
	col := m.currentColumn()
	if col == nil || len(col.Sessions) == 0 {
		m.rowCursor = 0
		return
	}
	if m.rowCursor >= len(col.Sessions) {
		m.rowCursor = len(col.Sessions) - 1
	}
	if m.rowCursor < 0 {
		m.rowCursor = 0
	}
}
