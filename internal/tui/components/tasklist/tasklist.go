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
	duplicateCounts   map[int]int // newest session index → count of older duplicates
	dismissedIDs      map[string]struct{}
	dismissedStore    *data.DismissedStore
	rowCursor         int
	loading           bool
	statusIcon        func(string) string
	animStatusIcon    func(string, int) string
	animFrame         int
	selectedSessionID string
	width             int
	height            int
	splitMode         bool
}

// New creates a new task list model
func New(titleStyle, headerStyle, rowStyle, rowSelectedStyle lipgloss.Style, statusIconFunc func(string) string) Model {
	return NewWithStore(titleStyle, headerStyle, rowStyle, rowSelectedStyle, statusIconFunc, nil, nil)
}

// NewWithStore creates a new task list model backed by a persistent dismissed store.
func NewWithStore(titleStyle, headerStyle, rowStyle, rowSelectedStyle lipgloss.Style, statusIconFunc func(string) string, animStatusIconFunc func(string, int) string, store *data.DismissedStore) Model {
	dismissed := map[string]struct{}{}
	if store != nil {
		dismissed = store.IDs()
	}
	return Model{
		titleStyle:       titleStyle,
		tableHeaderStyle: headerStyle,
		tableRowStyle:    rowStyle,
		tableRowSelected: rowSelectedStyle,
		sessions:         []data.Session{},
		deEmphasizedIdx:  map[int]struct{}{},
		dismissedIDs:     dismissed,
		dismissedStore:   store,
		rowCursor:        0,
		loading:          true,
		statusIcon:       statusIconFunc,
		animStatusIcon:   animStatusIconFunc,
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

// SetSplitMode sets whether the list is rendered in split-pane mode.
// When split mode is active, the inline detail panel is not shown.
func (m *Model) SetSplitMode(split bool) {
	m.splitMode = split
}

// View renders the sessions as a focused single-column list
func (m Model) View() string {
	if m.loading {
		return m.titleStyle.Render("🔄 Loading sessions...\n\nFetching your agent sessions, one moment.")
	}

	if len(m.sessions) == 0 {
		return m.titleStyle.Render("✨ All quiet on the agent front.\n\nNo sessions match this filter — your agents are either napping or haven't checked in yet.\nPress 'r' to refresh, or tab to try another filter.")
	}

	return m.renderFocusedList()
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

func (m Model) renderRow(sessionIdx int, session data.Session, selected bool, width int) string {
	style := m.tableRowStyle
	if selected {
		style = m.tableRowSelected
	} else if m.isDeEmphasized(sessionIdx) {
		style = style.Faint(true)
	}

	icon := m.currentStatusIcon(session.Status)

	titleMax := width - 8
	if titleMax < 3 {
		titleMax = 3
	}
	title := truncate(sessionTitle(session), titleMax)
	badge := sessionBadge(session, m.isDeEmphasized(sessionIdx), m.duplicateCounts[sessionIdx])

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

	if dur := compactDuration(session); dur != "" {
		meta += " • ⏱ " + dur
	}

	return style.Render(titleLine + "\n" + meta)
}

// SetTasks updates sessions with sorting and de-emphasis
func (m *Model) SetTasks(sessions []data.Session) {
	m.loading = false
	if selected := m.SelectedTask(); selected != nil {
		m.selectedSessionID = selected.ID
	}

	// Filter out dismissed sessions
	filtered := make([]data.Session, 0, len(sessions))
	for _, session := range sessions {
		if _, dismissed := m.dismissedIDs[session.ID]; !dismissed {
			filtered = append(filtered, session)
		}
	}
	sessions = filtered

	type rankedSession struct {
		session      data.Session
		origIdx      int
		deEmphasized bool
		sortPriority int
	}

	deEmphasizedInputIdx, inputDupCounts := quietDuplicateIndices(sessions)
	ranked := make([]rankedSession, 0, len(sessions))
	for i, session := range sessions {
		_, deEmphasized := deEmphasizedInputIdx[i]
		ranked = append(ranked, rankedSession{
			session:      session,
			origIdx:      i,
			deEmphasized: deEmphasized,
			sortPriority: sessionSortPriority(session, deEmphasized),
		})
	}

	// Sort: most recent first, with de-emphasized items sinking to the bottom
	sort.SliceStable(ranked, func(i, j int) bool {
		// De-emphasized always last
		if ranked[i].deEmphasized != ranked[j].deEmphasized {
			return !ranked[i].deEmphasized
		}
		// Within same emphasis level: most recent first
		ti := ranked[i].session.UpdatedAt
		tj := ranked[j].session.UpdatedAt
		if !ti.IsZero() && !tj.IsZero() {
			return ti.After(tj)
		}
		if !ti.IsZero() {
			return true
		}
		if !tj.IsZero() {
			return false
		}
		return false
	})

	m.sessions = make([]data.Session, len(ranked))
	m.deEmphasizedIdx = map[int]struct{}{}
	m.duplicateCounts = map[int]int{}
	for i, candidate := range ranked {
		m.sessions[i] = candidate.session
		if candidate.deEmphasized {
			m.deEmphasizedIdx[i] = struct{}{}
		}
		if count, ok := inputDupCounts[candidate.origIdx]; ok {
			m.duplicateCounts[i] = count
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

// DismissSelected removes the currently selected session from view
func (m *Model) DismissSelected() {
	selected := m.SelectedTask()
	if selected == nil || selected.ID == "" {
		return
	}
	m.dismissedIDs[selected.ID] = struct{}{}
	if m.dismissedStore != nil {
		m.dismissedStore.Add(selected.ID)
	}
	// Remove from current sessions
	newSessions := make([]data.Session, 0, len(m.sessions)-1)
	for _, s := range m.sessions {
		if s.ID != selected.ID {
			newSessions = append(newSessions, s)
		}
	}
	m.sessions = newSessions
	// Clamp cursor
	if m.rowCursor >= len(m.sessions) {
		m.rowCursor = len(m.sessions) - 1
	}
	if m.rowCursor < 0 {
		m.rowCursor = 0
	}
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

func isActiveStatus(status string) bool {
	return data.StatusIsActive(status) || strings.EqualFold(strings.TrimSpace(status), "needs-input")
}

func sessionBadge(session data.Session, deEmphasized bool, duplicateCount int) string {
	if strings.EqualFold(strings.TrimSpace(session.Status), "needs-input") {
		badge := "🧑 waiting on you"
		if duplicateCount > 0 {
			badge += fmt.Sprintf(" (+%d older)", duplicateCount)
		}
		return badge
	}
	if strings.EqualFold(strings.TrimSpace(session.Status), "failed") {
		badge := "🚨 failed"
		if duplicateCount > 0 {
			badge += fmt.Sprintf(" (+%d older)", duplicateCount)
		}
		return badge
	}
	if deEmphasized {
		if !session.UpdatedAt.IsZero() {
			return fmt.Sprintf("↺ quiet duplicate · %s ago", formatTime(session.UpdatedAt))
		}
		return "↺ quiet duplicate"
	}
	if data.SessionNeedsAttention(session) {
		badge := fmt.Sprintf("⏸ idle %s", formatIdleDuration(time.Since(session.UpdatedAt)))
		if duplicateCount > 0 {
			badge += fmt.Sprintf(" (+%d older)", duplicateCount)
		}
		return badge
	}
	if !isActiveStatus(session.Status) || session.UpdatedAt.IsZero() {
		return ""
	}
	badge := "• in progress"
	if duplicateCount > 0 {
		badge += fmt.Sprintf(" (+%d older)", duplicateCount)
	}
	return badge
}

func formatIdleDuration(d time.Duration) string {
	if d < time.Hour {
		return fmt.Sprintf("~%dm", int(d.Minutes()))
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if m == 0 {
		return fmt.Sprintf("~%dh", h)
	}
	return fmt.Sprintf("~%dh%dm", h, m)
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

func quietDuplicateIndices(sessions []data.Session) (map[int]struct{}, map[int]int) {
	grouped := map[string][]int{}
	for i, session := range sessions {
		if !isQuietDuplicateSession(session) {
			continue
		}
		key := quietDuplicateKey(session)
		grouped[key] = append(grouped[key], i)
	}

	deEmphasized := map[int]struct{}{}
	counts := map[int]int{}
	for _, indexes := range grouped {
		if len(indexes) < 2 {
			continue
		}

		sort.SliceStable(indexes, func(i, j int) bool {
			return sessions[indexes[i]].UpdatedAt.After(sessions[indexes[j]].UpdatedAt)
		})

		counts[indexes[0]] = len(indexes) - 1
		for _, idx := range indexes[1:] {
			deEmphasized[idx] = struct{}{}
		}
	}

	return deEmphasized, counts
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

// compactDuration returns a short duration string for the metadata line.
// Returns empty string when telemetry is nil or duration is zero.
func compactDuration(session data.Session) string {
	if session.Telemetry == nil || session.Telemetry.Duration <= 0 {
		return ""
	}
	d := session.Telemetry.Duration
	if d < time.Minute {
		return "< 1m"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if minutes > 0 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
	return fmt.Sprintf("%dh", hours)
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

// SetSize updates the available rendering size for responsive layout.
func (m *Model) SetSize(width, height int) {
	if width > 0 {
		m.width = width
	}
	if height > 0 {
		m.height = height
	}
}

// SetAnimFrame updates the animation frame counter for animated status icons.
func (m *Model) SetAnimFrame(frame int) {
	m.animFrame = frame
}

// currentStatusIcon returns the animated icon if available, otherwise the static icon.
func (m Model) currentStatusIcon(status string) string {
	if m.animStatusIcon != nil {
		return m.animStatusIcon(status, m.animFrame)
	}
	return m.statusIcon(status)
}

func (m Model) pageSize() int {
	// Use full height (minus minimal chrome)
	available := m.height - 4
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
