package taskdetail

import (
	"strings"
	"testing"

	"github.com/maxbeizer/gh-agent-viz/internal/data"
)

func TestParseSessionDeps_NoRelationships(t *testing.T) {
	session := &data.Session{
		ID:         "s1",
		Title:      "Fix auth bug",
		Repository: "github/app",
		Branch:     "feature/auth-bug",
	}
	others := []data.Session{
		{ID: "s2", Title: "Update docs", Repository: "github/docs", Branch: "main"},
	}

	graph := ParseSessionDeps(session, others)
	rendered := RenderDepGraph(graph, 80)
	if rendered != "" {
		t.Fatalf("expected empty render for unrelated sessions, got: %q", rendered)
	}
}

func TestParseSessionDeps_NilSession(t *testing.T) {
	graph := ParseSessionDeps(nil, []data.Session{})
	rendered := RenderDepGraph(graph, 80)
	if rendered != "" {
		t.Fatalf("expected empty render for nil session, got: %q", rendered)
	}
}

func TestParseSessionDeps_SinglePRReference(t *testing.T) {
	session := &data.Session{
		ID:         "s1",
		PRNumber:   42,
		Repository: "github/app",
		Branch:     "feature/auth-bug",
	}
	others := []data.Session{
		{ID: "s2", PRNumber: 43, Repository: "github/app", Branch: "feature/auth-tests"},
	}

	graph := ParseSessionDeps(session, others)
	rendered := RenderDepGraph(graph, 80)
	if !strings.Contains(rendered, "PR #42") {
		t.Fatalf("expected PR #42 in render, got: %q", rendered)
	}
	if !strings.Contains(rendered, "PR #43") {
		t.Fatalf("expected PR #43 in render, got: %q", rendered)
	}
	if !strings.Contains(rendered, "──→") {
		t.Fatalf("expected arrow in render, got: %q", rendered)
	}
}

func TestParseSessionDeps_LinearChain(t *testing.T) {
	session := &data.Session{
		ID:         "s1",
		PRNumber:   42,
		Repository: "github/app",
		Branch:     "feature/auth-bug",
	}
	others := []data.Session{
		{ID: "s1", PRNumber: 42, Repository: "github/app", Branch: "feature/auth-bug"},
		{ID: "s2", PRNumber: 43, Repository: "github/app", Branch: "feature/auth-tests"},
		{ID: "s3", PRNumber: 44, Repository: "github/app", Branch: "feature/auth-docs"},
	}

	graph := ParseSessionDeps(session, others)
	rendered := RenderDepGraph(graph, 80)
	if !strings.Contains(rendered, "PR #42") {
		t.Fatalf("expected PR #42 in render, got: %q", rendered)
	}
	if !strings.Contains(rendered, "PR #43") {
		t.Fatalf("expected PR #43 in render, got: %q", rendered)
	}
	if !strings.Contains(rendered, "PR #44") {
		t.Fatalf("expected PR #44 in render, got: %q", rendered)
	}
}

func TestParseSessionDeps_FanOut(t *testing.T) {
	session := &data.Session{
		ID:         "s1",
		Title:      "Fix auth bug",
		PRNumber:   42,
		Repository: "github/app",
		Branch:     "feature/auth-bug",
	}
	others := []data.Session{
		{ID: "s2", Title: "Add auth tests", PRNumber: 43, Repository: "github/app", Branch: "feature/auth-tests"},
		{ID: "s3", Title: "Update docs", PRNumber: 44, Repository: "github/app", Branch: "feature/auth-docs"},
		{ID: "s4", Title: "Auth cleanup", PRNumber: 45, Repository: "github/app", Branch: "feature/auth-cleanup"},
	}

	graph := ParseSessionDeps(session, others)
	rendered := RenderDepGraph(graph, 80)

	if !strings.Contains(rendered, "PR #42") {
		t.Fatalf("expected PR #42 in render, got: %q", rendered)
	}
	// Should have tree connectors for fan-out
	if !strings.Contains(rendered, "├──") {
		t.Fatalf("expected ├── connector in fan-out render, got: %q", rendered)
	}
	if !strings.Contains(rendered, "└──") {
		t.Fatalf("expected └── connector in fan-out render, got: %q", rendered)
	}
}

func TestTruncateLabel(t *testing.T) {
	tests := []struct {
		label    string
		maxWidth int
		want     string
	}{
		{"short", 10, "short"},
		{"a very long label that exceeds", 15, "a very long ..."},
		{"abc", 3, "abc"},
		{"abcd", 3, "abc"},
	}

	for _, tt := range tests {
		got := truncateLabel(tt.label, tt.maxWidth)
		if got != tt.want {
			t.Errorf("truncateLabel(%q, %d) = %q, want %q", tt.label, tt.maxWidth, got, tt.want)
		}
	}
}

func TestRenderDepGraph_EmptyGraph(t *testing.T) {
	graph := &DepGraph{}
	rendered := RenderDepGraph(graph, 80)
	if rendered != "" {
		t.Fatalf("expected empty string for empty graph, got: %q", rendered)
	}
}

func TestRenderDepGraph_NilGraph(t *testing.T) {
	rendered := RenderDepGraph(nil, 80)
	if rendered != "" {
		t.Fatalf("expected empty string for nil graph, got: %q", rendered)
	}
}

func TestBranchGroupPrefix(t *testing.T) {
	tests := []struct {
		branch string
		want   string
	}{
		{"feature/auth-bug", "feature/auth"},
		{"feature/auth-tests", "feature/auth"},
		{"fix/login-flow", "fix/login"},
		{"main", ""},
		{"master", ""},
		{"", ""},
		{"feature/auth", "feature"},
		{"simple", "simple"},
	}

	for _, tt := range tests {
		got := branchGroupPrefix(tt.branch)
		if got != tt.want {
			t.Errorf("branchGroupPrefix(%q) = %q, want %q", tt.branch, got, tt.want)
		}
	}
}

func TestSessionLabel(t *testing.T) {
	tests := []struct {
		session *data.Session
		want    string
	}{
		{&data.Session{PRNumber: 42}, "PR #42"},
		{&data.Session{Title: "Fix bug"}, "Fix bug"},
		{&data.Session{ID: "abc-123"}, "abc-123"},
		{&data.Session{PRNumber: 10, Title: "Some title"}, "PR #10"},
	}

	for _, tt := range tests {
		got := sessionLabel(tt.session)
		if got != tt.want {
			t.Errorf("sessionLabel(%+v) = %q, want %q", tt.session, got, tt.want)
		}
	}
}
