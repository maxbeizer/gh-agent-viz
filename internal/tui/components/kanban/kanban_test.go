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
			return "🧑"
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
	return []data.Session{
		{ID: "1", Status: "running", Title: "Fix auth bug", Repository: "owner/repo", CreatedAt: time.Now().Add(-12 * time.Minute)},
		{ID: "2", Status: "needs-input", Title: "Review PR", Repository: "owner/repo", CreatedAt: time.Now().Add(-5 * time.Minute)},
		{ID: "3", Status: "completed", Title: "Migrate DB", Repository: "owner/repo", CreatedAt: time.Now().Add(-1 * time.Hour)},
		{ID: "4", Status: "completed", Title: "Add tests", Repository: "owner/repo", CreatedAt: time.Now().Add(-2 * time.Hour)},
		{ID: "5", Status: "failed", Title: "Deploy fix", Repository: "owner/repo", CreatedAt: time.Now().Add(-30 * time.Minute)},
		{ID: "6", Status: "queued", Title: "Queued task", Repository: "owner/repo", CreatedAt: time.Now().Add(-1 * time.Minute)},
		{ID: "7", Status: "active", Title: "Active task", Repository: "owner/repo", CreatedAt: time.Now().Add(-3 * time.Minute)},
	}
}

func TestSetSessions_DistributesIntoColumns(t *testing.T) {
	m := newTestModel()
	m.SetSessions(testSessions())

	cols := m.Columns()
	if len(cols) != 4 {
		t.Fatalf("expected 4 columns, got %d", len(cols))
	}

	// Running column: running + queued + active = 3
	if len(cols[0].Sessions) != 3 {
		t.Errorf("RUNNING column: expected 3 sessions, got %d", len(cols[0].Sessions))
	}
	// Needs input: 1
	if len(cols[1].Sessions) != 1 {
		t.Errorf("NEEDS INPUT column: expected 1 session, got %d", len(cols[1].Sessions))
	}
	// Completed: 2
	if len(cols[2].Sessions) != 2 {
		t.Errorf("COMPLETED column: expected 2 sessions, got %d", len(cols[2].Sessions))
	}
	// Failed: 1
	if len(cols[3].Sessions) != 1 {
		t.Errorf("FAILED column: expected 1 session, got %d", len(cols[3].Sessions))
	}
}

func TestMoveColumn_ClampsAtBounds(t *testing.T) {
	m := newTestModel()
	m.SetSessions(testSessions())

	// Start at column 0
	if m.ColCursor() != 0 {
		t.Fatalf("expected initial column cursor 0, got %d", m.ColCursor())
	}

	// Move left past beginning
	m.MoveColumn(-1)
	if m.ColCursor() != 0 {
		t.Errorf("expected column cursor clamped at 0, got %d", m.ColCursor())
	}

	// Move right
	m.MoveColumn(1)
	if m.ColCursor() != 1 {
		t.Errorf("expected column cursor 1, got %d", m.ColCursor())
	}

	// Move to last
	m.MoveColumn(10)
	if m.ColCursor() != 3 {
		t.Errorf("expected column cursor clamped at 3, got %d", m.ColCursor())
	}
}

func TestMoveRow_ClampsAtBounds(t *testing.T) {
	m := newTestModel()
	m.SetSessions(testSessions())

	// Column 0 (RUNNING) has 3 sessions
	m.MoveRow(1)
	if m.RowCursor() != 1 {
		t.Errorf("expected row cursor 1, got %d", m.RowCursor())
	}

	m.MoveRow(1)
	if m.RowCursor() != 2 {
		t.Errorf("expected row cursor 2, got %d", m.RowCursor())
	}

	// Clamp at end
	m.MoveRow(1)
	if m.RowCursor() != 2 {
		t.Errorf("expected row cursor clamped at 2, got %d", m.RowCursor())
	}

	// Move up past beginning
	m.MoveRow(-10)
	if m.RowCursor() != 0 {
		t.Errorf("expected row cursor clamped at 0, got %d", m.RowCursor())
	}
}

func TestMoveRow_EmptyColumn(t *testing.T) {
	m := newTestModel()
	// Only put sessions in RUNNING column
	m.SetSessions([]data.Session{
		{ID: "1", Status: "running", Title: "Test"},
	})

	// Move to NEEDS INPUT column (empty)
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
	if s.ID != "1" {
		t.Errorf("expected session ID '1', got %q", s.ID)
	}

	// Move to second row
	m.MoveRow(1)
	s = m.SelectedSession()
	if s == nil || s.Status != "queued" {
		t.Errorf("expected queued session at row 1")
	}

	// Move to NEEDS INPUT column
	m.MoveColumn(1)
	s = m.SelectedSession()
	if s == nil || s.ID != "2" {
		t.Errorf("expected session ID '2' in NEEDS INPUT column, got %v", s)
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
	// 4 columns, 3 gaps of 1 space each = 97 available, 97/4 = 24
	expected := (100 - 3) / 4
	if w != expected {
		t.Errorf("expected column width %d, got %d", expected, w)
	}
}

func TestColumnWidth_MinWidth(t *testing.T) {
	m := newTestModel()
	m.SetSize(40, 24) // 40 - 3 = 37, 37/4 = 9 which is < 20
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
	if !strings.Contains(view, "RUNNING") {
		t.Error("expected RUNNING column in view")
	}
	if !strings.Contains(view, "NEEDS INPUT") {
		t.Error("expected NEEDS INPUT column in view")
	}
	if !strings.Contains(view, "COMPLETED") {
		t.Error("expected COMPLETED column in view")
	}
	if !strings.Contains(view, "FAILED") {
		t.Error("expected FAILED column in view")
	}
}

func TestView_EmptyColumnsShowPlaceholder(t *testing.T) {
	m := newTestModel()
	m.SetSize(120, 24)
	// Only running sessions
	m.SetSessions([]data.Session{
		{ID: "1", Status: "running", Title: "Only runner"},
	})

	view := m.View()
	if !strings.Contains(view, "(no sessions)") {
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

	// Move to row 2 in RUNNING column (3 items)
	m.MoveRow(2)
	if m.RowCursor() != 2 {
		t.Fatalf("expected row 2, got %d", m.RowCursor())
	}

	// Move to NEEDS INPUT column (1 item) — row should clamp to 0
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
