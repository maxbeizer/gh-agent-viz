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

func runningColumnIDs(model Model) []string {
	ids := make([]string, 0, len(model.columnSessionIdx[0]))
	for _, idx := range model.columnSessionIdx[0] {
		ids = append(ids, model.sessions[idx].ID)
	}
	return ids
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

func TestSetTasks_PrioritizesNeedsInputAheadOfFreshRunning(t *testing.T) {
	model := newModel()
	now := time.Now()
	model.SetTasks([]data.Session{
		{ID: "fresh", Status: "running", Title: "Fresh run", UpdatedAt: now.Add(-2 * time.Minute)},
		{ID: "needs", Status: "needs-input", Title: "Need operator", UpdatedAt: now.Add(-40 * time.Minute)},
	})

	ids := runningColumnIDs(model)
	if len(ids) != 2 {
		t.Fatalf("expected 2 running sessions, got %d", len(ids))
	}
	if ids[0] != "needs" {
		t.Fatalf("expected needs-input session first, got order %v", ids)
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

	ids := runningColumnIDs(model)
	if len(ids) != 3 {
		t.Fatalf("expected 3 running sessions, got %d", len(ids))
	}
	if ids[len(ids)-1] != "dup-old" {
		t.Fatalf("expected oldest quiet duplicate to be de-emphasized to end, got order %v", ids)
	}

	view := model.View()
	if !strings.Contains(view, "↺ quiet duplicate") {
		t.Fatalf("expected quiet duplicate badge in view, got: %s", view)
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
	model.SetSize(140, 36)
	model.SetTasks(tasks)

	view := model.View()
	if !strings.Contains(view, "Source: agent") {
		t.Error("expected session summary to contain source label")
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

func TestNeedsInputStatusStaysInRunningColumn(t *testing.T) {
	model := newModel()
	model.SetTasks([]data.Session{
		{ID: "1", Status: "needs-input", Title: "Waiting input", UpdatedAt: time.Now()},
	})

	if len(model.columnSessionIdx[0]) != 1 {
		t.Fatalf("expected needs-input session in running column, got %+v", model.columnSessionIdx)
	}
}

func TestView_NarrowModeShowsSingleLaneHint(t *testing.T) {
	model := newModel()
	model.SetSize(70, 24)
	model.SetTasks([]data.Session{
		{ID: "1", Status: "running", Title: "Running Task", UpdatedAt: time.Now()},
		{ID: "2", Status: "completed", Title: "Done Task", UpdatedAt: time.Now()},
	})

	view := model.View()
	if !strings.Contains(view, "COMPACT VIEW") {
		t.Fatalf("expected narrow mode hint, got: %s", view)
	}
}

func TestView_VeryNarrowWidthUsesCompactRows(t *testing.T) {
	model := newModel()
	model.SetSize(12, 24)
	model.SetTasks([]data.Session{
		{ID: "1", Status: "running", Title: "Long Session Title", UpdatedAt: time.Now()},
	})

	view := model.View()
	if !strings.Contains(view, "COMPACT VIEW") {
		t.Fatalf("expected narrow mode hint, got: %s", view)
	}
	if strings.Contains(view, "• just now") {
		t.Fatalf("expected compact row without meta overflow, got: %s", view)
	}
}

func TestView_VeryShortHeightHidesFlightDeck(t *testing.T) {
	model := newModel()
	model.SetSize(120, 20)
	model.SetTasks([]data.Session{
		{ID: "1", Status: "running", Title: "Running Task", UpdatedAt: time.Now()},
	})

	view := model.View()
	if strings.Contains(view, "SESSION SUMMARY") || strings.Contains(view, "Session Summary") {
		t.Fatalf("expected no selected session panel in very short layout, got: %s", view)
	}
}

func TestView_MediumHeightUsesCompactFlightDeck(t *testing.T) {
	model := newModel()
	model.SetSize(120, 28)
	model.SetTasks([]data.Session{
		{ID: "1", Status: "running", Title: "Running Task", UpdatedAt: time.Now()},
	})

	view := model.View()
	if !strings.Contains(view, "Session Summary •") {
		t.Fatalf("expected compact selected session panel in medium layout, got: %s", view)
	}
}

func TestView_WideColumnShowsRepoAndBranch(t *testing.T) {
	model := newModel()
	model.SetSize(300, 36)
	model.SetTasks([]data.Session{
		{
			ID:         "1",
			Status:     "running",
			Title:      "Readable Session",
			Repository: "owner/repository-name",
			Branch:     "feature/super-readable-output",
			UpdatedAt:  time.Now(),
		},
	})

	view := model.View()
	if !strings.Contains(view, "Repository: owner/repository-name @ feature/super-readable-output") {
		t.Fatalf("expected expanded repo+branch details, got: %s", view)
	}
}

func TestView_ShowsAttentionChip(t *testing.T) {
	model := newModel()
	model.SetTasks([]data.Session{
		{
			ID:        "1",
			Status:    "needs-input",
			Title:     "Needs human input",
			UpdatedAt: time.Now(),
		},
	})

	view := model.View()
	if !strings.Contains(view, "needs action 1") {
		t.Fatalf("expected attention chip, got: %s", view)
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
	if !strings.Contains(view, "Needs your action: waiting on your input") {
		t.Fatalf("expected explicit input-needed reason, got: %s", view)
	}
	if !strings.Contains(view, "Needs your action: running but quiet") {
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
	if !strings.Contains(view, "SESSION SUMMARY") {
		t.Fatalf("expected selected session heading, got: %s", view)
	}
	if !strings.Contains(view, "Repository: not available") {
		t.Fatalf("expected friendly repository fallback, got: %s", view)
	}
	if !strings.Contains(view, "Branch: not available") {
		t.Fatalf("expected friendly branch fallback, got: %s", view)
	}
	if !strings.Contains(view, "Last update: not recorded") {
		t.Fatalf("expected friendly timestamp fallback, got: %s", view)
	}
	if strings.Contains(view, "no-repo") || strings.Contains(view, "unknown") {
		t.Fatalf("expected no sentinel placeholders, got: %s", view)
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
	if !strings.Contains(view, "Available actions: enter details • l logs") {
		t.Fatalf("expected selected-session actions to include logs, got: %s", view)
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
