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

func TestSetTasksGroupsIntoColumns(t *testing.T) {
	model := newModel()
	model.SetTasks([]data.Session{
		{ID: "1", Status: "running", Title: "Running 1", UpdatedAt: time.Now()},
		{ID: "2", Status: "completed", Title: "Done 1", UpdatedAt: time.Now()},
		{ID: "3", Status: "failed", Title: "Failed 1", UpdatedAt: time.Now()},
		{ID: "4", Status: "queued", Title: "Running 2", UpdatedAt: time.Now()},
	})

	if got := len(model.columnSessionIdx[0]); got != 2 {
		t.Fatalf("expected 2 running tasks, got %d", got)
	}
	if got := len(model.columnSessionIdx[1]); got != 1 {
		t.Fatalf("expected 1 done task, got %d", got)
	}
	if got := len(model.columnSessionIdx[2]); got != 1 {
		t.Fatalf("expected 1 failed task, got %d", got)
	}
}

func TestMoveCursorWithinActiveColumn(t *testing.T) {
	model := newModel()
	model.SetTasks([]data.Session{
		{ID: "1", Status: "running", Title: "Running 1", UpdatedAt: time.Now()},
		{ID: "2", Status: "running", Title: "Running 2", UpdatedAt: time.Now()},
		{ID: "3", Status: "completed", Title: "Done 1", UpdatedAt: time.Now()},
	})

	model.MoveCursor(1)
	if model.rowCursor[model.activeColumn] != 1 {
		t.Fatalf("expected row cursor to move to 1, got %d", model.rowCursor[model.activeColumn])
	}

	model.MoveCursor(10)
	if model.rowCursor[model.activeColumn] != 1 {
		t.Fatalf("expected row cursor capped at last item, got %d", model.rowCursor[model.activeColumn])
	}
}

func TestMoveColumnAndSelectedTask(t *testing.T) {
	model := newModel()
	model.SetTasks([]data.Session{
		{ID: "1", Status: "running", Title: "Running 1", UpdatedAt: time.Now()},
		{ID: "2", Status: "completed", Title: "Done 1", UpdatedAt: time.Now()},
		{ID: "3", Status: "failed", Title: "Failed 1", UpdatedAt: time.Now()},
	})

	model.MoveColumn(1)
	selected := model.SelectedTask()
	if selected == nil || selected.ID != "2" {
		t.Fatalf("expected done task selected, got %#v", selected)
	}

	model.MoveColumn(1)
	selected = model.SelectedTask()
	if selected == nil || selected.ID != "3" {
		t.Fatalf("expected failed task selected, got %#v", selected)
	}
}

func TestViewShowsKanbanColumns(t *testing.T) {
	model := newModel()
	model.SetTasks([]data.Session{
		{ID: "1", Status: "running", Title: "Running Task", Repository: "owner/repo", Source: data.SourceAgentTask, UpdatedAt: time.Now()},
		{ID: "2", Status: "completed", Title: "Done Task", Repository: "owner/repo", Source: data.SourceLocalCopilot, UpdatedAt: time.Now()},
	})

	view := model.View()
	if !strings.Contains(view, "Running") {
		t.Fatal("expected running column header in view")
	}
	if !strings.Contains(view, "Done") {
		t.Fatal("expected done column header in view")
	}
	if !strings.Contains(view, "Failed") {
		t.Fatal("expected failed column header in view")
	}
	if !strings.Contains(view, "Running Task") {
		t.Fatal("expected running task in view")
	}
	if !strings.Contains(view, "Done Task") {
		t.Fatal("expected done task in view")
	}
}

func TestViewEmptyAndLoadingStates(t *testing.T) {
	model := newModel()
	if got := model.View(); !strings.Contains(got, "No sessions found") {
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

	// Set initial sessions and select the second one in running column
	initialTasks := []data.Session{
		{ID: "1", Status: "running", Title: "Task 1", UpdatedAt: time.Now()},
		{ID: "2", Status: "running", Title: "Task 2", UpdatedAt: time.Now()},
		{ID: "3", Status: "running", Title: "Task 3", UpdatedAt: time.Now()},
	}
	model.SetTasks(initialTasks)
	model.MoveCursor(1) // Select task 2

	// Refresh with reordered sessions (simulating new fetch)
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

func TestView_ShowsSourceBadge(t *testing.T) {
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
	model.SetTasks(tasks)

	view := model.View()
	if !strings.Contains(view, "agent") {
		t.Error("expected view to contain session source")
	}
}

func TestView_ShowsTaskCount(t *testing.T) {
	model := New(
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		func(status string) string { return "" },
	)

	tasks := []data.Session{
		{ID: "1", Status: "running", Title: "Task 1"},
		{ID: "2", Status: "completed", Title: "Task 2"},
		{ID: "3", Status: "failed", Title: "Task 3"},
	}
	model.SetTasks(tasks)

	view := model.View()
	if !strings.Contains(view, "Running (1)") || !strings.Contains(view, "Done (1)") || !strings.Contains(view, "Failed (1)") {
		t.Error("expected view to show per-column task counts")
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
	if !strings.Contains(view, "No sessions found") {
		t.Error("expected improved empty state message")
	}
	if !strings.Contains(view, "refresh with 'r'") {
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

func TestNeedsInputStatusStaysInRunningColumn(t *testing.T) {
	model := newModel()
	model.SetTasks([]data.Session{
		{ID: "1", Status: "needs-input", Title: "Waiting input", UpdatedAt: time.Now()},
	})

	if len(model.columnSessionIdx[0]) != 1 {
		t.Fatalf("expected needs-input session in running column, got %+v", model.columnSessionIdx)
	}
}
