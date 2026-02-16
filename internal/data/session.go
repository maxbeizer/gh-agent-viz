package data

import (
	"strings"
	"time"
)

// SessionSource represents where a session originated from
type SessionSource string

const (
	SourceAgentTask    SessionSource = "agent-task"
	SourceLocalCopilot SessionSource = "local-copilot"
)

// AttentionStaleThreshold is the quiet-window threshold for active sessions.
const AttentionStaleThreshold = 20 * time.Minute

// SessionTelemetry holds derived usage metrics for a session
type SessionTelemetry struct {
	Duration          time.Duration // elapsed time from created to last activity
	ConversationTurns int           // number of conversation exchanges
	UserMessages      int           // messages from user
	AssistantMessages int           // messages from assistant
}

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
	Telemetry  *SessionTelemetry `json:"telemetry,omitempty"`
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

// SessionNeedsAttention indicates whether a session likely requires operator action.
func SessionNeedsAttention(session Session) bool {
	status := strings.ToLower(strings.TrimSpace(session.Status))
	if status == "needs-input" || status == "failed" {
		return true
	}

	if !StatusIsActive(session.Status) || session.UpdatedAt.IsZero() {
		return false
	}

	return time.Since(session.UpdatedAt) >= AttentionStaleThreshold
}

// StatusIsActive determines if a status string represents an active session.
func StatusIsActive(status string) bool {
	normalized := strings.ToLower(strings.TrimSpace(status))
	return normalized == "running" || normalized == "queued" || normalized == "active" || normalized == "open" || normalized == "in progress"
}
