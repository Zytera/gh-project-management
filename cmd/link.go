package cmd

import (
	"context"
	"fmt"

	"github.com/Zytera/gh-project-management/internal/config"
	"github.com/Zytera/gh-project-management/internal/gh"
	"github.com/spf13/cobra"
)

var linkCmd = &cobra.Command{
	Use:   "link",
	Short: "Manage issue linking (parent-child relationships)",
	Long: `Manage parent-child relationships between issues.

This command uses GitHub's API to create hierarchical relationships by updating
issue bodies with tasklist items that GitHub recognizes as parent-child links.

The typical hierarchy is:
  Epic → User Story → Task → Subtask`,
}

var linkAddCmd = &cobra.Command{
	Use:   "add <parent-issue> <child-issue>",
	Short: "Link a child issue to a parent issue",
	Long: `Link a child issue to a parent issue, establishing a parent-child relationship.

Issue references can be specified as:
  - #123 (issue number in configured default repo)
  - 123 (issue number in configured default repo)

Note: Both issues must be in the same repository.

Examples:
  # Link User Story #45 to Epic #44
  gh project-management link add #44 #45

  # Link Task #48 to User Story #45
  gh project-management link add 45 48`,
	Args: cobra.ExactArgs(2),
	RunE: runLinkAdd,
}

var linkRemoveCmd = &cobra.Command{
	Use:   "remove <parent-issue> <child-issue>",
	Short: "Remove a child issue from a parent issue",
	Long: `Remove a child issue from a parent issue, breaking the parent-child relationship.

Issue references can be specified as:
  - #123 (issue number in configured default repo)
  - 123 (issue number in configured default repo)

Examples:
  # Remove User Story #45 from Epic #44
  gh project-management link remove #44 #45`,
	Args: cobra.ExactArgs(2),
	RunE: runLinkRemove,
}

func runLinkAdd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	parentIssueRef := args[0]
	childIssueRef := args[1]

	// Parse issue references
	_, _, parentNumber, err := gh.ParseIssueReference(parentIssueRef, cfg.Owner, cfg.DefaultRepo)
	if err != nil {
		return fmt.Errorf("invalid parent issue reference: %w", err)
	}

	_, _, childNumber, err := gh.ParseIssueReference(childIssueRef, cfg.Owner, cfg.DefaultRepo)
	if err != nil {
		return fmt.Errorf("invalid child issue reference: %w", err)
	}

	fmt.Printf("Linking issue #%d as child of issue #%d...\n", childNumber, parentNumber)

	// Add the sub-issue relationship
	err = gh.AddSubIssue(ctx, cfg.Owner, cfg.DefaultRepo, parentNumber, childNumber)
	if err != nil {
		return fmt.Errorf("failed to link issues: %w", err)
	}

	fmt.Printf("✓ Successfully linked #%d to #%d\n", childNumber, parentNumber)
	return nil
}

func runLinkRemove(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	parentIssueRef := args[0]
	childIssueRef := args[1]

	// Parse issue references
	_, _, parentNumber, err := gh.ParseIssueReference(parentIssueRef, cfg.Owner, cfg.DefaultRepo)
	if err != nil {
		return fmt.Errorf("invalid parent issue reference: %w", err)
	}

	_, _, childNumber, err := gh.ParseIssueReference(childIssueRef, cfg.Owner, cfg.DefaultRepo)
	if err != nil {
		return fmt.Errorf("invalid child issue reference: %w", err)
	}

	fmt.Printf("Removing link between #%d and #%d...\n", childNumber, parentNumber)

	// Remove the sub-issue relationship
	err = gh.RemoveSubIssue(ctx, cfg.Owner, cfg.DefaultRepo, parentNumber, childNumber)
	if err != nil {
		return fmt.Errorf("failed to remove link: %w", err)
	}

	fmt.Printf("✓ Successfully removed link between #%d and #%d\n", childNumber, parentNumber)
	return nil
}

func init() {
	linkCmd.AddCommand(linkAddCmd)
	linkCmd.AddCommand(linkRemoveCmd)
	rootCmd.AddCommand(linkCmd)
}
