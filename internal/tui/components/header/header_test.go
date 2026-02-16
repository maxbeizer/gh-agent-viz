package header

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func newTestModel(title string, filter *string) Model {
	s := lipgloss.NewStyle()
	return New(s, s, s, s, title, filter)
}

func TestNew(t *testing.T) {
	title := "Agent Task Viewer"
	filter := "running"

	model := newTestModel(title, &filter)

	if model.title != title {
		t.Errorf("expected title '%s', got '%s'", title, model.title)
	}
	if model.filter == nil {
		t.Fatal("expected filter to be non-nil")
	}
	if *model.filter != filter {
		t.Errorf("expected filter '%s', got '%s'", filter, *model.filter)
	}
}

func TestNew_NilFilter(t *testing.T) {
	title := "Agent Task Viewer"

	model := newTestModel(title, nil)

	if model.title != title {
		t.Errorf("expected title '%s', got '%s'", title, model.title)
	}
	if model.filter != nil {
		t.Error("expected filter to be nil")
	}
}

func TestView_RendersTitle(t *testing.T) {
	title := "My Application"

	model := newTestModel(title, nil)
	view := model.View()

	if !strings.Contains(view, title) {
		t.Errorf("expected view to contain title '%s', got: %s", title, view)
	}
}

func TestView_WithFilter(t *testing.T) {
	title := "Tasks"
	filter := "completed"

	model := newTestModel(title, &filter)
	model.SetCounts(FilterCounts{All: 5, Completed: 2})
	view := model.View()

	if !strings.Contains(view, title) {
		t.Errorf("expected view to contain title '%s'", title)
	}
	// Active tab should be highlighted
	if !strings.Contains(view, "DONE") {
		t.Errorf("expected view to contain DONE tab for completed filter, got: %s", view)
	}
}

func TestView_WithTabCounts(t *testing.T) {
	title := "Tasks"
	filter := "all"

	model := newTestModel(title, &filter)
	model.SetCounts(FilterCounts{All: 10, Attention: 3, Active: 5, Completed: 4, Failed: 1})
	view := model.View()

	if !strings.Contains(view, "(10)") {
		t.Errorf("expected ALL count of 10, got: %s", view)
	}
	if !strings.Contains(view, "(3)") {
		t.Errorf("expected ACTION count of 3, got: %s", view)
	}
}

func TestView_EndsWithNewline(t *testing.T) {
	title := "Tasks"

	model := newTestModel(title, nil)
	view := model.View()

	if !strings.HasSuffix(view, "\n") {
		t.Error("expected view to end with newline")
	}
}
