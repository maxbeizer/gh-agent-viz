package diffview

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FileDiff represents a single file's diff content
type FileDiff struct {
	Path      string
	Additions int
	Deletions int
	Patch     string // unified diff content
}

// Model represents the diff view component state
type Model struct {
	files    []FileDiff
	viewport viewport.Model
	width    int
	height   int
	ready    bool
}

// Styles for diff rendering
var (
	addStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))  // green
	delStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // red
	hunkStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Faint(true)
	headerStyle = lipgloss.NewStyle().Bold(true)
	sepStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
)

// New creates a new diff view model
func New(width, height int) Model {
	vp := viewport.New(width, height)
	return Model{
		viewport: vp,
		width:    width,
		height:   height,
	}
}

// SetDiffs updates the file diffs and re-renders the viewport content
func (m *Model) SetDiffs(files []FileDiff) {
	m.files = files
	m.viewport.SetContent(m.renderDiffs())
	m.viewport.GotoTop()
	m.ready = true
}

// SetSize updates the viewport dimensions
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.Width = width
	m.viewport.Height = height
	if m.ready {
		m.viewport.SetContent(m.renderDiffs())
	}
}

// View renders the diff view
func (m Model) View() string {
	if !m.ready || len(m.files) == 0 {
		return "No diffs available"
	}
	return m.viewport.View()
}

// Update handles messages for the diff view
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// renderDiffs produces the colored diff output for all files
func (m Model) renderDiffs() string {
	var sb strings.Builder
	for i, f := range m.files {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(renderFileHeader(f))
		sb.WriteString("\n")
		sb.WriteString(renderPatch(f.Patch))
	}
	return sb.String()
}

// renderFileHeader renders the file name with addition/deletion counts
func renderFileHeader(f FileDiff) string {
	stats := formatStats(f.Additions, f.Deletions)
	title := headerStyle.Render(fmt.Sprintf("📄 %s  %s", f.Path, stats))
	sep := sepStyle.Render(strings.Repeat("─", 40))
	return title + "\n" + sep
}

// formatStats formats the +N -M counters
func formatStats(additions, deletions int) string {
	var parts []string
	if additions > 0 {
		parts = append(parts, fmt.Sprintf("+%d", additions))
	}
	if deletions > 0 {
		parts = append(parts, fmt.Sprintf("-%d", deletions))
	}
	if len(parts) == 0 {
		return ""
	}
	return "(" + strings.Join(parts, " ") + ")"
}

// RenderPatch renders a unified diff patch with syntax coloring.
// Exported for testing.
func RenderPatch(patch string) string {
	return renderPatch(patch)
}

func renderPatch(patch string) string {
	lines := strings.Split(patch, "\n")
	var sb strings.Builder
	for _, line := range lines {
		rendered := renderDiffLine(line)
		sb.WriteString(rendered)
		sb.WriteString("\n")
	}
	return sb.String()
}

// renderDiffLine applies color to a single diff line
func renderDiffLine(line string) string {
	switch {
	case strings.HasPrefix(line, "+++ ") || strings.HasPrefix(line, "--- "):
		// file header lines in unified diff — render dimmed
		return hunkStyle.Render(line)
	case strings.HasPrefix(line, "+"):
		return addStyle.Render(line)
	case strings.HasPrefix(line, "-"):
		return delStyle.Render(line)
	case strings.HasPrefix(line, "@@"):
		return hunkStyle.Render(line)
	default:
		return line
	}
}

// ParseUnifiedDiff parses raw `gh pr diff` output into FileDiff structs
func ParseUnifiedDiff(raw string) []FileDiff {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	// Split on "diff --git" boundaries
	parts := strings.Split(raw, "diff --git ")
	var files []FileDiff
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		fd := parseFileDiff(part)
		files = append(files, fd)
	}
	return files
}

// parseFileDiff parses a single file's diff section
func parseFileDiff(section string) FileDiff {
	lines := strings.Split(section, "\n")
	fd := FileDiff{}

	// Extract path from the first line: "a/path b/path"
	if len(lines) > 0 {
		fd.Path = extractPath(lines[0])
	}

	// Look for --- a/ and +++ b/ to refine path, count +/- lines
	var patchLines []string
	inPatch := false
	for _, line := range lines {
		if strings.HasPrefix(line, "--- a/") || strings.HasPrefix(line, "--- /dev/null") {
			inPatch = true
		}
		if strings.HasPrefix(line, "+++ b/") {
			fd.Path = strings.TrimPrefix(line, "+++ b/")
		}

		if inPatch {
			patchLines = append(patchLines, line)
			// Count additions and deletions (only actual content lines, not headers)
			if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++ ") {
				fd.Additions++
			} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "--- ") {
				fd.Deletions++
			}
		}
	}

	fd.Patch = strings.Join(patchLines, "\n")
	return fd
}

// extractPath pulls the file path from the "a/path b/path" line
func extractPath(headerLine string) string {
	// Format: "a/some/path b/some/path"
	parts := strings.SplitN(headerLine, " ", 2)
	if len(parts) < 1 {
		return ""
	}
	path := parts[0]
	path = strings.TrimPrefix(path, "a/")
	return path
}
