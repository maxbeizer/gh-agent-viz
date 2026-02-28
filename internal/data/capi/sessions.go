package capi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Session is the public type returned to callers, containing the fields
// we care about for display and status tracking.
type Session struct {
	ID              string
	Name            string
	State           string
	HeadRef         string
	Model           string
	RepoID          uint64
	ResourceType    string
	ResourceID      int64
	PremiumRequests float64
	CreatedAt       string
	LastUpdatedAt   string
	CompletedAt     string
	Error           *SessionError
}

// SessionError surfaces error details from failed sessions.
type SessionError struct {
	Code    string
	Message string
}

// ListSessions fetches the user's most recent agent sessions.
func (c *Client) ListSessions(ctx context.Context, limit int) ([]Session, error) {
	if limit <= 0 {
		limit = defaultPageSize
	}

	var all []apiSession
	seen := make(map[int64]struct{})

	for page := 1; ; page++ {
		u := baseCAPIURL + "/agents/sessions"
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
		if err != nil {
			return nil, err
		}
		q := req.URL.Query()
		q.Set("page_size", strconv.Itoa(defaultPageSize))
		q.Set("page_number", strconv.Itoa(page))
		q.Set("sort", "last_updated_at,desc")
		req.URL.RawQuery = q.Encode()

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("capi list sessions: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("capi list sessions: %s: %s", resp.Status, strings.TrimSpace(string(body)))
		}

		var data sessionsResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return nil, fmt.Errorf("capi decode sessions: %w", err)
		}

		for _, s := range data.Sessions {
			if _, dup := seen[s.ResourceID]; dup {
				continue
			}
			if s.ResourceID != 0 {
				seen[s.ResourceID] = struct{}{}
			}
			all = append(all, s)
			if len(all) >= limit {
				break
			}
		}

		if len(data.Sessions) < defaultPageSize || len(all) >= limit {
			break
		}
	}

	if len(all) > limit {
		all = all[:limit]
	}

	return toSessions(all), nil
}

// GetSession fetches a single session by ID.
func (c *Client) GetSession(ctx context.Context, id string) (*Session, error) {
	if id == "" {
		return nil, fmt.Errorf("session ID is required")
	}

	u := fmt.Sprintf("%s/agents/sessions/%s", baseCAPIURL, url.PathEscape(id))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("capi get session: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("capi get session: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var raw apiSession
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("capi decode session: %w", err)
	}

	result := toSession(raw)
	return &result, nil
}

// GetSessionLogs fetches the event log for a session.
func (c *Client) GetSessionLogs(ctx context.Context, id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("session ID is required")
	}

	u := fmt.Sprintf("%s/agents/sessions/%s/logs", baseCAPIURL, url.PathEscape(id))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("capi get session logs: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("session not found: %s", id)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("capi get session logs: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("capi read session logs: %w", err)
	}
	return strings.TrimSpace(string(body)), nil
}

func toSessions(raw []apiSession) []Session {
	out := make([]Session, len(raw))
	for i, s := range raw {
		out[i] = toSession(s)
	}
	return out
}

func toSession(s apiSession) Session {
	result := Session{
		ID:              s.ID,
		Name:            s.Name,
		State:           s.State,
		HeadRef:         s.HeadRef,
		Model:           s.Model,
		RepoID:          s.RepoID,
		ResourceType:    s.ResourceType,
		ResourceID:      s.ResourceID,
		PremiumRequests: s.PremiumRequests,
		CreatedAt:       formatTime(s.CreatedAt),
		LastUpdatedAt:   formatTime(s.LastUpdatedAt),
		CompletedAt:     formatTime(s.CompletedAt),
	}
	if s.Error != nil {
		result.Error = &SessionError{Code: s.Error.Code, Message: s.Error.Message}
	}
	return result
}

func formatTime(t interface{ IsZero() bool; Format(string) string }) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02T15:04:05Z07:00")
}
