package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/maxbeizer/gh-agent-viz/internal/data"
)

func TestResumeSession_ValidRunningSession(t *testing.T) {
	m := NewModel("", false)

	// Create a running local session
	task := &data.Session{
		ID:     "test-session-123",
		Status: "running",
		Title:  "Test Running Task",
		Source: data.SourceLocalCopilot,
	}

	// Test that running sessions are resumable
	cmd := m.resumeSession(task)
	if cmd == nil {
		t.Error("expected cmd to be non-nil for running session")
	}
}

func TestResumeSession_ValidQueuedSession(t *testing.T) {
	m := NewModel("", false)

	// Create a queued local session
	task := &data.Session{
		ID:     "test-session-456",
		Status: "queued",
		Title:  "Test Queued Task",
		Source: data.SourceLocalCopilot,
	}

	// Test that queued sessions are resumable
	cmd := m.resumeSession(task)
	if cmd == nil {
		t.Error("expected cmd to be non-nil for queued session")
	}
}

func TestResumeSession_ValidNeedsInputSession(t *testing.T) {
	m := NewModel("", false)

	task := &data.Session{
		ID:     "test-session-needs-input",
		Status: "needs-input",
		Title:  "Needs Input",
		Source: data.SourceLocalCopilot,
	}

	cmd := m.resumeSession(task)
	if cmd == nil {
		t.Error("expected cmd to be non-nil for needs-input session")
	}
}

func TestResumeSession_CompletedSession(t *testing.T) {
	m := NewModel("", false)

	// Create a completed local session
	task := &data.Session{
		ID:     "test-session-789",
		Status: "completed",
		Title:  "Test Completed Task",
		Source: data.SourceLocalCopilot,
	}

	// Test that completed sessions return an error
	cmd := m.resumeSession(task)
	if cmd == nil {
		t.Fatal("expected cmd to be non-nil")
	}

	// Execute the command to get the message
	msg := cmd()
	errResult, ok := msg.(errMsg)
	if !ok {
		t.Fatal("expected errMsg for completed session")
	}

	if errResult.err == nil {
		t.Error("expected error for completed session")
	}
	if !strings.Contains(errResult.err.Error(), "resumable") {
		t.Errorf("expected error to mention resumable states, got: %s", errResult.err)
	}
}

func TestResumeSession_FailedSession(t *testing.T) {
	m := NewModel("", false)

	// Create a failed local session
	task := &data.Session{
		ID:     "test-session-999",
		Status: "failed",
		Title:  "Test Failed Task",
		Source: data.SourceLocalCopilot,
	}

	// Test that failed sessions return an error
	cmd := m.resumeSession(task)
	if cmd == nil {
		t.Fatal("expected cmd to be non-nil")
	}

	// Execute the command to get the message
	msg := cmd()
	errResult, ok := msg.(errMsg)
	if !ok {
		t.Fatal("expected errMsg for failed session")
	}

	if errResult.err == nil {
		t.Error("expected error for failed session")
	}
}

func TestResumeSession_NilTask(t *testing.T) {
	m := NewModel("", false)

	// Test nil task
	cmd := m.resumeSession(nil)
	if cmd == nil {
		t.Fatal("expected cmd to be non-nil")
	}

	// Execute the command to get the message
	msg := cmd()
	errResult, ok := msg.(errMsg)
	if !ok {
		t.Fatal("expected errMsg for nil task")
	}

	if errResult.err == nil {
		t.Error("expected error for nil task")
	}
}

func TestResumeSession_EmptySessionID(t *testing.T) {
	m := NewModel("", false)

	// Create a local session with empty ID
	task := &data.Session{
		ID:     "",
		Status: "running",
		Title:  "Test Task No ID",
		Source: data.SourceLocalCopilot,
	}

	// Test that empty ID returns an error
	cmd := m.resumeSession(task)
	if cmd == nil {
		t.Fatal("expected cmd to be non-nil")
	}

	// Execute the command to get the message
	msg := cmd()
	errResult, ok := msg.(errMsg)
	if !ok {
		t.Fatal("expected errMsg for empty session ID")
	}

	if errResult.err == nil {
		t.Error("expected error for empty session ID")
	}
}

func TestResumeSession_NormalizesStatusCase(t *testing.T) {
	m := NewModel("", false)

	// Test that mixed-case status is normalized
	tests := []string{"Running", "RUNNING", "  running  ", "  QUEUED  ", "Needs-Input"}
	for _, status := range tests {
		task := &data.Session{
			ID:     "test-session",
			Status: status,
			Source: data.SourceLocalCopilot,
		}
		cmd := m.resumeSession(task)
		if cmd == nil {
			t.Errorf("expected cmd to be non-nil for status %q", status)
		}
		// The command will fail at exec.Command level, but it should NOT
		// return a status validation error
		msg := cmd()
		if errResult, ok := msg.(errMsg); ok {
			if strings.Contains(errResult.err.Error(), "resumable") {
				t.Errorf("status %q should be accepted after normalization, got: %s", status, errResult.err)
			}
		}
	}
}

func TestResumeSession_DetailViewResume(t *testing.T) {
	m := NewModel("", false)
	m.taskList.SetTasks([]data.Session{
		{
			ID:     "local-1",
			Status: "running",
			Title:  "Local Session",
			Source: data.SourceLocalCopilot,
		},
	})
	m.viewMode = ViewModeDetail

	// Press 's' in detail view
	_, cmd := m.handleDetailKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if cmd == nil {
		t.Error("expected resume command to work in detail view")
	}
}

func TestResumeSession_LogViewResume(t *testing.T) {
	m := NewModel("", false)
	m.taskList.SetTasks([]data.Session{
		{
			ID:     "local-1",
			Status: "queued",
			Title:  "Local Session",
			Source: data.SourceLocalCopilot,
		},
	})
	m.viewMode = ViewModeLog

	// Press 's' in log view
	_, cmd := m.handleLogKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if cmd == nil {
		t.Error("expected resume command to work in log view")
	}
}

func TestUpdateFooterHints_DetailViewShowsResumeForResumableSession(t *testing.T) {
	m := NewModel("", false)
	m.viewMode = ViewModeDetail
	m.taskList.SetTasks([]data.Session{
		{
			ID:     "local-1",
			Status: "needs-input",
			Title:  "Local",
			Source: data.SourceLocalCopilot,
		},
	})

	m.updateFooterHints()
	footerView := m.footer.View()
	if !strings.Contains(footerView, "resume session") {
		t.Fatalf("expected resume hint in detail view for resumable session, got: %s", footerView)
	}
}

func TestCycleFilterForwardAndBackward(t *testing.T) {
	m := NewModel("", false)
	m.ctx.StatusFilter = "all"

	m.cycleFilter(1)
	if m.ctx.StatusFilter != "attention" {
		t.Fatalf("expected attention after forward cycle, got %q", m.ctx.StatusFilter)
	}

	m.cycleFilter(-1)
	if m.ctx.StatusFilter != "all" {
		t.Fatalf("expected all after backward cycle, got %q", m.ctx.StatusFilter)
	}

	m.cycleFilter(-1)
	if m.ctx.StatusFilter != "failed" {
		t.Fatalf("expected failed when cycling backward from all, got %q", m.ctx.StatusFilter)
	}
}

func TestHandleListKeys_AttentionToggle(t *testing.T) {
	m := NewModel("", false)
	m.ctx.StatusFilter = "all"

	updated, cmd := m.handleListKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if cmd == nil {
		t.Fatal("expected fetch command when toggling attention mode")
	}
	updatedModel := updated.(Model)
	if updatedModel.ctx.StatusFilter != "attention" {
		t.Fatalf("expected attention filter, got %q", updatedModel.ctx.StatusFilter)
	}
}

func TestHandleListKeys_LocalSessionLogShowsHelpfulError(t *testing.T) {
	m := NewModel("", false)
	m.taskList.SetTasks([]data.Session{
		{
			ID:         "local-1",
			Status:     "running",
			Title:      "Local Session",
			Repository: "owner/repo",
			Source:     data.SourceLocalCopilot,
		},
	})

	updated, cmd := m.handleListKeys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if cmd != nil {
		t.Fatal("expected no log fetch command for local session")
	}

	updatedModel := updated.(Model)
	if updatedModel.viewMode != ViewModeList {
		t.Fatalf("expected list view to remain active, got %v", updatedModel.viewMode)
	}
	if updatedModel.ctx.Error == nil {
		t.Fatal("expected helpful error for local session logs")
	}
}

func TestView_NotReadyShowsWhimsicalStartupText(t *testing.T) {
	m := NewModel("", false)
	view := m.View()
	if view != "Loading board..." {
		t.Fatalf("expected startup text, got %q", view)
	}
}

func TestUpdateFooterHints_LocalSessionShowsOnlyAvailableActions(t *testing.T) {
	m := NewModel("", false)
	m.viewMode = ViewModeList
	m.taskList.SetTasks([]data.Session{
		{
			ID:     "local-1",
			Status: "running",
			Title:  "Local",
			Source: data.SourceLocalCopilot,
		},
	})

	m.updateFooterHints()
	footerView := m.footer.View()
	if !strings.Contains(footerView, "resume session") {
		t.Fatalf("expected resume hint for resumable local session, got: %s", footerView)
	}
	if strings.Contains(footerView, "show task logs") {
		t.Fatalf("expected logs hint to be hidden for local session, got: %s", footerView)
	}
	if strings.Contains(footerView, "open PR URL") {
		t.Fatalf("expected open PR hint to be hidden for local session, got: %s", footerView)
	}
}

func TestUpdateFooterHints_AgentSessionWithoutPRHidesOpenPRHint(t *testing.T) {
	m := NewModel("", false)
	m.viewMode = ViewModeList
	m.taskList.SetTasks([]data.Session{
		{
			ID:         "agent-1",
			Status:     "running",
			Title:      "Agent",
			Repository: "owner/repo",
			Source:     data.SourceAgentTask,
		},
	})

	m.updateFooterHints()
	footerView := m.footer.View()
	if !strings.Contains(footerView, "show task logs") {
		t.Fatalf("expected logs hint for agent session, got: %s", footerView)
	}
	if strings.Contains(footerView, "open PR URL") {
		t.Fatalf("expected open PR hint to be hidden when PR is not linked, got: %s", footerView)
	}
}
