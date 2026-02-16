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

func TestSetTasksSortsByRecency(t *testing.T) {
	model := newModel()
	now := time.Now()
	model.SetTasks([]data.Session{
		{ID: "1", Status: "running", Title: "Running 1", UpdatedAt: now.Add(-2 * time.Hour)},
		{ID: "2", Status: "completed", Title: "Done 1", UpdatedAt: now.Add(-1 * time.Hour)},
		{ID: "3", Status: "failed", Title: "Failed 1", UpdatedAt: now},
		{ID: "4", Status: "queued", Title: "Running 2", UpdatedAt: now.Add(-3 * time.Hour)},
	})

	if len(model.sessions) != 4 {
		t.Fatalf("expected 4 sessions, got %d", len(model.sessions))
	}
	// Most recent first
	if model.sessions[0].ID != "3" {
		t.Fatalf("expected most recent session (failed, just now) first, got %s", model.sessions[0].ID)
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
	// Default state is loading
	if got := model.View(); !strings.Contains(got, "Loading sessions") {
		t.Fatalf("expected loading state on init, got: %s", got)
	}

	// After data arrives, loading is false — show empty state
	model.loading = false
	if got := model.View(); !strings.Contains(got, "All quiet on the agent front") {
		t.Fatalf("expected empty state, got: %s", got)
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

func TestSetTasks_SortsByRecencyNotPriority(t *testing.T) {
	model := newModel()
	now := time.Now()
	model.SetTasks([]data.Session{
		{ID: "fresh", Status: "running", Title: "Fresh run", UpdatedAt: now.Add(-2 * time.Minute)},
		{ID: "needs", Status: "needs-input", Title: "Need operator", UpdatedAt: now.Add(-40 * time.Minute)},
	})

	if len(model.sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(model.sessions))
	}
	// Most recent first, regardless of status
	if model.sessions[0].ID != "fresh" {
		t.Fatalf("expected most recent session first, got order: %s, %s", model.sessions[0].ID, model.sessions[1].ID)
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
	// Inline detail panel was removed in favor of split-pane layout.
	// The list view should render the row with the task title.
	if !strings.Contains(view, "Task 1") {
		t.Error("expected list view to contain task title")
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
	model.loading = false

	view := model.View()
	if !strings.Contains(view, "All quiet on the agent front") {
		t.Error("expected whimsical empty state message")
	}
	if !strings.Contains(view, "refresh") {
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
	// Inline detail panel was removed; verify the row renders with the title.
	if !strings.Contains(view, "Fallback Session") {
		t.Fatalf("expected list view to contain session title, got: %s", view)
	}
	// "not recorded" still appears in the row meta line for zero timestamps
	if !strings.Contains(view, "not recorded") {
		t.Fatalf("expected friendly timestamp fallback in row meta, got: %s", view)
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
	// Inline detail panel was removed; verify the row renders with the title.
	if !strings.Contains(view, "Agent Session") {
		t.Fatalf("expected list view to contain session title, got: %s", view)
	}
	// The inline "o open PR" action text is no longer rendered in list view.
	if strings.Contains(view, "o open PR") {
		t.Fatalf("expected open PR action to not appear in list view, got: %s", view)
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

func TestDismissSelected_RemovesFromView(t *testing.T) {
	model := newModel()
	now := time.Now()
	model.SetTasks([]data.Session{
		{ID: "1", Status: "running", Title: "Keep", UpdatedAt: now},
		{ID: "2", Status: "failed", Title: "Dismiss Me", UpdatedAt: now.Add(-time.Hour)},
	})

	// Select the second session
	model.MoveCursor(1)
	selected := model.SelectedTask()
	if selected == nil || selected.ID != "2" {
		t.Fatalf("expected session 2 selected, got %v", selected)
	}

	// Dismiss it
	model.DismissSelected()
	if len(model.sessions) != 1 {
		t.Fatalf("expected 1 session after dismiss, got %d", len(model.sessions))
	}
	if model.sessions[0].ID != "1" {
		t.Fatalf("expected session 1 remaining, got %s", model.sessions[0].ID)
	}
}

func TestDismissSelected_PersistsAcrossRefresh(t *testing.T) {
	model := newModel()
	now := time.Now()
	model.SetTasks([]data.Session{
		{ID: "1", Status: "running", Title: "Keep", UpdatedAt: now},
		{ID: "2", Status: "failed", Title: "Dismiss Me", UpdatedAt: now.Add(-time.Hour)},
	})

	model.MoveCursor(1)
	model.DismissSelected()

	// Simulate refresh with same data
	model.SetTasks([]data.Session{
		{ID: "1", Status: "running", Title: "Keep", UpdatedAt: now},
		{ID: "2", Status: "failed", Title: "Dismiss Me", UpdatedAt: now.Add(-time.Hour)},
	})

	if len(model.sessions) != 1 {
		t.Fatalf("dismissed session should stay hidden after refresh, got %d sessions", len(model.sessions))
	}
}

func TestFormatIdleDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{20 * time.Minute, "~20m"},
		{45 * time.Minute, "~45m"},
		{59 * time.Minute, "~59m"},
		{60 * time.Minute, "~1h"},
		{90 * time.Minute, "~1h30m"},
		{2 * time.Hour, "~2h"},
		{2*time.Hour + 15*time.Minute, "~2h15m"},
	}
	for _, tc := range tests {
		got := formatIdleDuration(tc.d)
		if got != tc.want {
			t.Errorf("formatIdleDuration(%v) = %q, want %q", tc.d, got, tc.want)
		}
	}
}

func TestSessionBadge_IdleShowsDuration(t *testing.T) {
	session := data.Session{
		Status:    "running",
		UpdatedAt: time.Now().Add(-30 * time.Minute),
	}
	badge := sessionBadge(session, false, 0)
	if !strings.HasPrefix(badge, "⏸ idle ~") {
		t.Fatalf("expected idle badge with duration, got %q", badge)
	}
	if strings.Contains(badge, "check progress") {
		t.Fatalf("old 'check progress' text should be gone, got %q", badge)
	}
}

func TestCompactDuration(t *testing.T) {
	tests := []struct {
		name string
		dur  time.Duration
		want string
	}{
		{"nil telemetry", 0, ""},
		{"30 seconds", 30 * time.Second, "< 1m"},
		{"5 minutes", 5 * time.Minute, "5m"},
		{"90 minutes", 90 * time.Minute, "1h30m"},
		{"2 hours", 2 * time.Hour, "2h"},
		{"2h30m", 2*time.Hour + 30*time.Minute, "2h30m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := data.Session{ID: "t"}
			if tt.dur > 0 {
				session.Telemetry = &data.SessionTelemetry{Duration: tt.dur}
			}
			got := compactDuration(session)
			if got != tt.want {
				t.Errorf("compactDuration(%v) = %q, want %q", tt.dur, got, tt.want)
			}
		})
	}
}

func TestView_MetaLineShowsDuration(t *testing.T) {
	model := newModel()
	model.SetSize(140, 30)
	model.SetTasks([]data.Session{
		{
			ID:         "1",
			Status:     "running",
			Title:      "Task With Duration",
			Repository: "owner/repo",
			UpdatedAt:  time.Now(),
			Telemetry: &data.SessionTelemetry{
				Duration: 12 * time.Minute,
			},
		},
	})

	view := model.View()
	if !strings.Contains(view, "⏱ 12m") {
		t.Fatalf("expected compact duration in metadata line, got: %s", view)
	}
}

func TestNewestDuplicateShowsCountIndicator(t *testing.T) {
	model := newModel()
	model.SetSize(160, 32)
	now := time.Now()
	model.SetTasks([]data.Session{
		{ID: "dup-1", Status: "running", Title: "Sync roadmap", Repository: "owner/repo", Branch: "main", Source: data.SourceAgentTask, UpdatedAt: now.Add(-70 * time.Minute)},
		{ID: "dup-2", Status: "running", Title: "Sync roadmap", Repository: "owner/repo", Branch: "main", Source: data.SourceAgentTask, UpdatedAt: now.Add(-50 * time.Minute)},
		{ID: "dup-3", Status: "running", Title: "Sync roadmap", Repository: "owner/repo", Branch: "main", Source: data.SourceAgentTask, UpdatedAt: now.Add(-40 * time.Minute)},
	})

	view := model.View()
	if !strings.Contains(view, "(+2 older)") {
		t.Fatalf("expected newest duplicate to show (+2 older) count, got: %s", view)
	}
}

func TestOlderDuplicateShowsTimeSince(t *testing.T) {
	model := newModel()
	model.SetSize(160, 32)
	now := time.Now()
	model.SetTasks([]data.Session{
		{ID: "dup-old", Status: "running", Title: "Sync roadmap", Repository: "owner/repo", Branch: "main", Source: data.SourceAgentTask, UpdatedAt: now.Add(-70 * time.Minute)},
		{ID: "dup-new", Status: "running", Title: "Sync roadmap", Repository: "owner/repo", Branch: "main", Source: data.SourceAgentTask, UpdatedAt: now.Add(-40 * time.Minute)},
	})

	view := model.View()
	if !strings.Contains(view, "↺ quiet duplicate ·") {
		t.Fatalf("expected older duplicate to show time-since suffix, got: %s", view)
	}
}

func TestNonDuplicateSessionsUnaffected(t *testing.T) {
	model := newModel()
	model.SetSize(160, 32)
	now := time.Now()
	model.SetTasks([]data.Session{
		{ID: "a", Status: "running", Title: "Unique task A", Repository: "owner/repo-a", UpdatedAt: now.Add(-5 * time.Minute)},
		{ID: "b", Status: "completed", Title: "Unique task B", Repository: "owner/repo-b", UpdatedAt: now.Add(-10 * time.Minute)},
	})

	view := model.View()
	if strings.Contains(view, "older)") {
		t.Fatalf("non-duplicate sessions should not show duplicate count, got: %s", view)
	}
	if strings.Contains(view, "↺ quiet duplicate") {
		t.Fatalf("non-duplicate sessions should not show duplicate badge, got: %s", view)
	}
}

func TestCycleGroupBy(t *testing.T) {
	model := newModel()
	if model.GroupByLabel() != "" {
		t.Fatalf("expected empty, got %q", model.GroupByLabel())
	}
	model.CycleGroupBy()
	if model.GroupByLabel() != "repo" {
		t.Fatalf("expected repo, got %q", model.GroupByLabel())
	}
	model.CycleGroupBy()
	if model.GroupByLabel() != "status" {
		t.Fatalf("expected status, got %q", model.GroupByLabel())
	}
	model.CycleGroupBy()
	if model.GroupByLabel() != "source" {
		t.Fatalf("expected source, got %q", model.GroupByLabel())
	}
	model.CycleGroupBy()
	if model.GroupByLabel() != "" {
		t.Fatalf("expected empty after wrap, got %q", model.GroupByLabel())
	}
}

func TestViewGroupedByRepository(t *testing.T) {
	model := newModel()
	model.SetSize(120, 40)
	now := time.Now()
	model.SetTasks([]data.Session{
		{ID: "1", Status: "running", Title: "Task A", Repository: "owner/repo-a", UpdatedAt: now},
		{ID: "2", Status: "completed", Title: "Task B", Repository: "owner/repo-b", UpdatedAt: now.Add(-time.Hour)},
		{ID: "3", Status: "running", Title: "Task C", Repository: "owner/repo-a", UpdatedAt: now.Add(-2 * time.Hour)},
	})
	model.CycleGroupBy()
	view := model.View()
	if !strings.Contains(view, "repository: owner/repo-a") {
		t.Fatalf("expected header for repo-a, got: %s", view)
	}
	if !strings.Contains(view, "repository: owner/repo-b") {
		t.Fatalf("expected header for repo-b, got: %s", view)
	}
}

func TestViewGroupedPreservesSortWithinGroups(t *testing.T) {
	model := newModel()
	model.SetSize(120, 40)
	now := time.Now()
	model.SetTasks([]data.Session{
		{ID: "1", Status: "running", Title: "Older Task", Repository: "owner/repo", UpdatedAt: now.Add(-2 * time.Hour)},
		{ID: "2", Status: "running", Title: "Newer Task", Repository: "owner/repo", UpdatedAt: now},
	})
	model.CycleGroupBy()
	view := model.View()
	if strings.Index(view, "Newer Task") > strings.Index(view, "Older Task") {
		t.Fatalf("expected newer before older within group")
	}
}

func TestViewGroupedByStatus(t *testing.T) {
	model := newModel()
	model.SetSize(120, 40)
	now := time.Now()
	model.SetTasks([]data.Session{
		{ID: "1", Status: "running", Title: "Running Task", UpdatedAt: now},
		{ID: "2", Status: "completed", Title: "Done Task", UpdatedAt: now.Add(-time.Hour)},
	})
	model.CycleGroupBy()
	model.CycleGroupBy()
	view := model.View()
	if !strings.Contains(view, "status: running") {
		t.Fatalf("expected running header, got: %s", view)
	}
}

func TestViewNoGroupShowsNoHeaders(t *testing.T) {
	model := newModel()
	model.SetSize(120, 40)
	model.SetTasks([]data.Session{
		{ID: "1", Status: "running", Title: "Task A", Repository: "owner/repo", UpdatedAt: time.Now()},
	})
	view := model.View()
	if strings.Contains(view, "repository:") || strings.Contains(view, "status:") {
		t.Fatalf("expected no headers in None mode, got: %s", view)
	}
}
