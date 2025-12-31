package cmd

import (
	"github.com/spf13/cobra"
)

var issueCmd = &cobra.Command{
	Use:   "issue",
	Short: "Manage issues (epics, stories, tasks, bugs, features)",
	Long: `Create and manage different types of issues in your GitHub project.

Issues can be created using templates from your repository or built-in defaults.
Use 'gh project-management issue create --type <type> --show-fields' to see
available fields for any issue type.`,
}

func init() {
	issueCmd.AddCommand(issueCreateCmd)
	rootCmd.AddCommand(issueCmd)
}
