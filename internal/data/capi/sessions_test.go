package capi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestListSessions(t *testing.T) {
	sessions := []apiSession{
		{
			ID:            "sess-1",
			Name:          "Fix bug",
			State:         "completed",
			ResourceType:  "pull",
			ResourceID:    100,
			HeadRef:       "copilot/fix-bug",
			Model:         "claude-sonnet-4",
			CreatedAt:     time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
			LastUpdatedAt: time.Date(2026, 1, 1, 13, 0, 0, 0, time.UTC),
		},
		{
			ID:            "sess-2",
			Name:          "Add feature",
			State:         "running",
			ResourceType:  "pull",
			ResourceID:    101,
			HeadRef:       "copilot/add-feature",
			PremiumRequests: 42.5,
			CreatedAt:     time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC),
			LastUpdatedAt: time.Date(2026, 1, 2, 11, 0, 0, 0, time.UTC),
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/agents/sessions" {
			http.NotFound(w, r)
			return
		}
		// Verify expected headers
		if got := r.Header.Get("Authorization"); got != "Bearer gho_test_token" {
			t.Errorf("Authorization = %q, want Bearer gho_test_token", got)
		}
		if got := r.Header.Get("Copilot-Integration-Id"); got != integrationID {
			t.Errorf("Copilot-Integration-Id = %q, want %q", got, integrationID)
		}
		if got := r.Header.Get("X-GitHub-Api-Version"); got != apiVersion {
			t.Errorf("X-GitHub-Api-Version = %q, want %q", got, apiVersion)
		}

		resp := sessionsResponse{Sessions: sessions}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	client := newTestClient(srv, "gho_test_token")
	result, err := client.ListSessions(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListSessions() error = %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("ListSessions() returned %d sessions, want 2", len(result))
	}

	if result[0].ID != "sess-1" {
		t.Errorf("result[0].ID = %q, want %q", result[0].ID, "sess-1")
	}
	if result[0].State != "completed" {
		t.Errorf("result[0].State = %q, want %q", result[0].State, "completed")
	}
	if result[1].Name != "Add feature" {
		t.Errorf("result[1].Name = %q, want %q", result[1].Name, "Add feature")
	}
	if result[1].PremiumRequests != 42.5 {
		t.Errorf("result[1].PremiumRequests = %f, want 42.5", result[1].PremiumRequests)
	}
}

func TestListSessionsDeduplicates(t *testing.T) {
	sessions := []apiSession{
		{ID: "sess-1", ResourceID: 100, State: "running"},
		{ID: "sess-2", ResourceID: 100, State: "completed"}, // duplicate resource
		{ID: "sess-3", ResourceID: 200, State: "queued"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(sessionsResponse{Sessions: sessions})
	}))
	defer srv.Close()

	client := newTestClient(srv, "gho_test_token")
	result, err := client.ListSessions(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListSessions() error = %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("got %d sessions, want 2 (deduplicated)", len(result))
	}
	if result[0].ID != "sess-1" || result[1].ID != "sess-3" {
		t.Errorf("unexpected session IDs: %s, %s", result[0].ID, result[1].ID)
	}
}

func TestGetSession(t *testing.T) {
	sess := apiSession{
		ID:    "sess-abc",
		Name:  "Test task",
		State: "running",
		Error: &apiError{Code: "timeout", Message: "timed out"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/agents/sessions/sess-abc" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(sess)
	}))
	defer srv.Close()

	client := newTestClient(srv, "gho_test_token")
	result, err := client.GetSession(context.Background(), "sess-abc")
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if result.Name != "Test task" {
		t.Errorf("Name = %q, want %q", result.Name, "Test task")
	}
	if result.Error == nil || result.Error.Code != "timeout" {
		t.Errorf("Error = %v, want timeout error", result.Error)
	}
}

func TestGetSessionNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	client := newTestClient(srv, "gho_test_token")
	_, err := client.GetSession(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("GetSession() expected error for not found")
	}
}

func TestGetSessionLogs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/agents/sessions/sess-log/logs" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte("line 1\nline 2\nline 3"))
	}))
	defer srv.Close()

	client := newTestClient(srv, "gho_test_token")
	logs, err := client.GetSessionLogs(context.Background(), "sess-log")
	if err != nil {
		t.Fatalf("GetSessionLogs() error = %v", err)
	}
	if logs != "line 1\nline 2\nline 3" {
		t.Errorf("logs = %q, want 3 lines", logs)
	}
}

func TestListSessionsHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer srv.Close()

	client := newTestClient(srv, "gho_test_token")
	_, err := client.ListSessions(context.Background(), 10)
	if err == nil {
		t.Fatal("ListSessions() expected error on 500")
	}
}

// newTestClient creates a Client pointing at a test server.
// This bypasses resolveToken() so tests don't need gh auth.
func newTestClient(srv *httptest.Server, token string) *Client {
	// Override the base URL by using a custom transport that rewrites the host
	transport := &testTransport{
		base:    srv.Client().Transport,
		token:   token,
		srvURL:  srv.URL,
	}
	return &Client{
		httpClient: &http.Client{Transport: transport},
	}
}

// testTransport rewrites requests to point at the test server while
// still adding the auth headers that capiTransport would add.
type testTransport struct {
	base   http.RoundTripper
	token  string
	srvURL string
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	req.Header.Set("Copilot-Integration-Id", integrationID)
	req.Header.Set("X-GitHub-Api-Version", apiVersion)

	// Rewrite URL to point at test server
	testURL := t.srvURL + req.URL.Path
	if req.URL.RawQuery != "" {
		testURL += "?" + req.URL.RawQuery
	}
	newReq, _ := http.NewRequestWithContext(req.Context(), req.Method, testURL, req.Body)
	newReq.Header = req.Header
	return t.base.RoundTrip(newReq)
}
