package mission

import (
"fmt"
"sort"
"strings"
"time"

"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
"github.com/maxbeizer/gh-agent-viz/internal/data"
"github.com/maxbeizer/gh-agent-viz/internal/tui/components/sparkline"
)

// SessionCard pairs a session with its derived last-action text.
type SessionCard struct {
Session    data.Session
LastAction string
}

// repoSummary holds aggregate counts for a single repository.
type repoSummary struct {
Name       string
Active     int
Idle       int
NeedsInput int
Done       int
Failed     int
MostRecent time.Time
}

// fleetStats holds global aggregate counts.
type fleetStats struct {
Total      int
Active     int
Idle       int
NeedsInput int
Done       int
Failed     int
}

// attentionItem is a session needing user action.
type attentionItem struct {
Session data.Session
Reason  string
}

// PanelFocus indicates which dashboard panel has keyboard focus.
type PanelFocus int

const (
PanelActive PanelFocus = iota
PanelRecent
PanelAttention
PanelRepos
PanelIdle
)

// Model represents the mission control summary dashboard.
type Model struct {
sessions   []data.Session
stats      fleetStats
repos      []repoSummary
attention  []attentionItem
focus      PanelFocus // which panel has keyboard focus
cursors    [5]int     // per-panel cursor: [active, attention, repos, recent, idle]
scrollOffsets [5]int  // per-panel scroll offset (first visible item index)
panelHeights  [5]int  // per-panel visible content height from last render
// Panel Y ranges for mouse click detection (set during render)
panelYRanges [3][2]int // [panel][start, end] row ranges
tokenUsage map[string]*data.TokenUsage
statusIcon func(string) string
animStatusIcon func(string, int) string
animFrame  int
width      int
height     int
titleStyle lipgloss.Style
cardStyle  lipgloss.Style
cardSelStyle lipgloss.Style
}

// New creates a new mission control model.
func New(
titleStyle lipgloss.Style,
cardStyle lipgloss.Style,
cardSelStyle lipgloss.Style,
statusIconFunc func(string) string,
animStatusIconFunc func(string, int) string,
) Model {
return Model{
statusIcon:     statusIconFunc,
animStatusIcon: animStatusIconFunc,
titleStyle:     titleStyle,
cardStyle:      cardStyle,
cardSelStyle:   cardSelStyle,
width:          80,
height:         24,
}
}

// SetSessions recomputes all dashboard data from sessions.
func (m *Model) SetSessions(sessions []data.Session) {
m.sessions = sessions
m.computeStats()
m.computeRepos()
m.computeAttention()
m.clampCursor()
}

// SetTokenUsage stores the token usage map for cost/model display.
func (m *Model) SetTokenUsage(usage map[string]*data.TokenUsage) {
	m.tokenUsage = usage
}

func (m *Model) computeStats() {
m.stats = fleetStats{Total: len(m.sessions)}
for _, s := range m.sessions {
status := strings.ToLower(strings.TrimSpace(s.Status))
switch {
case status == "needs-input":
m.stats.NeedsInput++
case status == "failed":
m.stats.Failed++
case status == "completed":
m.stats.Done++
case data.SessionIsActiveNotIdle(s):
m.stats.Active++
case data.StatusIsActive(s.Status):
m.stats.Idle++
default:
m.stats.Done++
}
}
}

func (m *Model) computeRepos() {
repoMap := map[string]*repoSummary{}
var order []string
for _, s := range m.sessions {
repo := strings.TrimSpace(s.Repository)
if repo == "" {
repo = "local"
}
r, exists := repoMap[repo]
if !exists {
r = &repoSummary{Name: repo}
repoMap[repo] = r
order = append(order, repo)
}
status := strings.ToLower(strings.TrimSpace(s.Status))
switch {
case status == "needs-input":
r.NeedsInput++
case status == "failed":
r.Failed++
case status == "completed":
r.Done++
case data.SessionIsActiveNotIdle(s):
r.Active++
case data.StatusIsActive(s.Status):
r.Idle++
default:
r.Done++
}
if s.UpdatedAt.After(r.MostRecent) {
r.MostRecent = s.UpdatedAt
}
}
m.repos = make([]repoSummary, 0, len(order))
for _, name := range order {
m.repos = append(m.repos, *repoMap[name])
}
sort.SliceStable(m.repos, func(i, j int) bool {
return m.repos[i].MostRecent.After(m.repos[j].MostRecent)
})
}

func (m *Model) computeAttention() {
m.attention = nil
for _, s := range m.sessions {
level := data.SessionAttentionLevel(s)
if level < data.AttentionWarning {
continue
}
status := strings.ToLower(strings.TrimSpace(s.Status))
var reason string
switch {
case status == "needs-input":
reason = "✋ Input needed"
case status == "failed":
reason = "❌ Failed"
case level == data.AttentionWarning && status == "queued":
reason = "🟡 Queued too long"
case level == data.AttentionWarning:
reason = "🟡 Possibly stuck (idle 4h+)"
default:
reason = "🟡 Needs review"
}
m.attention = append(m.attention, attentionItem{Session: s, Reason: reason})
}
}

// SetSize sets the available rendering dimensions.
func (m *Model) SetSize(width, height int) {
m.width = width
m.height = height
}

// SetAnimFrame updates the animation frame counter.
func (m *Model) SetAnimFrame(frame int) {
m.animFrame = frame
}

// MoveCursor moves the cursor within the focused panel.
func (m *Model) MoveCursor(delta int) {
maxLen := m.focusedPanelLen()
if maxLen == 0 { return }
m.cursors[m.focus] += delta
if m.cursors[m.focus] < 0 { m.cursors[m.focus] = 0 }
if m.cursors[m.focus] >= maxLen { m.cursors[m.focus] = maxLen - 1 }
m.ensureVisible()
}

// ensureVisible adjusts the scroll offset so the cursor is within the visible window.
func (m *Model) ensureVisible() {
p := m.focus
visible := m.panelHeights[p]
if visible <= 0 { visible = 5 } // sane default before first render

// Attention panel uses 2 lines per item; others use 1.
linesPerItem := 1
if p == PanelAttention { linesPerItem = 2 }

maxVisible := visible / linesPerItem
if maxVisible < 1 { maxVisible = 1 }

cursor := m.cursors[p]
if cursor < m.scrollOffsets[p] {
m.scrollOffsets[p] = cursor
}
if cursor >= m.scrollOffsets[p] + maxVisible {
m.scrollOffsets[p] = cursor - maxVisible + 1
}
if m.scrollOffsets[p] < 0 { m.scrollOffsets[p] = 0 }
}

// CyclePanel moves focus to the next/previous panel.
func (m *Model) CyclePanel(delta int) {
// Only cycle through rendered panels: Attention, Active, Recent.
panels := []PanelFocus{PanelAttention, PanelActive, PanelRecent}
current := 0
for i, p := range panels {
if p == m.focus { current = i; break }
}
next := (current + delta + len(panels)) % len(panels)
m.focus = panels[next]
m.clampCursor()
m.ensureVisible()
}

// SetFocus sets the focused panel directly.
func (m *Model) SetFocus(panel PanelFocus) {
m.focus = panel
m.clampCursor()
m.ensureVisible()
}

// Focus returns the currently focused panel.
func (m *Model) Focus() PanelFocus {
return m.focus
}

// FocusPanelAtY sets focus based on a mouse click Y coordinate.
func (m *Model) FocusPanelAtY(y int) {
for i, r := range m.panelYRanges {
if y >= r[0] && y <= r[1] {
m.focus = PanelFocus(i)
return
}
}
}

func (m *Model) focusedPanelLen() int {
switch m.focus {
case PanelActive:
return len(m.activeSessions())
case PanelAttention:
return len(m.attention)
case PanelRepos:
return len(m.repos)
case PanelRecent:
return len(m.recentCompletions(8))
case PanelIdle:
return 0
}
return 0
}

// Cursor returns the current cursor position.
func (m *Model) Cursor() int {
return m.cursors[m.focus]
}

// Cards returns session cards (kept for interface compat).
func (m *Model) Cards() []SessionCard {
cards := make([]SessionCard, len(m.sessions))
for i, s := range m.sessions {
cards[i] = SessionCard{Session: s, LastAction: DeriveLastAction(s)}
}
return cards
}

// SelectedSession returns the session at the cursor in the focused panel, or nil.
func (m *Model) SelectedSession() *data.Session {
switch m.focus {
case PanelActive:
active := m.activeSessions()
idx := m.cursors[PanelActive]
if idx >= 0 && idx < len(active) {
s := active[idx]
return &s
}
case PanelAttention:
idx := m.cursors[PanelAttention]
if idx >= 0 && idx < len(m.attention) {
s := m.attention[idx].Session
return &s
}
case PanelRecent:
recent := m.recentCompletions(8)
idx := m.cursors[PanelRecent]
if idx >= 0 && idx < len(recent) {
s := recent[idx]
return &s
}
case PanelIdle:
idle := m.idleSessions()
idx := m.cursors[PanelIdle]
if idx >= 0 && idx < len(idle) {
s := idle[idx]
return &s
}
case PanelRepos:
return nil
}
return nil
}

// SelectedRepo returns the repository name at the repos cursor position.
func (m *Model) SelectedRepo() string {
idx := m.cursors[PanelRepos]
if len(m.repos) == 0 || idx < 0 || idx >= len(m.repos) {
return ""
}
return m.repos[idx].Name
}

// View renders the summary dashboard.
func (m *Model) View() string {
if len(m.sessions) == 0 {
return m.titleStyle.Render("  No sessions found — run gh agent-viz --demo to explore")
}

// Narrow terminals: fall back to single-column layout
if m.width < 100 {
return m.viewSingleColumn()
}

return m.viewMultiPane()
}

// renderPanel returns a bordered panel with a title header line above the box.
func renderPanel(title string, content string, width, _ int) string {
borderColor := compat.AdaptiveColor{Light: lipgloss.Color("249"), Dark: lipgloss.Color("240")}
titleColor := compat.AdaptiveColor{Light: lipgloss.Color("24"), Dark: lipgloss.Color("75")}

titleRendered := lipgloss.NewStyle().
Bold(true).
Foreground(titleColor).
Render(" " + title)

box := lipgloss.NewStyle().
Border(lipgloss.RoundedBorder()).
BorderForeground(borderColor).
Width(width - 2).
Padding(0, 1).
Render(content)

return titleRendered + "\n" + box
}

// panelChrome is the number of lines each panel adds beyond content:
// 1 title line + 2 border lines (top + bottom of the rounded box).
const panelChrome = 3

// allocateBudget distributes contentBudget lines across panels proportionally
// based on each panel's requested line count. Every panel gets at least minLines.
// Returns the allocated height for each panel.
func allocateBudget(requested []int, contentBudget, minLines int) []int {
	n := len(requested)
	alloc := make([]int, n)

	totalRequested := 0
	for _, r := range requested {
		totalRequested += r
	}

	// Everything fits — give each panel what it asked for.
	if totalRequested <= contentBudget {
		copy(alloc, requested)
		return alloc
	}

	// Distribute proportionally, then assign remainders by largest fraction.
	remaining := contentBudget
	for i, r := range requested {
		a := r * contentBudget / totalRequested
		if a < minLines { a = minLines }
		alloc[i] = a
		remaining -= a
	}
	// Distribute remaining lines to panels that still need more.
	for remaining > 0 {
		bestIdx := -1
		bestNeed := 0
		for i, r := range requested {
			need := r - alloc[i]
			if need > bestNeed {
				bestNeed = need
				bestIdx = i
			}
		}
		if bestIdx < 0 { break }
		alloc[bestIdx]++
		remaining--
	}

	return alloc
}

// truncateWithIndicator truncates lines to maxLines and appends an overflow
// indicator if any lines were hidden. The indicator counts toward maxLines.
func truncateWithIndicator(lines []string, maxLines, totalItems int) []string {
	if len(lines) <= maxLines {
		return lines
	}
	hidden := totalItems - (maxLines - 1) // -1 for the indicator line
	if hidden < 1 { hidden = 1 }
	indicator := lipgloss.NewStyle().Faint(true).Render(
		fmt.Sprintf("  ▼ %d more", hidden))
	result := make([]string, maxLines)
	copy(result, lines[:maxLines-1])
	result[maxLines-1] = indicator
	return result
}

// windowLines returns a visible slice of lines based on scroll offset, adding
// "▲ N above" / "▼ N below" indicators when content extends beyond the window.
// linesPerItem indicates how many lines each logical item occupies (1 for most
// panels, 2 for Active which has a metadata line per session).
func windowLines(lines []string, scrollOffset, maxLines, totalItems, linesPerItem int) []string {
	if len(lines) <= maxLines {
		return lines
	}

	// Convert item-based scroll offset to line offset.
	lineOffset := scrollOffset * linesPerItem
	if lineOffset > len(lines) { lineOffset = len(lines) }

	dim := lipgloss.NewStyle().Faint(true)

	hasAbove := lineOffset > 0

	// After reserving a line for "above" indicator, check if remaining
	// lines fit in the budget.
	availForContent := maxLines
	if hasAbove { availForContent-- }
	hasBelow := len(lines) - lineOffset > availForContent

	// Reserve lines for indicators.
	contentLines := maxLines
	if hasAbove { contentLines-- }
	if hasBelow { contentLines-- }
	if contentLines < 1 { contentLines = 1 }

	end := lineOffset + contentLines
	if end > len(lines) { end = len(lines) }
	visible := lines[lineOffset:end]

	var result []string
	if hasAbove {
		above := scrollOffset
		result = append(result, dim.Render(fmt.Sprintf("  ▲ %d above", above)))
	}
	result = append(result, visible...)
	if hasBelow {
		belowItems := totalItems - scrollOffset - (len(visible) / linesPerItem)
		if belowItems < 1 { belowItems = 1 }
		result = append(result, dim.Render(fmt.Sprintf("  ▼ %d more", belowItems)))
	}

	return result
}

// viewMultiPane renders the 2-column dashboard with Attention as the primary left panel.
func (m *Model) viewMultiPane() string {
dim := lipgloss.NewStyle().Faint(true)
sessionStyle := lipgloss.NewStyle().Foreground(compat.AdaptiveColor{Light: lipgloss.Color("236"), Dark: lipgloss.Color("252")})
cursorStyle := lipgloss.NewStyle().Bold(true)

totalWidth := m.width - 2
leftWidth := totalWidth * 60 / 100
rightWidth := totalWidth - leftWidth

availHeight := m.height - 6
if availHeight < 12 { availHeight = 12 }

focusColor := compat.AdaptiveColor{Light: lipgloss.Color("30"), Dark: lipgloss.Color("73")}

// ── BUILD ALL CONTENT LINES FIRST (no truncation yet) ──

// Active sessions
activeSessions := m.activeSessions()
var activeLines []string
rInnerW := rightWidth - 8
if rInnerW < 30 { rInnerW = 30 }
for i, s := range activeSessions {
icon := m.statusIcon(s.Status)
if m.animStatusIcon != nil && data.SessionIsActiveNotIdle(s) {
icon = m.animStatusIcon(s.Status, m.animFrame)
}
title := s.Title
repo := shortRepo(s.Repository)
age := ""
if !s.CreatedAt.IsZero() {
age = formatDuration(time.Since(s.CreatedAt))
}
rightText := repo
if age != "" { rightText = repo + "  " + age }
maxT := rInnerW - len(rightText) - 8
if maxT < 10 { maxT = 10 }
if len(title) > maxT { title = title[:maxT-1] + "…" }

gutter := "  "
titleRender := sessionStyle.Render(title)
if m.focus == PanelActive && i == m.cursors[PanelActive] {
gutter = "▎ "
titleRender = cursorStyle.Render(title)
}
left := fmt.Sprintf("%s%s %s", gutter, icon, titleRender)
right := dim.Render(rightText)
pad := rInnerW - lipgloss.Width(left) - lipgloss.Width(right)
if pad < 1 { pad = 1 }
activeLines = append(activeLines, left + strings.Repeat(" ", pad) + right)
}
if len(activeLines) == 0 {
activeLines = append(activeLines, dim.Render("  no active sessions"))
}

// Attention — split needs-input (with messages) from failed (collapsed)
var attnLines []string
innerW := leftWidth - 6

var inputItems []attentionItem
var failedItems []attentionItem
for _, item := range m.attention {
st := strings.ToLower(strings.TrimSpace(item.Session.Status))
if st == "failed" {
failedItems = append(failedItems, item)
} else {
inputItems = append(inputItems, item)
}
}

// Needs-input items: 2 lines each (title + assistant message)
for i, item := range inputItems {
title := item.Session.Title
maxT := innerW * 2 / 3
if maxT < 10 { maxT = 10 }
if len(title) > maxT { title = title[:maxT-1] + "…" }
repo := shortRepo(item.Session.Repository)
ago := formatAge(item.Session.UpdatedAt)

gutter := "  "
titleRender := sessionStyle.Render(title)
if m.focus == PanelAttention && i == m.cursors[PanelAttention] {
gutter = "▎ "
titleRender = cursorStyle.Render(title)
}

left := fmt.Sprintf("%s✋ %s", gutter, titleRender)
right := dim.Render(fmt.Sprintf("%s  %s", repo, ago))
pad := innerW - lipgloss.Width(left) - lipgloss.Width(right)
if pad < 1 { pad = 1 }
attnLines = append(attnLines, left + strings.Repeat(" ", pad) + right)

if item.Session.LastAssistantMessage != "" {
msgText := item.Session.LastAssistantMessage
maxMsg := innerW - 6
if maxMsg < 20 { maxMsg = 20 }
if len(msgText) > maxMsg { msgText = msgText[:maxMsg-1] + "…" }
attnLines = append(attnLines, gutter + "  " + dim.Render("💬 " + msgText))
} else {
attnLines = append(attnLines, gutter + "  " + dim.Render("waiting for your response"))
}
}

// Failed: collapsed summary (not one line per failure)
if len(failedItems) > 0 {
if len(inputItems) > 0 {
attnLines = append(attnLines, "")
}
failedStyle := lipgloss.NewStyle().Foreground(compat.AdaptiveColor{Light: lipgloss.Color("160"), Dark: lipgloss.Color("203")})
attnLines = append(attnLines, failedStyle.Render(fmt.Sprintf("  ❌ %d failed sessions", len(failedItems))))
maxShow := 3
if len(failedItems) < maxShow { maxShow = len(failedItems) }
for j := 0; j < maxShow; j++ {
title := failedItems[j].Session.Title
maxT := innerW - 10
if maxT < 10 { maxT = 10 }
if len(title) > maxT { title = title[:maxT-1] + "…" }
ago := formatAge(failedItems[j].Session.UpdatedAt)
attnLines = append(attnLines, dim.Render(fmt.Sprintf("     %s  %s", title, ago)))
}
if len(failedItems) > maxShow {
attnLines = append(attnLines, dim.Render(fmt.Sprintf("     … and %d more", len(failedItems)-maxShow)))
}
}

if len(attnLines) == 0 {
attnLines = append(attnLines, dim.Render("  all clear ✨"))
}

// Recent completions
recentDone := m.recentCompletions(8)
var recentLines []string
for i, s := range recentDone {
icon := "✅"
if strings.EqualFold(s.Status, "failed") { icon = "❌" }
title := s.Title
maxT := rInnerW - 20
if maxT < 10 { maxT = 10 }
if len(title) > maxT { title = title[:maxT-1] + "…" }
ago := formatAge(s.UpdatedAt)
pr := ""
if s.PRNumber > 0 { pr = dim.Render(fmt.Sprintf(" PR #%d", s.PRNumber)) }
gutter := "  "
if m.focus == PanelRecent && i == m.cursors[PanelRecent] {
gutter = "▎ "
title = cursorStyle.Render(title)
}
recentLines = append(recentLines, fmt.Sprintf("%s%s %s%s  %s", gutter, icon, title, pr, dim.Render(ago)))
}
if len(recentLines) == 0 {
recentLines = append(recentLines, dim.Render("  no completions yet"))
}

// Fleet summary
var fleetLines []string
summaryParts := []string{}
if m.stats.Active > 0 { summaryParts = append(summaryParts, fmt.Sprintf("● %d active", m.stats.Active)) }
if m.stats.Idle > 0 { summaryParts = append(summaryParts, fmt.Sprintf("💤 %d idle", m.stats.Idle)) }
if m.stats.NeedsInput > 0 { summaryParts = append(summaryParts, fmt.Sprintf("✋ %d input", m.stats.NeedsInput)) }
if m.stats.Done > 0 { summaryParts = append(summaryParts, fmt.Sprintf("✅ %d done", m.stats.Done)) }
if m.stats.Failed > 0 { summaryParts = append(summaryParts, fmt.Sprintf("❌ %d fail", m.stats.Failed)) }
totalTokens := int64(0)
for _, s := range m.sessions {
if s.Telemetry != nil { totalTokens += s.Telemetry.InputTokens }
}
if totalTokens > 0 {
summaryParts = append(summaryParts, fmt.Sprintf("🪙 %s", data.FormatTokenCount(totalTokens)))
}
// Pulse animation indicator
pulseFrames := []string{"◐", "◓", "◑", "◒"}
pulseChar := pulseFrames[m.animFrame % len(pulseFrames)]
pulseStyle := lipgloss.NewStyle().Foreground(compat.AdaptiveColor{Light: lipgloss.Color("28"), Dark: lipgloss.Color("42")})
fleetLines = append(fleetLines, strings.Join(summaryParts, "  "))
barWidth := rightWidth - 6
if barWidth > 50 { barWidth = 50 }
if barWidth > 0 && m.stats.Total > 0 {
fleetLines = append(fleetLines, m.renderBar(barWidth))
}
todayDone, todayTokens := m.todayStats()
if todayDone > 0 || todayTokens > 0 {
var todayParts []string
if todayDone > 0 { todayParts = append(todayParts, fmt.Sprintf("✅ %d completed", todayDone)) }
if todayTokens > 0 { todayParts = append(todayParts, fmt.Sprintf("🪙 %s", data.FormatTokenCount(todayTokens))) }
fleetLines = append(fleetLines, dim.Render("today: " + strings.Join(todayParts, "  ")))
}

// Cost estimate
totalCost := data.TotalCost(m.tokenUsage)
if totalCost > 0 {
fleetLines = append(fleetLines, dim.Render("est. cost: " + data.FormatCost(totalCost)))
}

// Activity panel content
var activityLines []string

// 24h activity heatmap
hourly := data.HourlyActivity(m.sessions)
hourlyFloats := make([]float64, 24)
for i, v := range hourly {
hourlyFloats[i] = float64(v)
}
heatmapStr := sparkline.RenderHeatmap(hourlyFloats, 24)
heatLabel := lipgloss.NewStyle().Bold(true).Foreground(compat.AdaptiveColor{Light: lipgloss.Color("172"), Dark: lipgloss.Color("214")}).Render("🔥 24h")
activityLines = append(activityLines, heatLabel + " " + heatmapStr + dim.Render("  0h─────────12h────────23h"))

// 7-day trend
daily7 := data.DailySessionCounts(m.sessions, 7)
daily7f := make([]float64, len(daily7))
for i, v := range daily7 { daily7f[i] = float64(v) }
trendStr := sparkline.Render(daily7f, 14)
arrow := sparkline.TrendArrow(daily7f)
trendLabel := lipgloss.NewStyle().Bold(true).Foreground(compat.AdaptiveColor{Light: lipgloss.Color("30"), Dark: lipgloss.Color("75")}).Render("📊 7d")
activityLines = append(activityLines, trendLabel + "  " + trendStr + " " + arrow)

// Model distribution
if m.tokenUsage != nil {
dist := data.ModelDistribution(m.tokenUsage)
if len(dist) > 0 {
activityLines = append(activityLines, "")
modelLabel := lipgloss.NewStyle().Bold(true).Foreground(compat.AdaptiveColor{Light: lipgloss.Color("99"), Dark: lipgloss.Color("141")}).Render("🤖 Models")
activityLines = append(activityLines, modelLabel)
type mc struct { name string; count int }
var models []mc
for k, v := range dist { models = append(models, mc{k, v}) }
sort.Slice(models, func(i, j int) bool { return models[i].count > models[j].count })
for _, mdl := range models {
activityLines = append(activityLines, dim.Render(fmt.Sprintf("  %s: %d sessions", mdl.name, mdl.count)))
}
}
}

// ── BUDGET-BASED HEIGHT ALLOCATION ──
//
// Left column: Attention only (gets full height)
// Right column: Fleet (fixed), Activity (fixed), Active, Recent

leftPanelCount := 1 // Attention
rightPanelCount := 4 // Fleet, Activity, Active, Recent

leftContentBudget := availHeight - (leftPanelCount * panelChrome)
if leftContentBudget < 3 { leftContentBudget = 3 }

// Right column: reserve fleet+activity lines first (fixed), budget the rest.
fleetFixed := len(fleetLines)
activityFixed := len(activityLines)
rightContentBudget := availHeight - (rightPanelCount * panelChrome) - fleetFixed - activityFixed
if rightContentBudget < 4 { rightContentBudget = 4 }

// Left: Attention gets it all
m.panelHeights[PanelAttention] = leftContentBudget
m.ensureVisible()
attnLines = windowLines(attnLines, m.scrollOffsets[PanelAttention], leftContentBudget, len(m.attention), 2)

// Right: Active and Recent split remaining budget
rightRequested := []int{len(activeLines), len(recentLines)}
rightAlloc := allocateBudget(rightRequested, rightContentBudget, 1)
m.panelHeights[PanelActive] = rightAlloc[0]
m.panelHeights[PanelRecent] = rightAlloc[1]
activeLines = windowLines(activeLines, m.scrollOffsets[PanelActive], rightAlloc[0], len(activeSessions), 1)
recentLines = windowLines(recentLines, m.scrollOffsets[PanelRecent], rightAlloc[1], len(recentDone), 1)

// ── RENDER PANELS ──

attnPanel := renderPanelFocused(
fmt.Sprintf("❶ Attention (%d)", len(m.attention)),
strings.Join(attnLines, "\n"), leftWidth, len(attnLines),
m.focus == PanelAttention, focusColor)

fleetTitle := pulseStyle.Render(pulseChar) + " Fleet"
fleetPanel := renderPanel(fleetTitle, strings.Join(fleetLines, "\n"), rightWidth, len(fleetLines))
activityPanel := renderPanel("📊 Activity", strings.Join(activityLines, "\n"), rightWidth, len(activityLines))

activePanel := renderPanelFocused(
fmt.Sprintf("❷ Active (%d)", len(activeSessions)),
strings.Join(activeLines, "\n"), rightWidth, len(activeLines),
m.focus == PanelActive, focusColor)

recentPanel := renderPanelFocused("❸ Recent",
strings.Join(recentLines, "\n"), rightWidth, len(recentLines),
m.focus == PanelRecent, focusColor)

// ── ASSEMBLE COLUMNS ──
leftCol := attnPanel

rightPanels := []string{fleetPanel, activityPanel, activePanel, recentPanel}
rightCol := lipgloss.JoinVertical(lipgloss.Left, rightPanels...)

result := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol)

// Safety clamp: ensure we never exceed the available height.
if lipgloss.Height(result) > availHeight {
lines := strings.Split(result, "\n")
if len(lines) > availHeight {
lines = lines[:availHeight]
}
result = strings.Join(lines, "\n")
}

return result
}

// shortRepo extracts just the repo name from "owner/repo".
func shortRepo(repo string) string {
if repo == "" { return "local" }
if parts := strings.SplitN(repo, "/", 2); len(parts) == 2 { return parts[1] }
return repo
}

// viewSingleColumn renders a tab-based layout for narrow terminals.
// Only the focused section is displayed at full height; a tab bar at
// the top shows all sections with counts for quick switching.
func (m *Model) viewSingleColumn() string {
w := m.width - 4
if w < 40 { w = 40 }

sessionStyle := lipgloss.NewStyle().Foreground(compat.AdaptiveColor{Light: lipgloss.Color("236"), Dark: lipgloss.Color("252")})
cursorStyle := lipgloss.NewStyle().Bold(true)
dim := lipgloss.NewStyle().Faint(true)
tabActive := lipgloss.NewStyle().Bold(true).Foreground(compat.AdaptiveColor{Light: lipgloss.Color("15"), Dark: lipgloss.Color("15")}).Background(compat.AdaptiveColor{Light: lipgloss.Color("24"), Dark: lipgloss.Color("62")}).Padding(0, 1)
tabInactive := lipgloss.NewStyle().Faint(true).Padding(0, 1)

availHeight := m.height - 6
if availHeight < 12 { availHeight = 12 }

// Compute counts for tab bar.
activeSessions := m.activeSessions()
recentDone := m.recentCompletions(8)

// Build tab bar.
type tabDef struct {
	panel PanelFocus
	label string
}
tabs := []tabDef{
	{PanelAttention, fmt.Sprintf("❶ Attn(%d)", len(m.attention))},
	{PanelActive, fmt.Sprintf("❷ Active(%d)", len(activeSessions))},
	{PanelRecent, fmt.Sprintf("❸ Recent(%d)", len(recentDone))},
	{PanelRepos, fmt.Sprintf("❹ Repos(%d)", len(m.repos))},
}

var tabParts []string
for _, tab := range tabs {
	if tab.panel == m.focus {
		tabParts = append(tabParts, tabActive.Render(tab.label))
	} else {
		tabParts = append(tabParts, tabInactive.Render(tab.label))
	}
}
tabBar := strings.Join(tabParts, " ")

// Fleet summary line.
summaryParts := []string{}
if m.stats.Active > 0 { summaryParts = append(summaryParts, fmt.Sprintf("● %d active", m.stats.Active)) }
if m.stats.Idle > 0 { summaryParts = append(summaryParts, fmt.Sprintf("💤 %d idle", m.stats.Idle)) }
if m.stats.Done > 0 { summaryParts = append(summaryParts, fmt.Sprintf("✅ %d done", m.stats.Done)) }
if m.stats.Failed > 0 { summaryParts = append(summaryParts, fmt.Sprintf("❌ %d fail", m.stats.Failed)) }
// Pulse animation
pulseFrames := []string{"◐", "◓", "◑", "◒"}
pulseChar := pulseFrames[m.animFrame % len(pulseFrames)]
pulseStyle := lipgloss.NewStyle().Foreground(compat.AdaptiveColor{Light: lipgloss.Color("28"), Dark: lipgloss.Color("42")})
fleetLine := "  " + pulseStyle.Render(pulseChar + " LIVE") + "  " + strings.Join(summaryParts, "  ")

// Chrome: fleet line + tab bar + blank line = 3 lines.
contentHeight := availHeight - 3
if contentHeight < 3 { contentHeight = 3 }

// Build content for the focused section.
var items []string
var totalCount int

switch m.focus {
case PanelActive:
	totalCount = len(activeSessions)
	for i, s := range activeSessions {
		icon := m.statusIcon(s.Status)
		if m.animStatusIcon != nil && data.SessionIsActiveNotIdle(s) {
			icon = m.animStatusIcon(s.Status, m.animFrame)
		}
		title := s.Title
		maxT := w - 10
		if maxT < 10 { maxT = 10 }
		if len(title) > maxT { title = title[:maxT-1] + "…" }
		gutter := "  "
		titleRender := sessionStyle.Render(title)
		if i == m.cursors[PanelActive] {
			gutter = "▎ "
			titleRender = cursorStyle.Render(title)
		}
		items = append(items, fmt.Sprintf("%s%s %s", gutter, icon, titleRender))
	}
	if len(items) == 0 {
		items = append(items, dim.Render("  no active sessions"))
	}

case PanelAttention:
	totalCount = len(m.attention)
	for i, item := range m.attention {
		title := item.Session.Title
		maxT := w / 2
		if maxT < 10 { maxT = 10 }
		if len(title) > maxT { title = title[:maxT-1] + "…" }
		ago := formatAge(item.Session.UpdatedAt)
		gutter := "  "
		titleRender := sessionStyle.Render(title)
		if i == m.cursors[PanelAttention] {
			gutter = "▎ "
			titleRender = cursorStyle.Render(title)
		}
		items = append(items, fmt.Sprintf("%s%s %s  %s", gutter, item.Reason, titleRender, dim.Render(ago)))
	}
	if len(items) == 0 {
		items = append(items, dim.Render("  all clear ✨"))
	}

case PanelRecent:
	totalCount = len(recentDone)
	for i, s := range recentDone {
		icon := "✅"
		if strings.EqualFold(s.Status, "failed") { icon = "❌" }
		title := s.Title
		maxT := w - 20
		if maxT < 10 { maxT = 10 }
		if len(title) > maxT { title = title[:maxT-1] + "…" }
		ago := formatAge(s.UpdatedAt)
		pr := ""
		if s.PRNumber > 0 { pr = dim.Render(fmt.Sprintf(" PR #%d", s.PRNumber)) }
		gutter := "  "
		if i == m.cursors[PanelRecent] {
			gutter = "▎ "
			title = cursorStyle.Render(title)
		}
		items = append(items, fmt.Sprintf("%s%s %s%s  %s", gutter, icon, title, pr, dim.Render(ago)))
	}
	if len(items) == 0 {
		items = append(items, dim.Render("  no completions yet"))
	}

case PanelRepos:
	totalCount = len(m.repos)
	for i, r := range m.repos {
		selected := i == m.cursors[PanelRepos]
		items = append(items, m.renderRepoRow(r, selected, w))
	}
	if len(items) == 0 {
		items = append(items, dim.Render("  no repos"))
	}
}

// Apply windowed scrolling to the content.
m.panelHeights[m.focus] = contentHeight
m.ensureVisible()
items = windowLines(items, m.scrollOffsets[m.focus], contentHeight, totalCount, 1)

return strings.Join([]string{
	fleetLine,
	tabBar,
	"",
	strings.Join(items, "\n"),
}, "\n")
}

// renderPanelFocused renders a panel with a highlighted border when focused.
func renderPanelFocused(title string, content string, width, _ int, focused bool, focusColor compat.AdaptiveColor) string {
borderColor := compat.AdaptiveColor{Light: lipgloss.Color("249"), Dark: lipgloss.Color("240")}
titleColor := compat.AdaptiveColor{Light: lipgloss.Color("24"), Dark: lipgloss.Color("75")}
if focused {
borderColor = focusColor
titleColor = focusColor
}

titleRendered := lipgloss.NewStyle().
Bold(true).
Foreground(titleColor).
Render(" " + title)

box := lipgloss.NewStyle().
Border(lipgloss.RoundedBorder()).
BorderForeground(borderColor).
Width(width - 2).
Padding(0, 1).
Render(content)

return titleRendered + "\n" + box
}

func (m *Model) renderBar(width int) string {
total := m.stats.Total
if total == 0 {
return ""
}
// Compute segment widths
segments := []struct {
count int
char  string
color string
}{
{m.stats.Active, "█", "42"},
{m.stats.Idle, "▓", "243"},
{m.stats.NeedsInput, "█", "214"},
{m.stats.Done, "░", "72"},
{m.stats.Failed, "█", "203"},
}

var parts []string
for _, seg := range segments {
if seg.count == 0 {
continue
}
segWidth := (seg.count * width) / total
if segWidth < 1 {
segWidth = 1
}
parts = append(parts, lipgloss.NewStyle().
Foreground(lipgloss.Color(seg.color)).
Render(strings.Repeat(seg.char, segWidth)))
}
return strings.Join(parts, "")
}

func (m *Model) renderRepoRow(r repoSummary, selected bool, width int) string {
gutter := "  "
if selected {
gutter = "▎ "
}

// Build counts
var counts []string
if r.Active > 0 {
counts = append(counts, fmt.Sprintf("● %d active", r.Active))
}
if r.Idle > 0 {
counts = append(counts, fmt.Sprintf("💤 %d idle", r.Idle))
}
if r.NeedsInput > 0 {
counts = append(counts, fmt.Sprintf("✋ %d", r.NeedsInput))
}
if r.Done > 0 {
counts = append(counts, fmt.Sprintf("✅ %d done", r.Done))
}
if r.Failed > 0 {
counts = append(counts, fmt.Sprintf("❌ %d", r.Failed))
}

name := r.Name
maxName := width / 3
if len(name) > maxName {
name = name[:maxName-1] + "…"
}

right := strings.Join(counts, "   ")
pad := width - len(gutter) - len(name) - len(right) - 4
if pad < 2 {
pad = 2
}

line := fmt.Sprintf("%s%-*s%s%s", gutter, len(name), name, strings.Repeat(" ", pad), right)

if selected {
return lipgloss.NewStyle().Bold(true).Render(line)
}
return line
}

func (m *Model) clampCursor() {
for i := range m.cursors {
max := 0
switch PanelFocus(i) {
case PanelActive:
max = len(m.activeSessions())
case PanelAttention:
max = len(m.attention)
case PanelRepos:
max = len(m.repos)
case PanelRecent:
max = len(m.recentCompletions(8))
case PanelIdle:
max = len(m.idleSessions())
}
if m.cursors[i] < 0 { m.cursors[i] = 0 }
if max > 0 && m.cursors[i] >= max { m.cursors[i] = max - 1 }
}
}

// activeSessions returns sessions that are currently running or waiting for input.
func (m *Model) activeSessions() []data.Session {
var active []data.Session
for _, s := range m.sessions {
if data.SessionIsActiveNotIdle(s) {
active = append(active, s)
}
}
return active
}

// idleSessions returns sessions with active status but idle for 20+ minutes.
func (m *Model) idleSessions() []data.Session {
var idle []data.Session
for _, s := range m.sessions {
if data.StatusIsActive(s.Status) && !data.SessionIsActiveNotIdle(s) {
idle = append(idle, s)
}
}
return idle
}

// todayStats returns completed count and tokens burned since midnight UTC.
func (m *Model) todayStats() (int, int64) {
now := time.Now().UTC()
todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
completed := 0
tokens := int64(0)
for _, s := range m.sessions {
if s.UpdatedAt.After(todayStart) {
status := strings.ToLower(strings.TrimSpace(s.Status))
if status == "completed" || status == "failed" {
completed++
}
if s.Telemetry != nil {
tokens += s.Telemetry.InputTokens
}
}
}
return completed, tokens
}

// recentCompletions returns the most recent n completed/failed sessions.
func (m *Model) recentCompletions(n int) []data.Session {
var done []data.Session
for _, s := range m.sessions {
status := strings.ToLower(strings.TrimSpace(s.Status))
if status == "completed" || status == "failed" {
done = append(done, s)
}
}
// Sort by UpdatedAt descending
sort.SliceStable(done, func(i, j int) bool {
return done[i].UpdatedAt.After(done[j].UpdatedAt)
})
if len(done) > n {
done = done[:n]
}
return done
}

func formatAge(t time.Time) string {
if t.IsZero() {
return ""
}
d := time.Since(t)
switch {
case d < time.Minute:
return "just now"
case d < time.Hour:
return fmt.Sprintf("%dm ago", int(d.Minutes()))
case d < 24*time.Hour:
return fmt.Sprintf("%dh ago", int(d.Hours()))
default:
return fmt.Sprintf("%dd ago", int(d.Hours()/24))
}
}

func formatDuration(d time.Duration) string {
switch {
case d < time.Minute:
return fmt.Sprintf("%ds", int(d.Seconds()))
case d < time.Hour:
return fmt.Sprintf("%dm", int(d.Minutes()))
case d < 24*time.Hour:
h := int(d.Hours())
m := int(d.Minutes()) % 60
if m > 0 { return fmt.Sprintf("%dh%dm", h, m) }
return fmt.Sprintf("%dh", h)
default:
return fmt.Sprintf("%dd", int(d.Hours()/24))
}
}

// DeriveLastAction returns a brief description of what the session is currently doing.
func DeriveLastAction(s data.Session) string {
status := strings.ToLower(strings.TrimSpace(s.Status))
switch status {
case "queued":
return "⏳ Waiting to start"
case "failed":
return "❌ Session failed"
case "completed":
if s.PRNumber > 0 {
return fmt.Sprintf("📤 PR #%d ready for review", s.PRNumber)
}
return "✅ Completed"
case "needs-input":
if s.Source == data.SourceLocalCopilot {
if msg := data.FetchLastAssistantMessage(s.ID); msg != "" {
truncated := msg
if len(truncated) > 80 {
truncated = truncated[:77] + "..."
}
return "❓ \"" + truncated + "\""
}
}
return "✋ Waiting for input"
case "running":
if s.Source == data.SourceLocalCopilot {
if action := data.FetchLastSessionAction(s); action != "" {
return action
}
}
return "● Working..."
default:
return "● Working..."
}
}
