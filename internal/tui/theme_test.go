package tui

import (
	"strings"
	"testing"
)

func TestStatusIcon_Running(t *testing.T) {
	icon := StatusIcon("running")
	expected := "●"
	if icon != expected {
		t.Errorf("expected icon '%s' for running status, got '%s'", expected, icon)
	}
}

func TestStatusIcon_Queued(t *testing.T) {
	icon := StatusIcon("queued")
	expected := "○"
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
	if theme.ThemeName() != "default" {
		t.Errorf("expected theme name 'default', got %q", theme.ThemeName())
	}
}

func TestNewThemeFromConfig_Presets(t *testing.T) {
	tests := []struct {
		input    string
		wantName string
	}{
		{"default", "default"},
		{"catppuccin-mocha", "catppuccin-mocha"},
		{"dracula", "dracula"},
		{"tokyo-night", "tokyo-night"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			theme := NewThemeFromConfig(tt.input)
			if theme == nil {
				t.Fatal("expected non-nil theme")
			}
			if theme.ThemeName() != tt.wantName {
				t.Errorf("expected theme name %q, got %q", tt.wantName, theme.ThemeName())
			}
		})
	}
}

func TestNewThemeFromConfig_EmptyFallsBackToDefault(t *testing.T) {
	theme := NewThemeFromConfig("")
	if theme == nil {
		t.Fatal("expected non-nil theme")
	}
	if theme.ThemeName() != "default" {
		t.Errorf("expected fallback theme name 'default', got %q", theme.ThemeName())
	}
}

func TestNewThemeFromConfig_UnknownFallsBackToDefault(t *testing.T) {
	theme := NewThemeFromConfig("nonexistent-theme")
	if theme == nil {
		t.Fatal("expected non-nil theme")
	}
	if theme.ThemeName() != "default" {
		t.Errorf("expected fallback theme name 'default', got %q", theme.ThemeName())
	}
}

func TestAnimatedStatusIcon_Running(t *testing.T) {
	// Frame 0 should contain the first pulsing dot character
	icon := AnimatedStatusIcon("running", 0)
	if !strings.Contains(icon, "●") {
		t.Errorf("expected ● in running frame 0, got %q", icon)
	}
	// Frame 2 should contain the ring character
	icon = AnimatedStatusIcon("running", 2)
	if !strings.Contains(icon, "◉") {
		t.Errorf("expected ◉ in running frame 2, got %q", icon)
	}
}

func TestAnimatedStatusIcon_Queued(t *testing.T) {
	icon := AnimatedStatusIcon("queued", 0)
	if !strings.Contains(icon, "◌") {
		t.Errorf("expected ◌ in queued frame 0, got %q", icon)
	}
	icon = AnimatedStatusIcon("queued", 1)
	if !strings.Contains(icon, "○") {
		t.Errorf("expected ○ in queued frame 1, got %q", icon)
	}
}

func TestAnimatedStatusIcon_WrapsFrames(t *testing.T) {
	// Frame 6 should wrap to frame 0 for running (6 frames)
	icon := AnimatedStatusIcon("running", 6)
	if !strings.Contains(icon, "●") {
		t.Errorf("expected ● for running frame 6 (wrap), got %q", icon)
	}
	// Frame 4 should wrap to frame 0 for queued (4 frames)
	icon = AnimatedStatusIcon("queued", 4)
	if !strings.Contains(icon, "◌") {
		t.Errorf("expected ◌ for queued frame 4 (wrap), got %q", icon)
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
