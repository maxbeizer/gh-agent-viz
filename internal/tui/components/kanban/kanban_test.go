package kanban

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/maxbeizer/gh-agent-viz/internal/data"
)

func newTestModel() Model {
	style := lipgloss.NewStyle()
	icon := func(s string) string {
		switch s {
		case "running":
			return "🟢"
		case "needs-input":
			return "✋"
		case "completed":
			return "✅"
		case "failed":
			return "❌"
		default:
			return "⚪"
		}
	}
	return New(style, style, style, style.Background(lipgloss.Color("237")), icon, nil)
}

func testSessions() []data.Session {
	now := time.Now()
	return []data.Session{
		{ID: "1", Status: "running", Title: "Fix auth bug", Repository: "owner/repo", CreatedAt: now.Add(-12 * time.Minute), UpdatedAt: now.Add(-2 * time.Minute)},
		{ID: "2", Status: "needs-input", Title: "Review PR", Repository: "owner/repo", CreatedAt: now.Add(-5 * time.Minute), UpdatedAt: now.Add(-1 * time.Minute)},
		{ID: "3", Status: "completed", Title: "Migrate DB", Repository: "owner/repo", CreatedAt: now.Add(-1 * time.Hour), UpdatedAt: now.Add(-30 * time.Minute)},
		{ID: "4", Status: "completed", Title: "Add tests", Repository: "owner/repo", CreatedAt: now.Add(-2 * time.Hour), UpdatedAt: now.Add(-1 * time.Hour)},
		{ID: "5", Status: "failed", Title: "Deploy fix", Repository: "owner/repo", CreatedAt: now.Add(-30 * time.Minute), UpdatedAt: now.Add(-25 * time.Minute)},
		{ID: "6", Status: "queued", Title: "Queued task", Repository: "owner/repo", CreatedAt: now.Add(-1 * time.Minute), UpdatedAt: now.Add(-30 * time.Second)},
		{ID: "7", Status: "running", Title: "Idle task", Repository: "owner/repo", CreatedAt: now.Add(-3 * time.Hour), UpdatedAt: now.Add(-40 * time.Minute)},
	}
}

func TestSetSessions_DistributesIntoColumns(t *testing.T) {
	m := newTestModel()
	m.SetSessions(testSessions())

	cols := m.Columns()
	if len(cols) != 3 {
		t.Fatalf("expected 3 columns, got %d", len(cols))
	}

	// IN PROGRESS: running(recent) + needs-input(recent) + queued = 3
	if len(cols[0].Sessions) != 3 {
		t.Errorf("IN PROGRESS column: expected 3 sessions, got %d", len(cols[0].Sessions))
	}
	// IDLE: running(idle 40m) = 1
	if len(cols[1].Sessions) != 1 {
		t.Errorf("IDLE column: expected 1 session, got %d", len(cols[1].Sessions))
	}
	// DONE: completed(2) + failed(1) = 3
	if len(cols[2].Sessions) != 3 {
		t.Errorf("DONE column: expected 3 sessions, got %d", len(cols[2].Sessions))
	}
}

func TestMoveColumn_ClampsAtBounds(t *testing.T) {
	m := newTestModel()
	m.SetSessions(testSessions())

	if m.ColCursor() != 0 {
		t.Fatalf("expected initial column cursor 0, got %d", m.ColCursor())
	}

	m.MoveColumn(-1)
	if m.ColCursor() != 0 {
		t.Errorf("expected column cursor clamped at 0, got %d", m.ColCursor())
	}

	m.MoveColumn(1)
	if m.ColCursor() != 1 {
		t.Errorf("expected column cursor 1, got %d", m.ColCursor())
	}

	m.MoveColumn(10)
	if m.ColCursor() != 2 {
		t.Errorf("expected column cursor clamped at 2, got %d", m.ColCursor())
	}
}

func TestMoveRow_ClampsAtBounds(t *testing.T) {
	m := newTestModel()
	m.SetSessions(testSessions())

	// IN PROGRESS has 3 sessions
	m.MoveRow(1)
	if m.RowCursor() != 1 {
		t.Errorf("expected row cursor 1, got %d", m.RowCursor())
	}

	m.MoveRow(1)
	if m.RowCursor() != 2 {
		t.Errorf("expected row cursor 2, got %d", m.RowCursor())
	}

	m.MoveRow(1)
	if m.RowCursor() != 2 {
		t.Errorf("expected row cursor clamped at 2, got %d", m.RowCursor())
	}

	m.MoveRow(-10)
	if m.RowCursor() != 0 {
		t.Errorf("expected row cursor clamped at 0, got %d", m.RowCursor())
	}
}

func TestMoveRow_EmptyColumn(t *testing.T) {
	m := newTestModel()
	now := time.Now()
	m.SetSessions([]data.Session{
		{ID: "1", Status: "running", Title: "Test", UpdatedAt: now},
	})

	// Move to IDLE column (empty)
	m.MoveColumn(1)
	m.MoveRow(1)
	if m.RowCursor() != 0 {
		t.Errorf("expected row cursor 0 in empty column, got %d", m.RowCursor())
	}
}

func TestSelectedSession_ReturnsCorrectSession(t *testing.T) {
	m := newTestModel()
	m.SetSessions(testSessions())

	s := m.SelectedSession()
	if s == nil {
		t.Fatal("expected non-nil selected session")
	}

	// Move to IDLE column
	m.MoveColumn(1)
	s = m.SelectedSession()
	if s == nil || s.ID != "7" {
		t.Errorf("expected idle session ID '7', got %v", s)
	}
}

func TestSelectedSession_EmptyColumn(t *testing.T) {
	m := newTestModel()
	m.SetSessions([]data.Session{})

	s := m.SelectedSession()
	if s != nil {
		t.Error("expected nil selected session for empty board")
	}
}

func TestColumnWidth_Calculation(t *testing.T) {
	m := newTestModel()
	m.SetSize(100, 24)
	m.SetSessions(testSessions())

	w := m.ColumnWidth()
	// 3 columns, 2 gaps of 1 space each = 98 available, 98/3 = 32
	expected := (100 - 2) / 3
	if w != expected {
		t.Errorf("expected column width %d, got %d", expected, w)
	}
}

func TestColumnWidth_MinWidth(t *testing.T) {
	m := newTestModel()
	m.SetSize(40, 24)
	m.SetSessions(testSessions())

	w := m.ColumnWidth()
	if w < 20 {
		t.Errorf("expected column width >= 20, got %d", w)
	}
}

func TestView_RendersAllColumns(t *testing.T) {
	m := newTestModel()
	m.SetSize(120, 24)
	m.SetSessions(testSessions())

	view := m.View()
	if !strings.Contains(view, "IN PROGRESS") {
		t.Error("expected IN PROGRESS column in view")
	}
	if !strings.Contains(view, "IDLE") {
		t.Error("expected IDLE column in view")
	}
	if !strings.Contains(view, "DONE") {
		t.Error("expected DONE column in view")
	}
}

func TestView_EmptyColumnsShowPlaceholder(t *testing.T) {
	m := newTestModel()
	m.SetSize(120, 24)
	now := time.Now()
	m.SetSessions([]data.Session{
		{ID: "1", Status: "running", Title: "Only runner", UpdatedAt: now},
	})

	view := m.View()
	if !strings.Contains(view, "nothing active") && !strings.Contains(view, "all agents busy") && !strings.Contains(view, "no completed sessions") {
		t.Error("expected placeholder text for empty columns")
	}
}

func TestSetAnimFrame(t *testing.T) {
	m := newTestModel()
	m.SetAnimFrame(42)
	if m.animFrame != 42 {
		t.Errorf("expected animFrame 42, got %d", m.animFrame)
	}
}

func TestMoveColumn_ResetsRowCursor(t *testing.T) {
	m := newTestModel()
	m.SetSessions(testSessions())

	// Move to row 2 in IN PROGRESS column (3 items)
	m.MoveRow(2)
	if m.RowCursor() != 2 {
		t.Fatalf("expected row 2, got %d", m.RowCursor())
	}

	// Move to IDLE column (1 item) — row should clamp to 0
	m.MoveColumn(1)
	if m.RowCursor() != 0 {
		t.Errorf("expected row cursor clamped to 0 when moving to smaller column, got %d", m.RowCursor())
	}
}

func TestView_CardShowsSessionTitle(t *testing.T) {
	m := newTestModel()
	m.SetSize(120, 24)
	m.SetSessions([]data.Session{
		{ID: "1", Status: "running", Title: "Fix auth bug", Repository: "owner/repo"},
	})

	view := m.View()
	if !strings.Contains(view, "Fix auth bug") {
		t.Error("expected session title in view")
	}
}

func TestFormatAge(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{30 * time.Second, "<1m"},
		{12 * time.Minute, "12m"},
		{90 * time.Minute, "1h"},
		{25 * time.Hour, "1d"},
	}
	for _, tt := range tests {
		got := formatAge(time.Now().Add(-tt.d))
		if got != tt.want {
			t.Errorf("formatAge(%v ago) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestFormatAge_ZeroTime(t *testing.T) {
	got := formatAge(time.Time{})
	if got != "—" {
		t.Errorf("formatAge(zero) = %q, want '—'", got)
	}
}
