package gitactivity

import (
	"strings"
	"testing"

	"github.com/maxbeizer/gh-agent-viz/internal/data"
)

func TestNew_DefaultState(t *testing.T) {
	m := New(80, 24)
	if !m.loading {
		t.Fatal("expected loading to be true initially")
	}
	view := m.View()
	if !strings.Contains(view, "Loading") {
		t.Fatalf("expected loading text in view, got: %q", view)
	}
}

func TestSetDiffResult_EmptyDiff(t *testing.T) {
	m := New(80, 24)
	m.SetDiffResult(&data.GitDiffResult{})
	view := m.View()
	if !strings.Contains(view, "No uncommitted changes") {
		t.Fatalf("expected 'No uncommitted changes' for empty diff, got: %q", view)
	}
}

func TestSetDiffResult_WithChanges(t *testing.T) {
	m := New(80, 24)
	m.SetDiffResult(&data.GitDiffResult{
		Diff:      "diff --git a/foo.go b/foo.go\n--- a/foo.go\n+++ b/foo.go\n@@ -1,3 +1,4 @@\n package foo\n+// new comment\n",
		StatLines: " foo.go | 1 +\n 1 file changed, 1 insertion(+)",
		FileCount: 1,
		Additions: 1,
		Deletions: 0,
	})
	if m.loading {
		t.Fatal("expected loading to be false after SetDiffResult")
	}
	view := m.View()
	if strings.Contains(view, "No uncommitted changes") {
		t.Fatal("should not show 'No uncommitted changes' when there are changes")
	}
}

func TestSetDiffResult_Nil(t *testing.T) {
	m := New(80, 24)
	m.SetDiffResult(nil)
	view := m.View()
	if !strings.Contains(view, "No uncommitted changes") {
		t.Fatalf("expected 'No uncommitted changes' for nil result, got: %q", view)
	}
}

func TestSetLoading(t *testing.T) {
	m := New(80, 24)
	m.SetLoading(false)
	if m.loading {
		t.Fatal("expected loading to be false")
	}
	m.SetLoading(true)
	if !m.loading {
		t.Fatal("expected loading to be true")
	}
}
