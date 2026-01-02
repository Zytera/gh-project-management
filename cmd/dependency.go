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
	Use:   "add <blocked-issue> <blocking-issue> [<blocking-issue2> ...]",
	Short: "Add dependencies to an issue",
	Long: `Add one or more dependencies where the first issue is blocked by the following issues.

Issue references can be specified as:
  - #123 (issue in configured default repo)
  - 123 (issue in configured default repo)
  - owner/repo#123 (issue in a specific repository)

Examples:
  # Issue #46 is blocked by issue #45 (both in default repo)
  gh project-management dependency add #46 #45

  # Issue #50 is blocked by multiple issues
  gh project-management dependency add #50 #45 #47 #48

  # Issue #50 in current repo blocked by issue #25 in another repo
  gh project-management dependency add #50 Zytera/backend#25`,
	Args: cobra.MinimumNArgs(2),
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
	blockingIssueRefs := args[1:]

	// Parse blocked issue reference
	blockedOwner, blockedRepo, blockedNumber, err := gh.ParseIssueReference(blockedIssueRef, cfg.Owner, cfg.DefaultRepo)
	if err != nil {
		return fmt.Errorf("invalid blocked issue reference: %w", err)
	}

	fmt.Printf("Adding dependencies to %s/%s#%d...\n", blockedOwner, blockedRepo, blockedNumber)

	// Track success count
	successCount := 0
	totalCount := len(blockingIssueRefs)

	// Add each blocking issue as a dependency
	for _, blockingIssueRef := range blockingIssueRefs {
		// Parse blocking issue reference
		blockingOwner, blockingRepo, blockingNumber, err := gh.ParseIssueReference(blockingIssueRef, cfg.Owner, cfg.DefaultRepo)
		if err != nil {
			fmt.Printf("⚠️  Warning: Invalid blocking issue reference '%s': %v\n", blockingIssueRef, err)
			continue
		}

		// Verify both issues are in the same repository
		if blockedOwner != blockingOwner || blockedRepo != blockingRepo {
			fmt.Printf("⚠️  Warning: Issue #%d must be in the same repository as #%d (skipping)\n", blockingNumber, blockedNumber)
			continue
		}

		// Add the dependency
		err = gh.AddBlockedBy(ctx, blockedOwner, blockedRepo, blockedNumber, blockingNumber)
		if err != nil {
			fmt.Printf("⚠️  Warning: Failed to add dependency on #%d: %v\n", blockingNumber, err)
			continue
		}

		fmt.Printf("  ✓ Blocked by #%d\n", blockingNumber)
		successCount++
	}

	// Summary
	if successCount == 0 {
		return fmt.Errorf("failed to add any dependencies")
	}

	fmt.Printf("\n✓ Successfully added %d/%d dependencies to issue #%d\n", successCount, totalCount, blockedNumber)
	return nil
}

func init() {
	dependencyCmd.AddCommand(dependencyAddCmd)
	rootCmd.AddCommand(dependencyCmd)
}
