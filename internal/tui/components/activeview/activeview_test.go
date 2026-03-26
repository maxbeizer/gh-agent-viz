package activeview

import (
"fmt"
"strings"
"testing"
"time"

"github.com/maxbeizer/gh-agent-viz/internal/data"
)

func plainIcon(status string) string { return "[" + status + "]" }

func makeSessions() []data.Session {
now := time.Now()
return []data.Session{
{ID: "1", Status: "running", Title: "Running task", Repository: "github/github", Branch: "feature/foo", UpdatedAt: now, CreatedAt: now.Add(-5 * time.Minute)},
{ID: "2", Status: "completed", Title: "Done task", Repository: "github/github", UpdatedAt: now.Add(-1 * time.Hour)},
{ID: "3", Status: "needs-input", Title: "Blocked task", Repository: "org/repo", PRNumber: 42, UpdatedAt: now.Add(-2 * time.Minute)},
{ID: "4", Status: "failed", Title: "Failed task", Repository: "org/other", UpdatedAt: now.Add(-30 * time.Second)},
{ID: "5", Status: "queued", Title: "Queued task", Repository: "github/github", UpdatedAt: now},
}
}

func TestSetSessions_FiltersToActiveOnly(t *testing.T) {
m := New(plainIcon, nil)
m.SetSessions(makeSessions())
if m.SessionCount() != 4 {
t.Errorf("expected 4 active sessions, got %d", m.SessionCount())
}
for i := 0; i < m.SessionCount(); i++ {
m.cursor = i
s := m.SelectedSession()
if s == nil {
t.Fatalf("nil session at index %d", i)
}
if strings.EqualFold(s.Status, "completed") {
t.Errorf("completed session should not appear in active view")
}
}
}

func TestSetSessions_SortOrder(t *testing.T) {
m := New(plainIcon, nil)
m.SetSessions(makeSessions())
if m.SessionCount() < 2 {
t.Fatal("need at least 2 sessions")
}
if m.sessions[0].Status != "needs-input" {
t.Errorf("expected needs-input first, got %q", m.sessions[0].Status)
}
if m.sessions[1].Status != "failed" {
t.Errorf("expected failed second, got %q", m.sessions[1].Status)
}
}

func TestMoveCursor_ClampsBounds(t *testing.T) {
m := New(plainIcon, nil)
m.SetSessions(makeSessions())
m.SetSize(120, 40)
m.MoveCursor(100)
if m.cursor != m.SessionCount()-1 {
t.Errorf("cursor should clamp to last item, got %d", m.cursor)
}
m.MoveCursor(-200)
if m.cursor != 0 {
t.Errorf("cursor should clamp to 0, got %d", m.cursor)
}
}

func TestSelectedSession_ReturnsCorrectSession(t *testing.T) {
m := New(plainIcon, nil)
m.SetSessions(makeSessions())
s := m.SelectedSession()
if s == nil {
t.Fatal("expected non-nil session")
}
if s.Status != "needs-input" {
t.Errorf("expected needs-input, got %q", s.Status)
}
}

func TestSelectedSession_EmptyList(t *testing.T) {
m := New(plainIcon, nil)
m.SetSessions(nil)
if m.SelectedSession() != nil {
t.Error("expected nil for empty session list")
}
}

func TestDismissSelected(t *testing.T) {
m := New(plainIcon, nil)
m.SetSessions(makeSessions())
initial := m.SessionCount()
m.DismissSelected()
if m.SessionCount() != initial-1 {
t.Errorf("expected %d sessions after dismiss, got %d", initial-1, m.SessionCount())
}
}

func TestDismissSelected_EmptyList(t *testing.T) {
m := New(plainIcon, nil)
m.SetSessions(nil)
m.DismissSelected()
if m.SessionCount() != 0 {
t.Errorf("expected 0 sessions, got %d", m.SessionCount())
}
}

func TestView_EmptyState(t *testing.T) {
m := New(plainIcon, nil)
m.SetSize(120, 30)
m.SetSessions([]data.Session{
{ID: "1", Status: "completed", Title: "Done task", UpdatedAt: time.Now()},
})
view := m.View()
if !strings.Contains(view, "All quiet") {
t.Error("empty state should show 'All quiet' message")
}
if !strings.Contains(view, "Just finished") {
t.Error("empty state should show 'Just finished' section")
}
}

func TestView_EmptyState_NoCompletions(t *testing.T) {
m := New(plainIcon, nil)
m.SetSize(120, 30)
m.SetSessions(nil)
view := m.View()
if !strings.Contains(view, "All quiet") {
t.Error("empty state should show 'All quiet' message")
}
if strings.Contains(view, "Just finished") {
t.Error("should not show 'Just finished' when no completions exist")
}
}

func TestView_ShowsSessionCount(t *testing.T) {
m := New(plainIcon, nil)
m.SetSize(120, 40)
m.SetSessions(makeSessions())
view := m.View()
if !strings.Contains(view, "4") {
t.Error("should show session count")
}
}

func TestView_HorizontalLayout(t *testing.T) {
m := New(plainIcon, nil)
m.SetSize(120, 40)
m.SetSessions(makeSessions())
if !m.useHorizontalLayout() {
t.Error("expected horizontal layout at width 120")
}
view := m.View()
if !strings.Contains(view, "running") {
t.Error("should show status breakdown in list panel")
}
if !strings.Contains(view, "Detail") {
t.Error("should show Detail panel title")
}
}

func TestView_VerticalLayout(t *testing.T) {
m := New(plainIcon, nil)
m.SetSize(80, 30)
m.SetSessions(makeSessions())
if m.useHorizontalLayout() {
t.Error("expected vertical layout at width 80")
}
}

func TestView_DetailShowsRepoBranch(t *testing.T) {
m := New(plainIcon, nil)
m.SetSize(120, 40)
m.SetSessions([]data.Session{
{ID: "1", Status: "running", Title: "Fix auth", Repository: "github/github", Branch: "feature/auth-fix", UpdatedAt: time.Now()},
})
view := m.View()
if !strings.Contains(view, "github/github") {
t.Error("detail should show repository")
}
if !strings.Contains(view, "feature/auth-fix") {
t.Error("detail should show branch")
}
}

func TestRecentCompletions_Limit(t *testing.T) {
m := New(plainIcon, nil)
now := time.Now()
m.allSessions = []data.Session{
{ID: "1", Status: "completed", Title: "A", UpdatedAt: now},
{ID: "2", Status: "completed", Title: "B", UpdatedAt: now.Add(-1 * time.Hour)},
{ID: "3", Status: "completed", Title: "C", UpdatedAt: now.Add(-2 * time.Hour)},
{ID: "4", Status: "completed", Title: "D", UpdatedAt: now.Add(-3 * time.Hour)},
}
recent := m.recentCompletions(3)
if len(recent) != 3 {
t.Errorf("expected 3 recent completions, got %d", len(recent))
}
if recent[0].Title != "A" {
t.Errorf("expected most recent first, got %q", recent[0].Title)
}
}

func TestDismiss_PersistsAcrossSetSessions(t *testing.T) {
m := New(plainIcon, nil)
sessions := makeSessions()
m.SetSessions(sessions)
dismissed := m.SelectedSession()
m.DismissSelected()
m.SetSessions(sessions)
for i := 0; i < m.SessionCount(); i++ {
m.cursor = i
s := m.SelectedSession()
if s.ID == dismissed.ID {
t.Error("dismissed session should not reappear after SetSessions")
}
}
}

func TestView_RespectsHeight(t *testing.T) {
m := New(plainIcon, nil)
m.SetSize(120, 20)
now := time.Now()
var sessions []data.Session
for i := 0; i < 20; i++ {
sessions = append(sessions, data.Session{
ID: fmt.Sprintf("s%d", i), Status: "running",
Title: fmt.Sprintf("Task %d", i), UpdatedAt: now,
})
}
m.SetSessions(sessions)
view := m.View()
lineCount := strings.Count(view, "\n") + 1
if lineCount > 25 {
t.Errorf("view should roughly fit terminal height, got %d lines", lineCount)
}
}

func TestView_ScrollIndicator(t *testing.T) {
m := New(plainIcon, nil)
m.SetSize(120, 20)
now := time.Now()
var sessions []data.Session
for i := 0; i < 10; i++ {
sessions = append(sessions, data.Session{
ID: fmt.Sprintf("s%d", i), Status: "running",
Title: fmt.Sprintf("Task %d", i), UpdatedAt: now,
})
}
m.SetSessions(sessions)
view := m.View()
if !strings.Contains(view, "1/10") {
t.Error("should show scroll position indicator")
}
}

func TestSetSessions_ExcludesIdleSessions(t *testing.T) {
m := New(plainIcon, nil)
now := time.Now()
sessions := []data.Session{
{ID: "1", Status: "running", Title: "Active", UpdatedAt: now},
{ID: "2", Status: "running", Title: "Idle", UpdatedAt: now.Add(-30 * time.Minute)},
{ID: "3", Status: "running", Title: "Very idle", UpdatedAt: now.Add(-2 * time.Hour)},
{ID: "4", Status: "needs-input", Title: "Waiting", UpdatedAt: now.Add(-1 * time.Hour)},
{ID: "5", Status: "failed", Title: "Failed", UpdatedAt: now.Add(-1 * time.Hour)},
}
m.SetSessions(sessions)
// Should include: Active (running+recent), Waiting (needs-input), Failed
// Should exclude: Idle and Very idle (running but stale >20min)
if m.SessionCount() != 3 {
t.Errorf("expected 3 sessions (active+waiting+failed), got %d", m.SessionCount())
for i := 0; i < m.SessionCount(); i++ {
m.cursor = i
s := m.SelectedSession()
t.Logf("  session %d: %s (%s)", i, s.Title, s.Status)
}
}
}
