package taskdetail

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/maxbeizer/gh-agent-viz/internal/data"
)

func TestView_UsesFriendlyFallbacks(t *testing.T) {
	model := New(
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		func(string) string { return "•" },
	)
	model.SetTask(&data.Session{
		ID: "session-1",
	})

	view := model.View()
	if !strings.Contains(view, "Untitled Session") {
		t.Fatalf("expected untitled fallback, got: %s", view)
	}
	if !strings.Contains(view, "Repository: not linked") {
		t.Fatalf("expected repository fallback, got: %s", view)
	}
	if !strings.Contains(view, "Branch:     not linked") {
		t.Fatalf("expected branch fallback, got: %s", view)
	}
	if !strings.Contains(view, "Created:    not recorded") || !strings.Contains(view, "Updated:    not recorded") {
		t.Fatalf("expected timestamp fallback, got: %s", view)
	}
}

func TestView_ShowsRecordedTimestamps(t *testing.T) {
	model := New(
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		func(string) string { return "•" },
	)
	now := time.Date(2025, time.January, 2, 3, 4, 5, 0, time.UTC)
	model.SetTask(&data.Session{
		ID:        "session-2",
		Title:     "Named Session",
		CreatedAt: now,
		UpdatedAt: now,
	})

	view := model.View()
	if !strings.Contains(view, "Created:    2025-01-02 03:04:05") {
		t.Fatalf("expected formatted created timestamp, got: %s", view)
	}
	if !strings.Contains(view, "Updated:    2025-01-02 03:04:05") {
		t.Fatalf("expected formatted updated timestamp, got: %s", view)
	}
}
