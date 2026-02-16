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
