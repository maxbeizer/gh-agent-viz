package data

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// LocalSessionWorkspace represents the structure of a workspace.yaml file
type LocalSessionWorkspace struct {
	ID                  string                   `yaml:"id"`
	SessionID           string                   `yaml:"session_id"`
	CreatedAt           string                   `yaml:"created_at"`
	UpdatedAt           string                   `yaml:"updated_at"`
	StartTime           string                   `yaml:"start_time"`
	LastActivity        string                   `yaml:"last_activity"`
	Status              string                   `yaml:"status"`
	Repository          string                   `yaml:"repository"`
	Branch              string                   `yaml:"branch"`
	Title               string                   `yaml:"title"`
	Summary             string                   `yaml:"summary"`
	ConversationHistory []map[string]interface{} `yaml:"conversation_history"`
}

// FetchLocalSessions retrieves local Copilot CLI sessions from ~/.copilot/session-state/
func FetchLocalSessions() ([]Session, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	sessionDir := filepath.Join(homeDir, ".copilot", "session-state")

	// Check if directory exists
	if _, err := os.Stat(sessionDir); os.IsNotExist(err) {
		// Not an error - just no local sessions
		return []Session{}, nil
	}

	// Read all subdirectories
	entries, err := os.ReadDir(sessionDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read session directory: %w", err)
	}

	var sessions []Session
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		workspaceFile := filepath.Join(sessionDir, entry.Name(), "workspace.yaml")
		session, err := parseWorkspaceFile(workspaceFile)
		if err != nil {
			// Tolerant parsing - log error but continue
			continue
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

// parseWorkspaceFile parses a single workspace.yaml file with tolerant error handling
func parseWorkspaceFile(path string) (Session, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Session{}, fmt.Errorf("failed to read workspace file: %w", err)
	}

	var workspace LocalSessionWorkspace
	if err := yaml.Unmarshal(data, &workspace); err != nil {
		// Try to extract what we can from a malformed file
		return parseWorkspaceFileFallback(data)
	}

	return convertLocalSessionToSession(workspace)
}

// parseWorkspaceFileFallback attempts best-effort parsing of malformed YAML
func parseWorkspaceFileFallback(data []byte) (Session, error) {
	// Try to extract key fields manually
	lines := strings.Split(string(data), "\n")
	session := Session{
		Source: SourceLocalCopilot,
		Status: "unknown",
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "id:") {
			session.ID = strings.Trim(strings.TrimPrefix(line, "id:"), `" `)
		} else if strings.HasPrefix(line, "summary:") && session.Title == "" {
			session.Title = strings.Trim(strings.TrimPrefix(line, "summary:"), `" `)
		} else if strings.HasPrefix(line, "updated_at:") {
			timeStr := strings.Trim(strings.TrimPrefix(line, "updated_at:"), `" `)
			if t, ok := parseAnyTime(timeStr); ok {
				session.UpdatedAt = t
			}
		} else if strings.HasPrefix(line, "created_at:") {
			timeStr := strings.Trim(strings.TrimPrefix(line, "created_at:"), `" `)
			if t, ok := parseAnyTime(timeStr); ok {
				session.CreatedAt = t
			}
		} else if strings.HasPrefix(line, "session_id:") {
			session.ID = strings.Trim(strings.TrimPrefix(line, "session_id:"), `" `)
		} else if strings.HasPrefix(line, "title:") {
			session.Title = strings.Trim(strings.TrimPrefix(line, "title:"), `" `)
		} else if strings.HasPrefix(line, "repository:") {
			session.Repository = strings.Trim(strings.TrimPrefix(line, "repository:"), `" `)
		} else if strings.HasPrefix(line, "branch:") {
			session.Branch = strings.Trim(strings.TrimPrefix(line, "branch:"), `" `)
		} else if strings.HasPrefix(line, "status:") {
			session.Status = strings.Trim(strings.TrimPrefix(line, "status:"), `" `)
		} else if strings.HasPrefix(line, "last_activity:") {
			timeStr := strings.Trim(strings.TrimPrefix(line, "last_activity:"), `" `)
			if t, ok := parseAnyTime(timeStr); ok {
				session.UpdatedAt = t
			}
		}
	}

	// If we got at least an ID, consider it a valid session
	if session.ID == "" {
		return Session{}, fmt.Errorf("failed to extract session ID from malformed file")
	}

	// Apply status normalization
	if session.UpdatedAt.IsZero() {
		session.UpdatedAt = session.CreatedAt
	}
	session.Status = DeriveLocalSessionStatus(session.Status, session.UpdatedAt)

	if session.Title == "" {
		if session.ID != "" {
			session.Title = fmt.Sprintf("Session %s", truncateTitle(session.ID))
		} else {
			session.Title = "Untitled Session"
		}
	}

	return session, nil
}

// convertLocalSessionToSession converts a LocalSessionWorkspace to a Session
func convertLocalSessionToSession(workspace LocalSessionWorkspace) (Session, error) {
	session := Session{
		ID:         workspace.ID,
		Title:      workspace.Title,
		Repository: workspace.Repository,
		Branch:     workspace.Branch,
		Source:     SourceLocalCopilot,
	}
	if session.ID == "" {
		session.ID = workspace.SessionID
	}
	if session.Title == "" {
		session.Title = workspace.Summary
	}

	// Parse timestamps from current format first, then legacy fields.
	session.CreatedAt = parseSessionTime(workspace.CreatedAt, workspace.StartTime)
	session.UpdatedAt = parseSessionTime(workspace.UpdatedAt, workspace.LastActivity)
	if session.UpdatedAt.IsZero() {
		session.UpdatedAt = session.CreatedAt
	}

	// Derive status from metadata
	session.Status = DeriveLocalSessionStatus(workspace.Status, session.UpdatedAt)

	// Use title from conversation if not set
	if session.Title == "" && len(workspace.ConversationHistory) > 0 {
		if userContent, ok := workspace.ConversationHistory[0]["content"].(string); ok {
			session.Title = truncateTitle(userContent)
		}
	}

	// Default title if still empty
	if session.Title == "" {
		if session.ID != "" {
			session.Title = fmt.Sprintf("Session %s", truncateTitle(session.ID))
		} else {
			session.Title = "Untitled Session"
		}
	}

	return session, nil
}

func parseSessionTime(primary, fallback string) time.Time {
	if primary != "" {
		if t, ok := parseAnyTime(primary); ok {
			return t
		}
	}
	if fallback != "" {
		if t, ok := parseAnyTime(fallback); ok {
			return t
		}
	}
	return time.Time{}
}

func parseAnyTime(value string) (time.Time, bool) {
	layouts := []string{time.RFC3339Nano, time.RFC3339}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// DeriveLocalSessionStatus derives a normalized status from session metadata
func DeriveLocalSessionStatus(rawStatus string, lastActivity time.Time) string {
	// First normalize any explicit status
	normalized := strings.ToLower(strings.TrimSpace(rawStatus))

	// Map explicit statuses
	switch normalized {
	case "completed", "finished", "done", "merged", "closed":
		return "completed"
	case "running", "in progress", "active", "open":
		return "running"
	case "failed", "error", "cancelled", "canceled":
		return "failed"
	case "queued", "pending", "waiting":
		return "queued"
	}

	// If no explicit status or unknown status, derive from last activity
	if lastActivity.IsZero() {
		return "unknown"
	}

	// If last activity was more than 24 hours ago, consider stale
	if time.Since(lastActivity) > 24*time.Hour {
		return "completed"
	}

	// Otherwise assume still running
	return "running"
}

// truncateTitle truncates a string to a reasonable title length
func truncateTitle(s string) string {
	const maxLen = 100
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// FetchAllSessions fetches both agent-task and local Copilot sessions
func FetchAllSessions(repo string) ([]Session, error) {
	var allSessions []Session

	// Fetch agent tasks
	agentTasks, err := FetchAgentTasks(repo)
	if err == nil {
		for _, task := range agentTasks {
			allSessions = append(allSessions, FromAgentTask(task))
		}
	}
	// Don't fail completely if agent tasks fail - we still want local sessions

	// Fetch local sessions
	localSessions, err := FetchLocalSessions()
	if err == nil {
		// Filter local sessions by repo if specified
		for _, session := range localSessions {
			if repo == "" || session.Repository == repo {
				allSessions = append(allSessions, session)
			}
		}
	}
	// Don't fail completely if local sessions fail - we might still have agent tasks

	return allSessions, nil
}
