package footer

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

func TestNew(t *testing.T) {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	keys := []key.Binding{
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
	}

	model := New(style, keys)

	if len(model.hints) != 2 {
		t.Errorf("expected 2 hints, got %d", len(model.hints))
	}
}

func TestNew_EmptyKeys(t *testing.T) {
	style := lipgloss.NewStyle()
	keys := []key.Binding{}

	model := New(style, keys)

	if len(model.hints) != 0 {
		t.Errorf("expected 0 hints, got %d", len(model.hints))
	}
}

func TestView_ContainsKeybindingHints(t *testing.T) {
	style := lipgloss.NewStyle()
	keys := []key.Binding{
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	}

	model := New(style, keys)
	view := model.View()

	if !strings.Contains(view, "quit") {
		t.Error("expected view to contain 'quit' hint")
	}
	if !strings.Contains(view, "refresh") {
		t.Error("expected view to contain 'refresh' hint")
	}
	if !strings.Contains(view, "select") {
		t.Error("expected view to contain 'select' hint")
	}
}

func TestView_HintsSeparatedByBullets(t *testing.T) {
	style := lipgloss.NewStyle()
	keys := []key.Binding{
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
	}

	model := New(style, keys)
	view := model.View()

	// The hints should be separated by bullet points
	if !strings.Contains(view, "•") {
		t.Error("expected hints to be separated by bullet points")
	}
}

func TestView_StartsWithNewline(t *testing.T) {
	style := lipgloss.NewStyle()
	keys := []key.Binding{
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	}

	model := New(style, keys)
	view := model.View()

	if !strings.HasPrefix(view, "\n") {
		t.Error("expected view to start with newline")
	}
}

func TestSetHints(t *testing.T) {
	style := lipgloss.NewStyle()
	initialKeys := []key.Binding{
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	}

	model := New(style, initialKeys)
	if len(model.hints) != 1 {
		t.Fatalf("expected 1 initial hint, got %d", len(model.hints))
	}

	// Update hints
	newKeys := []key.Binding{
		key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "action1")),
		key.NewBinding(key.WithKeys("b"), key.WithHelp("b", "action2")),
		key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "action3")),
	}
	model.SetHints(newKeys)

	if len(model.hints) != 3 {
		t.Errorf("expected 3 hints after update, got %d", len(model.hints))
	}

	view := model.View()
	if !strings.Contains(view, "action1") {
		t.Error("expected view to contain new hint 'action1'")
	}
	if !strings.Contains(view, "action2") {
		t.Error("expected view to contain new hint 'action2'")
	}
	if !strings.Contains(view, "action3") {
		t.Error("expected view to contain new hint 'action3'")
	}
}

func TestView_EmptyHints(t *testing.T) {
	style := lipgloss.NewStyle()
	keys := []key.Binding{}

	model := New(style, keys)
	view := model.View()

	// View should still be generated even with no hints
	if !strings.HasPrefix(view, "\n") {
		t.Error("expected view to start with newline even with empty hints")
	}
}

func TestView_TruncatesWhenExceedingWidth(t *testing.T) {
	style := lipgloss.NewStyle()
	keys := []key.Binding{
		key.NewBinding(key.WithKeys("↑/↓"), key.WithHelp("↑/↓", "move")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "logs")),
		key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "dismiss")),
		key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "group")),
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	}

	model := New(style, keys)
	model.SetWidth(30)
	view := model.View()

	if !strings.Contains(view, "…") {
		t.Error("expected truncated view to contain ellipsis")
	}
	if !strings.Contains(view, "move") {
		t.Error("expected first hint to be preserved")
	}
}

func TestView_NoTruncationWhenFits(t *testing.T) {
	style := lipgloss.NewStyle()
	keys := []key.Binding{
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	}

	model := New(style, keys)
	model.SetWidth(80)
	view := model.View()

	if strings.Contains(view, "…") {
		t.Error("expected no ellipsis when hints fit")
	}
	if !strings.Contains(view, "quit") {
		t.Error("expected hint to be present")
	}
}

func TestView_NoTruncationWhenWidthZero(t *testing.T) {
	style := lipgloss.NewStyle()
	keys := []key.Binding{
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
	}

	model := New(style, keys)
	// width defaults to 0; should not truncate
	view := model.View()

	if strings.Contains(view, "…") {
		t.Error("expected no truncation when width is zero")
	}
}
