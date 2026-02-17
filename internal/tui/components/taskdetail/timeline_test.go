package taskdetail

import (
	"strings"
	"testing"
	"time"

	"github.com/maxbeizer/gh-agent-viz/internal/data"
)

func withFrozenNow(t time.Time, fn func()) {
	orig := nowFunc
	nowFunc = func() time.Time { return t }
	defer func() { nowFunc = orig }()
	fn()
}

func TestRenderTimeline_NilSession(t *testing.T) {
	got := RenderTimeline(nil, 24)
	if got != "" {
		t.Errorf("expected empty string for nil session, got %q", got)
	}
}

func TestRenderTimeline_ZeroCreatedAt(t *testing.T) {
	s := &data.Session{Status: "running"}
	got := RenderTimeline(s, 24)
	if got != "" {
		t.Errorf("expected empty string for zero CreatedAt, got %q", got)
	}
}

func TestRenderTimeline_ZeroWidth(t *testing.T) {
	s := &data.Session{
		Status:    "running",
		CreatedAt: time.Now().Add(-time.Hour),
	}
	got := RenderTimeline(s, 0)
	if got != "" {
		t.Errorf("expected empty string for zero width, got %q", got)
	}
}

func TestRenderTimeline_RunningSession(t *testing.T) {
	now := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)
	s := &data.Session{
		Status:    "running",
		CreatedAt: now.Add(-2 * time.Hour),
		UpdatedAt: now.Add(-30 * time.Minute),
	}

	withFrozenNow(now, func() {
		got := RenderTimeline(s, 24)
		if got == "" {
			t.Fatal("expected non-empty timeline")
		}
		// Should contain block characters
		if !strings.ContainsRune(got, blockFull) {
			t.Errorf("expected █ in running session timeline, got %q", got)
		}
		if !strings.ContainsRune(got, blockIdle) {
			t.Errorf("expected ░ in timeline with idle period, got %q", got)
		}
		if !strings.Contains(got, "→ now") {
			t.Errorf("expected '→ now' label, got %q", got)
		}
		// Right edge should be █ for active session
		bar := strings.Split(got, "  ")[0]
		runes := []rune(bar)
		if runes[len(runes)-1] != blockFull {
			t.Errorf("expected right edge to be █ for running session, got %c", runes[len(runes)-1])
		}
	})
}

func TestRenderTimeline_CompletedSession(t *testing.T) {
	now := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)
	s := &data.Session{
		Status:    "completed",
		CreatedAt: now.Add(-4 * time.Hour),
		UpdatedAt: now.Add(-2 * time.Hour),
	}

	withFrozenNow(now, func() {
		got := RenderTimeline(s, 24)
		if got == "" {
			t.Fatal("expected non-empty timeline")
		}
		if !strings.ContainsRune(got, blockActive) {
			t.Errorf("expected ▓ in completed session timeline, got %q", got)
		}
		if !strings.ContainsRune(got, blockIdle) {
			t.Errorf("expected ░ in completed session with idle tail, got %q", got)
		}
		if !strings.Contains(got, "4h ago") {
			t.Errorf("expected '4h ago' label, got %q", got)
		}
	})
}

func TestRenderTimeline_VeryShortSession(t *testing.T) {
	now := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)
	s := &data.Session{
		Status:    "running",
		CreatedAt: now.Add(-10 * time.Second),
		UpdatedAt: now.Add(-5 * time.Second),
	}

	withFrozenNow(now, func() {
		got := RenderTimeline(s, 24)
		if got == "" {
			t.Fatal("expected non-empty timeline for short session")
		}
		if !strings.Contains(got, "<1m ago") {
			t.Errorf("expected '<1m ago' for short session, got %q", got)
		}
	})
}

func TestRenderTimeline_JustCreated(t *testing.T) {
	now := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)
	s := &data.Session{
		Status:    "running",
		CreatedAt: now,
	}

	withFrozenNow(now, func() {
		got := RenderTimeline(s, 16)
		if got == "" {
			t.Fatal("expected non-empty timeline for just-created session")
		}
		bar := strings.Split(got, "  ")[0]
		if len([]rune(bar)) != 16 {
			t.Errorf("expected bar width 16, got %d", len([]rune(bar)))
		}
	})
}

func TestRenderTimeline_NarrowWidth(t *testing.T) {
	now := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)
	s := &data.Session{
		Status:    "completed",
		CreatedAt: now.Add(-1 * time.Hour),
		UpdatedAt: now.Add(-30 * time.Minute),
	}

	withFrozenNow(now, func() {
		got := RenderTimeline(s, 4)
		if got == "" {
			t.Fatal("expected non-empty timeline for narrow width")
		}
		bar := strings.Split(got, "  ")[0]
		if len([]rune(bar)) != 4 {
			t.Errorf("expected bar width 4, got %d", len([]rune(bar)))
		}
	})
}

func TestRenderTimeline_WideWidth(t *testing.T) {
	now := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)
	s := &data.Session{
		Status:    "running",
		CreatedAt: now.Add(-3 * time.Hour),
		UpdatedAt: now.Add(-1 * time.Hour),
	}

	withFrozenNow(now, func() {
		got := RenderTimeline(s, 80)
		if got == "" {
			t.Fatal("expected non-empty timeline for wide width")
		}
		bar := strings.Split(got, "  ")[0]
		if len([]rune(bar)) != 80 {
			t.Errorf("expected bar width 80, got %d", len([]rune(bar)))
		}
	})
}

func TestRenderTimeline_OnlyCreatedNoUpdated(t *testing.T) {
	now := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)
	s := &data.Session{
		Status:    "queued",
		CreatedAt: now.Add(-30 * time.Minute),
	}

	withFrozenNow(now, func() {
		got := RenderTimeline(s, 24)
		if got == "" {
			t.Fatal("expected non-empty timeline for session with only CreatedAt")
		}
		// All chars except right edge should be idle since activeEnd == created
		bar := strings.Split(got, "  ")[0]
		runes := []rune(bar)
		// First char maps to created time (position 0 == created), should be active
		// since posTime == activeEnd == created
		if runes[0] != blockFull {
			t.Errorf("expected first char to be █ (at created time), got %c", runes[0])
		}
	})
}

func TestFormatRelative(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{30 * time.Second, "<1m ago"},
		{5 * time.Minute, "5m ago"},
		{90 * time.Minute, "1h ago"},
		{2 * time.Hour, "2h ago"},
		{25 * time.Hour, "1d ago"},
		{48 * time.Hour, "2d ago"},
	}

	for _, tt := range tests {
		got := formatRelative(tt.d)
		if got != tt.want {
			t.Errorf("formatRelative(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}
