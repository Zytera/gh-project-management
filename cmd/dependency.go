package cmd

import (
	"context"
	"fmt"

	"github.com/Zytera/gh-project-management/internal/config"
	"github.com/Zytera/gh-project-management/internal/gh"
	"github.com/spf13/cobra"
)

var dependencyCmd = &cobra.Command{
	Use:   "dependency",
	Short: "Manage issue dependencies",
	Long: `Manage dependencies between issues.

Dependencies establish blocked-by relationships where one issue is blocked by another.
This is useful for tracking which issues must be completed before others can proceed.`,
}

var dependencyAddCmd = &cobra.Command{
	Use:   "add <blocked-issue> <blocking-issue>",
	Short: "Add a dependency between issues",
	Long: `Add a dependency where the first issue is blocked by the second issue.

Issue references can be specified as:
  - #123 (issue in configured default repo)
  - 123 (issue in configured default repo)
  - owner/repo#123 (issue in a specific repository)

Examples:
  # Issue #46 is blocked by issue #45 (both in default repo)
  gh project-management dependency add #46 #45

  # Issue #50 in current repo blocked by issue #25 in another repo
  gh project-management dependency add #50 Zytera/backend#25`,
	Args: cobra.ExactArgs(2),
	RunE: runDependencyAdd,
}

func runDependencyAdd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	blockedIssueRef := args[0]
	blockingIssueRef := args[1]

	// Parse issue references
	blockedOwner, blockedRepo, blockedNumber, err := gh.ParseIssueReference(blockedIssueRef, cfg.Owner, cfg.DefaultRepo)
	if err != nil {
		return fmt.Errorf("invalid blocked issue reference: %w", err)
	}

	blockingOwner, blockingRepo, blockingNumber, err := gh.ParseIssueReference(blockingIssueRef, cfg.Owner, cfg.DefaultRepo)
	if err != nil {
		return fmt.Errorf("invalid blocking issue reference: %w", err)
	}

	// Verify both issues are in the same repository
	if blockedOwner != blockingOwner || blockedRepo != blockingRepo {
		return fmt.Errorf("both issues must be in the same repository (blocked: %s/%s, blocking: %s/%s)",
			blockedOwner, blockedRepo, blockingOwner, blockingRepo)
	}

	fmt.Printf("Adding dependency: %s/%s#%d is blocked by #%d...\n",
		blockedOwner, blockedRepo, blockedNumber, blockingNumber)

	// Add the dependency
	err = gh.AddBlockedBy(ctx, blockedOwner, blockedRepo, blockedNumber, blockingNumber)
	if err != nil {
		return fmt.Errorf("failed to add dependency: %w", err)
	}

	fmt.Printf("✓ Successfully added dependency: #%d is now blocked by #%d\n", blockedNumber, blockingNumber)
	return nil
}

func init() {
	dependencyCmd.AddCommand(dependencyAddCmd)
	rootCmd.AddCommand(dependencyCmd)
}
