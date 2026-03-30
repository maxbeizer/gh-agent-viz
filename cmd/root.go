package cmd

import (
	"fmt"
	"os"
	"runtime/pprof"

	tea "charm.land/bubbletea/v2"
	"github.com/maxbeizer/gh-agent-viz/internal/data"
	"github.com/maxbeizer/gh-agent-viz/internal/tui"
	"github.com/spf13/cobra"
)

var (
	repoFlag     string
	debugFlag    bool
	demoFlag     bool
	snapshotFlag string
	profileFlag  string
)

// Version is set by goreleaser at build time.
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "gh-agent-viz",
	Short: "Interactive terminal UI for visualizing GitHub Copilot agent sessions",
	Long: `gh-agent-viz is a GitHub CLI extension that provides an interactive
terminal UI for visualizing and managing GitHub Copilot coding agent sessions.

	View agent task status, browse details, and review logs in an easy-to-use TUI.`,
	Run: func(cmd *cobra.Command, args []string) {
		data.SetDebug(debugFlag)

		// Start CPU profiling if requested
		if profileFlag != "" {
			f, err := os.Create(profileFlag)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not create profile: %v\n", err)
				os.Exit(1)
			}
			if err := pprof.StartCPUProfile(f); err != nil {
				fmt.Fprintf(os.Stderr, "Could not start profile: %v\n", err)
				os.Exit(1)
			}
			defer func() {
				pprof.StopCPUProfile()
				f.Close()
				fmt.Fprintf(os.Stderr, "CPU profile written to %s\nAnalyze with: go tool pprof -http=:8080 %s\n", profileFlag, profileFlag)
			}()
		}

		// Create the Bubble Tea program
		model := tui.NewModel(repoFlag, debugFlag, demoFlag, snapshotFlag, Version)
		p := tea.NewProgram(model)

		// Run the program
		if _, err := p.Run(); err != nil {
			// Stop profiling before exit so the profile is flushed
			if profileFlag != "" {
				pprof.StopCPUProfile()
			}
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
	// Version (GoReleaser injects the real version at build time)
	rootCmd.Version = Version

	// Add flags
	rootCmd.Flags().StringVarP(&repoFlag, "repo", "R", "", "Scope to a specific repository (format: owner/repo)")
	rootCmd.Flags().BoolVar(&debugFlag, "debug", false, "Enable debug diagnostics and write command logs to ~/.gh-agent-viz-debug.log")
	rootCmd.Flags().BoolVar(&demoFlag, "demo", false, "Run with fake demo data for screenshots and recordings")
	rootCmd.Flags().StringVar(&snapshotFlag, "snapshot", "", "Write a JSON snapshot of TUI state after initial load and exit")
	rootCmd.Flags().StringVar(&profileFlag, "profile", "", "Write a CPU profile to the given file (analyze with: go tool pprof)")
}
