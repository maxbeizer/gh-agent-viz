package toast

import (
	"strings"
	"testing"
	"time"
)

func TestNew_DefaultValues(t *testing.T) {
	m := New()
	if m.HasToasts() {
		t.Fatal("new model should have no toasts")
	}
	if m.Count() != 0 {
		t.Fatalf("expected 0 toasts, got %d", m.Count())
	}
	if m.View() != "" {
		t.Fatalf("expected empty view, got %q", m.View())
	}
}

func TestPush_AddsToast(t *testing.T) {
	m := New()
	m.Push("🟢", "Fix auth bug", "running → completed")

	if !m.HasToasts() {
		t.Fatal("expected toasts after push")
	}
	if m.Count() != 1 {
		t.Fatalf("expected 1 toast, got %d", m.Count())
	}
}

func TestPush_MultipleToasts(t *testing.T) {
	m := New()
	m.Push("🟢", "Task 1", "running → completed")
	m.Push("❌", "Task 2", "running → failed")

	if m.Count() != 2 {
		t.Fatalf("expected 2 toasts, got %d", m.Count())
	}
}

func TestPush_MaxToastEviction(t *testing.T) {
	m := New()
	m.Push("🟢", "Task 1", "a → b")
	m.Push("🟢", "Task 2", "a → b")
	m.Push("🟢", "Task 3", "a → b")
	m.Push("🟢", "Task 4", "a → b") // should evict Task 1

	if m.Count() != 3 {
		t.Fatalf("expected 3 toasts (max), got %d", m.Count())
	}

	view := m.View()
	if strings.Contains(view, "Task 1") {
		t.Fatal("expected Task 1 to be evicted")
	}
	if !strings.Contains(view, "Task 4") {
		t.Fatal("expected Task 4 to be present")
	}
}

func TestTick_RemovesExpiredToasts(t *testing.T) {
	m := New()
	m.ttl = 50 * time.Millisecond

	m.Push("🟢", "Expiring", "a → b")
	if m.Count() != 1 {
		t.Fatal("expected 1 toast before expiry")
	}

	time.Sleep(60 * time.Millisecond)
	m.Tick()

	if m.Count() != 0 {
		t.Fatalf("expected 0 toasts after expiry, got %d", m.Count())
	}
	if m.HasToasts() {
		t.Fatal("expected HasToasts to be false after expiry")
	}
}

func TestTick_KeepsUnexpiredToasts(t *testing.T) {
	m := New()
	m.ttl = 1 * time.Second

	m.Push("🟢", "Still alive", "a → b")
	m.Tick()

	if m.Count() != 1 {
		t.Fatalf("expected 1 toast still alive, got %d", m.Count())
	}
}

func TestTick_MixedExpiry(t *testing.T) {
	m := New()
	m.ttl = 50 * time.Millisecond

	m.Push("🟢", "Old toast", "a → b")
	time.Sleep(60 * time.Millisecond)
	m.Push("❌", "New toast", "c → d")

	m.Tick()

	if m.Count() != 1 {
		t.Fatalf("expected 1 toast (only new one), got %d", m.Count())
	}

	view := m.View()
	if strings.Contains(view, "Old toast") {
		t.Fatal("expected old toast to be expired")
	}
	if !strings.Contains(view, "New toast") {
		t.Fatal("expected new toast to be present")
	}
}

func TestView_EmptyState(t *testing.T) {
	m := New()
	if m.View() != "" {
		t.Fatalf("expected empty string for no toasts, got %q", m.View())
	}
}

func TestView_ContainsToastContent(t *testing.T) {
	m := New()
	m.Push("🟢", "Fix auth bug", "running → completed")

	view := m.View()
	if !strings.Contains(view, "🟢") {
		t.Fatal("expected icon in view")
	}
	if !strings.Contains(view, "Fix auth bug") {
		t.Fatal("expected title in view")
	}
	if !strings.Contains(view, "running → completed") {
		t.Fatal("expected message in view")
	}
}

func TestView_MultipleToastsStacked(t *testing.T) {
	m := New()
	m.Push("🟢", "Task A", "running → completed")
	m.Push("❌", "Task B", "running → failed")

	view := m.View()
	if !strings.Contains(view, "Task A") {
		t.Fatal("expected Task A in stacked view")
	}
	if !strings.Contains(view, "Task B") {
		t.Fatal("expected Task B in stacked view")
	}
}

func TestPush_TruncatesLongTitle(t *testing.T) {
	m := New()
	m.Push("🟢", "This is a very long session title that should be truncated", "a → b")

	view := m.View()
	if strings.Contains(view, "truncated") {
		t.Fatal("expected long title to be truncated")
	}
	if !strings.Contains(view, "…") {
		t.Fatal("expected ellipsis in truncated title")
	}
}

func TestSetWidth(t *testing.T) {
	m := New()
	m.SetWidth(50)
	m.Push("🟢", "Test", "a → b")

	// Should not panic
	view := m.View()
	if view == "" {
		t.Fatal("expected non-empty view after SetWidth")
	}
}
