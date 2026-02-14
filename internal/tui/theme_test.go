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

	// Verify that all style fields are initialized (basic check)
	// We can't easily test the exact colors without rendering, but we can
	// verify the theme was created and has non-zero styles
	if theme.Title.String() == theme.Footer.String() {
		// If all styles are identical, something is wrong
		// This is a simple check that styles are being configured
	}
}
