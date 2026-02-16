package tui

import (
	"testing"
)

func TestStatusIcon_Running(t *testing.T) {
	icon := StatusIcon("running")
	expected := "🟢"
	if icon != expected {
		t.Errorf("expected icon '%s' for running status, got '%s'", expected, icon)
	}
}

func TestStatusIcon_Queued(t *testing.T) {
	icon := StatusIcon("queued")
	expected := "🟡"
	if icon != expected {
		t.Errorf("expected icon '%s' for queued status, got '%s'", expected, icon)
	}
}

func TestStatusIcon_NeedsInput(t *testing.T) {
	icon := StatusIcon("needs-input")
	expected := "🧑"
	if icon != expected {
		t.Errorf("expected icon '%s' for needs-input status, got '%s'", expected, icon)
	}
}

func TestStatusIcon_Completed(t *testing.T) {
	icon := StatusIcon("completed")
	expected := "✅"
	if icon != expected {
		t.Errorf("expected icon '%s' for completed status, got '%s'", expected, icon)
	}
}

func TestStatusIcon_Failed(t *testing.T) {
	icon := StatusIcon("failed")
	expected := "❌"
	if icon != expected {
		t.Errorf("expected icon '%s' for failed status, got '%s'", expected, icon)
	}
}

func TestStatusIcon_Unknown(t *testing.T) {
	icon := StatusIcon("unknown")
	expected := "⚪"
	if icon != expected {
		t.Errorf("expected icon '%s' for unknown status, got '%s'", expected, icon)
	}
}

func TestStatusIcon_EmptyString(t *testing.T) {
	icon := StatusIcon("")
	expected := "⚪"
	if icon != expected {
		t.Errorf("expected default icon '%s' for empty status, got '%s'", expected, icon)
	}
}

func TestStatusIcon_InvalidStatus(t *testing.T) {
	icon := StatusIcon("not-a-valid-status")
	expected := "⚪"
	if icon != expected {
		t.Errorf("expected default icon '%s' for invalid status, got '%s'", expected, icon)
	}
}

func TestNewTheme(t *testing.T) {
	theme := NewTheme()
	if theme == nil {
		t.Fatal("expected non-nil theme")
	}

	// Verify the theme was successfully created
	// lipgloss.Style fields contain functions and cannot be compared directly,
	// so we just verify the function completed successfully and returned a theme
}

func TestAnimatedStatusIcon_Running(t *testing.T) {
	// Frame 0 should return first braille frame
	icon := AnimatedStatusIcon("running", 0)
	if icon != "⠋" {
		t.Errorf("expected ⠋ for running frame 0, got %q", icon)
	}
	// Frame 1 should return second braille frame
	icon = AnimatedStatusIcon("running", 1)
	if icon != "⠙" {
		t.Errorf("expected ⠙ for running frame 1, got %q", icon)
	}
}

func TestAnimatedStatusIcon_Queued(t *testing.T) {
	icon := AnimatedStatusIcon("queued", 0)
	if icon != "⠿" {
		t.Errorf("expected ⠿ for queued frame 0, got %q", icon)
	}
	icon = AnimatedStatusIcon("queued", 1)
	if icon != "⠷" {
		t.Errorf("expected ⠷ for queued frame 1, got %q", icon)
	}
}

func TestAnimatedStatusIcon_WrapsFrames(t *testing.T) {
	// Frame 10 should wrap to frame 0 for running (10 frames)
	icon := AnimatedStatusIcon("running", 10)
	if icon != "⠋" {
		t.Errorf("expected ⠋ for running frame 10 (wrap), got %q", icon)
	}
	// Frame 6 should wrap to frame 0 for queued (6 frames)
	icon = AnimatedStatusIcon("queued", 6)
	if icon != "⠿" {
		t.Errorf("expected ⠿ for queued frame 6 (wrap), got %q", icon)
	}
}

func TestAnimatedStatusIcon_StaticForOtherStatuses(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"completed", "✅"},
		{"failed", "❌"},
		{"needs-input", "🧑"},
		{"unknown", "⚪"},
	}
	for _, tt := range tests {
		icon := AnimatedStatusIcon(tt.status, 5)
		if icon != tt.expected {
			t.Errorf("expected %q for %s, got %q", tt.expected, tt.status, icon)
		}
	}
}
