package tooltimeline

import (
	"strings"
	"testing"
)

func TestToolIcon_Bash(t *testing.T) {
	if icon := ToolIcon("bash"); icon != "🔧" {
		t.Errorf("expected 🔧 for bash, got %s", icon)
	}
	if icon := ToolIcon("run_terminal_command"); icon != "🔧" {
		t.Errorf("expected 🔧 for run_terminal_command, got %s", icon)
	}
}

func TestToolIcon_Edit(t *testing.T) {
	if icon := ToolIcon("edit_file"); icon != "✏️" {
		t.Errorf("expected ✏️ for edit_file, got %s", icon)
	}
	if icon := ToolIcon("write"); icon != "✏️" {
		t.Errorf("expected ✏️ for write, got %s", icon)
	}
	if icon := ToolIcon("create_file"); icon != "✏️" {
		t.Errorf("expected ✏️ for create_file, got %s", icon)
	}
}

func TestToolIcon_Read(t *testing.T) {
	if icon := ToolIcon("read_file"); icon != "📄" {
		t.Errorf("expected 📄 for read_file, got %s", icon)
	}
	if icon := ToolIcon("view"); icon != "📄" {
		t.Errorf("expected 📄 for view, got %s", icon)
	}
}

func TestToolIcon_Search(t *testing.T) {
	if icon := ToolIcon("grep"); icon != "🔍" {
		t.Errorf("expected 🔍 for grep, got %s", icon)
	}
	if icon := ToolIcon("search_code"); icon != "🔍" {
		t.Errorf("expected 🔍 for search_code, got %s", icon)
	}
	if icon := ToolIcon("glob"); icon != "🔍" {
		t.Errorf("expected 🔍 for glob, got %s", icon)
	}
}

func TestToolIcon_Git(t *testing.T) {
	if icon := ToolIcon("git_commit"); icon != "📤" {
		t.Errorf("expected 📤 for git_commit, got %s", icon)
	}
}

func TestToolIcon_Test(t *testing.T) {
	if icon := ToolIcon("run_test"); icon != "🧪" {
		t.Errorf("expected 🧪 for run_test, got %s", icon)
	}
}

func TestToolIcon_Default(t *testing.T) {
	if icon := ToolIcon("unknown_tool"); icon != "⚙️" {
		t.Errorf("expected ⚙️ for unknown_tool, got %s", icon)
	}
}

func TestToolIcon_CaseInsensitive(t *testing.T) {
	if icon := ToolIcon("BASH"); icon != "🔧" {
		t.Errorf("expected 🔧 for BASH, got %s", icon)
	}
}

func TestNew(t *testing.T) {
	m := New(80, 24)
	if m.width != 80 || m.height != 24 {
		t.Errorf("expected 80x24, got %dx%d", m.width, m.height)
	}
}

func TestSetEvents_Empty(t *testing.T) {
	m := New(80, 24)
	m.SetEvents(nil)
	view := m.View()
	if !strings.Contains(view, "No tool executions") {
		t.Errorf("expected empty state message, got: %s", view)
	}
}

func TestSetEvents_WithEvents(t *testing.T) {
	m := New(80, 24)
	events := []ToolEvent{
		{Timestamp: "2025-01-15T09:15:00Z", ToolName: "grep", Icon: "🔍"},
		{Timestamp: "2025-01-15T09:15:30Z", ToolName: "view", Icon: "📄"},
		{Timestamp: "2025-01-15T09:16:00Z", ToolName: "edit", Icon: "✏️"},
	}
	m.SetEvents(events)
	view := m.View()
	if !strings.Contains(view, "Tool Timeline") {
		t.Errorf("expected 'Tool Timeline' in view, got: %s", view)
	}
	if !strings.Contains(view, "3 executions") {
		t.Errorf("expected '3 executions' in view, got: %s", view)
	}
	if !strings.Contains(view, "grep") {
		t.Errorf("expected 'grep' in view, got: %s", view)
	}
}

func TestSetSize(t *testing.T) {
	m := New(80, 24)
	m.SetSize(120, 40)
	if m.width != 120 || m.height != 40 {
		t.Errorf("expected 120x40, got %dx%d", m.width, m.height)
	}
}

func TestFormatTimelineTimestamp(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"2025-01-15T09:15:00Z", "09:15"},
		{"2025-01-15T14:30:45.123Z", "14:30"},
		{"invalid", "invalid"},
	}
	for _, tt := range tests {
		result := formatTimelineTimestamp(tt.input)
		if result != tt.expected {
			t.Errorf("formatTimelineTimestamp(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestRapidSequenceGrouping(t *testing.T) {
	m := New(80, 24)
	events := []ToolEvent{
		{Timestamp: "2025-01-15T09:15:00Z", ToolName: "grep", Icon: "🔍"},
		{Timestamp: "2025-01-15T09:15:30Z", ToolName: "view", Icon: "📄"},
	}
	m.SetEvents(events)
	view := m.View()
	// Both share the same 09:15 minute — second should show dot grouping
	if !strings.Contains(view, "·") {
		t.Errorf("expected rapid-sequence dot grouping, got: %s", view)
	}
}
