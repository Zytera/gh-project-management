package cmd

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/Zytera/gh-project-management/internal/config"
	"github.com/Zytera/gh-project-management/internal/gh"
	"github.com/Zytera/gh-project-management/internal/styles"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
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
	Args:  cobra.ExactArgs(1),
	RunE:  runContextAdd,
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
}

func runContextList(cmd *cobra.Command, args []string) error {
	globalConfig, err := config.LoadGlobal()
	if err != nil {
		return err
	}

	if len(globalConfig.Contexts) == 0 {
		fmt.Println("No contexts configured. Run 'gh project-management init' to set up your first project.")
		return nil
	}

	fmt.Println("Available contexts:")
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "CURRENT\tNAME\tORGANIZATION\tPROJECT\tDEFAULT REPO")

	for name, ctx := range globalConfig.Contexts {
		current := " "
		if name == globalConfig.CurrentContext {
			current = "*"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s (#%s)\t%s\n", current, name, ctx.Org, ctx.ProjectName, ctx.ProjectID, ctx.DefaultRepo)
	}

	w.Flush()
	return nil
}

func runContextCurrent(cmd *cobra.Command, args []string) error {
	globalConfig, err := config.LoadGlobal()
	if err != nil {
		return err
	}

	if globalConfig.CurrentContext == "" {
		fmt.Println("No current context set.")
		return nil
	}

	ctx, exists := globalConfig.Contexts[globalConfig.CurrentContext]
	if !exists {
		return fmt.Errorf("current context '%s' not found", globalConfig.CurrentContext)
	}

	fmt.Printf("Current context: %s\n\n", styles.HeaderStyle.Render(globalConfig.CurrentContext))
	fmt.Printf("  Organization:    %s\n", ctx.Org)
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

	globalConfig, err := config.LoadGlobal()
	if err != nil {
		return err
	}

	if _, exists := globalConfig.Contexts[contextName]; !exists {
		return fmt.Errorf("context '%s' not found. Use 'gh project-management context list' to see available contexts", contextName)
	}

	globalConfig.CurrentContext = contextName

	if err := config.Save(globalConfig); err != nil {
		return err
	}

	fmt.Printf("✓ Switched to context '%s'\n", contextName)
	return nil
}

func collectContextConfiguration() (*config.Context, error) {
	var (
		org         string
		projectID   string
		projectName string
		defaultRepo string
		teamRepos   = make(map[string]string)
	)

	// Step 1: Fetch and select organization
	fmt.Println()
	fmt.Println("🔍 Fetching your organizations...")

	orgs, err := gh.ListOrganizations()
	if err != nil {
		return nil, fmt.Errorf("error fetching organizations: %w", err)
	}

	if len(orgs) == 0 {
		return nil, fmt.Errorf("no organizations found. You need to belong to at least one organization")
	}

	var selectedOrgIndex int
	orgOptions := make([]huh.Option[int], len(orgs))
	for i, o := range orgs {
		displayName := o.Login
		if o.Name != "" {
			displayName = fmt.Sprintf("%s (%s)", o.Login, o.Name)
		}
		orgOptions[i] = huh.NewOption(displayName, i)
	}

	orgForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Select organization").
				Description("Choose the GitHub organization for this project").
				Options(orgOptions...).
				Value(&selectedOrgIndex),
		),
	)

	if err := orgForm.Run(); err != nil {
		return nil, fmt.Errorf("error selecting organization: %w", err)
	}

	org = orgs[selectedOrgIndex].Login

	// Fetch repositories for autocomplete
	fmt.Println()
	fmt.Printf("🔍 Fetching repositories for %s...\n", org)

	repos, err := gh.ListOrgRepositories(org)
	if err != nil {
		return nil, fmt.Errorf("error fetching repositories: %w", err)
	}

	repoNames := gh.GetRepositoryNames(repos)

	// Step 2: Fetch and select project
	fmt.Println()
	fmt.Printf("🔍 Fetching projects for %s...\n", org)

	projects, err := gh.ListOrgProjects(org)
	if err != nil {
		return nil, fmt.Errorf("error fetching projects: %w", err)
	}

	if len(projects) == 0 {
		return nil, fmt.Errorf("no projects found for organization %s", org)
	}

	var selectedProjectIndex int
	projectOptions := make([]huh.Option[int], len(projects))
	for i, p := range projects {
		projectOptions[i] = huh.NewOption(gh.FormatProjectDisplay(p), i)
	}

	projectForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Select project").
				Description("Choose the GitHub Project to manage").
				Options(projectOptions...).
				Value(&selectedProjectIndex),
		),
	)

	if err := projectForm.Run(); err != nil {
		return nil, fmt.Errorf("error selecting project: %w", err)
	}

	selectedProject := projects[selectedProjectIndex]
	projectID = strconv.Itoa(selectedProject.Number)
	projectName = selectedProject.Title

	// Step 3: Default repository
	fmt.Println()

	repoForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Default repository").
				Description("Repository for Epics/User Stories and issue templates").
				Suggestions(repoNames).
				Value(&defaultRepo).
				Validate(func(s string) error {
					if len(s) == 0 {
						return errors.New("default repository cannot be empty")
					}
					return nil
				}),
		),
	)

	if err := repoForm.Run(); err != nil {
		return nil, fmt.Errorf("error getting default repo: %w", err)
	}

	// Step 4: Team repositories
	fmt.Println()
	fmt.Println("📦 Now let's configure team repositories where tasks will be transferred.")
	fmt.Println()

	addMore := true
	for addMore {
		var teamName, repoName string
		var continueAdding bool

		teamForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Team name").
					Description("Team identifier (e.g., 'Backend', 'App', 'Web', 'Auth')").
					Value(&teamName).
					Validate(func(s string) error {
						if len(s) == 0 {
							return errors.New("team name cannot be empty")
						}
						// Check for duplicate team names (case-insensitive)
						for existingTeam := range teamRepos {
							if strings.EqualFold(existingTeam, s) {
								return fmt.Errorf("team '%s' already exists", existingTeam)
							}
						}
						return nil
					}),

				huh.NewInput().
					Title("Repository name").
					Description("Repository for team tasks").
					Suggestions(repoNames).
					Value(&repoName).
					Validate(func(s string) error {
						if len(s) == 0 {
							return errors.New("repository name cannot be empty")
						}
						return nil
					}),

				huh.NewConfirm().
					Title("Add another team?").
					Value(&continueAdding),
			),
		)

		if err := teamForm.Run(); err != nil {
			return nil, fmt.Errorf("error running team form: %w", err)
		}

		teamRepos[teamName] = repoName
		addMore = continueAdding
	}

	if len(teamRepos) == 0 {
		return nil, fmt.Errorf("at least one team repository is required")
	}

	// Create and validate configuration
	ctx := &config.Context{
		Org:         org,
		ProjectID:   projectID,
		ProjectName: projectName,
		DefaultRepo: defaultRepo,
		TeamRepos:   teamRepos,
	}

	if err := ctx.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return ctx, nil
}

func runContextAdd(cmd *cobra.Command, args []string) error {
	contextName := args[0]

	globalConfig, err := config.LoadGlobal()
	if err != nil {
		return err
	}

	if _, exists := globalConfig.Contexts[contextName]; exists {
		return fmt.Errorf("context '%s' already exists", contextName)
	}

	// Collect configuration via the shared interactive form
	ctx, err := collectContextConfiguration()
	if err != nil {
		return err
	}

	globalConfig.Contexts[contextName] = *ctx

	// If this is the first context, set it as current
	if globalConfig.CurrentContext == "" {
		globalConfig.CurrentContext = contextName
	}

	if err := config.Save(globalConfig); err != nil {
		return fmt.Errorf("error saving configuration: %w", err)
	}

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

	globalConfig, err := config.LoadGlobal()
	if err != nil {
		return err
	}

	if _, exists := globalConfig.Contexts[contextName]; !exists {
		return fmt.Errorf("context '%s' not found", contextName)
	}

	// Confirm deletion
	var confirm bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Delete context '%s'?", contextName)).
				Description("This action cannot be undone").
				Value(&confirm),
		),
	)

	if err := confirmForm.Run(); err != nil {
		return fmt.Errorf("error running confirmation: %w", err)
	}

	if !confirm {
		fmt.Println("Deletion cancelled")
		return nil
	}

	delete(globalConfig.Contexts, contextName)

	// If we deleted the current context, clear it
	if globalConfig.CurrentContext == contextName {
		globalConfig.CurrentContext = ""
		// If there are other contexts, suggest one
		if len(globalConfig.Contexts) > 0 {
			for name := range globalConfig.Contexts {
				fmt.Printf("Note: Current context was deleted. Use 'gh project-management context use %s' to set a new current context\n", name)
				break
			}
		}
	}

	if err := config.Save(globalConfig); err != nil {
		return fmt.Errorf("error saving configuration: %w", err)
	}

	fmt.Printf("✓ Context '%s' deleted\n", contextName)
	return nil
}
