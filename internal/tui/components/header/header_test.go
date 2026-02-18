package header

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func newTestModel(title string, filter *string) Model {
	s := lipgloss.NewStyle()
	return New(s, s, s, s, title, filter, false)
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

func TestView_RendersTabs(t *testing.T) {
	model := newTestModel("My Application", nil)
	view := model.View()

	if !strings.Contains(view, "RUNNING") {
		t.Errorf("expected view to contain RUNNING tab, got: %s", view)
	}
}

func TestView_WithFilter(t *testing.T) {
	title := "Tasks"
	filter := "completed"

	model := newTestModel(title, &filter)
	model.SetCounts(FilterCounts{All: 5, Completed: 2})
	view := model.View()

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
		t.Errorf("expected ATTENTION count of 3, got: %s", view)
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

func newTestModelWithBanner(title string, filter *string, useAscii bool) Model {
	s := lipgloss.NewStyle()
	m := New(s, s, s, s, title, filter, useAscii)
	m.SetSize(80, 24)
	return m
}

func TestView_BannerRendersWhenEnabled(t *testing.T) {
	model := newTestModelWithBanner("Tasks", nil, true)
	view := model.View()

	if !strings.Contains(view, "A G E N T   V I Z") {
		t.Errorf("expected view to contain ASCII banner when enabled, got: %s", view)
	}
}

func TestView_PlainTitleWhenBannerDisabled(t *testing.T) {
	model := newTestModelWithBanner("⚡ Agent Sessions", nil, false)
	view := model.View()

	if strings.Contains(view, "A G E N T   V I Z") {
		t.Error("expected view NOT to contain ASCII banner when disabled")
	}
	if !strings.Contains(view, "RUNNING") {
		t.Errorf("expected view to contain tabs when banner is disabled, got: %s", view)
	}
}

func TestView_BannerHiddenOnShortTerminal(t *testing.T) {
	model := newTestModelWithBanner("Tasks", nil, true)
	model.SetSize(80, 14) // below minHeightForBanner
	view := model.View()

	if strings.Contains(view, "A G E N T   V I Z") {
		t.Error("expected banner to be hidden when terminal height < 15")
	}
}

func TestView_BannerHiddenOnNarrowTerminal(t *testing.T) {
	model := newTestModelWithBanner("Tasks", nil, true)
	model.SetSize(20, 24) // too narrow for banner
	view := model.View()

	if strings.Contains(view, "A G E N T   V I Z") {
		t.Error("expected banner to be hidden when terminal is too narrow")
	}
}

func TestView_BannerShownAtExactMinHeight(t *testing.T) {
	model := newTestModelWithBanner("Tasks", nil, true)
	model.SetSize(80, 15) // exactly minHeightForBanner
	view := model.View()

	if !strings.Contains(view, "A G E N T   V I Z") {
		t.Error("expected banner to be shown at exactly minimum height")
	}
}
