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
	if len(model.sessions) != 0 {
		t.Errorf("expected empty sessions list, got %d sessions", len(model.sessions))
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

	sessions := []data.Session{
		{ID: "1", Status: "running", Title: "Task 1", Source: data.SourceAgentTask},
		{ID: "2", Status: "completed", Title: "Task 2", Source: data.SourceAgentTask},
		{ID: "3", Status: "failed", Title: "Task 3", Source: data.SourceLocalCopilot},
	}

	model.SetTasks(sessions)

	if len(model.sessions) != 3 {
		t.Errorf("expected 3 sessions, got %d", len(model.sessions))
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

	// Set initial sessions and move cursor
	initialSessions := []data.Session{
		{ID: "1", Status: "running", Title: "Task 1", Source: data.SourceAgentTask},
		{ID: "2", Status: "completed", Title: "Task 2", Source: data.SourceAgentTask},
		{ID: "3", Status: "failed", Title: "Task 3", Source: data.SourceLocalCopilot},
	}
	model.SetTasks(initialSessions)
	model.MoveCursor(2)

	// Now set new sessions with fewer items
	newSessions := []data.Session{
		{ID: "4", Status: "running", Title: "Task 4", Source: data.SourceAgentTask},
	}
	model.SetTasks(newSessions)

	if model.cursor >= len(newSessions) {
		t.Errorf("cursor should be adjusted to within bounds, got cursor=%d for %d sessions", model.cursor, len(newSessions))
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

	sessions := []data.Session{
		{ID: "1", Status: "running", Title: "Task 1", Source: data.SourceAgentTask},
		{ID: "2", Status: "completed", Title: "Task 2", Source: data.SourceAgentTask},
		{ID: "3", Status: "failed", Title: "Task 3", Source: data.SourceLocalCopilot},
	}
	model.SetTasks(sessions)

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

	sessions := []data.Session{
		{ID: "1", Status: "running", Title: "Task 1", Source: data.SourceAgentTask},
		{ID: "2", Status: "completed", Title: "Task 2", Source: data.SourceLocalCopilot},
	}
	model.SetTasks(sessions)

	selected := model.SelectedTask()
	if selected == nil {
		t.Fatal("expected selected session, got nil")
	}
	if selected.ID != "1" {
		t.Errorf("expected selected session ID '1', got '%s'", selected.ID)
	}

	model.MoveCursor(1)
	selected = model.SelectedTask()
	if selected == nil {
		t.Fatal("expected selected session, got nil")
	}
	if selected.ID != "2" {
		t.Errorf("expected selected session ID '2', got '%s'", selected.ID)
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
		t.Error("expected nil for empty session list")
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
	if !strings.Contains(view, "No sessions found") {
		t.Errorf("expected message about no sessions, got: %s", view)
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

	sessions := []data.Session{
		{
			ID:         "abc123",
			Status:     "running",
			Title:      "Fix bug in handler",
			Repository: "owner/repo",
			UpdatedAt:  time.Now().Add(-30 * time.Minute),
			Source:     data.SourceAgentTask,
		},
	}
	model.SetTasks(sessions)

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

	allSessions := []data.Session{
		{ID: "1", Status: "running", Title: "Task 1", Source: data.SourceAgentTask},
		{ID: "2", Status: "completed", Title: "Task 2", Source: data.SourceAgentTask},
		{ID: "3", Status: "failed", Title: "Task 3", Source: data.SourceLocalCopilot},
		{ID: "4", Status: "running", Title: "Task 4", Source: data.SourceLocalCopilot},
	}

	// Test filtering for running sessions
	var runningSessions []data.Session
	for _, session := range allSessions {
		if session.Status == "running" {
			runningSessions = append(runningSessions, session)
		}
	}

	model.SetTasks(runningSessions)
	if len(model.sessions) != 2 {
		t.Errorf("expected 2 running sessions, got %d", len(model.sessions))
	}

	// Test filtering for completed sessions
	var completedSessions []data.Session
	for _, session := range allSessions {
		if session.Status == "completed" {
			completedSessions = append(completedSessions, session)
		}
	}

	model.SetTasks(completedSessions)
	if len(model.sessions) != 1 {
		t.Errorf("expected 1 completed session, got %d", len(model.sessions))
	}
}
