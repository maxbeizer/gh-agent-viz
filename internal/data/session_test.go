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
