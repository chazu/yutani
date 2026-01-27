// Package cli provides the Yutani command-line interface.
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version information
	Version   = "0.1.0-alpha"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "yutani",
	Short: "Yutani TUI Framework CLI",
	Long: `Yutani CLI - Command-line tools for the Yutani TUI framework.

The Yutani CLI provides tools for:
  - Starting and managing servers
  - Inspecting running applications
  - Profiling performance
  - Debugging widget hierarchies
  - Generating project scaffolding`,
	Version: Version,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Set version template
	rootCmd.SetVersionTemplate(fmt.Sprintf(`Yutani CLI
Version:    %s
Build Time: %s
Git Commit: %s
`, Version, BuildTime, GitCommit))

	// Add commands
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(pingCmd)
	rootCmd.AddCommand(inspectCmd)
	rootCmd.AddCommand(profileCmd)
	rootCmd.AddCommand(debugCmd)
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(widgetCmd)
	rootCmd.AddCommand(sessionCmd)
	rootCmd.AddCommand(injectCmd)
}

