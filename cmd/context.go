package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/Zytera/gh-project-management/internal/config"
	contextTUI "github.com/Zytera/gh-project-management/internal/tui/context"
	contextPkg "github.com/Zytera/gh-project-management/pkg/context"
	"github.com/spf13/cobra"
)

var (
	// Flags for context add
	contextOwnerType   string
	contextOwner       string
	contextProjectID   string
	contextProjectName string
	contextDefaultRepo string
	contextTeamRepos   []string
)

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Manage project contexts",
	Long:  `Manage project contexts for gh-project-management. Contexts allow you to work with multiple projects.`,
}

var contextListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured contexts",
	RunE:  runContextList,
}

var contextCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Display the current context",
	RunE:  runContextCurrent,
}

var contextUseCmd = &cobra.Command{
	Use:   "use <context-name>",
	Short: "Switch to a different context",
	Args:  cobra.ExactArgs(1),
	RunE:  runContextUse,
}

var contextAddCmd = &cobra.Command{
	Use:   "add <context-name>",
	Short: "Add a new project context",
	Long:  `Add a new project context. Can be used interactively or with flags.`,
	Example: `  # Interactive mode
  gh project-management context add mycontext

  # With flags (organization project)
  gh project-management context add mycontext \
    --owner-type org \
    --owner Zytera \
    --project-id 3 \
    --project-name "Medapsis" \
    --default-repo project-management \
    --team-repos Backend=backend,App=mobile-app,Web=web-app,Auth=auth

  # With flags (personal project)
  gh project-management context add myproject \
    --owner-type user \
    --owner myusername \
    --project-id 1 \
    --project-name "My Project" \
    --default-repo my-repo \
    --team-repos Team1=repo1,Team2=repo2`,
	Args: cobra.ExactArgs(1),
	RunE: runContextAdd,
}

var contextDeleteCmd = &cobra.Command{
	Use:   "delete <context-name>",
	Short: "Delete a context",
	Args:  cobra.ExactArgs(1),
	RunE:  runContextDelete,
}

func init() {
	rootCmd.AddCommand(contextCmd)
	contextCmd.AddCommand(contextListCmd)
	contextCmd.AddCommand(contextCurrentCmd)
	contextCmd.AddCommand(contextUseCmd)
	contextCmd.AddCommand(contextAddCmd)
	contextCmd.AddCommand(contextDeleteCmd)

	// Add flags for context add
	contextAddCmd.Flags().StringVar(&contextOwnerType, "owner-type", "", "Owner type: 'user' or 'org'")
	contextAddCmd.Flags().StringVar(&contextOwner, "owner", "", "Owner name (organization or username)")
	contextAddCmd.Flags().StringVar(&contextProjectID, "project-id", "", "Project ID")
	contextAddCmd.Flags().StringVar(&contextProjectName, "project-name", "", "Project name")
	contextAddCmd.Flags().StringVar(&contextDefaultRepo, "default-repo", "", "Default repository")
	contextAddCmd.Flags().StringSliceVar(&contextTeamRepos, "team-repos", []string{}, "Team repositories in format 'team=repo' (e.g., Backend=backend,App=mobile-app)")
}

// parseTeamRepos converts a slice of "team=repo" strings to a map
func parseTeamRepos(teamRepoSlice []string) (map[string]string, error) {
	teamRepos := make(map[string]string)
	for _, tr := range teamRepoSlice {
		parts := strings.SplitN(tr, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid format '%s', expected 'team=repo'", tr)
		}
		team := strings.TrimSpace(parts[0])
		repo := strings.TrimSpace(parts[1])
		if team == "" || repo == "" {
			return nil, fmt.Errorf("team and repo cannot be empty in '%s'", tr)
		}
		teamRepos[team] = repo
	}
	return teamRepos, nil
}

func runContextList(cmd *cobra.Command, args []string) error {
	globalConfig, err := contextPkg.ListContexts()
	if err != nil {
		return err
	}

	if len(globalConfig.Contexts) == 0 {
		fmt.Println("No contexts configured. Run 'gh project-management context add <name>' to configure your first project.")
		return nil
	}

	fmt.Println("Available contexts:")
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "CURRENT\tNAME\tOWNER\tPROJECT\tDEFAULT REPO")

	for name, ctx := range globalConfig.Contexts {
		current := " "
		if name == globalConfig.CurrentContext {
			current = "*"
		}
		ownerDisplay := ctx.Owner
		if ctx.OwnerType == "user" {
			ownerDisplay = ctx.Owner + " (personal)"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s (#%s)\t%s\n", current, name, ownerDisplay, ctx.ProjectName, ctx.ProjectID, ctx.DefaultRepo)
	}

	w.Flush()
	return nil
}

func runContextCurrent(cmd *cobra.Command, args []string) error {
	ctx, name, err := contextPkg.GetCurrentContext()
	if err != nil {
		fmt.Println("No current context set.")
		return nil
	}

	fmt.Printf("Current context: \033[1m%s\033[0m\n\n", name)
	ownerLabel := "Organization"
	if ctx.OwnerType == "user" {
		ownerLabel = "Owner (personal)"
	}
	fmt.Printf("  %s:    %s\n", ownerLabel, ctx.Owner)
	fmt.Printf("  Project:         %s (#%s)\n", ctx.ProjectName, ctx.ProjectID)
	fmt.Printf("  Default repo:    %s\n", ctx.DefaultRepo)
	fmt.Printf("  Team repos:\n")
	for team, repo := range ctx.TeamRepos {
		fmt.Printf("    %s → %s\n", team, repo)
	}

	return nil
}

func runContextUse(cmd *cobra.Command, args []string) error {
	contextName := args[0]

	if err := contextPkg.SwitchContext(contextName); err != nil {
		return err
	}

	fmt.Printf("✓ Switched to context '%s'\n", contextName)
	return nil
}

func runContextAdd(cmd *cobra.Command, args []string) error {
	contextName := args[0]

	// Check if all required flags are provided for non-interactive mode
	hasAllFlags := contextOwnerType != "" && contextOwner != "" && contextProjectID != "" &&
		contextProjectName != "" && contextDefaultRepo != "" && len(contextTeamRepos) > 0

	if hasAllFlags {
		// Validate and convert owner type
		var ownerType config.OwnerType
		switch contextOwnerType {
		case "user":
			ownerType = config.OwnerTypeUser
		case "org":
			ownerType = config.OwnerTypeOrg
		default:
			return fmt.Errorf("invalid owner-type '%s', must be 'user' or 'org'", contextOwnerType)
		}

		// Parse team repos from flags
		teamRepos, err := parseTeamRepos(contextTeamRepos)
		if err != nil {
			return fmt.Errorf("invalid team-repos format: %w", err)
		}

		// Non-interactive mode with flags
		params := contextPkg.AddContextParams{
			Name:        contextName,
			OwnerType:   ownerType,
			Owner:       contextOwner,
			ProjectID:   contextProjectID,
			ProjectName: contextProjectName,
			DefaultRepo: contextDefaultRepo,
			TeamRepos:   teamRepos,
		}

		if err := contextPkg.AddContext(params); err != nil {
			return err
		}

		fmt.Printf("\n✓ Context '%s' added successfully\n", contextName)
		return nil
	}

	// Interactive mode
	ctx, err := contextTUI.CollectContextConfiguration()
	if err != nil {
		return err
	}

	params := contextPkg.AddContextParams{
		Name:        contextName,
		OwnerType:   ctx.OwnerType,
		Owner:       ctx.Owner,
		ProjectID:   ctx.ProjectID,
		ProjectName: ctx.ProjectName,
		DefaultRepo: ctx.DefaultRepo,
		TeamRepos:   ctx.TeamRepos,
	}

	if err := contextPkg.AddContext(params); err != nil {
		return err
	}

	globalConfig, _ := contextPkg.ListContexts()
	fmt.Printf("\n✓ Context '%s' added successfully\n", contextName)
	if globalConfig.CurrentContext == contextName {
		fmt.Printf("✓ Context '%s' set as current\n", contextName)
	} else {
		fmt.Printf("Use 'gh project-management context use %s' to switch to this context\n", contextName)
	}

	return nil
}

func runContextDelete(cmd *cobra.Command, args []string) error {
	contextName := args[0]

	// Confirm deletion
	confirm, err := contextTUI.ConfirmContextDeletion(contextName)
	if err != nil {
		return err
	}

	if !confirm {
		fmt.Println("Deletion cancelled")
		return nil
	}

	if err := contextPkg.DeleteContext(contextName); err != nil {
		return err
	}

	fmt.Printf("✓ Context '%s' deleted\n", contextName)

	// Suggest another context if available
	globalConfig, _ := contextPkg.ListContexts()
	if len(globalConfig.Contexts) > 0 && globalConfig.CurrentContext == "" {
		for name := range globalConfig.Contexts {
			fmt.Printf("Note: Use 'gh project-management context use %s' to set a new current context\n", name)
			break
		}
	}

	return nil
}
