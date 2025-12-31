package issue

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/Zytera/gh-project-management/internal/config"
	"github.com/Zytera/gh-project-management/internal/gh"
	"github.com/Zytera/gh-project-management/internal/git"
	"github.com/Zytera/gh-project-management/internal/templates"
)

// assignIssueToProject assigns an issue to the configured project
func assignIssueToProject(ctx context.Context, cfg *config.Config, issue *gh.Issue) error {
	// Parse project number
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

	// Add issue to project
	_, err = gh.AddIssueToProject(ctx, projectNodeID, issue.ID)
	if err != nil {
		return fmt.Errorf("failed to add issue to project: %w", err)
	}

	return nil
}

// CreateDynamicIssueParams contains parameters for creating an issue with a dynamic template
type CreateDynamicIssueParams struct {
	Config    *config.Config
	IssueType string
	Title     string
	Fields    map[string]string
}

// CreateDynamicIssue creates an issue using a dynamic template (from repo or default)
func CreateDynamicIssue(ctx context.Context, params CreateDynamicIssueParams) (*gh.Issue, error) {

	_, _, err := GetTemplate(ctx, params.Config.Owner, params.Config.DefaultRepo, params.IssueType)
	return nil, err
	// // Get template (from repo or default)
	// var template *templates.IssueTemplate
	// var err error

	// // Try to get template from repo first
	// repoTemplate, _, repoErr := getTemplate(ctx, params.Config.Owner, params.Config.DefaultRepo, params.IssueType)

	// if repoErr == nil && repoTemplate != nil {
	// 	// Use template from repo
	// 	template = repoTemplate
	// } else {
	// 	// Fall back to default template
	// 	template, err = templates.GetDefaultTemplate(params.IssueType)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to get template for type %s: %w", params.IssueType, err)
	// 	}
	// }

	// // Validate required fields
	// if err := templates.ValidateFields(template, params.Fields); err != nil {
	// 	return nil, fmt.Errorf("field validation failed: %w", err)
	// }

	// // Build issue body from template
	// body, err := templates.BuildBodyFromTemplate(template, params.Fields)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to build issue body: %w", err)
	// }

	// // Create issue
	// issue, err := gh.CreateIssue(ctx, params.Config.Owner, params.Config.DefaultRepo, params.Title, body)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to create issue: %w", err)
	// }

	// // Assign to project
	// if err := assignIssueToProject(ctx, params.Config, issue); err != nil {
	// 	return nil, fmt.Errorf("failed to assign issue to project: %w", err)
	// }

	// // TODO: Set Issue Type

	// return issue, nil
}

func GetTemplate(ctx context.Context, owner, repo, issueType string) (*templates.IssueTemplate, string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	repoName, err := git.GetRepoName(currentDir)

	if err != nil {
		return nil, "", fmt.Errorf("failed to get repo name: %w", err)
	}
	println(repo + repoName)
	if repoName == repo {
		fmt.Println("Using templates from current repository: " + repoName)
		return getTemplateFromCurrentDirectory(ctx, owner, repo, issueType)
	}

	return gh.GetTemplateFromRepo(ctx, owner, repo, issueType)
}

func getTemplateFromCurrentDirectory(ctx context.Context, owner, repo, issueType string) (*templates.IssueTemplate, string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	templateDir := currentDir + "/.github/ISSUE_TEMPLATE"
	files, err := os.ReadDir(templateDir)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read template directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		// Check if file has .yml or .yaml extension
		name := file.
		if len(name) > 4 && (name[len(name)-4:] == ".yml" || (len(name) > 5 && name[len(name)-5:] == ".yaml")) {
			// TODO: Parse and match template with issueType
			fmt.Printf("Found template file: %s\n", name)
		}
	}

	return nil, "", fmt.Errorf("template not found in current directory")
}
