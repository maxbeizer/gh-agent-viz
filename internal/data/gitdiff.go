package data

import (
	"fmt"
	"os/exec"
	"strings"
)

// GitDiffResult holds the combined diff output and summary stats
type GitDiffResult struct {
	Diff      string // combined unified diff (unstaged + staged)
	StatLines string // human-readable stat summary (like git diff --stat)
	FileCount int    // number of files changed
	Additions int    // total lines added
	Deletions int    // total lines removed
}

// FetchSessionGitDiff runs git diff in the given working directory,
// combining both unstaged and staged changes.
func FetchSessionGitDiff(workDir string) (*GitDiffResult, error) {
	if workDir == "" {
		return nil, fmt.Errorf("no working directory specified")
	}

	// Get unstaged changes
	unstaged, err := runGitDiff(workDir, false)
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}

	// Get staged changes
	staged, err := runGitDiff(workDir, true)
	if err != nil {
		return nil, fmt.Errorf("git diff --cached failed: %w", err)
	}

	// Combine diffs
	combined := unstaged
	if staged != "" {
		if combined != "" {
			combined += "\n"
		}
		combined += staged
	}

	// Get stat summary
	stat, _ := runGitDiffStat(workDir)

	result := &GitDiffResult{
		Diff:      combined,
		StatLines: stat,
	}
	result.FileCount, result.Additions, result.Deletions = parseNumstat(workDir)

	return result, nil
}

// runGitDiff executes git diff and returns the output
func runGitDiff(workDir string, cached bool) (string, error) {
	args := []string{"-C", workDir, "diff", "--no-color"}
	if cached {
		args = append(args, "--cached")
	}
	cmd := execCommand("git", args...)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git diff error: %s", string(exitErr.Stderr))
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// runGitDiffStat returns the --stat summary
func runGitDiffStat(workDir string) (string, error) {
	// Combine unstaged + staged stats
	cmd := execCommand("git", "-C", workDir, "diff", "--stat", "--no-color", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		// Fall back to just unstaged if HEAD doesn't exist yet
		cmd = execCommand("git", "-C", workDir, "diff", "--stat", "--no-color")
		out, err = cmd.Output()
		if err != nil {
			return "", err
		}
	}
	return strings.TrimSpace(string(out)), nil
}

// parseNumstat extracts file count, additions, deletions from git diff --numstat
func parseNumstat(workDir string) (files, adds, dels int) {
	cmd := execCommand("git", "-C", workDir, "diff", "--numstat", "--no-color", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		cmd = execCommand("git", "-C", workDir, "diff", "--numstat", "--no-color")
		out, err = cmd.Output()
		if err != nil {
			return 0, 0, 0
		}
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		files++
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			var a, d int
			fmt.Sscanf(parts[0], "%d", &a)
			fmt.Sscanf(parts[1], "%d", &d)
			adds += a
			dels += d
		}
	}
	return
}
