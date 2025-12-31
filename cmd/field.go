package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Zytera/gh-project-management/internal/config"
	"github.com/Zytera/gh-project-management/internal/gh"
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
)

var (
	teamValue       string
	priorityValue   string
	typeValue       string
	noTransfer      bool
	teamRepoMapping = map[string]string{
		"Backend": "backend",
		"App":     "mobile-app",
		"Web":     "web-app",
		"Auth":    "auth",
	}
)

var fieldCmd = &cobra.Command{
	Use:   "field",
	Short: "Manage custom fields for issues",
	Long: `Manage custom fields (Team, Priority, Type) for issues in GitHub Projects.

Custom fields should be set BEFORE transferring issues to other repositories.
The Team field can be inferred from the issue context and determines the target repository.`,
}

var fieldSetCmd = &cobra.Command{
	Use:   "set <issue-number>",
	Short: "Set custom fields for an issue",
	Long: `Set custom fields (Team, Priority, Type) for an issue in the project.

Available fields:
  --team         Team responsible (Backend, App, Web, Auth) - auto-transfers to team repo
  --priority     Priority level (Critical, High, Medium, Low)
  --type         Issue type (Epic, User Story, Story, Task, Bug, Feature)
  --no-transfer  Prevent automatic transfer when Team field is set

Examples:
  # Set team field (automatically transfers to Backend repo)
  gh project-management field set #48 --team Backend

  # Set multiple fields (automatically transfers)
  gh project-management field set 48 --team Backend --priority High --type Task

  # Set team but prevent automatic transfer
  gh project-management field set 48 --team Backend --priority High --no-transfer`,
	Args: cobra.ExactArgs(1),
	RunE: runFieldSet,
}

func runFieldSet(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Parse issue reference
	issueRef := args[0]
	owner, repo, issueNumber, err := gh.ParseIssueReference(issueRef, cfg.Owner, cfg.DefaultRepo)
	if err != nil {
		return fmt.Errorf("invalid issue reference: %w", err)
	}

	// Check that at least one field is being set
	if teamValue == "" && priorityValue == "" && typeValue == "" {
		return fmt.Errorf("at least one field must be specified (--team, --priority, or --type)")
	}

	// Get project number
	projectNumber, err := strconv.Atoi(cfg.ProjectID)
	if err != nil {
		return fmt.Errorf("invalid project ID '%s': %w", cfg.ProjectID, err)
	}

	// Get project node ID
	var projectNodeID string
	if cfg.OwnerType == config.OwnerTypeOrg {
		projectNodeID, err = gh.GetProjectNodeID(ctx, cfg.Owner, projectNumber)
	} else {
		projectNodeID, err = gh.GetUserProjectNodeID(ctx, projectNumber)
	}
	if err != nil {
		return fmt.Errorf("failed to get project node ID: %w", err)
	}

	// Get issue node ID
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	issueNodeID, err := gh.GetIssueNodeID(ctx, *client, owner, repo, issueNumber)
	if err != nil {
		return fmt.Errorf("failed to get issue node ID: %w", err)
	}

	// Get project item ID for the issue
	projectItemID, err := gh.GetProjectItemID(ctx, projectNodeID, issueNodeID)
	if err != nil {
		return fmt.Errorf("failed to get project item ID: %w", err)
	}

	// Get project fields
	fields, err := gh.GetProjectFields(ctx, projectNodeID)
	if err != nil {
		return fmt.Errorf("failed to get project fields: %w", err)
	}

	fmt.Printf("Setting custom fields for issue #%d...\n", issueNumber)

	// Set Team field if specified
	if teamValue != "" {
		teamField := gh.FindFieldByName(fields, "Team")
		if teamField == nil {
			return fmt.Errorf("Team field not found in project")
		}

		// Find the option ID for the team value
		var optionID string
		for _, opt := range teamField.Options {
			if opt.Name == teamValue {
				optionID = opt.ID
				break
			}
		}
		if optionID == "" {
			return fmt.Errorf("team value '%s' not found in Team field options", teamValue)
		}

		err = gh.UpdateProjectItemField(ctx, projectNodeID, projectItemID, teamField.ID, optionID)
		if err != nil {
			return fmt.Errorf("failed to set Team field: %w", err)
		}
		fmt.Printf("  ✓ Team: %s\n", teamValue)
	}

	// Set Priority field if specified
	if priorityValue != "" {
		priorityField := gh.FindFieldByName(fields, "Priority")
		if priorityField == nil {
			return fmt.Errorf("Priority field not found in project")
		}

		// Find the option ID for the priority value
		var optionID string
		for _, opt := range priorityField.Options {
			if opt.Name == priorityValue {
				optionID = opt.ID
				break
			}
		}
		if optionID == "" {
			return fmt.Errorf("priority value '%s' not found in Priority field options", priorityValue)
		}

		err = gh.UpdateProjectItemField(ctx, projectNodeID, projectItemID, priorityField.ID, optionID)
		if err != nil {
			return fmt.Errorf("failed to set Priority field: %w", err)
		}
		fmt.Printf("  ✓ Priority: %s\n", priorityValue)
	}

	// Set Type field if specified
	if typeValue != "" {
		typeField := gh.FindFieldByName(fields, "Type")
		if typeField == nil {
			return fmt.Errorf("Type field not found in project")
		}

		// Find the option ID for the type value
		var optionID string
		for _, opt := range typeField.Options {
			if opt.Name == typeValue {
				optionID = opt.ID
				break
			}
		}
		if optionID == "" {
			return fmt.Errorf("type value '%s' not found in Type field options", typeValue)
		}

		err = gh.UpdateProjectItemField(ctx, projectNodeID, projectItemID, typeField.ID, optionID)
		if err != nil {
			return fmt.Errorf("failed to set Type field: %w", err)
		}
		fmt.Printf("  ✓ Type: %s\n", typeValue)
	}

	fmt.Printf("\n✓ Successfully updated custom fields for issue #%d\n", issueNumber)

	// Auto-transfer if team was set and not disabled
	if teamValue != "" && !noTransfer {
		targetRepo, exists := teamRepoMapping[teamValue]
		if !exists {
			fmt.Printf("\n⚠️  Warning: No repository mapping found for team '%s', skipping transfer\n", teamValue)
			return nil
		}

		// Only transfer if we're in the default repo (project-management)
		if repo != cfg.DefaultRepo {
			fmt.Printf("\n⚠️  Warning: Issue is already in %s/%s, not in default repo. Skipping transfer.\n", owner, repo)
			return nil
		}

		fmt.Printf("\n🚀 Auto-transferring to %s/%s based on Team field...\n", cfg.Owner, targetRepo)

		sourceRepo := fmt.Sprintf("%s/%s", cfg.Owner, cfg.DefaultRepo)
		_, err = gh.TransferIssue(ctx, issueNumber, cfg.Owner, targetRepo, sourceRepo)
		if err != nil {
			return fmt.Errorf("failed to auto-transfer issue: %w", err)
		}

		fmt.Printf("\n✓ Successfully transferred issue #%d to %s/%s\n", issueNumber, cfg.Owner, targetRepo)
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("1. Note the new issue number from the output above")
		fmt.Println("2. Update parent issue body with cross-repo reference: ### <owner>/<repo>#<new-number>")
	}

	return nil
}

func init() {
	fieldSetCmd.Flags().StringVar(&teamValue, "team", "", "Team value (Backend, App, Web, Auth) - automatically transfers to team repo")
	fieldSetCmd.Flags().StringVar(&priorityValue, "priority", "", "Priority value (Critical, High, Medium, Low)")
	fieldSetCmd.Flags().StringVar(&typeValue, "type", "", "Type value (Epic, User Story, Story, Task, Bug, Feature)")
	fieldSetCmd.Flags().BoolVar(&noTransfer, "no-transfer", false, "Prevent automatic transfer when Team field is set")

	fieldCmd.AddCommand(fieldSetCmd)
	rootCmd.AddCommand(fieldCmd)
}
