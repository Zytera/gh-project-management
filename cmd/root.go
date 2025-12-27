package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:     "gh-project-managment",
	Short:   "TODO",
	Long:    `TODO`,
	Version: Version,
}

func Execute() int {
	// Add subcommands here (will be added in next tasks)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
