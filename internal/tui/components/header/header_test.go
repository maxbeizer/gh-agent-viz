package header

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestNew(t *testing.T) {
	titleStyle := lipgloss.NewStyle().Bold(true)
	title := "Agent Task Viewer"
	filter := "running"

	model := New(titleStyle, title, &filter)

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
	titleStyle := lipgloss.NewStyle().Bold(true)
	title := "Agent Task Viewer"

	model := New(titleStyle, title, nil)

	if model.title != title {
		t.Errorf("expected title '%s', got '%s'", title, model.title)
	}
	if model.filter != nil {
		t.Error("expected filter to be nil")
	}
}

func TestView_RendersTitle(t *testing.T) {
	titleStyle := lipgloss.NewStyle()
	title := "My Application"
	
	model := New(titleStyle, title, nil)
	view := model.View()

	if !strings.Contains(view, title) {
		t.Errorf("expected view to contain title '%s', got: %s", title, view)
	}
}

func TestView_WithFilter(t *testing.T) {
	titleStyle := lipgloss.NewStyle()
	title := "Tasks"
	filter := "completed"

	model := New(titleStyle, title, &filter)
	view := model.View()

	if !strings.Contains(view, title) {
		t.Errorf("expected view to contain title '%s'", title)
	}
	if !strings.Contains(view, "Filter:") {
		t.Error("expected view to contain 'Filter:' label")
	}
	if !strings.Contains(view, filter) {
		t.Errorf("expected view to contain filter '%s'", filter)
	}
}

func TestView_WithEmptyFilter(t *testing.T) {
	titleStyle := lipgloss.NewStyle()
	title := "Tasks"
	filter := ""

	model := New(titleStyle, title, &filter)
	view := model.View()

	if !strings.Contains(view, title) {
		t.Errorf("expected view to contain title '%s'", title)
	}
	// Empty filter should not show the filter label
	if strings.Contains(view, "Filter:") {
		t.Error("did not expect view to contain 'Filter:' label when filter is empty")
	}
}

func TestView_WithNilFilter(t *testing.T) {
	titleStyle := lipgloss.NewStyle()
	title := "Tasks"

	model := New(titleStyle, title, nil)
	view := model.View()

	if !strings.Contains(view, title) {
		t.Errorf("expected view to contain title '%s'", title)
	}
	if strings.Contains(view, "Filter:") {
		t.Error("did not expect view to contain 'Filter:' label when filter is nil")
	}
}

func TestView_EndsWithNewline(t *testing.T) {
	titleStyle := lipgloss.NewStyle()
	title := "Tasks"

	model := New(titleStyle, title, nil)
	view := model.View()

	if !strings.HasSuffix(view, "\n") {
		t.Error("expected view to end with newline")
	}
}
