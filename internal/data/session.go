package data

import (
	"fmt"
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

// AttentionStaleMax is the upper bound — sessions idle longer than this
// are considered abandoned and no longer need attention.
const AttentionStaleMax = 4 * time.Hour

// SessionTelemetry holds derived usage metrics for a session
type SessionTelemetry struct {
	Duration          time.Duration // elapsed time from created to last activity
	ConversationTurns int           // number of conversation exchanges
	UserMessages      int           // messages from user
	AssistantMessages int           // messages from assistant
	// Token usage from CLI logs
	Model        string // last model used
	InputTokens  int64  // total prompt tokens
	OutputTokens int64  // total completion tokens
	CachedTokens int64  // total cached prompt tokens
	ModelCalls   int    // number of model API calls
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
	HasLog     bool              `json:"-"` // true when a viewable log exists (e.g. events.jsonl)
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

// SessionNeedsAttention indicates whether a session requires operator action.
// Only true for sessions explicitly waiting on user input or that have failed.
// Idle sessions are informational, not actionable.
func SessionNeedsAttention(session Session) bool {
	status := strings.ToLower(strings.TrimSpace(session.Status))
	return status == "needs-input" || status == "failed"
}

// StatusIsActive determines if a status string represents an active session.
func StatusIsActive(status string) bool {
	normalized := strings.ToLower(strings.TrimSpace(status))
	return normalized == "running" || normalized == "queued" || normalized == "active" || normalized == "open" || normalized == "in progress"
}

// SessionIsActiveNotIdle returns true for sessions actively working (not idle 20+ min).
func SessionIsActiveNotIdle(session Session) bool {
	if !StatusIsActive(session.Status) && !strings.EqualFold(strings.TrimSpace(session.Status), "needs-input") {
		return false
	}
	if session.UpdatedAt.IsZero() {
		return true
	}
	return time.Since(session.UpdatedAt) < AttentionStaleThreshold
}

// IsDefaultBranch returns true for main/master/empty branch names.
func IsDefaultBranch(branch string) bool {
	b := strings.ToLower(strings.TrimSpace(branch))
	return b == "" || b == "main" || b == "master"
}

// FormatTokenCount formats a token count as a human-readable string (e.g., "2.7M", "11.7K", "437").
func FormatTokenCount(n int64) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}
