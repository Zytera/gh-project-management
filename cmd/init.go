package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Zytera/gh-project-management/internal/config"
	"github.com/Zytera/gh-project-management/internal/gh"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize gh-project-management configuration",
	Long:  `Interactive setup to configure your first project context.`,
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	// Check if config already exists
	globalConfig, err := config.LoadGlobal()
	if err != nil {
		return err
	}

	if len(globalConfig.Contexts) > 0 {
		fmt.Println("⚠️  Configuration already exists. Use 'gh project-management context add' to add a new project.")
		return nil
	}

	fmt.Println("⚙️  No configuration found. Let's set up your first project!")
	fmt.Println()

	// Collect configuration via interactive form
	var (
		contextName string
		org         string
		projectID   string
		projectName string
		defaultRepo string
		teamRepos   = make(map[string]string)
	)

	// Step 1: Context name
	contextForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Context name").
				Description("A short name for this project configuration (e.g., 'project-management')").
				Value(&contextName).
				Validate(func(s string) error {
					if len(s) == 0 {
						return errors.New("context name cannot be empty")
					}
					return nil
				}),
		),
	)

	if err := contextForm.Run(); err != nil {
		return fmt.Errorf("error running form: %w", err)
	}

	// Step 2: Fetch and select organization
	fmt.Println()
	fmt.Println("🔍 Fetching your organizations...")

	orgs, err := gh.ListOrganizations()
	if err != nil {
		return fmt.Errorf("error fetching organizations: %w", err)
	}

	if len(orgs) == 0 {
		return fmt.Errorf("no organizations found. You need to belong to at least one organization")
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
		return fmt.Errorf("error selecting organization: %w", err)
	}

	org = orgs[selectedOrgIndex].Login

	// Fetch repositories for autocomplete
	fmt.Println()
	fmt.Printf("🔍 Fetching repositories for %s...\n", org)

	repos, err := gh.ListOrgRepositories(org)
	if err != nil {
		return fmt.Errorf("error fetching repositories: %w", err)
	}

	repoNames := gh.GetRepositoryNames(repos)

	// Step 3: Fetch and select project
	fmt.Println()
	fmt.Printf("🔍 Fetching projects for %s...\n", org)

	projects, err := gh.ListOrgProjects(org)
	if err != nil {
		return fmt.Errorf("error fetching projects: %w", err)
	}

	if len(projects) == 0 {
		return fmt.Errorf("no projects found for organization %s", org)
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
		return fmt.Errorf("error selecting project: %w", err)
	}

	selectedProject := projects[selectedProjectIndex]
	projectID = strconv.Itoa(selectedProject.Number)
	projectName = selectedProject.Title

	// Step 4: Default repository
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
		return fmt.Errorf("error getting default repo: %w", err)
	}

	// Step 2: Team repositories
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
			return fmt.Errorf("error running team form: %w", err)
		}

		teamRepos[teamName] = repoName
		addMore = continueAdding
	}

	if len(teamRepos) == 0 {
		return fmt.Errorf("at least one team repository is required")
	}

	// Create and save configuration
	ctx := config.Context{
		Org:         org,
		ProjectID:   projectID,
		ProjectName: projectName,
		DefaultRepo: defaultRepo,
		TeamRepos:   teamRepos,
	}

	if err := ctx.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	globalConfig.Contexts[contextName] = ctx
	globalConfig.CurrentContext = contextName

	if err := config.Save(globalConfig); err != nil {
		return fmt.Errorf("error saving configuration: %w", err)
	}

	configPath, _ := config.GetConfigPath()
	fmt.Printf("\n✓ Configuration saved to %s\n", configPath)
	fmt.Printf("✓ Context '%s' set as current\n\n", contextName)
	fmt.Println("You can now use commands like:")
	fmt.Println("  gh project-management create-epic")
	fmt.Println("  gh project-management create-user-story")
	fmt.Println("  gh project-management create-task")

	return nil
}
