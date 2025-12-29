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
	Org         string
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
		Org:         params.Org,
		ProjectID:   params.ProjectID,
		ProjectName: params.ProjectName,
		DefaultRepo: params.DefaultRepo,
		TeamRepos:   params.TeamRepos,
	}

	if err := ctx.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Ensure Team field exists in the project
	if params.ProjectID != "" && len(params.TeamRepos) > 0 {
		// Get project node ID
		projects, err := gh.ListOrgProjects(params.Org)
		if err == nil {
			for _, p := range projects {
				if fmt.Sprintf("%d", p.Number) == params.ProjectID {
					bgCtx := context.Background()
					_, _ = gh.EnsureTeamField(bgCtx, p.ID, params.TeamRepos)
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
