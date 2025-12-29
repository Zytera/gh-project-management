package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/Zytera/gh-project-management/internal/config"
	"github.com/spf13/cobra"
)

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "gh-project-management",
	Short: "Manage GitHub Projects with hierarchical issues",
	Long: `gh-project-management is a GitHub CLI extension for managing projects
with hierarchical issues (Epics, User Stories, Tasks).

Use 'gh project-management context add <name>' to configure your first project.`,
	Version: Version,
}

func Execute() int {
	// Check if the command being executed needs configuration
	// Commands that don't need config: context, help, version
	args := os.Args[1:]
	needsConfig := true

	if len(args) > 0 {
		cmd := args[0]
		// Commands that don't require configuration
		if cmd == "context" || cmd == "help" || cmd == "--help" || cmd == "-h" || cmd == "version" || cmd == "--version" || cmd == "-v" {
			needsConfig = false
		}
	}

	var ctx context.Context
	if needsConfig {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, "\nRun 'gh project-management context add <name>' to configure your first project.")
			return 1
		}
		ctx = context.WithValue(context.Background(), config.ConfigKey{}, cfg)
	} else {
		ctx = context.Background()
	}

	rootCmd.SetContext(ctx)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
