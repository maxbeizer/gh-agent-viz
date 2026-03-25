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

	// Should include running, needs-input, failed, queued — NOT completed
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

	// needs-input should come first, then failed, then others
	if m.SessionCount() < 2 {
		t.Fatal("need at least 2 sessions")
	}
	first := m.sessions[0]
	if first.Status != "needs-input" {
		t.Errorf("expected needs-input first, got %q", first.Status)
	}
	second := m.sessions[1]
	if second.Status != "failed" {
		t.Errorf("expected failed second, got %q", second.Status)
	}
}

func TestMoveCursor_ClampsBounds(t *testing.T) {
	m := New(plainIcon, nil)
	m.SetSessions(makeSessions())
	m.SetSize(80, 40)

	// Move past end
	m.MoveCursor(100)
	if m.cursor != m.SessionCount()-1 {
		t.Errorf("cursor should clamp to last item, got %d", m.cursor)
	}

	// Move before start
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
	// First should be needs-input (highest priority)
	if s.Status != "needs-input" {
		t.Errorf("expected needs-input, got %q", s.Status)
	}
}

func TestSelectedSession_EmptyList(t *testing.T) {
	m := New(plainIcon, nil)
	m.SetSessions(nil)

	s := m.SelectedSession()
	if s != nil {
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
	m.DismissSelected() // should not panic
	if m.SessionCount() != 0 {
		t.Errorf("expected 0 sessions, got %d", m.SessionCount())
	}
}

func TestView_EmptyState(t *testing.T) {
	m := New(plainIcon, nil)
	m.SetSize(80, 24)

	// Only completed sessions — view should be empty
	m.SetSessions([]data.Session{
		{ID: "1", Status: "completed", Title: "Done task", UpdatedAt: time.Now()},
	})

	view := m.View()
	if !strings.Contains(view, "All quiet") {
		t.Error("empty state should show 'All quiet' message")
	}
	if !strings.Contains(view, "Recently finished") {
		t.Error("empty state should show 'Recently finished' section")
	}
}

func TestView_EmptyState_NoCompletions(t *testing.T) {
	m := New(plainIcon, nil)
	m.SetSize(80, 24)
	m.SetSessions(nil)

	view := m.View()
	if !strings.Contains(view, "All quiet") {
		t.Error("empty state should show 'All quiet' message")
	}
	if strings.Contains(view, "Recently finished") {
		t.Error("should not show 'Recently finished' when no completions exist")
	}
}

func TestView_RendersFocusedCard(t *testing.T) {
	m := New(plainIcon, nil)
	m.SetSize(120, 40)
	m.SetSessions(makeSessions())

	view := m.View()
	// Focused card should show action hints
	if !strings.Contains(view, "[enter] details") {
		t.Error("focused card should show action hints")
	}
	if !strings.Contains(view, "[o] open PR") {
		t.Error("focused card should show open PR hint")
	}
}

func TestView_UnfocusedCardNoHints(t *testing.T) {
	m := New(plainIcon, nil)
	m.SetSize(120, 40)
	sessions := makeSessions()
	m.SetSessions(sessions)

	// Move cursor to ensure first card is not focused when we check second
	// The view output contains one set of hints for the focused card.
	view := m.View()
	hintCount := strings.Count(view, "[enter] details")
	if hintCount != 1 {
		t.Errorf("expected exactly 1 set of action hints (focused card), got %d", hintCount)
	}
}

func TestView_ShowsActiveCount(t *testing.T) {
	m := New(plainIcon, nil)
	m.SetSize(80, 40)
	m.SetSessions(makeSessions())

	view := m.View()
	if !strings.Contains(view, "4 active") {
		t.Error("header should show active count")
	}
}

func TestRecentCompletions_Limit(t *testing.T) {
	m := New(plainIcon, nil)
	now := time.Now()
	sessions := []data.Session{
		{ID: "1", Status: "completed", Title: "A", UpdatedAt: now},
		{ID: "2", Status: "completed", Title: "B", UpdatedAt: now.Add(-1 * time.Hour)},
		{ID: "3", Status: "completed", Title: "C", UpdatedAt: now.Add(-2 * time.Hour)},
		{ID: "4", Status: "completed", Title: "D", UpdatedAt: now.Add(-3 * time.Hour)},
	}
	m.allSessions = sessions
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

	// Dismiss the first session (needs-input)
	dismissed := m.SelectedSession()
	m.DismissSelected()

	// Re-set sessions — dismissed one should stay dismissed
	m.SetSessions(sessions)
	for i := 0; i < m.SessionCount(); i++ {
		m.cursor = i
		s := m.SelectedSession()
		if s.ID == dismissed.ID {
			t.Error("dismissed session should not reappear after SetSessions")
		}
	}
}

func TestView_RespectsHeightConstraint(t *testing.T) {
m := New(plainIcon, nil)
m.SetSize(80, 20)

now := time.Now()
var sessions []data.Session
for i := 0; i < 20; i++ {
sessions = append(sessions, data.Session{
ID:        fmt.Sprintf("s%d", i),
Status:    "running",
Title:     fmt.Sprintf("Task %d", i),
UpdatedAt: now,
})
}
m.SetSessions(sessions)

view := m.View()
lineCount := strings.Count(view, "\n") + 1
if lineCount > 20 {
t.Errorf("view should fit in 20 lines, got %d lines", lineCount)
}
}

func TestView_ScrollIndicator(t *testing.T) {
m := New(plainIcon, nil)
m.SetSize(80, 20)

now := time.Now()
var sessions []data.Session
for i := 0; i < 10; i++ {
sessions = append(sessions, data.Session{
ID:        fmt.Sprintf("s%d", i),
Status:    "running",
Title:     fmt.Sprintf("Task %d", i),
UpdatedAt: now,
})
}
m.SetSessions(sessions)

view := m.View()
if !strings.Contains(view, "below") {
t.Error("should show scroll indicator when sessions don't all fit")
}
}

func TestView_ShowsRepoBranch(t *testing.T) {
m := New(plainIcon, nil)
m.SetSize(120, 40)
m.SetSessions([]data.Session{
{
ID:         "1",
Status:     "running",
Title:      "Fix auth",
Repository: "github/github",
Branch:     "feature/auth-fix",
UpdatedAt:  time.Now(),
},
})

view := m.View()
if !strings.Contains(view, "github/github") {
t.Error("card should show repository")
}
if !strings.Contains(view, "feature/auth-fix") {
t.Error("card should show branch on line 2")
}
}
