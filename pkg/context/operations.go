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

// EnsureCustomFields ensures Team and Priority custom fields exist in the project
// Uses cached verification status to avoid unnecessary API calls
func EnsureCustomFields(ctx context.Context, projectID string, teams map[string]string) error {
	// Check if fields have already been verified
	verified, err := config.AreCustomFieldsVerified()
	if err != nil {
		return fmt.Errorf("failed to check custom fields verification status: %w", err)
	}

	// If already verified, skip API calls
	if verified {
		return nil
	}

	// Get project node ID from project number
	// Note: projectID in config is the project number, we need the node ID
	var projects []gh.Project
	currentCtx, _, err := GetCurrentContext()
	if err != nil {
		return fmt.Errorf("failed to get current context: %w", err)
	}

	if currentCtx.OwnerType == config.OwnerTypeOrg {
		projects, err = gh.ListOrgProjects(currentCtx.Owner)
	} else {
		projects, err = gh.ListUserProjects()
	}

	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	// Find the project node ID
	var nodeID string
	for _, p := range projects {
		if fmt.Sprintf("%d", p.Number) == projectID {
			nodeID = p.ID
			break
		}
	}

	if nodeID == "" {
		return fmt.Errorf("project #%s not found", projectID)
	}

	// Ensure Team field exists
	if _, err := gh.EnsureTeamField(ctx, nodeID, teams); err != nil {
		return fmt.Errorf("failed to ensure Team field: %w", err)
	}

	// Ensure Priority field exists
	if _, err := gh.EnsurePriorityField(ctx, nodeID); err != nil {
		return fmt.Errorf("failed to ensure Priority field: %w", err)
	}

	// Mark fields as verified
	if err := config.MarkCustomFieldsVerified(); err != nil {
		return fmt.Errorf("failed to mark custom fields as verified: %w", err)
	}

	return nil
}
