package conversation

import (
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	m := New(80, 24)
	if m.width != 80 || m.height != 24 {
		t.Errorf("expected 80x24, got %dx%d", m.width, m.height)
	}
}

func TestSetMessages_Empty(t *testing.T) {
	m := New(80, 24)
	m.SetMessages(nil)
	view := m.View()
	if !strings.Contains(view, "No conversation events found for this session") {
		t.Errorf("expected empty state message, got %q", view)
	}
}

func TestSetMessages_UserMessage(t *testing.T) {
	m := New(80, 24)
	m.SetMessages([]ChatMessage{
		{Role: RoleUser, Content: "Fix the auth bug", Timestamp: "2026-01-15T09:15:00Z"},
	})
	view := m.View()
	if !strings.Contains(view, "You") {
		t.Error("expected 'You' header in user message")
	}
	if !strings.Contains(view, "Fix the auth bug") {
		t.Error("expected message content in output")
	}
}

func TestSetMessages_AgentMessage(t *testing.T) {
	m := New(80, 24)
	m.SetMessages([]ChatMessage{
		{Role: RoleAssistant, Content: "I'll fix it", Timestamp: "2026-01-15T09:15:00Z", Tools: []string{"bash", "edit"}},
	})
	view := m.View()
	if !strings.Contains(view, "Agent") {
		t.Error("expected 'Agent' header in assistant message")
	}
	if !strings.Contains(view, "I'll fix it") {
		t.Error("expected message content in output")
	}
	if !strings.Contains(view, "bash") || !strings.Contains(view, "edit") {
		t.Error("expected tool names in output")
	}
}

func TestSetMessages_MultipleTurns(t *testing.T) {
	m := New(100, 30)
	m.SetMessages([]ChatMessage{
		{Role: RoleUser, Content: "Hello", Timestamp: "2026-01-15T09:15:00Z"},
		{Role: RoleAssistant, Content: "Hi there!", Timestamp: "2026-01-15T09:15:05Z"},
		{Role: RoleUser, Content: "Now add tests", Timestamp: "2026-01-15T09:17:00Z"},
	})
	view := m.View()
	if strings.Count(view, "You") < 2 {
		t.Error("expected two user messages")
	}
	if !strings.Contains(view, "Agent") {
		t.Error("expected agent message")
	}
}

func TestTimeSeparator_LargeGap(t *testing.T) {
	m := New(80, 40)
	m.SetMessages([]ChatMessage{
		{Role: RoleUser, Content: "First", Timestamp: "2026-01-15T09:00:00Z"},
		{Role: RoleUser, Content: "Second", Timestamp: "2026-01-15T09:10:00Z"},
	})
	view := m.View()
	// 10 minute gap > 5 minute threshold, so separator should appear
	if !strings.Contains(view, "09:10") {
		t.Error("expected time separator for 10-minute gap")
	}
}

func TestTimeSeparator_SmallGap(t *testing.T) {
	m := New(80, 40)
	m.SetMessages([]ChatMessage{
		{Role: RoleUser, Content: "First", Timestamp: "2026-01-15T09:00:00Z"},
		{Role: RoleUser, Content: "Second", Timestamp: "2026-01-15T09:02:00Z"},
	})
	view := m.View()
	// 2 minute gap < 5 minute threshold — no separator with the time between messages
	// The timestamps still appear in the bubble headers
	lines := strings.Split(view, "\n")
	separatorFound := false
	for _, line := range lines {
		if strings.Contains(line, "──") && strings.Contains(line, "09:02") {
			separatorFound = true
		}
	}
	if separatorFound {
		t.Error("did not expect time separator for 2-minute gap")
	}
}

func TestSetSize(t *testing.T) {
	m := New(80, 24)
	m.SetMessages([]ChatMessage{
		{Role: RoleUser, Content: "hello", Timestamp: "2026-01-15T09:00:00Z"},
	})
	m.SetSize(120, 40)
	if m.width != 120 || m.height != 40 {
		t.Errorf("expected 120x40, got %dx%d", m.width, m.height)
	}
	// Should still render after resize
	view := m.View()
	if !strings.Contains(view, "hello") {
		t.Error("expected content after resize")
	}
}

func TestScrolling(t *testing.T) {
	m := New(80, 5) // Small viewport to force scrolling
	var msgs []ChatMessage
	for i := 0; i < 20; i++ {
		msgs = append(msgs, ChatMessage{
			Role:      RoleUser,
			Content:   "Line of chat content for scrolling test",
			Timestamp: "2026-01-15T09:00:00Z",
		})
	}
	m.SetMessages(msgs)

	// These should not panic
	m.LineDown()
	m.LineUp()
	m.HalfPageDown()
	m.HalfPageUp()
	m.GotoBottom()
	m.GotoTop()
}

func TestSystemMessage(t *testing.T) {
	m := New(80, 24)
	m.SetMessages([]ChatMessage{
		{Role: RoleSystem, Content: "Session started", Timestamp: "2026-01-15T09:00:00Z"},
	})
	view := m.View()
	if !strings.Contains(view, "Session started") {
		t.Error("expected system message content")
	}
}

func TestFormatToolLine(t *testing.T) {
	line := formatToolLine([]string{"bash", "edit", "grep"})
	if !strings.Contains(line, "bash") || !strings.Contains(line, "edit") || !strings.Contains(line, "grep") {
		t.Errorf("expected all tools in line, got %q", line)
	}
	if !strings.Contains(line, "•") {
		t.Error("expected bullet separator between tools")
	}
}

func TestFormatToolLine_Dedup(t *testing.T) {
	line := formatToolLine([]string{"bash", "bash", "edit"})
	if strings.Count(line, "bash") != 1 {
		t.Error("expected duplicate tools to be deduplicated")
	}
}

func TestWordWrap(t *testing.T) {
	result := wordWrap("hello world this is a test", 10)
	lines := strings.Split(result, "\n")
	for _, l := range lines {
		if len(l) > 12 { // some slack for word boundaries
			t.Errorf("line too long: %q", l)
		}
	}
}

func TestWordWrap_PreservesNewlines(t *testing.T) {
	result := wordWrap("line1\nline2", 80)
	if !strings.Contains(result, "line1") || !strings.Contains(result, "line2") {
		t.Error("expected both lines preserved")
	}
}
