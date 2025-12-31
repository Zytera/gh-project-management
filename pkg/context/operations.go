package context

import (
	"context"
	"fmt"

	"github.com/Zytera/gh-project-management/internal/config"
	"github.com/Zytera/gh-project-management/internal/gh"
)

// AddContextParams contains parameters for adding a new context
type AddContextParams struct {
	Name        string
	OwnerType   config.OwnerType
	Owner       string
	ProjectID   string
	ProjectName string
	DefaultRepo string
	TeamRepos   map[string]string
}

// AddContext adds a new context to the configuration
func AddContext(params AddContextParams) error {
	globalConfig, err := config.LoadGlobal()
	if err != nil {
		return err
	}

	if _, exists := globalConfig.Contexts[params.Name]; exists {
		return fmt.Errorf("context '%s' already exists", params.Name)
	}

	// Create context
	ctx := &config.Context{
		OwnerType:   params.OwnerType,
		Owner:       params.Owner,
		ProjectID:   params.ProjectID,
		ProjectName: params.ProjectName,
		DefaultRepo: params.DefaultRepo,
		TeamRepos:   params.TeamRepos,
	}

	if err := ctx.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Ensure Team and Priority fields exist in the project
	if params.ProjectID != "" && len(params.TeamRepos) > 0 {
		// Get project node ID
		var projects []gh.Project
		if params.OwnerType == config.OwnerTypeOrg {
			projects, err = gh.ListOrgProjects(params.Owner)
		} else {
			projects, err = gh.ListUserProjects()
		}

		if err == nil {
			for _, p := range projects {
				if fmt.Sprintf("%d", p.Number) == params.ProjectID {
					bgCtx := context.Background()
					_, _ = gh.EnsureTeamField(bgCtx, p.ID, params.TeamRepos)
					_, _ = gh.EnsurePriorityField(bgCtx, p.ID)
					break
				}
			}
		}
	}

	globalConfig.Contexts[params.Name] = *ctx

	// If this is the first context, set it as current
	if globalConfig.CurrentContext == "" {
		globalConfig.CurrentContext = params.Name
	}

	return config.Save(globalConfig)
}

// DeleteContext deletes a context from the configuration
func DeleteContext(name string) error {
	globalConfig, err := config.LoadGlobal()
	if err != nil {
		return err
	}

	if _, exists := globalConfig.Contexts[name]; !exists {
		return fmt.Errorf("context '%s' not found", name)
	}

	delete(globalConfig.Contexts, name)

	// If we deleted the current context, clear it
	if globalConfig.CurrentContext == name {
		globalConfig.CurrentContext = ""
	}

	return config.Save(globalConfig)
}

// SwitchContext switches to a different context
func SwitchContext(name string) error {
	globalConfig, err := config.LoadGlobal()
	if err != nil {
		return err
	}

	if _, exists := globalConfig.Contexts[name]; !exists {
		return fmt.Errorf("context '%s' not found", name)
	}

	globalConfig.CurrentContext = name
	return config.Save(globalConfig)
}

// ListContexts returns all configured contexts
func ListContexts() (*config.GlobalConfig, error) {
	return config.LoadGlobal()
}

// GetCurrentContext returns the current context
func GetCurrentContext() (*config.Context, string, error) {
	globalConfig, err := config.LoadGlobal()
	if err != nil {
		return nil, "", err
	}

	if globalConfig.CurrentContext == "" {
		return nil, "", fmt.Errorf("no current context set")
	}

	ctx, exists := globalConfig.Contexts[globalConfig.CurrentContext]
	if !exists {
		return nil, "", fmt.Errorf("current context '%s' not found", globalConfig.CurrentContext)
	}

	return &ctx, globalConfig.CurrentContext, nil
}

// UpdateContextParams contains parameters for updating a context
type UpdateContextParams struct {
	ContextName  string
	ProjectID    *string           // Optional: new project ID
	ProjectName  *string           // Optional: new project name
	DefaultRepo  *string           // Optional: new default repository
	TeamRepos    map[string]string // Optional: teams to add/update (merged with existing)
	ReplaceTeams bool              // If true, replace all teams instead of merging
}

// UpdateContext updates an existing context configuration
// and verifies/creates custom fields if teams are modified
func UpdateContext(params UpdateContextParams) error {
	globalConfig, err := config.LoadGlobal()
	if err != nil {
		return err
	}

	ctx, exists := globalConfig.Contexts[params.ContextName]
	if !exists {
		return fmt.Errorf("context '%s' not found", params.ContextName)
	}

	teamsModified := false

	// Update project ID if provided
	if params.ProjectID != nil {
		ctx.ProjectID = *params.ProjectID
	}

	// Update project name if provided
	if params.ProjectName != nil {
		ctx.ProjectName = *params.ProjectName
	}

	// Update default repo if provided
	if params.DefaultRepo != nil {
		ctx.DefaultRepo = *params.DefaultRepo
	}

	// Update team repos if provided
	if len(params.TeamRepos) > 0 {
		teamsModified = true
		if params.ReplaceTeams {
			// Replace all teams
			ctx.TeamRepos = params.TeamRepos
		} else {
			// Merge new teams with existing ones
			if ctx.TeamRepos == nil {
				ctx.TeamRepos = make(map[string]string)
			}
			for team, repo := range params.TeamRepos {
				ctx.TeamRepos[team] = repo
			}
		}
	}

	// Validate the updated context
	if err := ctx.Validate(); err != nil {
		return fmt.Errorf("invalid configuration after update: %w", err)
	}

	// Verify custom fields if teams were modified
	if teamsModified && ctx.ProjectID != "" && len(ctx.TeamRepos) > 0 {
		// Get project node ID
		var projects []gh.Project
		if ctx.OwnerType == config.OwnerTypeOrg {
			projects, err = gh.ListOrgProjects(ctx.Owner)
		} else {
			projects, err = gh.ListUserProjects()
		}

		if err == nil {
			for _, p := range projects {
				if fmt.Sprintf("%d", p.Number) == ctx.ProjectID {
					bgCtx := context.Background()
					// Ensure Team field has all teams (existing + new)
					if _, err := gh.EnsureTeamField(bgCtx, p.ID, ctx.TeamRepos); err != nil {
						return fmt.Errorf("failed to update Team field: %w", err)
					}
					break
				}
			}
		} else {
			return fmt.Errorf("failed to list projects: %w", err)
		}
	}

	// Save updated context
	globalConfig.Contexts[params.ContextName] = ctx
	return config.Save(globalConfig)
}
