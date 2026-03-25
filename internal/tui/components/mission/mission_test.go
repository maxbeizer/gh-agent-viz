package mission

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/maxbeizer/gh-agent-viz/internal/data"
)

func plainIcon(status string) string {
	switch status {
	case "running":
		return "●"
	case "queued":
		return "○"
	case "completed":
		return "✅"
	case "failed":
		return "❌"
	case "needs-input":
		return "✋"
	default:
		return "⚪"
	}
}

func newTestModel() Model {
	title := lipgloss.NewStyle()
	card := lipgloss.NewStyle()
	cardSel := lipgloss.NewStyle().Bold(true)
	return New(title, card, cardSel, plainIcon, nil)
}

func TestSetSessions(t *testing.T) {
	m := newTestModel()
	sessions := []data.Session{
		{ID: "1", Status: "running", Title: "Fix auth bug", Repository: "owner/repo"},
		{ID: "2", Status: "completed", Title: "Update docs", Repository: "owner/docs"},
	}
	m.SetSessions(sessions)

	if len(m.Cards()) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(m.Cards()))
	}
	if m.Cards()[0].Session.Title != "Fix auth bug" {
		t.Fatalf("expected first card title 'Fix auth bug', got %q", m.Cards()[0].Session.Title)
	}
}

func TestCursorNavigation(t *testing.T) {
	m := newTestModel()
	now := time.Now()
	sessions := []data.Session{
		{ID: "1", Status: "running", Title: "One", Repository: "owner/repo-a", UpdatedAt: now},
		{ID: "2", Status: "running", Title: "Two", Repository: "owner/repo-b", UpdatedAt: now},
		{ID: "3", Status: "running", Title: "Three", Repository: "owner/repo-c", UpdatedAt: now},
	}
	m.SetSessions(sessions)

	if m.Cursor() != 0 {
		t.Fatalf("expected initial cursor 0, got %d", m.Cursor())
	}

	m.MoveCursor(1)
	if m.Cursor() != 1 {
		t.Fatalf("expected cursor 1 after down, got %d", m.Cursor())
	}

	m.MoveCursor(1)
	if m.Cursor() != 2 {
		t.Fatalf("expected cursor 2 after second down, got %d", m.Cursor())
	}

	// Should clamp at end
	m.MoveCursor(1)
	if m.Cursor() != 2 {
		t.Fatalf("expected cursor clamped at 2, got %d", m.Cursor())
	}

	m.MoveCursor(-1)
	if m.Cursor() != 1 {
		t.Fatalf("expected cursor 1 after up, got %d", m.Cursor())
	}

	// Should clamp at start
	m.MoveCursor(-5)
	if m.Cursor() != 0 {
		t.Fatalf("expected cursor clamped at 0, got %d", m.Cursor())
	}
}

func TestCursorNavigationEmpty(t *testing.T) {
	m := newTestModel()
	m.SetSessions(nil)

	m.MoveCursor(1)
	if m.Cursor() != 0 {
		t.Fatalf("expected cursor 0 on empty, got %d", m.Cursor())
	}
}

func TestSelectedSession_FromActivePanel(t *testing.T) {
	m := newTestModel()
	now := time.Now()
	sessions := []data.Session{
		{ID: "1", Status: "running", Title: "First", Repository: "owner/repo", UpdatedAt: now},
	}
	m.SetSessions(sessions)

	// Default focus is Active panel — should select the running session
	m.SetFocus(PanelActive)
	s := m.SelectedSession()
	if s == nil {
		t.Fatal("expected session from active panel")
	}
	if s.ID != "1" {
		t.Fatalf("expected session ID '1', got %q", s.ID)
	}

	// Repos panel returns repo
	m.SetFocus(PanelRepos)
	repo := m.SelectedRepo()
	if repo != "owner/repo" {
		t.Fatalf("expected repo 'owner/repo', got %q", repo)
	}
}

func TestSelectedSessionEmpty(t *testing.T) {
	m := newTestModel()
	m.SetSessions(nil)
	if m.SelectedSession() != nil {
		t.Fatal("expected nil on empty")
	}
}

func TestViewEmpty(t *testing.T) {
	m := newTestModel()
	m.SetSessions(nil)
	view := m.View()
	if !strings.Contains(view, "No sessions") {
		t.Fatalf("expected empty state text, got %q", view)
	}
}

func TestViewRendersSummary(t *testing.T) {
	m := newTestModel()
	m.SetSize(100, 30)
	now := time.Now()
	sessions := []data.Session{
		{ID: "1", Status: "running", Title: "Fix auth bug", Repository: "owner/repo", UpdatedAt: now},
		{ID: "2", Status: "completed", Title: "Update docs", Repository: "owner/docs", PRNumber: 42, UpdatedAt: now.Add(-1 * time.Hour)},
	}
	m.SetSessions(sessions)

	view := m.View()

	// Should show repo names
	if !strings.Contains(view, "owner/repo") {
		t.Fatal("expected repo owner/repo in view")
	}
	if !strings.Contains(view, "owner/docs") {
		t.Fatal("expected repo owner/docs in view")
	}
	// Should show aggregate stats
	if !strings.Contains(view, "active") {
		t.Fatal("expected 'active' in summary")
	}
	if !strings.Contains(view, "done") {
		t.Fatal("expected 'done' in summary")
	}
	// Should have section headers
	if !strings.Contains(view, "Repos") {
		t.Fatal("expected Repos section")
	}
}

func TestDeriveLastAction_Queued(t *testing.T) {
	s := data.Session{Status: "queued"}
	action := DeriveLastAction(s)
	if action != "⏳ Waiting to start" {
		t.Fatalf("expected queued action, got %q", action)
	}
}

func TestDeriveLastAction_Failed(t *testing.T) {
	s := data.Session{Status: "failed"}
	action := DeriveLastAction(s)
	if action != "❌ Session failed" {
		t.Fatalf("expected failed action, got %q", action)
	}
}

func TestDeriveLastAction_CompletedNoPR(t *testing.T) {
	s := data.Session{Status: "completed"}
	action := DeriveLastAction(s)
	if action != "✅ Completed" {
		t.Fatalf("expected completed action, got %q", action)
	}
}

func TestDeriveLastAction_CompletedWithPR(t *testing.T) {
	s := data.Session{Status: "completed", PRNumber: 42}
	action := DeriveLastAction(s)
	if action != "📤 PR #42 ready for review" {
		t.Fatalf("expected PR action, got %q", action)
	}
}

func TestDeriveLastAction_Running(t *testing.T) {
	s := data.Session{Status: "running", Source: data.SourceAgentTask}
	action := DeriveLastAction(s)
	if action != "● Working..." {
		t.Fatalf("expected working action, got %q", action)
	}
}

func TestDeriveLastAction_NeedsInput(t *testing.T) {
	s := data.Session{Status: "needs-input", Source: data.SourceAgentTask}
	action := DeriveLastAction(s)
	if action != "✋ Waiting for input" {
		t.Fatalf("expected needs-input action, got %q", action)
	}
}

func TestFormatAge(t *testing.T) {
	tests := []struct {
		name   string
		t      time.Time
		expect string
	}{
		{"zero", time.Time{}, ""},
		{"recent", time.Now().Add(-30 * time.Second), "just now"},
		{"minutes", time.Now().Add(-12 * time.Minute), "12m ago"},
		{"hours", time.Now().Add(-3 * time.Hour), "3h ago"},
		{"days", time.Now().Add(-48 * time.Hour), "2d ago"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatAge(tt.t)
			if got != tt.expect {
				t.Fatalf("expected %q, got %q", tt.expect, got)
			}
		})
	}
}

func TestAnimFrameUpdates(t *testing.T) {
	m := newTestModel()
	m.SetAnimFrame(5)
	if m.animFrame != 5 {
		t.Fatalf("expected animFrame 5, got %d", m.animFrame)
	}
}

func TestSelectedRepo_ReturnsCorrectRepoAtCursor(t *testing.T) {
	m := newTestModel()
	now := time.Now()
	m.SetSessions([]data.Session{
		{ID: "1", Status: "running", Title: "Task A", Repository: "owner/repo-a", UpdatedAt: now},
		{ID: "2", Status: "running", Title: "Task B", Repository: "owner/repo-b", UpdatedAt: now.Add(-time.Hour)},
	})
	m.SetFocus(PanelRepos)
	m.MoveCursor(1)
	got := m.SelectedRepo()
	if got != "owner/repo-b" {
		t.Fatalf("expected 'owner/repo-b', got %q", got)
	}
}

func TestSelectedRepo_EmptyOnEmptyModel(t *testing.T) {
	m := newTestModel()
	m.SetSessions(nil)
	if got := m.SelectedRepo(); got != "" {
		t.Fatalf("expected empty string on empty model, got %q", got)
	}
}

func TestView_ContainsReposSection(t *testing.T) {
	m := newTestModel()
	m.SetSize(100, 30)
	now := time.Now()
	m.SetSessions([]data.Session{
		{ID: "1", Status: "running", Title: "Task", Repository: "owner/repo", UpdatedAt: now},
	})
	view := m.View()
	if !strings.Contains(view, "Repos") {
		t.Fatal("expected 'Repos' section header in view")
	}
}

func TestView_AttentionSectionWithNeedsInput(t *testing.T) {
	m := newTestModel()
	m.SetSize(100, 30)
	now := time.Now()
	m.SetSessions([]data.Session{
		{ID: "1", Status: "needs-input", Title: "Waiting task", Repository: "owner/repo", Source: data.SourceAgentTask, UpdatedAt: now},
	})
	view := m.View()
	if !strings.Contains(view, "Attention") {
		t.Fatal("expected 'Attention' section when needs-input session exists")
	}
}

func TestSetSize(t *testing.T) {
	m := newTestModel()
	m.SetSize(120, 40)
	if m.width != 120 || m.height != 40 {
		t.Fatalf("expected 120x40, got %dx%d", m.width, m.height)
	}
}

func TestActiveCountMatchesStats(t *testing.T) {
	m := newTestModel()
	m.SetSize(140, 30)
	now := time.Now()
	sessions := []data.Session{
		// needs-input session (stale) — should NOT be counted as active
		{ID: "1", Status: "needs-input", Title: "Waiting for input", UpdatedAt: now.Add(-time.Hour)},
		// running session (recent) — should be counted as active
		{ID: "2", Status: "running", Title: "Working", UpdatedAt: now},
		// completed session
		{ID: "3", Status: "completed", Title: "Done task", UpdatedAt: now.Add(-2 * time.Hour)},
	}
	m.SetSessions(sessions)

	activeCount := len(m.activeSessions())
	statsActive := m.stats.Active

	if activeCount != statsActive {
		t.Fatalf("active panel count (%d) != stats bar count (%d)", activeCount, statsActive)
	}
}

func TestNeedsInputNotInActivePanel(t *testing.T) {
	m := newTestModel()
	now := time.Now()
	sessions := []data.Session{
		{ID: "1", Status: "needs-input", Title: "Stale input", UpdatedAt: now.Add(-time.Hour)},
	}
	m.SetSessions(sessions)

	if len(m.activeSessions()) != 0 {
		t.Fatal("needs-input session should not appear in active panel")
	}
	if m.stats.NeedsInput != 1 {
		t.Fatalf("expected 1 needs-input in stats, got %d", m.stats.NeedsInput)
	}
	if len(m.attention) != 1 {
		t.Fatalf("expected needs-input in attention panel, got %d items", len(m.attention))
	}
}

func TestAllocateBudget_EverythingFits(t *testing.T) {
	requested := []int{3, 2, 1}
	alloc := allocateBudget(requested, 10, 1)
	// Total requested = 6, budget = 10 → everything fits
	for i, r := range requested {
		if alloc[i] != r {
			t.Fatalf("panel %d: expected %d, got %d", i, r, alloc[i])
		}
	}
}

func TestAllocateBudget_Proportional(t *testing.T) {
	requested := []int{20, 10, 10}
	alloc := allocateBudget(requested, 12, 1)
	total := 0
	for _, a := range alloc {
		total += a
		if a < 1 {
			t.Fatalf("allocation below minimum: %d", a)
		}
	}
	if total > 12 {
		t.Fatalf("total allocation %d exceeds budget 12", total)
	}
	// Active panel should get the most since it requested the most
	if alloc[0] <= alloc[1] || alloc[0] <= alloc[2] {
		t.Fatalf("expected panel 0 to get most space, got %v", alloc)
	}
}

func TestAllocateBudget_MinLines(t *testing.T) {
	requested := []int{1, 1, 100}
	alloc := allocateBudget(requested, 6, 1)
	for i, a := range alloc {
		if a < 1 {
			t.Fatalf("panel %d below minimum: %d", i, a)
		}
	}
}

func TestTruncateWithIndicator_NoTruncation(t *testing.T) {
	lines := []string{"a", "b", "c"}
	result := truncateWithIndicator(lines, 5, 3)
	if len(result) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(result))
	}
}

func TestTruncateWithIndicator_Truncates(t *testing.T) {
	lines := []string{"a", "b", "c", "d", "e"}
	result := truncateWithIndicator(lines, 3, 5)
	if len(result) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(result))
	}
	last := result[2]
	if !strings.Contains(last, "more") {
		t.Fatalf("expected overflow indicator, got %q", last)
	}
}

func TestViewMultiPane_FitsHeight(t *testing.T) {
	m := newTestModel()
	m.SetSize(140, 30)

	// Create many sessions to trigger overflow
	now := time.Now()
	var sessions []data.Session
	for i := 0; i < 50; i++ {
		sessions = append(sessions, data.Session{
			ID:         fmt.Sprintf("s%d", i),
			Status:     "completed",
			Title:      fmt.Sprintf("Session %d", i),
			Repository: fmt.Sprintf("owner/repo-%d", i%10),
			UpdatedAt:  now.Add(-time.Duration(i) * time.Hour),
		})
	}
	// Add some active and idle sessions
	sessions = append(sessions, data.Session{
		ID: "active1", Status: "running", Title: "Active task",
		Repository: "owner/repo-0", UpdatedAt: now,
	})
	for i := 0; i < 20; i++ {
		sessions = append(sessions, data.Session{
			ID:         fmt.Sprintf("idle%d", i),
			Status:     "running",
			Title:      fmt.Sprintf("Idle session %d", i),
			Repository: fmt.Sprintf("owner/repo-%d", i%5),
			UpdatedAt:  now.Add(-time.Hour), // idle = active status but old
		})
	}
	m.SetSessions(sessions)

	view := m.View()
	lineCount := strings.Count(view, "\n") + 1
	maxAllowed := 30 - 6 // availHeight
	if lineCount > maxAllowed {
		t.Fatalf("view has %d lines, exceeds max allowed %d", lineCount, maxAllowed)
	}
}

func TestScrollFollowsCursor(t *testing.T) {
	m := newTestModel()
	m.SetSize(140, 30)

	now := time.Now()
	var sessions []data.Session
	// Create 20 repos worth of completed sessions so repos panel overflows
	for i := 0; i < 30; i++ {
		sessions = append(sessions, data.Session{
			ID:         fmt.Sprintf("s%d", i),
			Status:     "completed",
			Title:      fmt.Sprintf("Session %d", i),
			Repository: fmt.Sprintf("owner/repo-%d", i),
			UpdatedAt:  now.Add(-time.Duration(i) * time.Hour),
		})
	}
	m.SetSessions(sessions)

	// First render to set panel heights
	m.View()

	// Focus on repos panel and navigate down past visible area
	m.SetFocus(PanelRepos)
	for i := 0; i < 25; i++ {
		m.MoveCursor(1)
	}

	// Cursor should be at 25
	if m.Cursor() != 25 {
		t.Fatalf("expected cursor at 25, got %d", m.Cursor())
	}

	// Scroll offset should have moved to keep cursor visible
	if m.scrollOffsets[PanelRepos] == 0 {
		t.Fatal("expected scroll offset to advance, but it's still 0")
	}

	// Re-render should show cursor's repo in the view
	view := m.View()
	// Repo names are "owner/repo-N" but may be truncated in display.
	// Check that the scrolled-to region is visible (not stuck at top).
	// After scrolling to cursor=25, we should NOT see repo-0 (it's scrolled away).
	if strings.Contains(view, "owner/repo-0 ") {
		t.Fatal("repo-0 should be scrolled out of view when cursor is at 25")
	}

	// Navigate back up
	for i := 0; i < 25; i++ {
		m.MoveCursor(-1)
	}
	if m.scrollOffsets[PanelRepos] != 0 {
		t.Fatalf("expected scroll offset back to 0, got %d", m.scrollOffsets[PanelRepos])
	}

	view = m.View()
	if !strings.Contains(view, "owner/repo-0") {
		t.Fatal("expected owner/repo-0 visible after scrolling back up")
	}
}

func TestWindowLines_NoOverflow(t *testing.T) {
	lines := []string{"a", "b", "c"}
	result := windowLines(lines, 0, 5, 3, 1)
	if len(result) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(result))
	}
}

func TestWindowLines_ScrolledDown(t *testing.T) {
	lines := []string{"a", "b", "c", "d", "e", "f", "g"}
	result := windowLines(lines, 3, 4, 7, 1)
	// Should have: ▲ indicator, d, e, ▼ indicator
	if len(result) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(result))
	}
	if !strings.Contains(result[0], "above") {
		t.Fatalf("expected above indicator, got %q", result[0])
	}
	if !strings.Contains(result[3], "more") {
		t.Fatalf("expected below indicator, got %q", result[3])
	}
}
