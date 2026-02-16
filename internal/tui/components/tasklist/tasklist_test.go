package tasklist

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/maxbeizer/gh-agent-viz/internal/data"
)

func newModel() Model {
	return New(
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		func(string) string { return "•" },
	)
}

func TestSetTasksSortsByPriority(t *testing.T) {
	model := newModel()
	model.SetTasks([]data.Session{
		{ID: "1", Status: "running", Title: "Running 1", UpdatedAt: time.Now()},
		{ID: "2", Status: "completed", Title: "Done 1", UpdatedAt: time.Now()},
		{ID: "3", Status: "failed", Title: "Failed 1", UpdatedAt: time.Now()},
		{ID: "4", Status: "queued", Title: "Running 2", UpdatedAt: time.Now()},
	})

	if len(model.sessions) != 4 {
		t.Fatalf("expected 4 sessions, got %d", len(model.sessions))
	}
	// Failed should be prioritized first (priority 0)
	if model.sessions[0].ID != "3" {
		t.Fatalf("expected failed session first, got %s", model.sessions[0].ID)
	}
}

func TestMoveCursorInFocusedList(t *testing.T) {
	model := newModel()
	model.SetTasks([]data.Session{
		{ID: "1", Status: "running", Title: "Running 1", UpdatedAt: time.Now()},
		{ID: "2", Status: "running", Title: "Running 2", UpdatedAt: time.Now()},
		{ID: "3", Status: "completed", Title: "Done 1", UpdatedAt: time.Now()},
	})

	model.MoveCursor(1)
	if model.rowCursor != 1 {
		t.Fatalf("expected row cursor to move to 1, got %d", model.rowCursor)
	}

	model.MoveCursor(10)
	if model.rowCursor != 2 {
		t.Fatalf("expected row cursor capped at last item, got %d", model.rowCursor)
	}
}

func TestSelectedTaskReturnsCursorSession(t *testing.T) {
	model := newModel()
	model.SetTasks([]data.Session{
		{ID: "1", Status: "running", Title: "Running 1", UpdatedAt: time.Now()},
		{ID: "2", Status: "completed", Title: "Done 1", UpdatedAt: time.Now()},
		{ID: "3", Status: "failed", Title: "Failed 1", UpdatedAt: time.Now()},
	})

	// First session should be selected by default (failed first due to priority)
	selected := model.SelectedTask()
	if selected == nil || selected.ID != "3" {
		t.Fatalf("expected failed task first due to priority, got %#v", selected)
	}

	model.MoveCursor(1)
	selected = model.SelectedTask()
	if selected == nil {
		t.Fatal("expected non-nil selected task")
	}
}

func TestViewShowsFocusedList(t *testing.T) {
	model := newModel()
	model.SetTasks([]data.Session{
		{ID: "1", Status: "running", Title: "Running Task", Repository: "owner/repo", Source: data.SourceAgentTask, UpdatedAt: time.Now()},
		{ID: "2", Status: "completed", Title: "Done Task", Repository: "owner/repo", Source: data.SourceLocalCopilot, UpdatedAt: time.Now()},
	})

	view := model.View()
	if !strings.Contains(view, "Running Task") {
		t.Fatal("expected running task in view")
	}
	if !strings.Contains(view, "Done Task") {
		t.Fatal("expected done task in view")
	}
}

func TestViewEmptyAndLoadingStates(t *testing.T) {
	model := newModel()
	if got := model.View(); !strings.Contains(got, "No sessions to show yet") {
		t.Fatalf("expected empty state, got: %s", got)
	}

	model.loading = true
	if got := model.View(); !strings.Contains(got, "Loading sessions") {
		t.Fatalf("expected loading state, got: %s", got)
	}
}

func TestSetTasks_PreservesSelection(t *testing.T) {
	model := New(
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		func(s string) string { return "" },
	)

	// Set initial sessions and select the second one
	initialTasks := []data.Session{
		{ID: "1", Status: "running", Title: "Task 1", UpdatedAt: time.Now()},
		{ID: "2", Status: "running", Title: "Task 2", UpdatedAt: time.Now()},
		{ID: "3", Status: "running", Title: "Task 3", UpdatedAt: time.Now()},
	}
	model.SetTasks(initialTasks)
	model.MoveCursor(1) // Select task 2

	// Refresh with reordered sessions
	refreshedTasks := []data.Session{
		{ID: "3", Status: "running", Title: "Task 3", UpdatedAt: time.Now().Add(-time.Minute)},
		{ID: "2", Status: "running", Title: "Task 2", UpdatedAt: time.Now()},
		{ID: "1", Status: "running", Title: "Task 1", UpdatedAt: time.Now().Add(-2 * time.Minute)},
	}
	model.SetTasks(refreshedTasks)

	selected := model.SelectedTask()
	if selected == nil || selected.ID != "2" {
		t.Errorf("expected selected task ID '2', got '%v'", selected)
	}
}

func TestSetTasks_PrioritizesNeedsInputFirst(t *testing.T) {
	model := newModel()
	now := time.Now()
	model.SetTasks([]data.Session{
		{ID: "fresh", Status: "running", Title: "Fresh run", UpdatedAt: now.Add(-2 * time.Minute)},
		{ID: "needs", Status: "needs-input", Title: "Need operator", UpdatedAt: now.Add(-40 * time.Minute)},
	})

	if len(model.sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(model.sessions))
	}
	if model.sessions[0].ID != "needs" {
		t.Fatalf("expected needs-input session first, got order: %s, %s", model.sessions[0].ID, model.sessions[1].ID)
	}
}

func TestSetTasks_DeEmphasizesOlderQuietDuplicates(t *testing.T) {
	model := newModel()
	model.SetSize(160, 32)
	now := time.Now()
	model.SetTasks([]data.Session{
		{ID: "dup-old", Status: "running", Title: "Sync roadmap", Repository: "owner/repo", Branch: "main", Source: data.SourceAgentTask, UpdatedAt: now.Add(-70 * time.Minute)},
		{ID: "dup-new", Status: "running", Title: "Sync roadmap", Repository: "owner/repo", Branch: "main", Source: data.SourceAgentTask, UpdatedAt: now.Add(-40 * time.Minute)},
		{ID: "active", Status: "running", Title: "Fresh implementation", UpdatedAt: now.Add(-3 * time.Minute)},
	})

	if len(model.sessions) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(model.sessions))
	}
	// Oldest quiet duplicate should be last
	if model.sessions[len(model.sessions)-1].ID != "dup-old" {
		t.Fatalf("expected oldest quiet duplicate last, got: %s", model.sessions[len(model.sessions)-1].ID)
	}

	view := model.View()
	if !strings.Contains(view, "↺ quiet duplicate") {
		t.Fatalf("expected quiet duplicate badge in view, got: %s", view)
	}
}

func TestView_ShowsSourceInfo(t *testing.T) {
	model := New(
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		func(status string) string { return "✓" },
	)

	tasks := []data.Session{
		{
			ID:         "1",
			Status:     "running",
			Title:      "Task 1",
			Repository: "owner/repo",
			Source:     data.SourceAgentTask,
			UpdatedAt:  time.Now(),
		},
	}
	model.SetSize(140, 36)
	model.SetTasks(tasks)

	view := model.View()
	if !strings.Contains(view, "Source: agent") {
		t.Error("expected inline detail to contain source label")
	}
}

func TestView_ImprovedEmptyState(t *testing.T) {
	model := New(
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		func(s string) string { return "" },
	)

	view := model.View()
	if !strings.Contains(view, "No sessions to show yet") {
		t.Error("expected improved empty state message")
	}
	if !strings.Contains(view, "Press 'r' to refresh") {
		t.Error("expected empty state to include helpful hints")
	}
}

func TestVisibleRangeKeepsCursorInView(t *testing.T) {
	start, end := visibleRange(20, 15, 6)
	if start > 15 || end <= 15 {
		t.Fatalf("expected cursor index 15 to be visible in range [%d,%d)", start, end)
	}
	if end-start != 6 {
		t.Fatalf("expected window size 6, got %d", end-start)
	}
}

func TestTruncate_HandlesUnicodeSafely(t *testing.T) {
	out := truncate("これは日本語です", 6)
	if !strings.HasSuffix(out, "...") {
		t.Fatalf("expected ellipsis suffix, got %q", out)
	}
	if strings.Contains(out, "�") {
		t.Fatalf("expected valid UTF-8 output, got %q", out)
	}
}

func TestView_CardShowsAttentionReason(t *testing.T) {
	model := newModel()
	model.SetSize(140, 36)
	model.SetTasks([]data.Session{
		{
			ID:        "1",
			Status:    "needs-input",
			Title:     "Waiting on operator",
			UpdatedAt: time.Now(),
		},
		{
			ID:        "2",
			Status:    "running",
			Title:     "Quiet but active",
			UpdatedAt: time.Now().Add(-30 * time.Minute),
		},
	})

	view := model.View()
	if !strings.Contains(view, "waiting on your input") {
		t.Fatalf("expected explicit input-needed reason, got: %s", view)
	}
	if !strings.Contains(view, "running but quiet") {
		t.Fatalf("expected explicit quiet-active reason, got: %s", view)
	}
}

func TestView_SelectedSessionUsesFriendlyFallbacks(t *testing.T) {
	model := newModel()
	model.SetSize(140, 36)
	model.SetTasks([]data.Session{
		{
			ID:        "1",
			Status:    "running",
			Title:     "Fallback Session",
			UpdatedAt: time.Time{},
		},
	})

	view := model.View()
	if !strings.Contains(view, "Repository: not available") {
		t.Fatalf("expected friendly repository fallback, got: %s", view)
	}
	if !strings.Contains(view, "Branch: not available") {
		t.Fatalf("expected friendly branch fallback, got: %s", view)
	}
	if !strings.Contains(view, "not recorded") {
		t.Fatalf("expected friendly timestamp fallback, got: %s", view)
	}
}

func TestView_SelectedSessionOmitsOpenPRActionWhenNoPRLinked(t *testing.T) {
	model := newModel()
	model.SetSize(140, 36)
	model.SetTasks([]data.Session{
		{
			ID:         "1",
			Status:     "running",
			Title:      "Agent Session",
			Repository: "owner/repo",
			Source:     data.SourceAgentTask,
			UpdatedAt:  time.Now(),
		},
	})

	view := model.View()
	if !strings.Contains(view, "l logs") {
		t.Fatalf("expected actions to include logs, got: %s", view)
	}
	if strings.Contains(view, "o open PR") {
		t.Fatalf("expected open PR action to be hidden when no PR is linked, got: %s", view)
	}
}

func TestSessionHasLinkedPR(t *testing.T) {
	if sessionHasLinkedPR(data.Session{Source: data.SourceLocalCopilot, PRURL: "https://github.com/maxbeizer/gh-agent-viz/pull/1"}) {
		t.Fatal("expected local sessions to never report linked PR")
	}
	if !sessionHasLinkedPR(data.Session{Source: data.SourceAgentTask, PRURL: "https://github.com/maxbeizer/gh-agent-viz/pull/1"}) {
		t.Fatal("expected agent task with PR URL to report linked PR")
	}
	if !sessionHasLinkedPR(data.Session{Source: data.SourceAgentTask, PRNumber: 42, Repository: "maxbeizer/gh-agent-viz"}) {
		t.Fatal("expected agent task with PR number and repo to report linked PR")
	}
	if sessionHasLinkedPR(data.Session{Source: data.SourceAgentTask, PRNumber: 42}) {
		t.Fatal("expected missing repository to prevent linked PR action")
	}
}

func TestView_ShowsGutterIndicator(t *testing.T) {
	model := newModel()
	model.SetSize(100, 30)
	model.SetTasks([]data.Session{
		{ID: "1", Status: "running", Title: "Selected Task", UpdatedAt: time.Now()},
		{ID: "2", Status: "completed", Title: "Other Task", UpdatedAt: time.Now()},
	})

	view := model.View()
	if !strings.Contains(view, "▎") {
		t.Fatalf("expected gutter indicator for selected row, got: %s", view)
	}
}

func TestMoveColumn_IsNoOpInFocusedMode(t *testing.T) {
	model := newModel()
	model.SetTasks([]data.Session{
		{ID: "1", Status: "running", Title: "Task 1", UpdatedAt: time.Now()},
	})

	model.MoveColumn(1)
	model.MoveColumn(-1)
	// Should not panic or change state
	selected := model.SelectedTask()
	if selected == nil || selected.ID != "1" {
		t.Fatal("MoveColumn should be a no-op in focused mode")
	}
}
