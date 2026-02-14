package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/maxbeizer/gh-agent-viz/internal/tui"
	"github.com/spf13/cobra"
)

var (
	repoFlag string
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "gh-agent-viz",
	Short: "Interactive terminal UI for visualizing GitHub Copilot agent sessions",
	Long: `gh-agent-viz is a GitHub CLI extension that provides an interactive
terminal UI for visualizing and managing GitHub Copilot coding agent sessions.

View agent task status, browse details, and review logs in an easy-to-use TUI.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create the Bubble Tea program
		model := tui.NewModel(repoFlag)
		p := tea.NewProgram(model, tea.WithAltScreen())

		// Run the program
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
			os.Exit(1)
		}
	},
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Add flags
	rootCmd.Flags().StringVarP(&repoFlag, "repo", "R", "", "Scope to a specific repository (format: owner/repo)")
}
