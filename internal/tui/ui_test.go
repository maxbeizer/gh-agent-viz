package tui

import (
	"testing"

	"github.com/maxbeizer/gh-agent-viz/internal/data"
)

func TestResumeSession_ValidRunningSession(t *testing.T) {
	m := NewModel("")

	// Create a running task
	task := &data.AgentTask{
		ID:     "test-session-123",
		Status: "running",
		Title:  "Test Running Task",
	}

	// Test that running sessions are resumable
	cmd := m.resumeSession(task)
	if cmd == nil {
		t.Error("expected cmd to be non-nil for running session")
	}
}

func TestResumeSession_ValidQueuedSession(t *testing.T) {
	m := NewModel("")

	// Create a queued task
	task := &data.AgentTask{
		ID:     "test-session-456",
		Status: "queued",
		Title:  "Test Queued Task",
	}

	// Test that queued sessions are resumable
	cmd := m.resumeSession(task)
	if cmd == nil {
		t.Error("expected cmd to be non-nil for queued session")
	}
}

func TestResumeSession_CompletedSession(t *testing.T) {
	m := NewModel("")

	// Create a completed task
	task := &data.AgentTask{
		ID:     "test-session-789",
		Status: "completed",
		Title:  "Test Completed Task",
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
	m := NewModel("")

	// Create a failed task
	task := &data.AgentTask{
		ID:     "test-session-999",
		Status: "failed",
		Title:  "Test Failed Task",
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
	m := NewModel("")

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
	m := NewModel("")

	// Create a task with empty ID
	task := &data.AgentTask{
		ID:     "",
		Status: "running",
		Title:  "Test Task No ID",
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
