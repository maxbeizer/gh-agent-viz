package data

import (
	"time"
)

// SessionSource represents where a session originated from
type SessionSource string

const (
	SourceAgentTask   SessionSource = "agent-task"
	SourceLocalCopilot SessionSource = "local-copilot"
)

// Session represents a unified model for both agent-task and local Copilot sessions
type Session struct {
	ID         string        `json:"id"`
	Status     string        `json:"status"`
	Title      string        `json:"title"`
	Repository string        `json:"repository"`
	Branch     string        `json:"branch"`
	PRURL      string        `json:"prUrl"`
	PRNumber   int           `json:"prNumber"`
	CreatedAt  time.Time     `json:"createdAt"`
	UpdatedAt  time.Time     `json:"updatedAt"`
	Source     SessionSource `json:"source"`
}

// FromAgentTask converts an AgentTask to a Session
func FromAgentTask(task AgentTask) Session {
	return Session{
		ID:         task.ID,
		Status:     task.Status,
		Title:      task.Title,
		Repository: task.Repository,
		Branch:     task.Branch,
		PRURL:      task.PRURL,
		PRNumber:   task.PRNumber,
		CreatedAt:  task.CreatedAt,
		UpdatedAt:  task.UpdatedAt,
		Source:     SourceAgentTask,
	}
}

// ToAgentTask converts a Session back to an AgentTask for backward compatibility
func (s Session) ToAgentTask() AgentTask {
	return AgentTask{
		ID:         s.ID,
		Status:     s.Status,
		Title:      s.Title,
		Repository: s.Repository,
		Branch:     s.Branch,
		PRURL:      s.PRURL,
		PRNumber:   s.PRNumber,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
	}
}
