package data

import (
	"bufio"
	"encoding/json"
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
	AwaitingUserInput   bool                     `yaml:"awaiting_user_input"`
	NeedsHumanInput     bool                     `yaml:"needs_human_input"`
	WaitingForUser      bool                     `yaml:"waiting_for_user"`
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

		// Check for events.jsonl to mark log availability
		eventsFile := filepath.Join(sessionDir, entry.Name(), "events.jsonl")
		if info, err := os.Stat(eventsFile); err == nil && info.Size() > 0 {
			session.HasLog = true
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
	needsInput := false

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
		} else if strings.HasPrefix(line, "awaiting_user_input:") || strings.HasPrefix(line, "needs_human_input:") || strings.HasPrefix(line, "waiting_for_user:") {
			value := strings.ToLower(strings.TrimSpace(strings.SplitN(line, ":", 2)[1]))
			if value == "true" {
				needsInput = true
			}
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
	if needsInput && isLocallyActiveStatus(session.Status) {
		session.Status = "needs-input"
	}

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
	if needsHumanInput(workspace) && isLocallyActiveStatus(session.Status) {
		session.Status = "needs-input"
	}

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

	// Derive telemetry from workspace metadata
	session.Telemetry = deriveSessionTelemetry(workspace, session.CreatedAt, session.UpdatedAt)

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
	case "needs-input", "needs input", "awaiting user input", "waiting for user", "input required":
		return "needs-input"
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

func needsHumanInput(workspace LocalSessionWorkspace) bool {
	if workspace.AwaitingUserInput || workspace.NeedsHumanInput || workspace.WaitingForUser {
		return true
	}
	if len(workspace.ConversationHistory) == 0 {
		return false
	}

	last := workspace.ConversationHistory[len(workspace.ConversationHistory)-1]
	role, _ := last["role"].(string)
	content, _ := last["content"].(string)
	if !strings.EqualFold(strings.TrimSpace(role), "assistant") {
		return false
	}
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return false
	}
	if strings.Contains(trimmed, "?") {
		return true
	}

	lower := strings.ToLower(trimmed)
	patterns := []string{
		"please choose",
		"which option",
		"what would you like",
		"can you confirm",
		"please provide",
		"pick one",
		"let me know",
	}
	for _, pattern := range patterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

func isLocallyActiveStatus(status string) bool {
	normalized := strings.ToLower(strings.TrimSpace(status))
	return normalized == "running" || normalized == "queued" || normalized == "needs-input"
}

// deriveSessionTelemetry computes usage metrics from workspace metadata
func deriveSessionTelemetry(workspace LocalSessionWorkspace, createdAt, updatedAt time.Time) *SessionTelemetry {
	telemetry := &SessionTelemetry{}

	// Duration
	if !createdAt.IsZero() && !updatedAt.IsZero() && updatedAt.After(createdAt) {
		telemetry.Duration = updatedAt.Sub(createdAt)
	}

	// Conversation metrics
	for _, entry := range workspace.ConversationHistory {
		role, _ := entry["role"].(string)
		switch strings.ToLower(strings.TrimSpace(role)) {
		case "user":
			telemetry.UserMessages++
		case "assistant":
			telemetry.AssistantMessages++
		}
	}
	telemetry.ConversationTurns = telemetry.UserMessages + telemetry.AssistantMessages

	return telemetry
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

// FetchLocalSessionLog reads events.jsonl for a local session and formats it
// as a human-readable conversation log.
func FetchLocalSessionLog(sessionID string) (string, error) {
	if sessionID == "" {
		return "", fmt.Errorf("session ID is required")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	eventsFile := filepath.Join(homeDir, ".copilot", "session-state", sessionID, "events.jsonl")
	f, err := os.Open(eventsFile)
	if err != nil {
		return "", fmt.Errorf("no event log found for this session")
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	// Allow large lines (some tool results can be big)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		var event struct {
			Type      string          `json:"type"`
			Timestamp string          `json:"timestamp"`
			Data      json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}

		line := formatEventLine(event.Type, event.Timestamp, event.Data)
		if line != "" {
			lines = append(lines, line)
		}
	}

	if len(lines) == 0 {
		return "No conversation events recorded for this session.", nil
	}

	header := "# Session Event Log\n\n"
	return header + strings.Join(lines, "\n"), nil
}

// formatEventLine renders a single event as a readable log line.
func formatEventLine(eventType, timestamp string, rawData json.RawMessage) string {
	ts := formatEventTimestamp(timestamp)

	switch eventType {
	case "session.start":
		return fmt.Sprintf("**%s** — 🚀 Session started\n", ts)

	case "user.message":
		var data struct {
			Content string `json:"content"`
		}
		if json.Unmarshal(rawData, &data) == nil && data.Content != "" {
			content := truncateLogContent(data.Content, 500)
			return fmt.Sprintf("**%s** — 👤 **User**\n\n%s\n", ts, content)
		}

	case "assistant.message":
		var data struct {
			Content string `json:"content"`
		}
		if json.Unmarshal(rawData, &data) == nil && data.Content != "" {
			content := truncateLogContent(data.Content, 500)
			return fmt.Sprintf("**%s** — 🤖 **Assistant**\n\n%s\n", ts, content)
		}

	case "tool.execution_start":
		var data struct {
			ToolName string `json:"toolName"`
		}
		if json.Unmarshal(rawData, &data) == nil && data.ToolName != "" {
			return fmt.Sprintf("`%s` 🔧 %s", ts, data.ToolName)
		}

	case "abort":
		return fmt.Sprintf("**%s** — ⛔ Aborted\n", ts)

	case "assistant.turn_start":
		return fmt.Sprintf("---\n**%s** — _Turn started_", ts)
	}

	return ""
}

func formatEventTimestamp(ts string) string {
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		t, err = time.Parse(time.RFC3339, ts)
		if err != nil {
			return ts
		}
	}
	return t.Format("15:04:05")
}

func truncateLogContent(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "\n\n_(truncated)_"
}

// SessionEvent represents a single parsed event from events.jsonl with full content.
type SessionEvent struct {
	Type      string // e.g. "user.message", "assistant.message", "tool.execution_start"
	Timestamp string // RFC3339 timestamp
	Role      string // "user" or "assistant" for messages
	Content   string // full content (not truncated)
	ToolName  string // for tool.execution_start events
}

// FetchSessionEvents reads events.jsonl for a local session and returns
// structured events without content truncation.
func FetchSessionEvents(sessionID string) ([]SessionEvent, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	eventsFile := filepath.Join(homeDir, ".copilot", "session-state", sessionID, "events.jsonl")
	f, err := os.Open(eventsFile)
	if err != nil {
		return nil, fmt.Errorf("no event log found for this session")
	}
	defer f.Close()

	var events []SessionEvent
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		var raw struct {
			Type      string          `json:"type"`
			Timestamp string          `json:"timestamp"`
			Data      json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &raw); err != nil {
			continue
		}

		ev := SessionEvent{
			Type:      raw.Type,
			Timestamp: raw.Timestamp,
		}

		switch raw.Type {
		case "user.message":
			var d struct {
				Content string `json:"content"`
			}
			if json.Unmarshal(raw.Data, &d) == nil {
				ev.Role = "user"
				ev.Content = d.Content
			}
		case "assistant.message":
			var d struct {
				Content string `json:"content"`
			}
			if json.Unmarshal(raw.Data, &d) == nil {
				ev.Role = "assistant"
				ev.Content = d.Content
			}
		case "tool.execution_start":
			var d struct {
				ToolName string `json:"toolName"`
			}
			if json.Unmarshal(raw.Data, &d) == nil {
				ev.ToolName = d.ToolName
			}
		}

		events = append(events, ev)
	}

	return events, nil
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
