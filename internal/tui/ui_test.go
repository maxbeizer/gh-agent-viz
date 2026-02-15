package tui

import (
	"testing"

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
	errMsg, ok := msg.(errMsg)
	if !ok {
		t.Fatal("expected errMsg for completed session")
	}

	if errMsg.err == nil {
		t.Error("expected error for completed session")
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
	errMsg, ok := msg.(errMsg)
	if !ok {
		t.Fatal("expected errMsg for failed session")
	}

	if errMsg.err == nil {
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
	errMsg, ok := msg.(errMsg)
	if !ok {
		t.Fatal("expected errMsg for nil task")
	}

	if errMsg.err == nil {
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
	errMsg, ok := msg.(errMsg)
	if !ok {
		t.Fatal("expected errMsg for empty session ID")
	}

	if errMsg.err == nil {
		t.Error("expected error for empty session ID")
	}
}
