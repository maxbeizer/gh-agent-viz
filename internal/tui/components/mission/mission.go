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
sessionTitle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "236", Dark: "252"})

var lines []string

// ── Fleet summary ──
lines = append(lines, "")
summaryParts := []string{}
if m.stats.Active > 0 {
summaryParts = append(summaryParts, fmt.Sprintf("● %d active", m.stats.Active))
}
if m.stats.Idle > 0 {
summaryParts = append(summaryParts, fmt.Sprintf("💤 %d idle", m.stats.Idle))
}
if m.stats.NeedsInput > 0 {
summaryParts = append(summaryParts, fmt.Sprintf("✋ %d input", m.stats.NeedsInput))
}
if m.stats.Done > 0 {
summaryParts = append(summaryParts, fmt.Sprintf("✅ %d done", m.stats.Done))
}
if m.stats.Failed > 0 {
summaryParts = append(summaryParts, fmt.Sprintf("❌ %d fail", m.stats.Failed))
}
totalTokens := int64(0)
for _, s := range m.sessions {
if s.Telemetry != nil {
totalTokens += s.Telemetry.InputTokens
}
}
if totalTokens > 0 {
summaryParts = append(summaryParts, fmt.Sprintf("🪙 %s tokens", data.FormatTokenCount(totalTokens)))
}
lines = append(lines, "  "+strings.Join(summaryParts, "  "))
barWidth := w - 4
if barWidth > 60 {
barWidth = 60
}
if barWidth > 0 && m.stats.Total > 0 {
lines = append(lines, "  "+m.renderBar(barWidth))
}

// ── Active sessions (what's running right now) ──
activeSessions := m.activeSessions()
if len(activeSessions) > 0 {
lines = append(lines, "")
lines = append(lines, "  "+sectionHead.Render("Active now"))
lines = append(lines, "  "+dim.Render(strings.Repeat("─", w-4)))
for _, s := range activeSessions {
icon := m.statusIcon(s.Status)
if m.animStatusIcon != nil && data.SessionIsActiveNotIdle(s) {
icon = m.animStatusIcon(s.Status, m.animFrame)
}
title := s.Title
maxTitle := w / 2
if maxTitle < 15 { maxTitle = 15 }
if len(title) > maxTitle { title = title[:maxTitle-1] + "…" }

action := DeriveLastAction(s)
maxAction := w/3
if len(action) > maxAction { action = action[:maxAction-1] + "…" }

repo := s.Repository
if repo == "" { repo = "local" }
if parts := strings.SplitN(repo, "/", 2); len(parts) == 2 { repo = parts[1] }

line := fmt.Sprintf("  %s %s", icon, sessionTitle.Render(title))
right := dim.Render(fmt.Sprintf("%s  %s", action, repo))
pad := w - lipgloss.Width(line) - lipgloss.Width(right)
if pad < 1 { pad = 1 }
lines = append(lines, line + strings.Repeat(" ", pad) + right)
}
}

// ── Needs your attention ──
if len(m.attention) > 0 {
lines = append(lines, "")
lines = append(lines, "  "+sectionHead.Render(fmt.Sprintf("Needs attention (%d)", len(m.attention))))
lines = append(lines, "  "+dim.Render(strings.Repeat("─", w-4)))
for _, item := range m.attention {
title := item.Session.Title
maxTitle := w / 3
if maxTitle < 15 { maxTitle = 15 }
if len(title) > maxTitle { title = title[:maxTitle-1] + "…" }

repo := item.Session.Repository
if repo == "" { repo = "local" }
if parts := strings.SplitN(repo, "/", 2); len(parts) == 2 { repo = parts[1] }

ago := formatAge(item.Session.UpdatedAt)
left := fmt.Sprintf("  %s %s", item.Reason, sessionTitle.Render(title))
right := dim.Render(fmt.Sprintf("%s  %s", repo, ago))
pad := w - lipgloss.Width(left) - lipgloss.Width(right)
if pad < 1 { pad = 1 }
lines = append(lines, left + strings.Repeat(" ", pad) + right)
}
}

// ── Repos ──
lines = append(lines, "")
lines = append(lines, "  "+sectionHead.Render("Repos"))
lines = append(lines, "  "+dim.Render(strings.Repeat("─", w-4)))
for i, r := range m.repos {
selected := i == m.cursor
lines = append(lines, m.renderRepoRow(r, selected, w))
}

// ── Recent completions ──
recentDone := m.recentCompletions(5)
if len(recentDone) > 0 {
lines = append(lines, "")
lines = append(lines, "  "+sectionHead.Render("Recent completions"))
lines = append(lines, "  "+dim.Render(strings.Repeat("─", w-4)))
for _, s := range recentDone {
icon := "✅"
if strings.EqualFold(s.Status, "failed") {
icon = "❌"
}
title := s.Title
maxTitle := w - 30
if maxTitle < 10 {
maxTitle = 10
}
if len(title) > maxTitle {
title = title[:maxTitle-1] + "…"
}
ago := formatAge(s.UpdatedAt)
prInfo := ""
if s.PRNumber > 0 {
prInfo = fmt.Sprintf(" PR #%d", s.PRNumber)
}
lines = append(lines, fmt.Sprintf("  %s %s%s  %s", icon, title, dim.Render(prInfo), dim.Render(ago)))
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
