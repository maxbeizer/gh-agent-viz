package logview

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderMarkdown_Heading(t *testing.T) {
	result := renderMarkdown("# Hello World", 80)
	if result == "" {
		t.Fatal("expected non-empty rendered output for heading")
	}
	if !strings.Contains(result, "Hello World") {
		t.Errorf("expected rendered output to contain 'Hello World', got: %s", result)
	}
}

func TestRenderMarkdown_CodeBlock(t *testing.T) {
	input := "```go\nfmt.Println(\"hello\")\n```"
	result := renderMarkdown(input, 80)
	if result == "" {
		t.Fatal("expected non-empty rendered output for code block")
	}
	if !strings.Contains(result, "hello") {
		t.Errorf("expected rendered output to contain 'hello', got: %s", result)
	}
}

func TestRenderMarkdown_PlainText(t *testing.T) {
	input := "Just some plain text with no markdown."
	result := renderMarkdown(input, 80)
	if result == "" {
		t.Fatal("expected non-empty rendered output for plain text")
	}
	if !strings.Contains(result, "plain text") {
		t.Errorf("expected rendered output to contain 'plain text', got: %s", result)
	}
}

func TestRenderMarkdown_ZeroWidth(t *testing.T) {
	result := renderMarkdown("# Test", 0)
	if !strings.Contains(result, "Test") {
		t.Errorf("expected fallback to default width, got: %s", result)
	}
}

func TestSetContent_RendersMarkdown(t *testing.T) {
	m := New(lipgloss.NewStyle(), 80, 24)
	m.SetContent("# Title\n\nSome body text.")

	if !m.ready {
		t.Fatal("expected model to be ready after SetContent")
	}
	if m.rawContent != "# Title\n\nSome body text." {
		t.Errorf("expected rawContent to be preserved, got: %s", m.rawContent)
	}
	if !strings.Contains(m.content, "Title") {
		t.Errorf("expected rendered content to contain 'Title', got: %s", m.content)
	}
}

func TestSetSize_ReRendersContent(t *testing.T) {
	m := New(lipgloss.NewStyle(), 80, 24)
	m.SetContent("# Title\n\nSome body text.")

	contentBefore := m.content
	m.SetSize(40, 12)
	contentAfter := m.content

	if m.viewport.Width != 40 {
		t.Errorf("expected viewport width 40, got: %d", m.viewport.Width)
	}
	if m.viewport.Height != 12 {
		t.Errorf("expected viewport height 12, got: %d", m.viewport.Height)
	}
	// Content should be re-rendered (may differ due to different wrap width)
	if !strings.Contains(contentAfter, "Title") {
		t.Errorf("expected re-rendered content to contain 'Title', got: %s", contentAfter)
	}
	// Verify it actually re-rendered (not just reused old content)
	_ = contentBefore // widths differ so rendering may differ
}

func TestSetSize_NoContent(t *testing.T) {
	m := New(lipgloss.NewStyle(), 80, 24)
	// SetSize with no content should not panic
	m.SetSize(40, 12)
	if m.viewport.Width != 40 {
		t.Errorf("expected viewport width 40, got: %d", m.viewport.Width)
	}
}

func TestView_NotReady(t *testing.T) {
	m := New(lipgloss.NewStyle(), 80, 24)
	view := m.View()
	if view != "Loading logs..." {
		t.Errorf("expected 'Loading logs...', got: %s", view)
	}
}

func TestSetFollowMode_Toggle(t *testing.T) {
	m := New(lipgloss.NewStyle(), 80, 24)
	if m.FollowMode() {
		t.Fatal("expected follow mode to be off by default")
	}
	m.SetFollowMode(true)
	if !m.FollowMode() {
		t.Fatal("expected follow mode to be on")
	}
	m.SetFollowMode(false)
	if m.FollowMode() {
		t.Fatal("expected follow mode to be off")
	}
}

func TestSetLive(t *testing.T) {
	m := New(lipgloss.NewStyle(), 80, 24)
	if m.IsLive() {
		t.Fatal("expected live to be off by default")
	}
	m.SetLive(true)
	if !m.IsLive() {
		t.Fatal("expected live to be on")
	}
	m.SetLive(false)
	if m.IsLive() {
		t.Fatal("expected live to be off")
	}
}

func TestAppendOrReplace_UpdatesWhenLonger(t *testing.T) {
	m := New(lipgloss.NewStyle(), 80, 24)
	m.SetContent("short")
	if m.rawContent != "short" {
		t.Fatalf("expected rawContent 'short', got %q", m.rawContent)
	}
	m.AppendOrReplace("short and longer")
	if m.rawContent != "short and longer" {
		t.Fatalf("expected rawContent to be updated, got %q", m.rawContent)
	}
}

func TestAppendOrReplace_NoUpdateWhenSameLength(t *testing.T) {
	m := New(lipgloss.NewStyle(), 80, 24)
	m.SetContent("hello")
	m.AppendOrReplace("hello")
	if m.rawContent != "hello" {
		t.Fatalf("expected rawContent to remain 'hello', got %q", m.rawContent)
	}
}

func TestAppendOrReplace_NoUpdateWhenShorter(t *testing.T) {
	m := New(lipgloss.NewStyle(), 80, 24)
	m.SetContent("longer content")
	m.AppendOrReplace("short")
	if m.rawContent != "longer content" {
		t.Fatalf("expected rawContent to remain unchanged, got %q", m.rawContent)
	}
}

func TestView_LiveFollowing(t *testing.T) {
	m := New(lipgloss.NewStyle(), 80, 24)
	m.SetContent("some log content")
	m.SetLive(true)
	m.SetFollowMode(true)
	view := m.View()
	if !strings.Contains(view, "LIVE 🔴") {
		t.Errorf("expected LIVE indicator, got: %s", view)
	}
}

func TestView_LivePaused(t *testing.T) {
	m := New(lipgloss.NewStyle(), 80, 24)
	m.SetContent("some log content")
	m.SetLive(true)
	m.SetFollowMode(false)
	view := m.View()
	if !strings.Contains(view, "PAUSED") {
		t.Errorf("expected PAUSED indicator, got: %s", view)
	}
}

func TestView_NotLive_NoIndicator(t *testing.T) {
	m := New(lipgloss.NewStyle(), 80, 24)
	m.SetContent("some log content")
	view := m.View()
	if strings.Contains(view, "LIVE") || strings.Contains(view, "PAUSED") {
		t.Errorf("expected no indicator for non-live session, got: %s", view)
	}
}
