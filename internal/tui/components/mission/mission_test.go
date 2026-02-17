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
		return "🧑"
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
	sessions := []data.Session{
		{ID: "1", Status: "running", Title: "One"},
		{ID: "2", Status: "running", Title: "Two"},
		{ID: "3", Status: "running", Title: "Three"},
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

func TestSelectedSession(t *testing.T) {
	m := newTestModel()
	sessions := []data.Session{
		{ID: "1", Status: "running", Title: "First"},
		{ID: "2", Status: "completed", Title: "Second"},
	}
	m.SetSessions(sessions)

	s := m.SelectedSession()
	if s == nil || s.ID != "1" {
		t.Fatal("expected selected session to be first")
	}

	m.MoveCursor(1)
	s = m.SelectedSession()
	if s == nil || s.ID != "2" {
		t.Fatal("expected selected session to be second after cursor move")
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

func TestViewRendersCards(t *testing.T) {
	m := newTestModel()
	m.SetSize(100, 30)
	sessions := []data.Session{
		{ID: "1", Status: "running", Title: "Fix auth bug", Repository: "owner/repo", CreatedAt: time.Now().Add(-12 * time.Minute)},
		{ID: "2", Status: "completed", Title: "Update docs", Repository: "owner/docs", PRNumber: 42},
	}
	m.SetSessions(sessions)

	view := m.View()

	if !strings.Contains(view, "Fix auth bug") {
		t.Fatal("expected running session title in view")
	}
	if !strings.Contains(view, "Update docs") {
		t.Fatal("expected completed session title in view")
	}
	if !strings.Contains(view, "owner/repo") {
		t.Fatal("expected repo in view")
	}
	// Should have separator between cards
	if !strings.Contains(view, "───") {
		t.Fatal("expected separator between cards")
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
	if action != "🧑 Waiting for input" {
		t.Fatalf("expected needs-input action, got %q", action)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name   string
		s      data.Session
		expect string
	}{
		{"completed", data.Session{Status: "completed"}, "done"},
		{"failed", data.Session{Status: "failed"}, "failed"},
		{"queued", data.Session{Status: "queued"}, "queued"},
		{"zero time", data.Session{Status: "running"}, ""},
		{"recent", data.Session{Status: "running", CreatedAt: time.Now().Add(-30 * time.Second)}, "⏱ <1m"},
		{"minutes", data.Session{Status: "running", CreatedAt: time.Now().Add(-12 * time.Minute)}, "⏱ 12m"},
		{"hours", data.Session{Status: "running", CreatedAt: time.Now().Add(-3 * time.Hour)}, "⏱ 3h"},
		{"days", data.Session{Status: "running", CreatedAt: time.Now().Add(-48 * time.Hour)}, "⏱ 2d"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.s)
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

func TestSetSize(t *testing.T) {
	m := newTestModel()
	m.SetSize(120, 40)
	if m.width != 120 || m.height != 40 {
		t.Fatalf("expected 120x40, got %dx%d", m.width, m.height)
	}
}
