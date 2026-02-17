package taskdetail

import (
	"fmt"
	"strings"

	"github.com/maxbeizer/gh-agent-viz/internal/data"
)

// DepNode represents a node in the dependency graph
type DepNode struct {
	Label string // e.g., "PR #42", "Issue #15", session title
	URL   string // for potential browser opening
}

// DepEdge represents a directed relationship
type DepEdge struct {
	From int // index into nodes
	To   int // index into nodes
}

// DepGraph holds the parsed dependency information
type DepGraph struct {
	Nodes []DepNode
	Edges []DepEdge
}

// ParseSessionDeps extracts dependency information from a session.
// Looks for PR references, issue references in the title/branch.
func ParseSessionDeps(session *data.Session, allSessions []data.Session) *DepGraph {
	if session == nil || len(allSessions) == 0 {
		return &DepGraph{}
	}

	graph := &DepGraph{}

	// Add the current session as the first node
	currentLabel := sessionLabel(session)
	graph.Nodes = append(graph.Nodes, DepNode{
		Label: currentLabel,
		URL:   session.PRURL,
	})

	currentIdx := 0
	repo := strings.TrimSpace(session.Repository)
	branch := strings.TrimSpace(session.Branch)
	branchPrefix := branchGroupPrefix(branch)

	for i := range allSessions {
		other := &allSessions[i]
		if other.ID == session.ID {
			continue
		}

		related := false

		// Same repository with related branch prefix
		otherRepo := strings.TrimSpace(other.Repository)
		otherBranch := strings.TrimSpace(other.Branch)
		if repo != "" && otherRepo != "" && repo == otherRepo {
			if branchPrefix != "" && branchGroupPrefix(otherBranch) == branchPrefix {
				related = true
			}
		}

		if !related {
			continue
		}

		otherLabel := sessionLabel(other)
		otherIdx := len(graph.Nodes)
		graph.Nodes = append(graph.Nodes, DepNode{
			Label: otherLabel,
			URL:   other.PRURL,
		})
		graph.Edges = append(graph.Edges, DepEdge{
			From: currentIdx,
			To:   otherIdx,
		})
	}

	return graph
}

// RenderDepGraph renders a dependency graph using box-drawing characters.
// Returns empty string if graph has no meaningful relationships.
func RenderDepGraph(graph *DepGraph, width int) string {
	if graph == nil || len(graph.Edges) == 0 {
		return ""
	}

	maxLabelWidth := width - 10
	if maxLabelWidth < 10 {
		maxLabelWidth = 10
	}

	// Collect targets for the current (first) node
	targets := []int{}
	for _, e := range graph.Edges {
		if e.From == 0 {
			targets = append(targets, e.To)
		}
	}

	if len(targets) == 0 {
		return ""
	}

	fromLabel := truncateLabel(graph.Nodes[0].Label, maxLabelWidth)

	var lines []string
	if len(targets) == 1 {
		toLabel := truncateLabel(graph.Nodes[targets[0]].Label, maxLabelWidth)
		lines = append(lines, fmt.Sprintf("  %s ──→ %s", fromLabel, toLabel))
	} else {
		// First target uses direct arrow
		toLabel := truncateLabel(graph.Nodes[targets[0]].Label, maxLabelWidth)
		lines = append(lines, fmt.Sprintf("  %s ──→ %s", fromLabel, toLabel))
		// Remaining targets use tree connectors
		padding := strings.Repeat(" ", len(fromLabel)+2)
		for i := 1; i < len(targets); i++ {
			toLabel = truncateLabel(graph.Nodes[targets[i]].Label, maxLabelWidth)
			connector := "├──"
			if i == len(targets)-1 {
				connector = "└──"
			}
			lines = append(lines, fmt.Sprintf("%s%s→ %s", padding, connector, toLabel))
		}
	}

	return strings.Join(lines, "\n")
}

// sessionLabel returns a display label for a session node.
func sessionLabel(s *data.Session) string {
	if s.PRNumber > 0 {
		return fmt.Sprintf("PR #%d", s.PRNumber)
	}
	title := strings.TrimSpace(s.Title)
	if title != "" {
		return title
	}
	return s.ID
}

// branchGroupPrefix extracts a grouping prefix from a branch name.
// e.g., "feature/auth-bug" → "feature/auth", "fix/login-flow" → "fix/login"
func branchGroupPrefix(branch string) string {
	branch = strings.TrimSpace(branch)
	if branch == "" || branch == "main" || branch == "master" {
		return ""
	}

	// Split on last hyphen to get the prefix group
	idx := strings.LastIndex(branch, "-")
	if idx > 0 {
		return branch[:idx]
	}

	// Split on last slash segment if no hyphen
	idx = strings.LastIndex(branch, "/")
	if idx > 0 {
		return branch[:idx]
	}

	return branch
}

// truncateLabel shortens a label to fit within maxWidth, adding ellipsis.
func truncateLabel(label string, maxWidth int) string {
	if len(label) <= maxWidth {
		return label
	}
	if maxWidth <= 3 {
		return label[:maxWidth]
	}
	return label[:maxWidth-3] + "..."
}
