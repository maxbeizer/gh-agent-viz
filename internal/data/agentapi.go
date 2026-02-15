package data

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

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
	jsonOutput, jsonErr := exec.Command("gh", "agent-task", "list", "--json").CombinedOutput()
	if jsonErr == nil {
		var tasks []AgentTask
		if err := json.Unmarshal(jsonOutput, &tasks); err != nil {
			return nil, fmt.Errorf("failed to parse agent tasks: %w", err)
		}
		if repo == "" {
			return tasks, nil
		}
		filtered := make([]AgentTask, 0, len(tasks))
		for _, task := range tasks {
			if task.Repository == repo {
				filtered = append(filtered, task)
			}
		}
		return filtered, nil
	}

	// Fallback for current gh agent-task CLI versions that don't support --json.
	if !strings.Contains(string(jsonOutput), "unknown flag: --json") {
		return nil, fmt.Errorf("failed to fetch agent tasks: %s", strings.TrimSpace(string(jsonOutput)))
	}

	output, err := exec.Command("gh", "agent-task", "list").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch agent tasks: %s", strings.TrimSpace(string(output)))
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	tasks := make([]AgentTask, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 5 {
			continue
		}

		taskRepo := strings.TrimSpace(fields[2])
		if repo != "" && taskRepo != repo {
			continue
		}

		id := strings.TrimPrefix(strings.TrimSpace(fields[1]), "#")
		prNumber, _ := strconv.Atoi(id)
		updatedAt, _ := time.Parse(time.RFC3339, strings.TrimSpace(fields[4]))

		tasks = append(tasks, AgentTask{
			ID:         id,
			Status:     normalizeStatus(strings.TrimSpace(fields[3])),
			Title:      strings.TrimSpace(fields[0]),
			Repository: taskRepo,
			PRURL:      fmt.Sprintf("https://github.com/%s/pull/%d", taskRepo, prNumber),
			PRNumber:   prNumber,
			UpdatedAt:  updatedAt,
		})
	}

	return tasks, nil
}

// FetchAgentTaskDetail retrieves detailed information for a specific agent task
func FetchAgentTaskDetail(id string, repo string) (*AgentTask, error) {
	if id == "" {
		return nil, fmt.Errorf("task id is required")
	}

	args := []string{"agent-task", "view", id, "--json"}
	if repo != "" {
		args = append(args, "-R", repo)
	}

	output, err := exec.Command("gh", args...).CombinedOutput()
	if err == nil {
		var task AgentTask
		if err := json.Unmarshal(output, &task); err != nil {
			return nil, fmt.Errorf("failed to parse agent task detail: %w", err)
		}
		return &task, nil
	}

	if !strings.Contains(string(output), "unknown flag: --json") && !strings.Contains(string(output), "session ID is required") {
		return nil, fmt.Errorf("failed to fetch agent task detail: %s", strings.TrimSpace(string(output)))
	}

	// Fallback: use PR metadata when session detail JSON is unavailable.
	prArgs := []string{"pr", "view", id, "--json", "number,title,headRefName,url,state,createdAt,updatedAt"}
	if repo != "" {
		prArgs = append(prArgs, "-R", repo)
	}

	prOutput, prErr := exec.Command("gh", prArgs...).CombinedOutput()
	if prErr != nil {
		return nil, fmt.Errorf("failed to fetch agent task detail: %s", strings.TrimSpace(string(prOutput)))
	}

	var pr struct {
		Number      int       `json:"number"`
		Title       string    `json:"title"`
		HeadRefName string    `json:"headRefName"`
		URL         string    `json:"url"`
		State       string    `json:"state"`
		CreatedAt   time.Time `json:"createdAt"`
		UpdatedAt   time.Time `json:"updatedAt"`
	}
	if err := json.Unmarshal(prOutput, &pr); err != nil {
		return nil, fmt.Errorf("failed to parse task detail: %w", err)
	}

	return &AgentTask{
		ID:         id,
		Status:     normalizeStatus(pr.State),
		Title:      pr.Title,
		Repository: repo,
		Branch:     pr.HeadRefName,
		PRURL:      pr.URL,
		PRNumber:   pr.Number,
		CreatedAt:  pr.CreatedAt,
		UpdatedAt:  pr.UpdatedAt,
	}, nil
}

// FetchAgentTaskLog retrieves the event log for a specific agent task
func FetchAgentTaskLog(id string, repo string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("task id is required")
	}

	args := []string{"agent-task", "view", id, "--log"}
	if repo != "" {
		args = append(args, "-R", repo)
	}

	output, err := exec.Command("gh", args...).CombinedOutput()
	if err != nil {
		trimmed := strings.TrimSpace(string(output))
		if strings.Contains(trimmed, "session ID is required") {
			return "", fmt.Errorf("agent logs require a session ID; open the task in the browser with 'o'")
		}
		return "", fmt.Errorf("failed to fetch agent task log: %s", trimmed)
	}

	return strings.TrimSpace(string(output)), nil
}

func normalizeStatus(status string) string {
	normalized := strings.ToLower(strings.TrimSpace(status))
	switch normalized {
	case "ready for review", "merged", "closed", "completed":
		return "completed"
	case "queued", "pending":
		return "queued"
	case "in progress", "running", "open":
		return "running"
	case "failed", "cancelled", "canceled":
		return "failed"
	default:
		return normalized
	}
}
