package help

import "testing"

func TestToggle(t *testing.T) {
	m := New()
	if m.Visible() {
		t.Fatal("expected help to start hidden")
	}
	m.Toggle()
	if !m.Visible() {
		t.Fatal("expected help to be visible after toggle")
	}
	m.Toggle()
	if m.Visible() {
		t.Fatal("expected help to be hidden after second toggle")
	}
}

func TestView_VisibleReturnsContent(t *testing.T) {
	m := New()
	m.SetSize(120, 40)
	m.Toggle()
	v := m.View()
	if v == "" {
		t.Fatal("expected non-empty view when visible")
	}
	if len(v) < 50 {
		t.Fatal("expected substantial content in help view")
	}
}

func TestView_HiddenReturnsEmpty(t *testing.T) {
	m := New()
	m.SetSize(120, 40)
	v := m.View()
	if v != "" {
		t.Fatalf("expected empty view when hidden, got %q", v)
	}
}
