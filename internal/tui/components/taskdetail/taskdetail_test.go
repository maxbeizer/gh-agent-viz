package taskdetail

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/maxbeizer/gh-agent-viz/internal/data"
)

func TestView_UsesFriendlyFallbacks(t *testing.T) {
	model := New(
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		func(string) string { return "•" },
	)
	model.SetTask(&data.Session{
		ID: "session-1",
	})

	view := model.View()
	if !strings.Contains(view, "Untitled Session") {
		t.Fatalf("expected untitled fallback, got: %s", view)
	}
	if !strings.Contains(view, "Repository: not available") {
		t.Fatalf("expected repository fallback, got: %s", view)
	}
	if !strings.Contains(view, "Branch:     not available") {
		t.Fatalf("expected branch fallback, got: %s", view)
	}
	if !strings.Contains(view, "Created:    not recorded") || !strings.Contains(view, "Updated:    not recorded") {
		t.Fatalf("expected timestamp fallback, got: %s", view)
	}
}

func TestView_ShowsRecordedTimestamps(t *testing.T) {
	model := New(
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		func(string) string { return "•" },
	)
	now := time.Date(2025, time.January, 2, 3, 4, 5, 0, time.UTC)
	model.SetTask(&data.Session{
		ID:        "session-2",
		Title:     "Named Session",
		CreatedAt: now,
		UpdatedAt: now,
	})

	view := model.View()
	if !strings.Contains(view, "Created:    2025-01-02 03:04:05") {
		t.Fatalf("expected formatted created timestamp, got: %s", view)
	}
	if !strings.Contains(view, "Updated:    2025-01-02 03:04:05") {
		t.Fatalf("expected formatted updated timestamp, got: %s", view)
	}
}

func TestView_ShowsTelemetry(t *testing.T) {
	model := New(
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		func(string) string { return "•" },
	)
	model.SetTask(&data.Session{
		ID:    "session-3",
		Title: "Session With Telemetry",
		Telemetry: &data.SessionTelemetry{
			Duration:          2*time.Hour + 30*time.Minute,
			ConversationTurns: 12,
			UserMessages:      5,
			AssistantMessages: 7,
		},
	})

	view := model.View()
	if !strings.Contains(view, "Session Stats") {
		t.Fatalf("expected Session Stats section header, got: %s", view)
	}
	if !strings.Contains(view, "⏱ Duration:") {
		t.Fatalf("expected duration with emoji prefix, got: %s", view)
	}
	if !strings.Contains(view, "2h 30m") {
		t.Fatalf("expected formatted duration, got: %s", view)
	}
	if !strings.Contains(view, "💬 Turns: 12") {
		t.Fatalf("expected conversation turn count with emoji, got: %s", view)
	}
	if !strings.Contains(view, "5 user · 7 assistant") {
		t.Fatalf("expected user/assistant breakdown with middle dot, got: %s", view)
	}
}

func TestView_NoTelemetryShowsNothing(t *testing.T) {
	model := New(
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		func(string) string { return "•" },
	)
	model.SetTask(&data.Session{
		ID:    "session-4",
		Title: "No Telemetry",
	})

	view := model.View()
	if strings.Contains(view, "Session Stats") {
		t.Fatalf("expected no Session Stats section without telemetry, got: %s", view)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{30 * time.Second, "30s"},
		{5 * time.Minute, "5m"},
		{90 * time.Minute, "1h 30m"},
		{2 * time.Hour, "2h"},
		{25 * time.Hour, "1d 1h"},
		{48 * time.Hour, "2d"},
	}

	for _, tt := range tests {
		got := formatDuration(tt.d)
		if got != tt.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestAttentionReason_NilSession(t *testing.T) {
	if got := attentionReason(nil); got != "" {
		t.Fatalf("expected empty for nil session, got %q", got)
	}
}

func TestAttentionReason_NeedsInput(t *testing.T) {
	s := &data.Session{Status: "needs-input"}
	got := attentionReason(s)
	if !strings.Contains(got, "waiting for your input") {
		t.Fatalf("expected needs-input reason, got %q", got)
	}
}

func TestAttentionReason_Failed(t *testing.T) {
	s := &data.Session{Status: "failed"}
	got := attentionReason(s)
	if !strings.Contains(got, "has failed") {
		t.Fatalf("expected failed reason, got %q", got)
	}
}

func TestAttentionReason_Idle_NoLongerAttention(t *testing.T) {
	s := &data.Session{
		Status:    "running",
		UpdatedAt: time.Now().Add(-30 * time.Minute),
	}
	got := attentionReason(s)
	if got != "" {
		t.Fatalf("idle sessions should not have attention reason, got %q", got)
	}
}

func TestAttentionReason_Stale_NoLongerAttention(t *testing.T) {
	s := &data.Session{
		Status:    "running",
		UpdatedAt: time.Now().Add(-5 * time.Hour),
	}
	got := attentionReason(s)
	if got != "" {
		t.Fatalf("stale sessions should not have attention reason, got %q", got)
	}
}

func TestAttentionReason_NoAttention(t *testing.T) {
	s := &data.Session{
		Status:    "running",
		UpdatedAt: time.Now().Add(-2 * time.Minute),
	}
	got := attentionReason(s)
	if got != "" {
		t.Fatalf("expected no attention reason for active session, got %q", got)
	}
}

func TestAttentionReason_CompletedSession(t *testing.T) {
	s := &data.Session{
		Status:    "completed",
		UpdatedAt: time.Now().Add(-2 * time.Hour),
	}
	got := attentionReason(s)
	if got != "" {
		t.Fatalf("expected no attention reason for completed session, got %q", got)
	}
}

func TestDetailValue(t *testing.T) {
	tests := []struct {
		value    string
		fallback string
		want     string
	}{
		{"hello", "default", "hello"},
		{"", "default", "default"},
		{"   ", "default", "default"},
		{"  trimmed  ", "default", "trimmed"},
	}

	for _, tt := range tests {
		got := detailValue(tt.value, tt.fallback)
		if got != tt.want {
			t.Errorf("detailValue(%q, %q) = %q, want %q", tt.value, tt.fallback, got, tt.want)
		}
	}
}

func TestDetailTitle(t *testing.T) {
	tests := []struct {
		title string
		want  string
	}{
		{"My Session", "My Session"},
		{"", "Untitled Session"},
		{"   ", "Untitled Session"},
	}

	for _, tt := range tests {
		got := detailTitle(tt.title)
		if got != tt.want {
			t.Errorf("detailTitle(%q) = %q, want %q", tt.title, got, tt.want)
		}
	}
}

func TestDetailTimestamp(t *testing.T) {
	tests := []struct {
		name string
		ts   time.Time
		want string
	}{
		{"zero time", time.Time{}, "not recorded"},
		{"specific time", time.Date(2025, 3, 14, 15, 9, 26, 0, time.UTC), "2025-03-14 15:09:26"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detailTimestamp(tt.ts)
			if got != tt.want {
				t.Errorf("detailTimestamp() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTimelineWidth(t *testing.T) {
	tests := []struct {
		available int
		want      int
	}{
		{0, 24},    // default
		{-1, 24},   // negative → default
		{30, 8},    // small → minimum 8
		{50, 22},   // mid
		{100, 48},  // large → capped at 48
		{200, 48},  // very large → capped at 48
	}

	for _, tt := range tests {
		got := timelineWidth(tt.available)
		if got != tt.want {
			t.Errorf("timelineWidth(%d) = %d, want %d", tt.available, got, tt.want)
		}
	}
}

func TestSectionDivider(t *testing.T) {
	// Zero or negative width should use default (40)
	result := sectionDivider(0)
	if result == "" {
		t.Fatal("expected non-empty divider")
	}

	// Positive width produces output
	result2 := sectionDivider(20)
	if result2 == "" {
		t.Fatal("expected non-empty divider for width 20")
	}
}

func TestFormatDuration_EdgeCases(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "0s"},
		{1 * time.Second, "1s"},
		{59 * time.Second, "59s"},
		{1 * time.Minute, "1m"},
		{1*time.Hour + 0*time.Minute, "1h"},
		{24 * time.Hour, "1d"},
		{49*time.Hour + 30*time.Minute, "2d 1h"},
	}

	for _, tt := range tests {
		got := formatDuration(tt.d)
		if got != tt.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestView_ShowsAttentionReason(t *testing.T) {
	model := New(
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		func(string) string { return "•" },
	)
	model.SetTask(&data.Session{
		ID:        "session-5",
		Status:    "needs-input",
		Title:     "Blocked Session",
	})
	view := model.View()
	if !strings.Contains(view, "waiting for your input") {
		t.Fatalf("expected attention reason in detail view, got: %s", view)
	}
}

func TestViewSplit_ShowsAttentionReason(t *testing.T) {
	model := New(
		lipgloss.NewStyle(),
		lipgloss.NewStyle(),
		func(string) string { return "•" },
	)
	model.SetTask(&data.Session{
		ID:     "session-6",
		Status: "failed",
		Title:  "Broken Session",
	})
	view := model.ViewSplit()
	if !strings.Contains(view, "has failed") {
		t.Fatalf("expected attention reason in split view, got: %s", view)
	}
}
