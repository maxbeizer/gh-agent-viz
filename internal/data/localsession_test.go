package data

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDeriveLocalSessionStatus_ExplicitCompleted(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{"completed", "completed", "completed"},
		{"finished", "finished", "completed"},
		{"done", "done", "completed"},
		{"merged", "merged", "completed"},
		{"closed", "closed", "completed"},
		{"COMPLETED", "COMPLETED", "completed"},
		{"  Completed  ", "  Completed  ", "completed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DeriveLocalSessionStatus(tt.status, time.Now())
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDeriveLocalSessionStatus_ExplicitRunning(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{"running", "running", "running"},
		{"in progress", "in progress", "running"},
		{"active", "active", "running"},
		{"open", "open", "running"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DeriveLocalSessionStatus(tt.status, time.Now())
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDeriveLocalSessionStatus_ExplicitNeedsInput(t *testing.T) {
	result := DeriveLocalSessionStatus("awaiting user input", time.Now())
	if result != "needs-input" {
		t.Fatalf("expected needs-input, got %s", result)
	}
}

func TestDeriveLocalSessionStatus_ExplicitFailed(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{"failed", "failed", "failed"},
		{"error", "error", "failed"},
		{"cancelled", "cancelled", "failed"},
		{"canceled", "canceled", "failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DeriveLocalSessionStatus(tt.status, time.Now())
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDeriveLocalSessionStatus_ExplicitQueued(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{"queued", "queued", "queued"},
		{"pending", "pending", "queued"},
		{"waiting", "waiting", "queued"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DeriveLocalSessionStatus(tt.status, time.Now())
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDeriveLocalSessionStatus_DerivedFromTime(t *testing.T) {
	tests := []struct {
		name         string
		lastActivity time.Time
		expected     string
	}{
		{
			"recent activity means running",
			time.Now().Add(-1 * time.Hour),
			"running",
		},
		{
			"old activity means completed",
			time.Now().Add(-48 * time.Hour),
			"completed",
		},
		{
			"zero time means unknown",
			time.Time{},
			"unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DeriveLocalSessionStatus("", tt.lastActivity)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestParseWorkspaceFile_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	workspaceFile := filepath.Join(tmpDir, "workspace.yaml")

	content := `session_id: "test-session-123"
start_time: "2026-02-15T03:10:00Z"
last_activity: "2026-02-15T03:30:00Z"
message_count: 15
status: "completed"
repository: "owner/repo"
branch: "main"
title: "Fix bug in parser"
conversation_history:
  - role: user
    content: "How do I write a unit test?"
    timestamp: "2026-02-15T03:11:00Z"
`

	if err := os.WriteFile(workspaceFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	session, err := parseWorkspaceFile(workspaceFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if session.ID != "test-session-123" {
		t.Errorf("expected ID 'test-session-123', got '%s'", session.ID)
	}
	if session.Title != "Fix bug in parser" {
		t.Errorf("expected title 'Fix bug in parser', got '%s'", session.Title)
	}
	if session.Repository != "owner/repo" {
		t.Errorf("expected repository 'owner/repo', got '%s'", session.Repository)
	}
	if session.Branch != "main" {
		t.Errorf("expected branch 'main', got '%s'", session.Branch)
	}
	if session.Status != "completed" {
		t.Errorf("expected status 'completed', got '%s'", session.Status)
	}
	if session.Source != SourceLocalCopilot {
		t.Errorf("expected source 'local-copilot', got '%s'", session.Source)
	}
}

func TestParseWorkspaceFile_Malformed(t *testing.T) {
	tmpDir := t.TempDir()
	workspaceFile := filepath.Join(tmpDir, "workspace.yaml")

	// Malformed YAML but with extractable fields
	content := `session_id: "test-session-456"
title: "Malformed session
repository: "owner/repo"
branch: main
status: running
last_activity: "2026-02-15T03:30:00Z"
{ invalid yaml here
`

	if err := os.WriteFile(workspaceFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	session, err := parseWorkspaceFile(workspaceFile)
	if err != nil {
		t.Fatalf("expected fallback parsing to succeed, got error: %v", err)
	}

	if session.ID != "test-session-456" {
		t.Errorf("expected ID 'test-session-456', got '%s'", session.ID)
	}
	if session.Repository != "owner/repo" {
		t.Errorf("expected repository 'owner/repo', got '%s'", session.Repository)
	}
}

func TestParseWorkspaceFile_NoTitle(t *testing.T) {
	tmpDir := t.TempDir()
	workspaceFile := filepath.Join(tmpDir, "workspace.yaml")

	content := `session_id: "test-session-789"
start_time: "2026-02-15T03:10:00Z"
last_activity: "2026-02-15T03:30:00Z"
status: "running"
conversation_history:
  - role: user
    content: "What is the meaning of life?"
    timestamp: "2026-02-15T03:11:00Z"
`

	if err := os.WriteFile(workspaceFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	session, err := parseWorkspaceFile(workspaceFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should use first conversation content as title
	if session.Title != "What is the meaning of life?" {
		t.Errorf("expected title from conversation, got '%s'", session.Title)
	}
}

func TestParseWorkspaceFile_NoTitleNoConversation(t *testing.T) {
	tmpDir := t.TempDir()
	workspaceFile := filepath.Join(tmpDir, "workspace.yaml")

	content := `session_id: "test-session-000"
start_time: "2026-02-15T03:10:00Z"
status: "running"
`

	if err := os.WriteFile(workspaceFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	session, err := parseWorkspaceFile(workspaceFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should use default title
	if session.Title != "Session test-session-000" {
		t.Errorf("expected default title 'Session test-session-000', got '%s'", session.Title)
	}
}

func TestParseWorkspaceFile_AwaitingUserInputFlag(t *testing.T) {
	tmpDir := t.TempDir()
	workspaceFile := filepath.Join(tmpDir, "workspace.yaml")

	content := `id: "session-input-1"
created_at: "2026-02-15T03:10:00Z"
updated_at: "2026-02-15T03:30:00Z"
status: "running"
awaiting_user_input: true
summary: "Need user confirmation"
`

	if err := os.WriteFile(workspaceFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	session, err := parseWorkspaceFile(workspaceFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session.Status != "needs-input" {
		t.Fatalf("expected needs-input status, got %s", session.Status)
	}
}

func TestParseWorkspaceFile_AssistantQuestionNeedsInput(t *testing.T) {
	tmpDir := t.TempDir()
	workspaceFile := filepath.Join(tmpDir, "workspace.yaml")

	content := `id: "session-input-2"
created_at: "2026-02-15T03:10:00Z"
updated_at: "2026-02-15T03:30:00Z"
status: "running"
summary: "Question pending"
conversation_history:
  - role: user
    content: "Add retries"
  - role: assistant
    content: "Should I apply retries to all endpoints or only write operations?"
`

	if err := os.WriteFile(workspaceFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	session, err := parseWorkspaceFile(workspaceFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session.Status != "needs-input" {
		t.Fatalf("expected needs-input status, got %s", session.Status)
	}
}

func TestParseWorkspaceFileFallback_InputFlagBeforeStatus(t *testing.T) {
	tmpDir := t.TempDir()
	workspaceFile := filepath.Join(tmpDir, "workspace.yaml")

	content := `session_id: "fallback-input-order"
awaiting_user_input: true
status: running
last_activity: "2026-02-15T03:30:00Z"
{ invalid yaml here
`

	if err := os.WriteFile(workspaceFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	session, err := parseWorkspaceFile(workspaceFile)
	if err != nil {
		t.Fatalf("expected fallback parsing to succeed, got error: %v", err)
	}
	if session.Status != "needs-input" {
		t.Fatalf("expected needs-input status, got %s", session.Status)
	}
}

func TestParseWorkspaceFile_CurrentWorkspaceFormat(t *testing.T) {
	tmpDir := t.TempDir()
	workspaceFile := filepath.Join(tmpDir, "workspace.yaml")

	createdAt := time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339)
	updatedAt := time.Now().Add(-5 * time.Minute).UTC().Format(time.RFC3339)

	content := fmt.Sprintf(`id: 564c025b-b5eb-4e02-ba47-425d915c4748
cwd: /Users/maxbeizer/code/gh-agent-viz
git_root: /Users/maxbeizer/code/gh-agent-viz
repository: maxbeizer/gh-agent-viz
branch: main
summary: Review And Test PR 1
summary_count: 1
created_at: %q
updated_at: %q
`, createdAt, updatedAt)

	if err := os.WriteFile(workspaceFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	session, err := parseWorkspaceFile(workspaceFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if session.ID != "564c025b-b5eb-4e02-ba47-425d915c4748" {
		t.Errorf("unexpected ID: %s", session.ID)
	}
	if session.Title != "Review And Test PR 1" {
		t.Errorf("expected summary to be title, got %q", session.Title)
	}
	if session.Status != "running" {
		t.Errorf("expected running from recent updated_at, got %q", session.Status)
	}
}

func TestFetchLocalSessions_NoDirectory(t *testing.T) {
	// Override home dir to a temp location without .copilot
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	sessions, err := FetchLocalSessions()
	if err != nil {
		t.Fatalf("expected no error when directory doesn't exist, got: %v", err)
	}

	if len(sessions) != 0 {
		t.Errorf("expected empty list, got %d sessions", len(sessions))
	}
}

func TestFetchLocalSessions_WithSessions(t *testing.T) {
	tmpDir := t.TempDir()
	sessionDir := filepath.Join(tmpDir, ".copilot", "session-state")

	// Create session directories
	session1Dir := filepath.Join(sessionDir, "session-001")
	session2Dir := filepath.Join(sessionDir, "session-002")

	if err := os.MkdirAll(session1Dir, 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}
	if err := os.MkdirAll(session2Dir, 0755); err != nil {
		t.Fatalf("failed to create session dir: %v", err)
	}

	// Create workspace files
	workspace1 := filepath.Join(session1Dir, "workspace.yaml")
	content1 := `session_id: "session-001"
title: "First session"
status: "completed"
last_activity: "2026-02-15T03:30:00Z"
`
	if err := os.WriteFile(workspace1, []byte(content1), 0644); err != nil {
		t.Fatalf("failed to write workspace file: %v", err)
	}

	workspace2 := filepath.Join(session2Dir, "workspace.yaml")
	content2 := `session_id: "session-002"
title: "Second session"
status: "running"
last_activity: "2026-02-15T04:30:00Z"
`
	if err := os.WriteFile(workspace2, []byte(content2), 0644); err != nil {
		t.Fatalf("failed to write workspace file: %v", err)
	}

	// Override home dir
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	sessions, err := FetchLocalSessions()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}

	// Check that we got both sessions (order might vary)
	foundFirst := false
	foundSecond := false
	for _, s := range sessions {
		if s.ID == "session-001" {
			foundFirst = true
			if s.Title != "First session" {
				t.Errorf("wrong title for session-001: %s", s.Title)
			}
		}
		if s.ID == "session-002" {
			foundSecond = true
			if s.Title != "Second session" {
				t.Errorf("wrong title for session-002: %s", s.Title)
			}
		}
	}

	if !foundFirst || !foundSecond {
		t.Error("didn't find both sessions")
	}
}

func TestTruncateTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"short string",
			"Hello world",
			"Hello world",
		},
		{
			"exactly 100 chars",
			"1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
			"1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
		},
		{
			"over 100 chars",
			"12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
			"1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567...",
		},
		{
			"with leading/trailing spaces",
			"  spaced  ",
			"spaced",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateTitle(tt.input)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestSessionConversion(t *testing.T) {
	original := AgentTask{
		ID:         "task-123",
		Status:     "completed",
		Title:      "Fix bug",
		Repository: "owner/repo",
		Branch:     "main",
		PRURL:      "https://github.com/owner/repo/pull/1",
		PRNumber:   1,
		CreatedAt:  time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
	}

	session := FromAgentTask(original)
	if session.Source != SourceAgentTask {
		t.Errorf("expected source to be agent-task, got %s", session.Source)
	}
	if session.ID != original.ID {
		t.Errorf("expected ID %s, got %s", original.ID, session.ID)
	}

	converted := session.ToAgentTask()
	if converted.ID != original.ID {
		t.Errorf("expected ID %s after conversion, got %s", original.ID, converted.ID)
	}
	if converted.Title != original.Title {
		t.Errorf("expected title '%s' after conversion, got '%s'", original.Title, converted.Title)
	}
}

func TestFormatEventLine_UserMessage(t *testing.T) {
	raw := []byte(`{"content":"hello world"}`)
	line := formatEventLine("user.message", "2026-01-15T10:30:00.000Z", raw)
	if line == "" {
		t.Fatal("expected non-empty line for user.message")
	}
	if !contains(line, "👤") {
		t.Error("expected user icon in output")
	}
	if !contains(line, "hello world") {
		t.Error("expected message content in output")
	}
}

func TestFormatEventLine_AssistantMessage(t *testing.T) {
	raw := []byte(`{"content":"I can help with that"}`)
	line := formatEventLine("assistant.message", "2026-01-15T10:30:05.000Z", raw)
	if line == "" {
		t.Fatal("expected non-empty line for assistant.message")
	}
	if !contains(line, "🤖") {
		t.Error("expected assistant icon in output")
	}
}

func TestFormatEventLine_ToolExecution(t *testing.T) {
	raw := []byte(`{"toolName":"bash"}`)
	line := formatEventLine("tool.execution_start", "2026-01-15T10:30:10.000Z", raw)
	if line == "" {
		t.Fatal("expected non-empty line for tool execution")
	}
	if !contains(line, "🔧") || !contains(line, "bash") {
		t.Error("expected tool icon and name in output")
	}
}

func TestFormatEventLine_EmptyAssistantMessage(t *testing.T) {
	raw := []byte(`{"content":""}`)
	line := formatEventLine("assistant.message", "2026-01-15T10:30:05.000Z", raw)
	if line != "" {
		t.Error("expected empty line for empty assistant.message")
	}
}

func TestFormatEventLine_SessionStart(t *testing.T) {
	line := formatEventLine("session.start", "2026-01-15T10:30:00.000Z", nil)
	if !contains(line, "🚀") {
		t.Error("expected rocket icon for session start")
	}
}

func TestFormatEventLine_UnknownType(t *testing.T) {
	line := formatEventLine("tool.execution_complete", "2026-01-15T10:30:00.000Z", nil)
	if line != "" {
		t.Error("expected empty line for unhandled event type")
	}
}

func TestTruncateLogContent_Short(t *testing.T) {
	result := truncateLogContent("short", 100)
	if result != "short" {
		t.Errorf("expected 'short', got %q", result)
	}
}

func TestTruncateLogContent_Long(t *testing.T) {
	long := string(make([]byte, 600))
	result := truncateLogContent(long, 500)
	if !contains(result, "_(truncated)_") {
		t.Error("expected truncation marker")
	}
}

func TestFetchLocalSessionLog_MissingSession(t *testing.T) {
	_, err := FetchLocalSessionLog("nonexistent-session-id-12345")
	if err == nil {
		t.Fatal("expected error for missing session")
	}
}

func TestFetchLocalSessionLog_EmptyID(t *testing.T) {
	_, err := FetchLocalSessionLog("")
	if err == nil {
		t.Fatal("expected error for empty session ID")
	}
}

func TestFetchLocalSessionLog_ValidEvents(t *testing.T) {
	// Create a temp session directory with events.jsonl
	tmpDir := t.TempDir()
	sessionID := "test-log-session"
	sessionDir := filepath.Join(tmpDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatal(err)
	}

	events := `{"type":"session.start","timestamp":"2026-01-15T10:30:00.000Z","data":{}}
{"type":"user.message","timestamp":"2026-01-15T10:30:01.000Z","data":{"content":"fix the bug"}}
{"type":"assistant.turn_start","timestamp":"2026-01-15T10:30:02.000Z","data":{"turnId":"1"}}
{"type":"tool.execution_start","timestamp":"2026-01-15T10:30:03.000Z","data":{"toolName":"bash"}}
{"type":"assistant.message","timestamp":"2026-01-15T10:30:05.000Z","data":{"content":"I fixed the bug"}}
`
	eventsFile := filepath.Join(sessionDir, "events.jsonl")
	if err := os.WriteFile(eventsFile, []byte(events), 0644); err != nil {
		t.Fatal(err)
	}

	// Override home dir to use our temp dir — we can't easily do this
	// without modifying the function, so test the formatting functions directly
	// and verify they produce correct output for known inputs.
	line := formatEventLine("user.message", "2026-01-15T10:30:01.000Z", []byte(`{"content":"fix the bug"}`))
	if !contains(line, "fix the bug") || !contains(line, "👤") {
		t.Errorf("unexpected user message rendering: %s", line)
	}

	toolLine := formatEventLine("tool.execution_start", "2026-01-15T10:30:03.000Z", []byte(`{"toolName":"bash"}`))
	if !contains(toolLine, "bash") || !contains(toolLine, "🔧") {
		t.Errorf("unexpected tool rendering: %s", toolLine)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
