package issue

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Zytera/gh-project-management/internal/config"
	"github.com/Zytera/gh-project-management/internal/gh"
	"github.com/Zytera/gh-project-management/internal/templates"
)

// assignIssueToProject assigns an issue to the configured project
// Returns the project item ID
func assignIssueToProject(ctx context.Context, cfg *config.Config, issue *gh.Issue) (string, error) {
	// Parse project number
	projectNumber, err := strconv.Atoi(cfg.ProjectID)
	if err != nil {
		return "", fmt.Errorf("invalid project ID '%s': %w", cfg.ProjectID, err)
	}

	// Get project node ID
	var projectNodeID string
	if cfg.OwnerType == config.OwnerTypeOrg {
		projectNodeID, err = gh.GetProjectNodeID(ctx, cfg.Owner, projectNumber)
	} else {
		projectNodeID, err = gh.GetUserProjectNodeID(ctx, projectNumber)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get project node ID: %w", err)
	}

	// Add issue to project and get the project item ID
	projectItemID, err := gh.AddIssueToProject(ctx, projectNodeID, issue.ID)
	if err != nil {
		return "", fmt.Errorf("failed to add issue to project: %w", err)
	}

	return projectItemID, nil
}

// CreateDynamicIssueParams contains parameters for creating an issue with a dynamic template
type CreateDynamicIssueParams struct {
	Config    *config.Config
	IssueType string
	Title     string
	Fields    map[string]string
}

// CreateDynamicIssueResult contains the result of creating an issue
type CreateDynamicIssueResult struct {
	Issue         *gh.Issue
	ProjectItemID string
}

// CreateDynamicIssue creates an issue using a dynamic template (from repo or default)
func CreateDynamicIssue(ctx context.Context, params CreateDynamicIssueParams) (*CreateDynamicIssueResult, error) {
	// Get template (from repo or default)
	var template *templates.IssueTemplate
	var err error

	template, _, err = GetTemplate(ctx, params.Config.Owner, params.Config.DefaultRepo, params.IssueType)
	if err != nil {
		return nil, fmt.Errorf("failed to get template for type %s: %w", params.IssueType, err)
	}

	// Validate required fields
	if err := templates.ValidateFields(template, params.Fields); err != nil {
		return nil, fmt.Errorf("field validation failed: %w", err)
	}

	// Build issue body from template
	body, err := templates.BuildBodyFromTemplate(template, params.Fields)
	if err != nil {
		return nil, fmt.Errorf("failed to build issue body: %w", err)
	}

	// Map issue type to GitHub issue type name
	issueTypeName := mapIssueTypeToGitHubType(params.IssueType)

	// Ensure issue type exists and get its ID
	var issueTypeID string
	if params.Config.OwnerType == config.OwnerTypeOrg && issueTypeName != "" {
		// EnsureIssueType will create the type if it doesn't exist
		issueTypeConfig, err := gh.EnsureIssueType(ctx, params.Config.Owner, issueTypeName, fmt.Sprintf("%s issue type", issueTypeName))
		if err != nil {
			// Log warning but continue - issue types might not be available for this org
			fmt.Printf("⚠️  Warning: Could not ensure issue type '%s': %v\n", issueTypeName, err)
		} else if issueTypeConfig != nil {
			issueTypeID = issueTypeConfig.ID
		}
	}

	// Create issue with issue type
	issue, err := gh.CreateIssue(ctx, params.Config.Owner, params.Config.DefaultRepo, params.Title, body, issueTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	// Assign to project and get project item ID
	projectItemID, err := assignIssueToProject(ctx, params.Config, issue)
	if err != nil {
		return nil, fmt.Errorf("failed to assign issue to project: %w", err)
	}

	return &CreateDynamicIssueResult{
		Issue:         issue,
		ProjectItemID: projectItemID,
	}, nil
}

// mapIssueTypeToGitHubType maps template issue type to GitHub issue type name
func mapIssueTypeToGitHubType(issueType string) string {
	switch strings.ToLower(issueType) {
	case "epic":
		return "Epic"
	case "story", "user_story", "user story":
		return "User Story"
	case "task":
		return "Task"
	case "bug":
		return "Bug"
	case "feature":
		return "Feature"
	default:
		// For custom types, capitalize first letter
		if issueType != "" {
			return strings.ToUpper(string(issueType[0])) + issueType[1:]
		}
		return ""
	}
}
