package data

import (
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
			name:    "active stale session",
			session: Session{Status: "running", UpdatedAt: time.Now().Add(-AttentionStaleThreshold - time.Minute)},
			want:    true,
		},
		{
			name:    "active fresh session",
			session: Session{Status: "running", UpdatedAt: time.Now().Add(-5 * time.Minute)},
			want:    false,
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
