package data

import (
	"fmt"
	"testing"
	"time"
)

func TestSessionNeedsAttention(t *testing.T) {
	cases := []struct {
		name    string
		session Session
		want    bool
	}{
		{
			name:    "needs-input status",
			session: Session{Status: "needs-input"},
			want:    true,
		},
		{
			name:    "failed status",
			session: Session{Status: "failed"},
			want:    true,
		},
		{
			name:    "active idle session is not attention",
			session: Session{Status: "running", UpdatedAt: time.Now().Add(-AttentionStaleThreshold - time.Minute)},
			want:    false,
		},
		{
			name:    "active fresh session",
			session: Session{Status: "running", UpdatedAt: time.Now().Add(-5 * time.Minute)},
			want:    false,
		},
		{
			name:    "active session idle 2h not attention",
			session: Session{Status: "running", UpdatedAt: time.Now().Add(-2 * time.Hour)},
			want:    false,
		},
		{
			name:    "active session idle 5h not attention",
			session: Session{Status: "running", UpdatedAt: time.Now().Add(-5 * time.Hour)},
			want:    false,
		},
		{
			name:    "needs-input idle 5h always needs attention",
			session: Session{Status: "needs-input", UpdatedAt: time.Now().Add(-5 * time.Hour)},
			want:    true,
		},
		{
			name:    "failed idle 5h always needs attention",
			session: Session{Status: "failed", UpdatedAt: time.Now().Add(-5 * time.Hour)},
			want:    true,
		},
		{
			name:    "completed session",
			session: Session{Status: "completed", UpdatedAt: time.Now().Add(-2 * time.Hour)},
			want:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := SessionNeedsAttention(tc.session)
			if got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestStatusIsActive_CaseInsensitive(t *testing.T) {
	for _, status := range []string{"Running", "QUEUED", "In Progress", "active", "OPEN"} {
		if !StatusIsActive(status) {
			t.Fatalf("expected %q to be active", status)
		}
	}
}

func TestStatusIsActive_Inactive(t *testing.T) {
	for _, status := range []string{"completed", "failed", "needs-input", "", "unknown"} {
		if StatusIsActive(status) {
			t.Fatalf("expected %q to NOT be active", status)
		}
	}
}

func TestIsDefaultBranch(t *testing.T) {
	tests := []struct {
		branch string
		want   bool
	}{
		{"main", true},
		{"master", true},
		{"Main", true},
		{"MASTER", true},
		{"", true},
		{"  main  ", true},
		{"feature/foo", false},
		{"develop", false},
		{"main-backup", false},
		{"my-master", false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("branch=%q", tt.branch), func(t *testing.T) {
			got := IsDefaultBranch(tt.branch)
			if got != tt.want {
				t.Errorf("IsDefaultBranch(%q) = %v, want %v", tt.branch, got, tt.want)
			}
		})
	}
}

func TestFormatTokenCount(t *testing.T) {
	tests := []struct {
		n    int64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{999, "999"},
		{1000, "1.0K"},
		{1500, "1.5K"},
		{11700, "11.7K"},
		{999999, "1000.0K"},
		{1000000, "1.0M"},
		{2700000, "2.7M"},
		{10000000, "10.0M"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("n=%d", tt.n), func(t *testing.T) {
			got := FormatTokenCount(tt.n)
			if got != tt.want {
				t.Errorf("FormatTokenCount(%d) = %q, want %q", tt.n, got, tt.want)
			}
		})
	}
}

func TestFromAgentTask_ToAgentTask_Roundtrip(t *testing.T) {
	original := AgentTask{
		ID:         "roundtrip-123",
		Status:     "running",
		Title:      "Roundtrip Test",
		Repository: "owner/repo",
		Branch:     "feature/x",
		PRURL:      "https://github.com/owner/repo/pull/42",
		PRNumber:   42,
		CreatedAt:  time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2025, 6, 15, 11, 0, 0, 0, time.UTC),
	}

	session := FromAgentTask(original)
	if session.Source != SourceAgentTask {
		t.Errorf("expected source %q, got %q", SourceAgentTask, session.Source)
	}

	roundtripped := session.ToAgentTask()

	if roundtripped.ID != original.ID {
		t.Errorf("ID mismatch: %q != %q", roundtripped.ID, original.ID)
	}
	if roundtripped.Status != original.Status {
		t.Errorf("Status mismatch: %q != %q", roundtripped.Status, original.Status)
	}
	if roundtripped.Title != original.Title {
		t.Errorf("Title mismatch: %q != %q", roundtripped.Title, original.Title)
	}
	if roundtripped.Repository != original.Repository {
		t.Errorf("Repository mismatch: %q != %q", roundtripped.Repository, original.Repository)
	}
	if roundtripped.Branch != original.Branch {
		t.Errorf("Branch mismatch: %q != %q", roundtripped.Branch, original.Branch)
	}
	if roundtripped.PRURL != original.PRURL {
		t.Errorf("PRURL mismatch: %q != %q", roundtripped.PRURL, original.PRURL)
	}
	if roundtripped.PRNumber != original.PRNumber {
		t.Errorf("PRNumber mismatch: %d != %d", roundtripped.PRNumber, original.PRNumber)
	}
	if !roundtripped.CreatedAt.Equal(original.CreatedAt) {
		t.Errorf("CreatedAt mismatch: %v != %v", roundtripped.CreatedAt, original.CreatedAt)
	}
	if !roundtripped.UpdatedAt.Equal(original.UpdatedAt) {
		t.Errorf("UpdatedAt mismatch: %v != %v", roundtripped.UpdatedAt, original.UpdatedAt)
	}
}

func TestSessionAttentionLevel(t *testing.T) {
	cases := []struct {
		name    string
		session Session
		want    AttentionLevel
	}{
		{
			name:    "needs-input is urgent",
			session: Session{Status: "needs-input"},
			want:    AttentionUrgent,
		},
		{
			name:    "failed is urgent",
			session: Session{Status: "failed"},
			want:    AttentionUrgent,
		},
		{
			name:    "running idle 5h is warning",
			session: Session{Status: "running", UpdatedAt: time.Now().Add(-5 * time.Hour)},
			want:    AttentionWarning,
		},
		{
			name:    "running idle 3h is none (below 4h threshold)",
			session: Session{Status: "running", UpdatedAt: time.Now().Add(-3 * time.Hour)},
			want:    AttentionNone,
		},
		{
			name:    "running idle 5min is none",
			session: Session{Status: "running", UpdatedAt: time.Now().Add(-5 * time.Minute)},
			want:    AttentionNone,
		},
		{
			name:    "queued 45min is warning",
			session: Session{Status: "queued", CreatedAt: time.Now().Add(-45 * time.Minute)},
			want:    AttentionWarning,
		},
		{
			name:    "queued 5min is none",
			session: Session{Status: "queued", CreatedAt: time.Now().Add(-5 * time.Minute)},
			want:    AttentionNone,
		},
		{
			name:    "completed is none",
			session: Session{Status: "completed"},
			want:    AttentionNone,
		},
		{
			name:    "active idle exactly 4h is warning",
			session: Session{Status: "active", UpdatedAt: time.Now().Add(-IdleWarningThreshold)},
			want:    AttentionWarning,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := SessionAttentionLevel(tc.session)
			if got != tc.want {
				t.Fatalf("SessionAttentionLevel: expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestSessionNeedsAnyAttention(t *testing.T) {
	// Urgent → true
	if !SessionNeedsAnyAttention(Session{Status: "failed"}) {
		t.Fatal("failed should need any attention")
	}
	// Warning → true
	if !SessionNeedsAnyAttention(Session{Status: "running", UpdatedAt: time.Now().Add(-5 * time.Hour)}) {
		t.Fatal("idle running should need any attention")
	}
	// None → false
	if SessionNeedsAnyAttention(Session{Status: "completed"}) {
		t.Fatal("completed should not need any attention")
	}
}

func TestAttentionLevelString(t *testing.T) {
	if AttentionUrgent.String() != "urgent" {
		t.Fatal("expected 'urgent'")
	}
	if AttentionWarning.String() != "warning" {
		t.Fatal("expected 'warning'")
	}
	if AttentionNone.String() != "none" {
		t.Fatal("expected 'none'")
	}
}
