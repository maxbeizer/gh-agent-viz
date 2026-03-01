package data

import (
	"os/exec"
	"strings"
	"testing"
)

func TestFetchSessionGitDiff_EmptyWorkDir(t *testing.T) {
	_, err := FetchSessionGitDiff("")
	if err == nil {
		t.Fatal("expected error for empty workDir")
	}
	if !strings.Contains(err.Error(), "no working directory") {
		t.Fatalf("expected 'no working directory' error, got: %s", err)
	}
}

func TestFetchSessionGitDiff_ValidRepo(t *testing.T) {
	// Use the current repo as a test target — it's a valid git repo
	// We just need to verify it doesn't error; actual diff content varies
	result, err := FetchSessionGitDiff(".")
	if err != nil {
		// If git isn't available, skip
		if _, pathErr := exec.LookPath("git"); pathErr != nil {
			t.Skip("git not available")
		}
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// FileCount, Additions, Deletions should be >= 0
	if result.FileCount < 0 || result.Additions < 0 || result.Deletions < 0 {
		t.Fatalf("unexpected negative stats: files=%d adds=%d dels=%d",
			result.FileCount, result.Additions, result.Deletions)
	}
}

func TestGitDiffResult_EmptyDiff(t *testing.T) {
	result := &GitDiffResult{}
	if result.Diff != "" {
		t.Fatal("expected empty diff")
	}
	if result.FileCount != 0 {
		t.Fatal("expected zero file count")
	}
}
