package tasklist

import (
	"testing"
	"time"

	"github.com/maxbeizer/gh-agent-viz/internal/data"
)

func TestDismissCompletedRemovesCompletedSessions(t *testing.T) {
	model := newModel()
	now := time.Now()
	model.SetTasks([]data.Session{
		{ID: "1", Status: "running", Title: "Running", UpdatedAt: now},
		{ID: "2", Status: "completed", Title: "Done 1", UpdatedAt: now.Add(-1 * time.Hour)},
		{ID: "3", Status: "completed", Title: "Done 2", UpdatedAt: now.Add(-2 * time.Hour)},
		{ID: "4", Status: "failed", Title: "Failed", UpdatedAt: now.Add(-3 * time.Hour)},
	})

	count := model.DismissCompleted()
	if count != 2 {
		t.Fatalf("expected 2 dismissed, got %d", count)
	}
	if len(model.sessions) != 2 {
		t.Fatalf("expected 2 remaining sessions, got %d", len(model.sessions))
	}
	for _, s := range model.sessions {
		if s.Status == "completed" {
			t.Fatalf("completed session %s should have been dismissed", s.ID)
		}
	}
}

func TestDismissCompletedReturnsZeroWhenNoneCompleted(t *testing.T) {
	model := newModel()
	now := time.Now()
	model.SetTasks([]data.Session{
		{ID: "1", Status: "running", Title: "Running", UpdatedAt: now},
		{ID: "2", Status: "failed", Title: "Failed", UpdatedAt: now.Add(-1 * time.Hour)},
	})

	count := model.DismissCompleted()
	if count != 0 {
		t.Fatalf("expected 0 dismissed, got %d", count)
	}
	if len(model.sessions) != 2 {
		t.Fatalf("expected 2 sessions unchanged, got %d", len(model.sessions))
	}
}

func TestDismissCompletedClampsCursor(t *testing.T) {
	model := newModel()
	now := time.Now()
	model.SetTasks([]data.Session{
		{ID: "1", Status: "completed", Title: "Done 1", UpdatedAt: now},
		{ID: "2", Status: "completed", Title: "Done 2", UpdatedAt: now.Add(-1 * time.Hour)},
		{ID: "3", Status: "running", Title: "Running", UpdatedAt: now.Add(-2 * time.Hour)},
	})
	// Move cursor to the end
	model.rowCursor = 2

	count := model.DismissCompleted()
	if count != 2 {
		t.Fatalf("expected 2 dismissed, got %d", count)
	}
	if model.rowCursor != 0 {
		t.Fatalf("expected cursor clamped to 0, got %d", model.rowCursor)
	}
}

func TestDismissCompletedHandlesCaseInsensitiveStatus(t *testing.T) {
	model := newModel()
	now := time.Now()
	model.SetTasks([]data.Session{
		{ID: "1", Status: "Completed", Title: "Done 1", UpdatedAt: now},
		{ID: "2", Status: "COMPLETED", Title: "Done 2", UpdatedAt: now.Add(-1 * time.Hour)},
		{ID: "3", Status: "running", Title: "Running", UpdatedAt: now.Add(-2 * time.Hour)},
	})

	count := model.DismissCompleted()
	if count != 2 {
		t.Fatalf("expected 2 dismissed, got %d", count)
	}
	if len(model.sessions) != 1 {
		t.Fatalf("expected 1 remaining session, got %d", len(model.sessions))
	}
}
