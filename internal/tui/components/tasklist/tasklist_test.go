package tasklist

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/maxbeizer/gh-agent-viz/internal/data"
)

func TestNew(t *testing.T) {
	titleStyle := lipgloss.NewStyle().Bold(true)
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	rowStyle := lipgloss.NewStyle().Padding(0, 1)
	rowSelectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("237"))
	statusIconFunc := func(status string) string { return "icon" }

	model := New(titleStyle, headerStyle, rowStyle, rowSelectedStyle, statusIconFunc)

	if model.cursor != 0 {
		t.Errorf("expected initial cursor to be 0, got %d", model.cursor)
	}
	if len(model.tasks) != 0 {
		t.Errorf("expected empty tasks list, got %d tasks", len(model.tasks))
	}
	if model.loading {
		t.Error("expected loading to be false initially")
	}
}

func TestSetTasks(t *testing.T) {
	model := New(
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		func(s string) string { return "" },
	)

	tasks := []data.AgentTask{
		{ID: "1", Status: "running", Title: "Task 1"},
		{ID: "2", Status: "completed", Title: "Task 2"},
		{ID: "3", Status: "failed", Title: "Task 3"},
	}

	model.SetTasks(tasks)

	if len(model.tasks) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(model.tasks))
	}
	if model.cursor != 0 {
		t.Errorf("expected cursor to be 0, got %d", model.cursor)
	}
}

func TestSetTasks_ResetsCursor(t *testing.T) {
	model := New(
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		func(s string) string { return "" },
	)

	// Set initial tasks and move cursor
	initialTasks := []data.AgentTask{
		{ID: "1", Status: "running", Title: "Task 1"},
		{ID: "2", Status: "completed", Title: "Task 2"},
		{ID: "3", Status: "failed", Title: "Task 3"},
	}
	model.SetTasks(initialTasks)
	model.MoveCursor(2)

	// Now set new tasks with fewer items
	newTasks := []data.AgentTask{
		{ID: "4", Status: "running", Title: "Task 4"},
	}
	model.SetTasks(newTasks)

	if model.cursor >= len(newTasks) {
		t.Errorf("cursor should be adjusted to within bounds, got cursor=%d for %d tasks", model.cursor, len(newTasks))
	}
}

func TestMoveCursor(t *testing.T) {
	model := New(
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		func(s string) string { return "" },
	)

	tasks := []data.AgentTask{
		{ID: "1", Status: "running", Title: "Task 1"},
		{ID: "2", Status: "completed", Title: "Task 2"},
		{ID: "3", Status: "failed", Title: "Task 3"},
	}
	model.SetTasks(tasks)

	// Move down
	model.MoveCursor(1)
	if model.cursor != 1 {
		t.Errorf("expected cursor to be 1, got %d", model.cursor)
	}

	// Move down again
	model.MoveCursor(1)
	if model.cursor != 2 {
		t.Errorf("expected cursor to be 2, got %d", model.cursor)
	}

	// Try to move past the end
	model.MoveCursor(1)
	if model.cursor != 2 {
		t.Errorf("cursor should not move past last item, got %d", model.cursor)
	}

	// Move up
	model.MoveCursor(-1)
	if model.cursor != 1 {
		t.Errorf("expected cursor to be 1, got %d", model.cursor)
	}

	// Move to start
	model.MoveCursor(-10)
	if model.cursor != 0 {
		t.Errorf("cursor should not move before first item, got %d", model.cursor)
	}
}

func TestSelectedTask(t *testing.T) {
	model := New(
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		func(s string) string { return "" },
	)

	tasks := []data.AgentTask{
		{ID: "1", Status: "running", Title: "Task 1"},
		{ID: "2", Status: "completed", Title: "Task 2"},
	}
	model.SetTasks(tasks)

	selected := model.SelectedTask()
	if selected == nil {
		t.Fatal("expected selected task, got nil")
	}
	if selected.ID != "1" {
		t.Errorf("expected selected task ID '1', got '%s'", selected.ID)
	}

	model.MoveCursor(1)
	selected = model.SelectedTask()
	if selected == nil {
		t.Fatal("expected selected task, got nil")
	}
	if selected.ID != "2" {
		t.Errorf("expected selected task ID '2', got '%s'", selected.ID)
	}
}

func TestSelectedTask_EmptyList(t *testing.T) {
	model := New(
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		func(s string) string { return "" },
	)

	selected := model.SelectedTask()
	if selected != nil {
		t.Error("expected nil for empty task list")
	}
}

func TestView_EmptyTasks(t *testing.T) {
	model := New(
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		func(s string) string { return "" },
	)

	view := model.View()
	if !strings.Contains(view, "No agent tasks found") {
		t.Errorf("expected message about no tasks, got: %s", view)
	}
}

func TestView_Loading(t *testing.T) {
	model := New(
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		func(s string) string { return "" },
	)
	model.loading = true

	view := model.View()
	if !strings.Contains(view, "Loading") {
		t.Errorf("expected loading message, got: %s", view)
	}
}

func TestView_WithTasks(t *testing.T) {
	model := New(
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		func(status string) string { return "📝" },
	)

	tasks := []data.AgentTask{
		{
			ID:         "abc123",
			Status:     "running",
			Title:      "Fix bug in handler",
			Repository: "owner/repo",
			UpdatedAt:  time.Now().Add(-30 * time.Minute),
		},
	}
	model.SetTasks(tasks)

	view := model.View()
	if !strings.Contains(view, "Repository") {
		t.Error("expected view to contain header with 'Repository'")
	}
	if !strings.Contains(view, "owner/repo") {
		t.Error("expected view to contain repository name")
	}
	if !strings.Contains(view, "Fix bug") {
		t.Error("expected view to contain task title")
	}
}

func TestFilterByStatus(t *testing.T) {
	model := New(
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		func(s string) string { return "" },
	)

	allTasks := []data.AgentTask{
		{ID: "1", Status: "running", Title: "Task 1"},
		{ID: "2", Status: "completed", Title: "Task 2"},
		{ID: "3", Status: "failed", Title: "Task 3"},
		{ID: "4", Status: "running", Title: "Task 4"},
	}

	// Test filtering for running tasks
	var runningTasks []data.AgentTask
	for _, task := range allTasks {
		if task.Status == "running" {
			runningTasks = append(runningTasks, task)
		}
	}

	model.SetTasks(runningTasks)
	if len(model.tasks) != 2 {
		t.Errorf("expected 2 running tasks, got %d", len(model.tasks))
	}

	// Test filtering for completed tasks
	var completedTasks []data.AgentTask
	for _, task := range allTasks {
		if task.Status == "completed" {
			completedTasks = append(completedTasks, task)
		}
	}

	model.SetTasks(completedTasks)
	if len(model.tasks) != 1 {
		t.Errorf("expected 1 completed task, got %d", len(model.tasks))
	}
}
