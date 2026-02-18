package data

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
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

	if cmdParts[0] != "gh" {
		fmt.Fprintf(os.Stderr, "wrong command: %v\n", cmdParts)
		os.Exit(1)
	}

	testMode := os.Getenv("TEST_SCENARIO")

	// Handle pr subcommands before agent-task guard
	if len(cmdParts) >= 2 && cmdParts[1] == "pr" {
		if testMode == "pr_diff_success" {
			fmt.Fprint(os.Stdout, "diff --git a/main.go b/main.go\nindex abc..def 100644\n--- a/main.go\n+++ b/main.go\n@@ -1,3 +1,4 @@\n package main\n-func old() {}\n+func new() {}\n+func extra() {}\n")
			os.Exit(0)
		}
		if testMode == "error" {
			fmt.Fprintln(os.Stderr, "command execution failed")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "unknown pr scenario: %s\n", testMode)
		os.Exit(1)
	}

	if cmdParts[1] != "agent-task" {
		fmt.Fprintf(os.Stderr, "wrong command: %v\n", cmdParts)
		os.Exit(1)
	}

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
		if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
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
		if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
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
		mockCmd.Env = append(os.Environ(),
			"GO_WANT_HELPER_PROCESS=1",
			"TEST_SCENARIO="+testScenario,
		)
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
	if firstTask.Source != "agent-task" {
		t.Errorf("wrong source: expected 'agent-task', got '%s'", firstTask.Source)
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
	if result.Source != "agent-task" {
		t.Errorf("wrong source: expected 'agent-task', got '%s'", result.Source)
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

func TestFetchPRDiff_ValidData(t *testing.T) {
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	execCommand = createMockExecCommand("pr_diff_success")

	result, err := FetchPRDiff(42, "owner/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "diff --git") {
		t.Error("expected diff output to contain 'diff --git'")
	}
	if !strings.Contains(result, "+func new()") {
		t.Error("expected diff output to contain added line")
	}
}

func TestFetchPRDiff_InvalidInputs(t *testing.T) {
	_, err := FetchPRDiff(0, "owner/repo")
	if err == nil {
		t.Error("expected error for PR number 0")
	}

	_, err = FetchPRDiff(1, "")
	if err == nil {
		t.Error("expected error for empty repo")
	}
}

func TestFetchPRDiff_CommandError(t *testing.T) {
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	execCommand = createMockExecCommand("error")

	_, err := FetchPRDiff(42, "owner/repo")
	if err == nil {
		t.Fatal("expected error when command fails, got none")
	}
}

func TestFetchPRForBranch_EmptyRepoAndBranch(t *testing.T) {
	num, url, err := FetchPRForBranch("", "")
	if num != 0 || url != "" || err != nil {
		t.Fatalf("expected (0, \"\", nil) for empty repo/branch, got (%d, %q, %v)", num, url, err)
	}
}

func TestFetchPRForBranch_MainBranch(t *testing.T) {
	num, url, err := FetchPRForBranch("owner/repo", "main")
	if num != 0 || url != "" || err != nil {
		t.Fatalf("expected (0, \"\", nil) for main branch, got (%d, %q, %v)", num, url, err)
	}
}

func TestFetchPRForBranch_MasterBranch(t *testing.T) {
	num, url, err := FetchPRForBranch("owner/repo", "master")
	if num != 0 || url != "" || err != nil {
		t.Fatalf("expected (0, \"\", nil) for master branch, got (%d, %q, %v)", num, url, err)
	}
}
