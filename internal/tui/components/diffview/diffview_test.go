package diffview

import (
	"strings"
	"testing"
)

func TestParseUnifiedDiff_BasicDiff(t *testing.T) {
	raw := `diff --git a/src/auth/handler.go b/src/auth/handler.go
index abc1234..def5678 100644
--- a/src/auth/handler.go
+++ b/src/auth/handler.go
@@ -15,7 +15,12 @@ func HandleAuth(...)
   func HandleAuth(w http.ResponseWriter, r *http.Request) {
-      token := r.Header.Get("Authorization")
+      token, err := extractToken(r)
+      if err != nil {
+          http.Error(w, "unauthorized", 401)
+          return
+      }
diff --git a/src/auth/handler_test.go b/src/auth/handler_test.go
new file mode 100644
--- /dev/null
+++ b/src/auth/handler_test.go
@@ -0,0 +1,3 @@
+func TestHandleAuth(t *testing.T) {
+    // test body
+}
`

	files := ParseUnifiedDiff(raw)
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}

	f1 := files[0]
	if f1.Path != "src/auth/handler.go" {
		t.Errorf("wrong path: %q", f1.Path)
	}
	if f1.Additions != 5 {
		t.Errorf("expected 5 additions, got %d", f1.Additions)
	}
	if f1.Deletions != 1 {
		t.Errorf("expected 1 deletion, got %d", f1.Deletions)
	}

	f2 := files[1]
	if f2.Path != "src/auth/handler_test.go" {
		t.Errorf("wrong path: %q", f2.Path)
	}
	if f2.Additions != 3 {
		t.Errorf("expected 3 additions, got %d", f2.Additions)
	}
	if f2.Deletions != 0 {
		t.Errorf("expected 0 deletions, got %d", f2.Deletions)
	}
}

func TestParseUnifiedDiff_EmptyInput(t *testing.T) {
	files := ParseUnifiedDiff("")
	if files != nil {
		t.Errorf("expected nil for empty input, got %v", files)
	}

	files = ParseUnifiedDiff("   \n  ")
	if files != nil {
		t.Errorf("expected nil for whitespace input, got %v", files)
	}
}

func TestRenderPatch_ColoredLines(t *testing.T) {
	patch := `--- a/file.go
+++ b/file.go
@@ -1,3 +1,3 @@
 context line
-removed line
+added line`

	rendered := RenderPatch(patch)

	// Additions should contain ANSI green (color 42)
	if !strings.Contains(rendered, "added line") {
		t.Error("rendered output should contain 'added line'")
	}
	if !strings.Contains(rendered, "removed line") {
		t.Error("rendered output should contain 'removed line'")
	}
	// Context lines should be unmodified
	if !strings.Contains(rendered, " context line") {
		t.Error("rendered output should preserve context lines")
	}
}

func TestFormatStats(t *testing.T) {
	tests := []struct {
		add, del int
		want     string
	}{
		{12, 3, "(+12 -3)"},
		{5, 0, "(+5)"},
		{0, 2, "(-2)"},
		{0, 0, ""},
	}

	for _, tt := range tests {
		got := formatStats(tt.add, tt.del)
		if got != tt.want {
			t.Errorf("formatStats(%d, %d) = %q, want %q", tt.add, tt.del, got, tt.want)
		}
	}
}

func TestRenderFileHeader(t *testing.T) {
	f := FileDiff{
		Path:      "src/main.go",
		Additions: 10,
		Deletions: 2,
	}
	header := renderFileHeader(f)
	if !strings.Contains(header, "src/main.go") {
		t.Error("header should contain file path")
	}
	if !strings.Contains(header, "+10 -2") {
		t.Error("header should contain addition/deletion counts")
	}
	if !strings.Contains(header, "📄") {
		t.Error("header should contain file icon")
	}
}

func TestModel_EmptyState(t *testing.T) {
	m := New(80, 24)
	view := m.View()
	if !strings.Contains(view, "No diffs available") {
		t.Errorf("expected empty state message, got %q", view)
	}
}

func TestModel_SetDiffs(t *testing.T) {
	m := New(80, 24)
	m.SetDiffs([]FileDiff{
		{
			Path:      "test.go",
			Additions: 1,
			Deletions: 0,
			Patch:     "+new line",
		},
	})
	view := m.View()
	if view == "No diffs available" {
		t.Error("expected rendered diff, got empty state")
	}
}

func TestExtractPath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"a/src/main.go b/src/main.go", "src/main.go"},
		{"a/file.txt b/file.txt", "file.txt"},
		{"", ""},
	}

	for _, tt := range tests {
		got := extractPath(tt.input)
		if got != tt.want {
			t.Errorf("extractPath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
