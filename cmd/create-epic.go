package cmd

import (
	"fmt"

	"github.com/Zytera/gh-project-management/internal/config"
	createissue "github.com/Zytera/gh-project-management/internal/tui/create-issue"
	"github.com/Zytera/gh-project-management/pkg/issue"
	"github.com/spf13/cobra"
)

var epicTitle string

var createEpicCmd = &cobra.Command{
	Use:   "create-epic",
	Short: "Create a new epic issue",
	Long:  `Create a new epic issue in the default repository. Can be used interactively or with flags.`,
	Example: `  # Interactive mode
  gh project-management create-epic

  # With title flag
  gh project-management create-epic --title "New Authentication System"`,
	Args: cobra.ExactArgs(0),
	RunE: runCreateEpic,
}

func init() {
	rootCmd.AddCommand(createEpicCmd)
	createEpicCmd.Flags().StringVarP(&epicTitle, "title", "t", "", "Title of the epic")
}

func runCreateEpic(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// If title is provided via flag, use non-interactive mode
	if epicTitle != "" {
		cfg := ctx.Value(config.ConfigKey{}).(*config.Config)
		createdIssue, err := issue.CreateEpic(ctx, cfg, epicTitle)
		if err != nil {
			return err
		}

		fmt.Printf("\n✓ Epic created successfully: %s\n", createdIssue.URL)
		fmt.Printf("  Title:  %s\n", createdIssue.Title)
		fmt.Printf("  Number: #%d\n", createdIssue.Number)
		return nil
	}

	// Otherwise, use interactive mode
	return createissue.CreateEpicForm(ctx)
}
