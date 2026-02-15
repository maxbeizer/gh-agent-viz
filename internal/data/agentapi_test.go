package data

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

// TestHelperProcess mocks the gh CLI command for testing
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	args := os.Args
	separatorIdx := -1
	for i, arg := range args {
		if arg == "--" {
			separatorIdx = i
			break
		}
	}
	if separatorIdx == -1 {
		fmt.Fprintln(os.Stderr, "missing separator")
		os.Exit(1)
	}

	cmdParts := args[separatorIdx+1:]
	if len(cmdParts) < 2 {
		fmt.Fprintln(os.Stderr, "insufficient args")
		os.Exit(1)
	}

	if cmdParts[0] != "gh" || cmdParts[1] != "agent-task" {
		fmt.Fprintf(os.Stderr, "wrong command: %v\n", cmdParts)
		os.Exit(1)
	}

	testMode := os.Getenv("TEST_SCENARIO")

	if testMode == "list_success" {
		result := []AgentTask{
			{
				ID:         "abc123",
				Status:     "completed",
				Title:      "Fix bug",
				Repository: "owner/repo",
				Branch:     "main",
				PRURL:      "https://github.com/owner/repo/pull/1",
				PRNumber:   1,
				CreatedAt:  time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				UpdatedAt:  time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
			},
		}
		json.NewEncoder(os.Stdout).Encode(result)
		os.Exit(0)
	}

	if testMode == "list_empty" {
		fmt.Fprintln(os.Stdout, "[]")
		os.Exit(0)
	}

	if testMode == "list_malformed" {
		fmt.Fprintln(os.Stdout, "{not valid json")
		os.Exit(0)
	}

	if testMode == "detail_success" {
		result := AgentTask{
			ID:         "abc123",
			Status:     "running",
			Title:      "Add feature",
			Repository: "owner/repo",
			Branch:     "feature",
			PRURL:      "https://github.com/owner/repo/pull/2",
			PRNumber:   2,
			CreatedAt:  time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			UpdatedAt:  time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
		}
		json.NewEncoder(os.Stdout).Encode(result)
		os.Exit(0)
	}

	if testMode == "log_success" {
		fmt.Fprintln(os.Stdout, "Log line 1")
		fmt.Fprintln(os.Stdout, "Log line 2")
		os.Exit(0)
	}

	if testMode == "error" {
		fmt.Fprintln(os.Stderr, "command execution failed")
		os.Exit(1)
	}

	if testMode == "check_repo_flag" {
		hasRepoFlag := false
		for i, arg := range cmdParts {
			if arg == "-R" && i+1 < len(cmdParts) {
				hasRepoFlag = true
				repoArg := cmdParts[i+1]
				fmt.Fprintf(os.Stdout, `[{"id":"test","status":"completed","title":"Found repo: %s"}]`, repoArg)
				os.Exit(0)
			}
		}
		if !hasRepoFlag {
			fmt.Fprintln(os.Stdout, `[{"id":"test","status":"completed","title":"No repo flag"}]`)
			os.Exit(0)
		}
	}

	fmt.Fprintf(os.Stderr, "unknown scenario: %s\n", testMode)
	os.Exit(1)
}

func createMockExecCommand(testScenario string) func(string, ...string) *exec.Cmd {
	return func(commandName string, commandArgs ...string) *exec.Cmd {
		fullArgs := []string{"-test.run=TestHelperProcess", "--", commandName}
		fullArgs = append(fullArgs, commandArgs...)
		mockCmd := exec.Command(os.Args[0], fullArgs...)
		mockCmd.Env = []string{
			"GO_WANT_HELPER_PROCESS=1",
			"TEST_SCENARIO=" + testScenario,
		}
		return mockCmd
	}
}

func TestFetchAgentTasks_ValidJSON(t *testing.T) {
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	execCommand = createMockExecCommand("list_success")

	result, err := FetchAgentTasks("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 task, got %d", len(result))
	}

	firstTask := result[0]
	if firstTask.ID != "abc123" {
		t.Errorf("wrong ID: expected 'abc123', got '%s'", firstTask.ID)
	}
	if firstTask.Status != "completed" {
		t.Errorf("wrong status: expected 'completed', got '%s'", firstTask.Status)
	}
	if firstTask.Title != "Fix bug" {
		t.Errorf("wrong title: expected 'Fix bug', got '%s'", firstTask.Title)
	}
}

func TestFetchAgentTasks_EmptyList(t *testing.T) {
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	execCommand = createMockExecCommand("list_empty")

	result, err := FetchAgentTasks("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected empty list, got %d items", len(result))
	}
}

func TestFetchAgentTasks_InvalidJSON(t *testing.T) {
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	execCommand = createMockExecCommand("list_malformed")

	_, err := FetchAgentTasks("")
	if err == nil {
		t.Fatal("expected error for invalid JSON, got none")
	}
}

func TestFetchAgentTasks_CommandError(t *testing.T) {
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	execCommand = createMockExecCommand("error")

	_, err := FetchAgentTasks("")
	if err == nil {
		t.Fatal("expected error when command fails, got none")
	}
}

func TestFetchAgentTasks_RepoScoping(t *testing.T) {
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	execCommand = createMockExecCommand("check_repo_flag")

	result, err := FetchAgentTasks("owner/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 task, got %d", len(result))
	}

	if result[0].Title != "Found repo: owner/repo" {
		t.Errorf("repo flag not passed correctly, got title: %s", result[0].Title)
	}
}

func TestFetchAgentTasks_NoRepoScoping(t *testing.T) {
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	execCommand = createMockExecCommand("check_repo_flag")

	result, err := FetchAgentTasks("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 task, got %d", len(result))
	}

	if result[0].Title != "No repo flag" {
		t.Errorf("unexpected behavior when no repo specified, got title: %s", result[0].Title)
	}
}

func TestFetchAgentTaskDetail_ValidData(t *testing.T) {
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	execCommand = createMockExecCommand("detail_success")

	result, err := FetchAgentTaskDetail("abc123", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "abc123" {
		t.Errorf("wrong ID: expected 'abc123', got '%s'", result.ID)
	}
	if result.Status != "running" {
		t.Errorf("wrong status: expected 'running', got '%s'", result.Status)
	}
	if result.Title != "Add feature" {
		t.Errorf("wrong title: expected 'Add feature', got '%s'", result.Title)
	}
}

func TestFetchAgentTaskDetail_CommandError(t *testing.T) {
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	execCommand = createMockExecCommand("error")

	_, err := FetchAgentTaskDetail("abc123", "")
	if err == nil {
		t.Fatal("expected error when command fails, got none")
	}
}

func TestFetchAgentTaskLog_ValidData(t *testing.T) {
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	execCommand = createMockExecCommand("log_success")

	result, err := FetchAgentTaskLog("abc123", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == "" {
		t.Error("expected non-empty log output")
	}

	if len(result) < 5 {
		t.Errorf("log output seems too short: %s", result)
	}
}

func TestFetchAgentTaskLog_CommandError(t *testing.T) {
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	execCommand = createMockExecCommand("error")

	_, err := FetchAgentTaskLog("abc123", "")
	if err == nil {
		t.Fatal("expected error when command fails, got none")
	}
}
