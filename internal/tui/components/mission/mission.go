package mission

import (
"fmt"
"sort"
"strings"
"time"

"github.com/charmbracelet/lipgloss"
"github.com/maxbeizer/gh-agent-viz/internal/data"
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
// Panel Y ranges for mouse click detection (set during render)
panelYRanges [3][2]int // [panel][start, end] row ranges
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
reason = "🔴 Waiting for input"
if s.Source == data.SourceLocalCopilot {
if msg := data.FetchLastAssistantMessage(s.ID); msg != "" {
if len(msg) > 60 {
msg = msg[:57] + "..."
}
reason = "🔴 \"" + msg + "\""
}
}
case status == "failed":
reason = "🔴 Failed"
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
}

// CyclePanel moves focus to the next/previous panel.
func (m *Model) CyclePanel(delta int) {
m.focus = PanelFocus((int(m.focus) + delta + 5) % 5)
m.clampCursor()
}

// SetFocus sets the focused panel directly.
func (m *Model) SetFocus(panel PanelFocus) {
m.focus = panel
m.clampCursor()
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
return len(m.idleSessions())
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
borderColor := lipgloss.AdaptiveColor{Light: "249", Dark: "240"}
titleColor := lipgloss.AdaptiveColor{Light: "24", Dark: "75"}

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

// viewMultiPane renders the btop-style 2-column dashboard.
func (m *Model) viewMultiPane() string {
dim := lipgloss.NewStyle().Faint(true)
sessionStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "236", Dark: "252"})
cursorStyle := lipgloss.NewStyle().Bold(true)

totalWidth := m.width - 2
leftWidth := totalWidth * 55 / 100
rightWidth := totalWidth - leftWidth

availHeight := m.height - 6
if availHeight < 12 { availHeight = 12 }

focusColor := lipgloss.AdaptiveColor{Light: "30", Dark: "73"}

// ── LEFT COLUMN ──

// Active sessions panel (primary — what you care about most)
activeSessions := m.activeSessions()
var activeLines []string
innerW := leftWidth - 6
for i, s := range activeSessions {
icon := m.statusIcon(s.Status)
if m.animStatusIcon != nil && data.SessionIsActiveNotIdle(s) {
icon = m.animStatusIcon(s.Status, m.animFrame)
}
title := s.Title
maxT := innerW * 2 / 3
if maxT < 10 { maxT = 10 }
if len(title) > maxT { title = title[:maxT-1] + "…" }
action := DeriveLastAction(s)
maxA := innerW / 3
if len(action) > maxA { action = action[:maxA-1] + "…" }

gutter := "  "
titleRender := sessionStyle.Render(title)
if m.focus == PanelActive && i == m.cursors[PanelActive] {
gutter = "▎ "
titleRender = cursorStyle.Render(title)
}
// Line 1: icon + title + action
left := fmt.Sprintf("%s%s %s", gutter, icon, titleRender)
right := dim.Render(action)
pad := innerW - lipgloss.Width(left) - lipgloss.Width(right)
if pad < 1 { pad = 1 }
activeLines = append(activeLines, left + strings.Repeat(" ", pad) + right)

// Line 2: duration + tokens + repo (dim metadata)
var meta []string
if !s.CreatedAt.IsZero() {
meta = append(meta, "⏱ "+formatDuration(time.Since(s.CreatedAt)))
}
if s.Telemetry != nil && s.Telemetry.InputTokens > 0 {
meta = append(meta, "🪙 "+data.FormatTokenCount(s.Telemetry.InputTokens))
}
repo := shortRepo(s.Repository)
meta = append(meta, repo)
activeLines = append(activeLines, gutter + "  " + dim.Render(strings.Join(meta, "  ")))
}
if len(activeLines) == 0 {
activeLines = append(activeLines, dim.Render("  no active sessions"))
}
activeHeight := len(activeLines)
maxActive := availHeight / 2
if activeHeight > maxActive { activeHeight = maxActive; activeLines = activeLines[:maxActive] }
if activeHeight < 2 { activeHeight = 2 }
activeTitle := fmt.Sprintf("Active (%d)", len(activeSessions))
activePanel := renderPanelFocused(activeTitle, strings.Join(activeLines, "\n"), leftWidth, activeHeight, m.focus == PanelActive, focusColor)

// Attention panel
var attnLines []string
for i, item := range m.attention {
title := item.Session.Title
maxT := innerW / 2
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
left := fmt.Sprintf("%s%s %s", gutter, item.Reason, titleRender)
right := dim.Render(fmt.Sprintf("%s %s", repo, ago))
pad := innerW - lipgloss.Width(left) - lipgloss.Width(right)
if pad < 1 { pad = 1 }
attnLines = append(attnLines, left + strings.Repeat(" ", pad) + right)
}
if len(attnLines) == 0 {
attnLines = append(attnLines, dim.Render("  all clear ✨"))
}
attnHeight := len(attnLines)
maxAttn := availHeight / 4
if attnHeight > maxAttn { attnHeight = maxAttn; attnLines = attnLines[:maxAttn] }
if attnHeight < 1 { attnHeight = 1 }
attnTitle := fmt.Sprintf("Attention (%d)", len(m.attention))
attnPanel := renderPanelFocused(attnTitle, strings.Join(attnLines, "\n"), leftWidth, attnHeight, m.focus == PanelAttention, focusColor)

// Idle sessions panel
idleSessions := m.idleSessions()
var idleLines []string
for i, s := range idleSessions {
title := s.Title
maxT := innerW * 2 / 3
if maxT < 10 { maxT = 10 }
if len(title) > maxT { title = title[:maxT-1] + "…" }
gutter := "  "
titleRender := sessionStyle.Render(title)
if m.focus == PanelIdle && i == m.cursors[PanelIdle] {
gutter = "▎ "
titleRender = cursorStyle.Render(title)
}
ago := formatAge(s.UpdatedAt)
repo := shortRepo(s.Repository)
left := fmt.Sprintf("%s💤 %s", gutter, titleRender)
right := dim.Render(fmt.Sprintf("%s  %s", repo, ago))
pad := innerW - lipgloss.Width(left) - lipgloss.Width(right)
if pad < 1 { pad = 1 }
idleLines = append(idleLines, left + strings.Repeat(" ", pad) + right)
}
// Attention
var idlePanel string
if len(idleLines) > 0 {
idleHeight := len(idleLines)
maxIdle := availHeight / 4
if idleHeight > maxIdle { idleHeight = maxIdle; idleLines = idleLines[:maxIdle] }
idlePanel = renderPanelFocused(fmt.Sprintf("Idle (%d)", len(idleSessions)), strings.Join(idleLines, "\n"), rightWidth, idleHeight, m.focus == PanelIdle, focusColor)
}

// ── RIGHT COLUMN ──

// Fleet summary panel
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
fleetLines = append(fleetLines, strings.Join(summaryParts, "  "))
barWidth := rightWidth - 6
if barWidth > 50 { barWidth = 50 }
if barWidth > 0 && m.stats.Total > 0 {
fleetLines = append(fleetLines, m.renderBar(barWidth))
}
// Today stats
todayDone, todayTokens := m.todayStats()
if todayDone > 0 || todayTokens > 0 {
var todayParts []string
if todayDone > 0 { todayParts = append(todayParts, fmt.Sprintf("✅ %d completed", todayDone)) }
if todayTokens > 0 { todayParts = append(todayParts, fmt.Sprintf("🪙 %s", data.FormatTokenCount(todayTokens))) }
fleetLines = append(fleetLines, dim.Render("today: " + strings.Join(todayParts, "  ")))
}
fleetHeight := len(fleetLines)
if fleetHeight < 2 { fleetHeight = 2 }
fleetPanel := renderPanel("Fleet", strings.Join(fleetLines, "\n"), rightWidth, fleetHeight)

// Repos panel
var repoLines []string
rInnerW := rightWidth - 6
for i, r := range m.repos {
selected := (m.focus == PanelRepos) && i == m.cursors[PanelRepos]
repoLines = append(repoLines, m.renderRepoRow(r, selected, rInnerW))
}
repoHeight := len(repoLines)
maxRepo := availHeight / 2
if maxRepo < 3 { maxRepo = 3 }
if repoHeight > maxRepo { repoHeight = maxRepo; repoLines = repoLines[:maxRepo] }
if repoHeight < 1 { repoHeight = 1 }
repoPanel := renderPanelFocused("Repos", strings.Join(repoLines, "\n"), rightWidth, repoHeight, m.focus == PanelRepos, focusColor)

// Recent completions (left column, below Active)
recentDone := m.recentCompletions(8)
var recentLines []string
recentInnerW := leftWidth - 6
for i, s := range recentDone {
icon := "✅"
if strings.EqualFold(s.Status, "failed") { icon = "❌" }
title := s.Title
maxT := recentInnerW - 20
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
recentHeight := len(recentLines)
if recentHeight < 1 { recentHeight = 1 }
maxRecent := availHeight / 3
if recentHeight > maxRecent { recentHeight = maxRecent; recentLines = recentLines[:maxRecent] }
recentPanel := renderPanelFocused("Recent", strings.Join(recentLines, "\n"), leftWidth, recentHeight, m.focus == PanelRecent, focusColor)

// ── ASSEMBLE COLUMNS ──
// Left: Active → Recent → Attention
leftCol := lipgloss.JoinVertical(lipgloss.Left, activePanel, recentPanel, attnPanel)

// Right: Fleet → Repos → Idle
rightPanels := []string{fleetPanel, repoPanel}
if len(idleLines) > 0 {
rightPanels = append(rightPanels, idlePanel)
}
rightCol := lipgloss.JoinVertical(lipgloss.Left, rightPanels...)

return lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol)
}

// shortRepo extracts just the repo name from "owner/repo".
func shortRepo(repo string) string {
if repo == "" { return "local" }
if parts := strings.SplitN(repo, "/", 2); len(parts) == 2 { return parts[1] }
return repo
}

// viewSingleColumn renders the fallback single-column layout for narrow terminals.
func (m *Model) viewSingleColumn() string {
w := m.width - 4
if w < 40 { w = 40 }

sectionHead := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "24", Dark: "75"})
sessionStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "236", Dark: "252"})

var lines []string
lines = append(lines, "")

// Fleet
summaryParts := []string{}
if m.stats.Active > 0 { summaryParts = append(summaryParts, fmt.Sprintf("● %d active", m.stats.Active)) }
if m.stats.Idle > 0 { summaryParts = append(summaryParts, fmt.Sprintf("💤 %d idle", m.stats.Idle)) }
if m.stats.Done > 0 { summaryParts = append(summaryParts, fmt.Sprintf("✅ %d done", m.stats.Done)) }
if m.stats.Failed > 0 { summaryParts = append(summaryParts, fmt.Sprintf("❌ %d fail", m.stats.Failed)) }
lines = append(lines, "  "+strings.Join(summaryParts, "  "))
barWidth := w - 4
if barWidth > 50 { barWidth = 50 }
if barWidth > 0 && m.stats.Total > 0 {
lines = append(lines, "  "+m.renderBar(barWidth))
}

// Active
activeSessions := m.activeSessions()
if len(activeSessions) > 0 {
lines = append(lines, "")
lines = append(lines, "  "+sectionHead.Render("Active now"))
for _, s := range activeSessions {
icon := m.statusIcon(s.Status)
if m.animStatusIcon != nil && data.SessionIsActiveNotIdle(s) {
icon = m.animStatusIcon(s.Status, m.animFrame)
}
title := s.Title
if len(title) > w/2 { title = title[:w/2-1] + "…" }
lines = append(lines, fmt.Sprintf("  %s %s", icon, sessionStyle.Render(title)))
}
}

// Attention
if len(m.attention) > 0 {
lines = append(lines, "")
lines = append(lines, "  "+sectionHead.Render(fmt.Sprintf("Attention (%d)", len(m.attention))))
for _, item := range m.attention {
title := item.Session.Title
if len(title) > w/2 { title = title[:w/2-1] + "…" }
lines = append(lines, fmt.Sprintf("  %s %s", item.Reason, sessionStyle.Render(title)))
}
}

// Repos
lines = append(lines, "")
lines = append(lines, "  "+sectionHead.Render("Repos"))
for i, r := range m.repos {
selected := (m.focus == PanelRepos) && i == m.cursors[PanelRepos]
lines = append(lines, m.renderRepoRow(r, selected, w))
}

return strings.Join(lines, "\n")
}

// renderPanelFocused renders a panel with a highlighted border when focused.
func renderPanelFocused(title string, content string, width, _ int, focused bool, focusColor lipgloss.AdaptiveColor) string {
borderColor := lipgloss.AdaptiveColor{Light: "249", Dark: "240"}
titleColor := lipgloss.AdaptiveColor{Light: "24", Dark: "75"}
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
if data.SessionIsActiveNotIdle(s) || strings.EqualFold(strings.TrimSpace(s.Status), "needs-input") {
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
