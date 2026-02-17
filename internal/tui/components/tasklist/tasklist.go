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

// groupByModes defines the cycle order for group-by modes.
var groupByModes = []string{"", "repository", "status", "source"}

// autoGroupThreshold is the session count at which auto-grouping by repo kicks in.
const autoGroupThreshold = 8

// Model represents the task list component state
type Model struct {
	titleStyle        lipgloss.Style
	tableHeaderStyle  lipgloss.Style
	tableRowStyle     lipgloss.Style
	tableRowSelected   lipgloss.Style
	sectionHeaderStyle lipgloss.Style
	sessions           []data.Session
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
	splitMode          bool
	groupBy            string // "", "repository", "status", "source"
	userSetGroupBy     bool   // true once user manually toggles via 'g'
	expandedGroup      int    // index of expanded group (-1 = none)
}

// New creates a new task list model
func New(titleStyle, headerStyle, rowStyle, rowSelectedStyle lipgloss.Style, statusIconFunc func(string) string) Model {
	return NewWithStore(titleStyle, headerStyle, rowStyle, rowSelectedStyle, lipgloss.NewStyle(), statusIconFunc, nil, nil)
}

// NewWithStore creates a new task list model backed by a persistent dismissed store.
func NewWithStore(titleStyle, headerStyle, rowStyle, rowSelectedStyle, sectionHeaderStyle lipgloss.Style, statusIconFunc func(string) string, animStatusIconFunc func(string, int) string, store *data.DismissedStore) Model {
	dismissed := map[string]struct{}{}
	if store != nil {
		dismissed = store.IDs()
	}
	return Model{
		titleStyle:         titleStyle,
		tableHeaderStyle:   headerStyle,
		tableRowStyle:      rowStyle,
		tableRowSelected:   rowSelectedStyle,
		sectionHeaderStyle: sectionHeaderStyle,
		sessions:         []data.Session{},
		dismissedIDs:     dismissed,
		dismissedStore:   store,
		rowCursor:        0,
		loading:          true,
		statusIcon:       statusIconFunc,
		animStatusIcon:   animStatusIconFunc,
		width:            80,
		height:           24,
		expandedGroup:    -1,
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
		emptyArt := `
    ╭──────────────────────────────╮
    │                              │
    │   ✨ All quiet out here      │
    │                              │
    │   No sessions match this     │
    │   filter. Press 'r' to       │
    │   refresh or tab to try      │
    │   another filter.            │
    │                              │
    ╰──────────────────────────────╯`
		return m.titleStyle.Render(emptyArt)
	}

	return m.renderFocusedList()
}

func (m Model) renderFocusedList() string {
	// Auto-group by repo when many sessions and user hasn't manually set grouping
	effectiveGroupBy := m.groupBy
	if effectiveGroupBy == "" && !m.userSetGroupBy && len(m.sessions) >= autoGroupThreshold {
		effectiveGroupBy = "repository"
	}

	if effectiveGroupBy != "" {
		return m.renderGroupedListWith(effectiveGroupBy)
	}

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
		style = m.statusSelectedStyle(session.Status)
	}

	icon := m.currentStatusIcon(session.Status)

	badge := sessionBadge(session, m.duplicateCounts[sessionIdx])

	// Gutter indicator: colored by status, brighter when selected
	gutter := m.statusGutter(session.Status, selected)

	if width < 40 {
		titleMax := width - 8
		if titleMax < 3 {
			titleMax = 3
		}
		title := truncate(sessionTitle(session), titleMax)
		titleLine := fmt.Sprintf("%s%s %s", gutter, icon, title)
		return style.Render(titleLine)
	}

	// Title line: left-aligned title, right-aligned badge
	badgeLen := len(badge)
	titleMax := width - 8 - badgeLen
	if badgeLen > 0 {
		titleMax -= 2 // space before badge
	}
	if titleMax < 10 {
		titleMax = 10
	}
	title := truncate(sessionTitle(session), titleMax)
	leftPart := fmt.Sprintf("%s%s %s", gutter, icon, title)
	if badge != "" {
		pad := width - len(leftPart) - badgeLen
		if pad < 1 {
			pad = 1
		}
		leftPart += strings.Repeat(" ", pad) + badge
	}

	// Meta line: repo + time, dimmed for visual hierarchy
	repo := truncate(rowRepository(session), width/2)
	metaText := fmt.Sprintf("    %s  %s", repo, formatTime(session.UpdatedAt))
	if dur := compactDuration(session); dur != "" {
		durStr := "⏱ " + dur
		pad := width - len(metaText) - len(durStr)
		if pad < 1 {
			pad = 1
		}
		metaText += strings.Repeat(" ", pad) + durStr
	}
	meta := lipgloss.NewStyle().Faint(true).Render(metaText)

	return style.Render(leftPart + "\n" + meta)
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
	}

	deEmphasizedInputIdx, inputDupCounts := quietDuplicateIndices(sessions)
	ranked := make([]rankedSession, 0, len(sessions))
	for i, session := range sessions {
		_, deEmphasized := deEmphasizedInputIdx[i]
		ranked = append(ranked, rankedSession{
			session:      session,
			origIdx:      i,
			deEmphasized: deEmphasized,
		})
	}

	// Sort: most recent first
	sort.SliceStable(ranked, func(i, j int) bool {
		ti := ranked[i].session.UpdatedAt
		tj := ranked[j].session.UpdatedAt
		if !ti.IsZero() && !tj.IsZero() {
			return ti.After(tj)
		}
		if !ti.IsZero() {
			return true
		}
		return false
	})

	m.sessions = make([]data.Session, 0, len(ranked))
	m.duplicateCounts = map[int]int{}
	for _, candidate := range ranked {
		if candidate.deEmphasized {
			continue // Hide quiet duplicates entirely
		}
		idx := len(m.sessions)
		m.sessions = append(m.sessions, candidate.session)
		if count, ok := inputDupCounts[candidate.origIdx]; ok {
			m.duplicateCounts[idx] = count
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

	gb := m.effectiveGroupBy()
	if gb == "" {
		// Ungrouped: simple cursor movement
		m.rowCursor += delta
		if m.rowCursor < 0 {
			m.rowCursor = 0
		}
		if m.rowCursor >= len(m.sessions) {
			m.rowCursor = len(m.sessions) - 1
		}
		return
	}

	// Grouped: navigate between visible items (expanded group sessions + collapsed group headers)
	groups := m.buildGroupsWith(gb)
	// Build flat list of navigable positions (session indices)
	var navigable []int
	for gi, g := range groups {
		if m.expandedGroup == gi {
			navigable = append(navigable, g.sessions...)
		} else {
			// Collapsed group: use first session index as representative
			if len(g.sessions) > 0 {
				navigable = append(navigable, g.sessions[0])
			}
		}
	}

	if len(navigable) == 0 {
		return
	}

	// Find current position in navigable list
	currentPos := 0
	for i, idx := range navigable {
		if idx == m.rowCursor {
			currentPos = i
			break
		}
	}

	newPos := currentPos + delta
	if newPos < 0 {
		newPos = 0
	}
	if newPos >= len(navigable) {
		newPos = len(navigable) - 1
	}
	m.rowCursor = navigable[newPos]
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

func sessionBadge(session data.Session, duplicateCount int) string {
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
	if isActiveStatus(session.Status) && !session.UpdatedAt.IsZero() {
		idle := time.Since(session.UpdatedAt)
		if idle >= data.AttentionStaleMax {
			return "💤 idle " + formatIdleDuration(idle)
		}
		if idle >= data.AttentionStaleThreshold {
			badge := "💤 idle " + formatIdleDuration(idle)
			if duplicateCount > 0 {
				badge += fmt.Sprintf(" (+%d older)", duplicateCount)
			}
			return badge
		}
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
	// Idle active sessions are candidates for quiet duplicate grouping
	if !isActiveStatus(session.Status) || session.UpdatedAt.IsZero() {
		return false
	}
	return time.Since(session.UpdatedAt) >= data.AttentionStaleThreshold
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
	title := strings.TrimSpace(session.Title)
	if title == "" {
		return "Untitled session"
	}
	// Replace "Session <UUID>" with a human-friendly fallback
	if isUUIDSessionTitle(title) {
		return "Untitled session"
	}
	return title
}

// isUUIDSessionTitle detects titles like "Session 7967abbc-163d-4975-9803-8d340b6eb590"
func isUUIDSessionTitle(title string) bool {
	if !strings.HasPrefix(title, "Session ") {
		return false
	}
	uuid := strings.TrimPrefix(title, "Session ")
	// UUID v4 is 36 chars: 8-4-4-4-12
	return len(uuid) == 36 && strings.Count(uuid, "-") == 4
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

// CycleGroupBy advances to the next group-by mode.
func (m *Model) CycleGroupBy() {
	m.userSetGroupBy = true
	for i, mode := range groupByModes {
		if mode == m.groupBy {
			m.groupBy = groupByModes[(i+1)%len(groupByModes)]
			return
		}
	}
	m.groupBy = groupByModes[0]
}

// GroupByLabel returns a human-readable label for the current group mode.
func (m Model) GroupByLabel() string {
	switch m.groupBy {
	case "repository":
		return "repo"
	case "status":
		return "status"
	case "source":
		return "source"
	default:
		return ""
	}
}

// ToggleGroupExpand expands or collapses the group containing the current cursor.
func (m *Model) ToggleGroupExpand() {
	groupIdx := m.groupIndexForCursor()
	if groupIdx < 0 {
		return
	}
	if m.expandedGroup == groupIdx {
		m.expandedGroup = -1
	} else {
		m.expandedGroup = groupIdx
	}
}

// IsGrouped returns true when grouping is active (manual or auto).
func (m Model) IsGrouped() bool {
	if m.groupBy != "" {
		return true
	}
	return !m.userSetGroupBy && len(m.sessions) >= autoGroupThreshold
}

// IsCursorOnCollapsedGroup returns true when the cursor is on a collapsed group header.
func (m Model) IsCursorOnCollapsedGroup() bool {
	gb := m.effectiveGroupBy()
	if gb == "" {
		return false
	}
	groups := m.buildGroupsWith(gb)
	for gi, g := range groups {
		if gi == m.expandedGroup {
			continue // expanded group — cursor is on a session row
		}
		for _, idx := range g.sessions {
			if idx == m.rowCursor {
				return true
			}
		}
	}
	return false
}

// effectiveGroupBy returns the active groupBy mode (manual or auto).
func (m Model) effectiveGroupBy() string {
	if m.groupBy != "" {
		return m.groupBy
	}
	if !m.userSetGroupBy && len(m.sessions) >= autoGroupThreshold {
		return "repository"
	}
	return ""
}

// groupIndexForCursor returns the group index containing the current cursor session.
func (m Model) groupIndexForCursor() int {
	gb := m.effectiveGroupBy()
	if gb == "" {
		return -1
	}
	groups := m.buildGroupsWith(gb)
	cursor := m.rowCursor
	for gi, g := range groups {
		for _, idx := range g.sessions {
			if idx == cursor {
				return gi
			}
		}
	}
	return -1
}

func sessionGroupKey(session data.Session, mode string) string {
	switch mode {
	case "repository":
		repo := strings.TrimSpace(session.Repository)
		if repo == "" {
			return "(no repository)"
		}
		return repo
	case "status":
		status := strings.TrimSpace(session.Status)
		if status == "" {
			return "(unknown)"
		}
		return status
	case "source":
		src := strings.TrimSpace(string(session.Source))
		if src == "" {
			return "(unknown)"
		}
		return src
	default:
		return ""
	}
}

type sessionGroup struct {
	label      string
	sessions   []int
	mostRecent time.Time
}

func (m Model) buildGroupsWith(groupBy string) []sessionGroup {
	groupMap := map[string]*sessionGroup{}
	var order []string
	for i, session := range m.sessions {
		key := sessionGroupKey(session, groupBy)
		g, exists := groupMap[key]
		if !exists {
			g = &sessionGroup{label: key}
			groupMap[key] = g
			order = append(order, key)
		}
		g.sessions = append(g.sessions, i)
		if session.UpdatedAt.After(g.mostRecent) {
			g.mostRecent = session.UpdatedAt
		}
	}
	groups := make([]sessionGroup, 0, len(order))
	for _, key := range order {
		groups = append(groups, *groupMap[key])
	}
	sort.SliceStable(groups, func(i, j int) bool {
		return groups[i].mostRecent.After(groups[j].mostRecent)
	})
	return groups
}

func (m Model) renderGroupedListWith(groupBy string) string {
	usableWidth := m.width - 4
	if usableWidth < 20 {
		usableWidth = 20
	}
	groups := m.buildGroupsWith(groupBy)
	rows := []string{}
	cursor := m.rowCursor
	if cursor >= len(m.sessions) {
		cursor = len(m.sessions) - 1
	}
	if cursor < 0 {
		cursor = 0
	}

	pageSize := m.pageSize()

	for gi, g := range groups {
		isExpanded := m.expandedGroup == gi

		if !isExpanded {
			// Collapsed: show header with count, highlight if cursor is here
			cursorInGroup := false
			for _, idx := range g.sessions {
				if idx == cursor {
					cursorInGroup = true
					break
				}
			}
			indicator := "▸"
			if cursorInGroup {
				indicator = "▎▸"
			}
			headerLine := fmt.Sprintf("  %s %s (%d)", indicator, g.label, len(g.sessions))
			rows = append(rows, m.sectionHeaderStyle.Render(headerLine))
			continue
		}

		// Expanded group
		headerLine := fmt.Sprintf("  ▾ %s (%d)", g.label, len(g.sessions))
		rows = append(rows, m.sectionHeaderStyle.Render(headerLine))

		// Paginate within group
		if len(g.sessions) <= pageSize {
			for _, idx := range g.sessions {
				session := m.sessions[idx]
				selected := idx == cursor
				rows = append(rows, m.renderRow(idx, session, selected, usableWidth))
			}
		} else {
			// Find cursor position within group
			cursorPos := 0
			for si, idx := range g.sessions {
				if idx == cursor {
					cursorPos = si
					break
				}
			}
			start, end := visibleRange(len(g.sessions), cursorPos, pageSize)
			if start > 0 {
				rows = append(rows, m.tableRowStyle.Render(
					fmt.Sprintf("    ↑ %d more", start)))
			}
			for si := start; si < end; si++ {
				idx := g.sessions[si]
				session := m.sessions[idx]
				selected := idx == cursor
				rows = append(rows, m.renderRow(idx, session, selected, usableWidth))
			}
			if end < len(g.sessions) {
				rows = append(rows, m.tableRowStyle.Render(
					fmt.Sprintf("    ↓ %d more", len(g.sessions)-end)))
			}
		}
	}
	return strings.Join(rows, "\n")
}

// statusColor returns a lipgloss.Color for the given session status.
func (m Model) statusColor(status string) lipgloss.Color {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "running":
		return lipgloss.Color("42")
	case "queued":
		return lipgloss.Color("222")
	case "needs-input":
		return lipgloss.Color("214")
	case "completed":
		return lipgloss.Color("72")
	case "failed":
		return lipgloss.Color("203")
	default:
		return lipgloss.Color("245")
	}
}

// statusSelectedStyle returns a row style tinted by status color.
func (m Model) statusSelectedStyle(status string) lipgloss.Style {
	base := m.tableRowSelected
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "running":
		return base.Background(lipgloss.Color("22"))
	case "needs-input":
		return base.Background(lipgloss.Color("94"))
	case "failed":
		return base.Background(lipgloss.Color("52"))
	case "completed":
		return base.Background(lipgloss.Color("23"))
	default:
		return base
	}
}

// statusGutter renders a colored gutter bar based on session status.
func (m Model) statusGutter(status string, selected bool) string {
	color := m.statusColor(status)
	if !selected {
		return lipgloss.NewStyle().Foreground(color).Faint(true).Render("▎") + " "
	}
	return lipgloss.NewStyle().Foreground(color).Render("▎") + " "
}
