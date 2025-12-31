package cmd

import (
	"context"
	"fmt"

	"github.com/Zytera/gh-project-management/internal/config"
	"github.com/Zytera/gh-project-management/internal/gh"
	"github.com/spf13/cobra"
)

var (
	targetRepo string
)

var transferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "Transfer issues between repositories",
	Long: `Transfer issues between repositories.

IMPORTANT: Follow this order when transferring issues:
1. Create issue in source repository
2. Link to parent with: gh project-management link add
3. Set custom fields with: gh project-management field set (especially Team)
4. Set dependencies with: gh project-management dependency add
5. Transfer to target repository (this command - can be inferred from Team field)

Note: Dependencies MUST be established before transferring, as they require issues
to be in the same repository. Custom fields (especially Team) should be set before
transfer to determine the target repository.`,
}

var transferIssueCmd = &cobra.Command{
	Use:   "issue <issue-number>",
	Short: "Transfer an issue to another repository",
	Long: `Transfer an issue to another repository.

The issue number will change after transfer. Make sure to note the new number
for updating parent issue references.

Examples:
  # Transfer issue #48 to backend repository
  gh project-management transfer issue 48 --target backend

  # Transfer issue #50 to mobile-app repository
  gh project-management transfer issue 50 --target mobile-app`,
	Args: cobra.ExactArgs(1),
	RunE: runTransferIssue,
}

func runTransferIssue(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Parse issue reference
	issueRef := args[0]
	_, _, issueNumber, err := gh.ParseIssueReference(issueRef, cfg.Owner, cfg.DefaultRepo)
	if err != nil {
		return fmt.Errorf("invalid issue reference: %w", err)
	}

	if targetRepo == "" {
		return fmt.Errorf("target repository is required (use --target flag)")
	}

	sourceRepo := fmt.Sprintf("%s/%s", cfg.Owner, cfg.DefaultRepo)

	fmt.Printf("Transferring issue #%d to %s/%s...\n", issueNumber, cfg.Owner, targetRepo)
	fmt.Println()
	fmt.Println("⚠️  REMINDER:")
	fmt.Println("   - Ensure custom fields are already set (Team, Priority, Type)")
	fmt.Println("   - Ensure dependencies are already set (they require same repo)")
	fmt.Println("   - Note the new issue number for updating parent references")
	fmt.Println()

	// Transfer the issue
	_, err = gh.TransferIssue(ctx, issueNumber, cfg.Owner, targetRepo, sourceRepo)
	if err != nil {
		return fmt.Errorf("failed to transfer issue: %w", err)
	}

	fmt.Printf("\n✓ Transfer initiated for issue #%d to %s/%s\n", issueNumber, cfg.Owner, targetRepo)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("1. Note the new issue number from the output above")
	fmt.Println("2. Update parent issue body with cross-repo reference: ### <owner>/<repo>#<new-number>")

	return nil
}

func init() {
	transferIssueCmd.Flags().StringVar(&targetRepo, "target", "", "Target repository name (required)")
	transferIssueCmd.MarkFlagRequired("target")

	transferCmd.AddCommand(transferIssueCmd)
	rootCmd.AddCommand(transferCmd)
}
