package context

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Zytera/gh-project-management/internal/config"
	"github.com/Zytera/gh-project-management/internal/gh"
	"github.com/charmbracelet/huh"
)

// CollectContextConfiguration runs an interactive form to collect all context configuration
func CollectContextConfiguration() (*config.Context, error) {
	var (
		ownerType   config.OwnerType
		owner       string
		projectID   string
		projectName string
		defaultRepo string
		teamRepos   = make(map[string]string)
	)

	// Step 1: Select owner type (user or org)
	fmt.Println()

	var ownerTypeSelection int
	ownerTypeForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Project owner type").
				Description("Select where your project is hosted").
				Options(
					huh.NewOption("Personal projects", 0),
					huh.NewOption("Organization projects", 1),
				).
				Value(&ownerTypeSelection),
		),
	)

	if err := ownerTypeForm.Run(); err != nil {
		return nil, fmt.Errorf("error selecting owner type: %w", err)
	}

	var repos []gh.Repository
	var repoNames []string
	var projects []gh.Project

	if ownerTypeSelection == 0 {
		// Personal projects
		ownerType = config.OwnerTypeUser

		fmt.Println()
		fmt.Println("🔍 Getting your username...")
		username, err := gh.GetCurrentUser()
		if err != nil {
			return nil, fmt.Errorf("error getting current user: %w", err)
		}
		owner = username

		// Fetch user repositories
		fmt.Println()
		fmt.Printf("🔍 Fetching your repositories...\n")
		repos, err = gh.ListUserRepositories()
		if err != nil {
			return nil, fmt.Errorf("error fetching repositories: %w", err)
		}
		repoNames = gh.GetRepositoryNames(repos)

		// Fetch user projects
		fmt.Println()
		fmt.Printf("🔍 Fetching your projects...\n")
		projects, err = gh.ListUserProjects()
		if err != nil {
			return nil, fmt.Errorf("error fetching projects: %w", err)
		}

		if len(projects) == 0 {
			return nil, fmt.Errorf("no projects found for user %s", owner)
		}
	} else {
		// Organization projects
		ownerType = config.OwnerTypeOrg

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

		owner = orgs[selectedOrgIndex].Login

		// Fetch organization repositories
		fmt.Println()
		fmt.Printf("🔍 Fetching repositories for %s...\n", owner)
		repos, err = gh.ListOrgRepositories(owner)
		if err != nil {
			return nil, fmt.Errorf("error fetching repositories: %w", err)
		}
		repoNames = gh.GetRepositoryNames(repos)

		// Fetch organization projects
		fmt.Println()
		fmt.Printf("🔍 Fetching projects for %s...\n", owner)
		projects, err = gh.ListOrgProjects(owner)
		if err != nil {
			return nil, fmt.Errorf("error fetching projects: %w", err)
		}

		if len(projects) == 0 {
			return nil, fmt.Errorf("no projects found for organization %s", owner)
		}
	}

	// Step 2: Select project

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
	projectNodeID := selectedProject.ID // GraphQL node ID for API calls

	// Step 3: Default repository
	fmt.Println()

	var selectedDefaultRepoIndex int
	repoNamesOptions := make([]huh.Option[int], len(repoNames))
	for i, p := range repoNames {
		repoNamesOptions[i] = huh.NewOption(p, i)
	}

	repoForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Default repository").
				Description("Repository for Epics/User Stories and issue templates").
				Options(repoNamesOptions...).
				Value(&selectedDefaultRepoIndex),
		),
	)

	defaultRepo = repoNames[selectedDefaultRepoIndex]

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

	// Step 5: Ensure Team custom field exists in the project
	fmt.Println()
	fmt.Println("🔧 Checking Team custom field in project...")

	bgCtx := context.Background()
	teamField, err := gh.EnsureTeamField(bgCtx, projectNodeID, teamRepos)
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to ensure Team custom field: %v\n", err)
		fmt.Println("You may need to create it manually in the project settings.")
	} else {
		if len(teamField.Options) == len(teamRepos) {
			fmt.Printf("✓ Team field '%s' found with %d options\n", teamField.Name, len(teamField.Options))
		} else {
			fmt.Printf("✓ Team field '%s' created with %d options\n", teamField.Name, len(teamField.Options))
		}
		for _, option := range teamField.Options {
			fmt.Printf("  • %s (%s)\n", option.Name, option.Color)
		}
	}

	// Create and validate configuration
	ctx := &config.Context{
		OwnerType:   ownerType,
		Owner:       owner,
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

// ConfirmContextDeletion shows a confirmation dialog for deleting a context
func ConfirmContextDeletion(contextName string) (bool, error) {
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
		return false, fmt.Errorf("error running confirmation: %w", err)
	}

	return confirm, nil
}
