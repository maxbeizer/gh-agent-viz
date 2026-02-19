package mission

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
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

func TestSelectedSession_ReturnsNil(t *testing.T) {
	m := newTestModel()
	now := time.Now()
	sessions := []data.Session{
		{ID: "1", Status: "running", Title: "First", Repository: "owner/repo", UpdatedAt: now},
	}
	m.SetSessions(sessions)

	// Summary view doesn't select individual sessions
	if m.SelectedSession() != nil {
		t.Fatal("expected nil — summary view doesn't select sessions")
	}

	// But SelectedRepo should work
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
	if !strings.Contains(view, "in progress") {
		t.Fatal("expected 'in progress' in summary")
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
	if !strings.Contains(view, "Needs your attention") {
		t.Fatal("expected 'Needs your attention' section when needs-input session exists")
	}
}

func TestSetSize(t *testing.T) {
	m := newTestModel()
	m.SetSize(120, 40)
	if m.width != 120 || m.height != 40 {
		t.Fatalf("expected 120x40, got %dx%d", m.width, m.height)
	}
}
