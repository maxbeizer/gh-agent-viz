package data

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// execCommand is a variable to allow mocking exec.Command in tests
var execCommand = exec.Command

// AgentTask represents a GitHub Copilot agent task session
type AgentTask struct {
	ID         string    `json:"id"`
	Status     string    `json:"status"`
	Title      string    `json:"title"`
	Repository string    `json:"repository"`
	Branch     string    `json:"branch"`
	PRURL      string    `json:"prUrl"`
	PRNumber   int       `json:"prNumber"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// FetchAgentTasks retrieves the list of agent tasks, optionally scoped to a repository
func FetchAgentTasks(repo string) ([]AgentTask, error) {
	args := []string{"agent-task", "list", "--json"}
	if repo != "" {
		args = append(args, "-R", repo)
	}

	cmd := execCommand("gh", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch agent tasks: %w", err)
	}

	var tasks []AgentTask
	if err := json.Unmarshal(output, &tasks); err != nil {
		return nil, fmt.Errorf("failed to parse agent tasks: %w", err)
	}

	return tasks, nil
}

// FetchAgentTaskDetail retrieves detailed information for a specific agent task
func FetchAgentTaskDetail(id string, repo string) (*AgentTask, error) {
	args := []string{"agent-task", "view", id, "--json"}
	if repo != "" {
		args = append(args, "-R", repo)
	}

	cmd := execCommand("gh", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch agent task detail: %w", err)
	}

	var task AgentTask
	if err := json.Unmarshal(output, &task); err != nil {
		return nil, fmt.Errorf("failed to parse agent task detail: %w", err)
	}

	return &task, nil
}

// FetchAgentTaskLog retrieves the event log for a specific agent task
func FetchAgentTaskLog(id string, repo string) (string, error) {
	args := []string{"agent-task", "view", id, "--log"}
	if repo != "" {
		args = append(args, "-R", repo)
	}

	cmd := execCommand("gh", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to fetch agent task log: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}
