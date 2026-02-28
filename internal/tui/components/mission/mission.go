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

// Model represents the mission control summary dashboard.
type Model struct {
sessions   []data.Session
stats      fleetStats
repos      []repoSummary
attention  []attentionItem
cursor     int // navigates repo rows
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
reason = "🟡 Possibly stuck (idle 2h+)"
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

// MoveCursor moves the cursor by delta (navigates repo rows).
func (m *Model) MoveCursor(delta int) {
if len(m.repos) == 0 {
return
}
m.cursor += delta
m.clampCursor()
}

// Cursor returns the current cursor position.
func (m *Model) Cursor() int {
return m.cursor
}

// Cards returns session cards (kept for interface compat).
func (m *Model) Cards() []SessionCard {
cards := make([]SessionCard, len(m.sessions))
for i, s := range m.sessions {
cards[i] = SessionCard{Session: s, LastAction: DeriveLastAction(s)}
}
return cards
}

// SelectedSession returns nil — summary view doesn't select individual sessions.
// Use SelectedRepo to get the focused repo name.
func (m *Model) SelectedSession() *data.Session {
return nil
}

// SelectedRepo returns the repository name at the cursor position.
func (m *Model) SelectedRepo() string {
if len(m.repos) == 0 || m.cursor < 0 || m.cursor >= len(m.repos) {
return ""
}
return m.repos[m.cursor].Name
}

// View renders the summary dashboard.
func (m *Model) View() string {
if len(m.sessions) == 0 {
return m.titleStyle.Render("  No sessions found — run gh agent-viz --demo to explore")
}

w := m.width - 4
if w < 40 {
w = 40
}

dim := lipgloss.NewStyle().Faint(true)
sectionHead := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "24", Dark: "75"})

var lines []string

// ── Fleet summary bar ──
lines = append(lines, "")
summaryParts := []string{}
if m.stats.Active > 0 {
summaryParts = append(summaryParts, fmt.Sprintf("● %d in progress", m.stats.Active))
}
if m.stats.Idle > 0 {
summaryParts = append(summaryParts, fmt.Sprintf("💤 %d idle", m.stats.Idle))
}
if m.stats.NeedsInput > 0 {
summaryParts = append(summaryParts, fmt.Sprintf("✋ %d waiting on you", m.stats.NeedsInput))
}
if m.stats.Done > 0 {
summaryParts = append(summaryParts, fmt.Sprintf("✅ %d done", m.stats.Done))
}
if m.stats.Failed > 0 {
summaryParts = append(summaryParts, fmt.Sprintf("❌ %d failed", m.stats.Failed))
}
lines = append(lines, "  "+strings.Join(summaryParts, "    "))

// ── Proportional bar ──
barWidth := w - 4
if barWidth > 60 {
barWidth = 60
}
if barWidth > 0 && m.stats.Total > 0 {
bar := m.renderBar(barWidth)
lines = append(lines, "  "+bar)
}
// Total token usage across all sessions
totalTokens := int64(0)
for _, s := range m.sessions {
if s.Telemetry != nil {
totalTokens += s.Telemetry.InputTokens
}
}
if totalTokens > 0 {
lines = append(lines, fmt.Sprintf("  🪙 %s tokens consumed", data.FormatTokenCount(totalTokens)))
}
lines = append(lines, "")

// ── Repos with activity ──
lines = append(lines, "  "+sectionHead.Render("Repos"))
lines = append(lines, "  "+dim.Render(strings.Repeat("─", w-4)))

for i, r := range m.repos {
selected := i == m.cursor
lines = append(lines, m.renderRepoRow(r, selected, w))
}

// ── Needs your attention ──
if len(m.attention) > 0 {
lines = append(lines, "")
lines = append(lines, "  "+sectionHead.Render("Needs your attention"))
lines = append(lines, "  "+dim.Render(strings.Repeat("─", w-4)))
for _, item := range m.attention {
repo := item.Session.Repository
if repo == "" {
repo = "local"
}
ago := formatAge(item.Session.UpdatedAt)
line := fmt.Sprintf("  %s  %-30s  %s  %s", item.Reason, "", repo, ago)
// Build it properly
reasonPart := item.Reason
if len(reasonPart) > w/2 {
reasonPart = reasonPart[:w/2-3] + "..."
}
right := dim.Render(fmt.Sprintf("%s  %s", repo, ago))
pad := w - len(reasonPart) - len(repo) - len(ago) - 6
if pad < 2 {
pad = 2
}
_ = line // suppress unused
lines = append(lines, fmt.Sprintf("  %s%s%s", reasonPart, strings.Repeat(" ", pad), right))
}
}

return strings.Join(lines, "\n")
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
if len(m.repos) == 0 {
m.cursor = 0
return
}
if m.cursor < 0 {
m.cursor = 0
}
if m.cursor >= len(m.repos) {
m.cursor = len(m.repos) - 1
}
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
