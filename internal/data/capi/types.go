package capi

import "time"

// apiSession is the raw JSON shape returned by the Copilot API.
type apiSession struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	UserID           int64     `json:"user_id"`
	AgentID          int64     `json:"agent_id"`
	State            string    `json:"state"`
	OwnerID          uint64    `json:"owner_id"`
	RepoID           uint64    `json:"repo_id"`
	ResourceType     string    `json:"resource_type"`
	ResourceID       int64     `json:"resource_id"`
	ResourceGlobalID string    `json:"resource_global_id"`
	LastUpdatedAt    time.Time `json:"last_updated_at,omitempty"`
	CreatedAt        time.Time `json:"created_at,omitempty"`
	CompletedAt      time.Time `json:"completed_at,omitempty"`
	EventURL         string    `json:"event_url"`
	EventType        string    `json:"event_type"`
	PremiumRequests  float64   `json:"premium_requests"`
	HeadRef          string    `json:"head_ref"`
	Model            string    `json:"model"`
	AgentTaskID      string    `json:"agent_task_id"`
	AgentType        string    `json:"agent_type"`
	WorkflowRunID    uint64    `json:"workflow_run_id,omitempty"`
	Error            *apiError `json:"error,omitempty"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// sessionsResponse wraps the paginated list endpoint response.
type sessionsResponse struct {
	Sessions []apiSession `json:"sessions"`
}
