package templates

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Zytera/gh-project-management/internal/gh"
	"github.com/Zytera/gh-project-management/internal/git"
)

func GetTemplate(ctx context.Context, owner string, repo string, issueType string) (*IssueTemplate, string, error) {

	var template *IssueTemplate
	var templateSource string
	var err error

	template, templateSource, err = getTemplateFromLocalRepo(ctx, owner, repo, issueType)
	if err == nil {
		return template, templateSource, nil
	}

	template, _, err = gh.GetTemplateFromRepo(ctx, owner, repo, issueType)
	if err == nil && template != nil {
		return template, fmt.Sprintf("repository (.github/ISSUE_TEMPLATE/%s)", GetTemplateFileName(issueType)), nil
	} else {
		// Fall back to default template
		template, err = GetDefaultTemplate(issueType)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get template for type '%s': %w\n\nAvailable default types: epic, user_story, task, bug, feature", issueType, err)
		}
		templateSource = "default embedded template"
	}

	return template, templateSource, nil
}

func getTemplateFromLocalRepo(ctx context.Context, owner, repo, issueType string) (*IssueTemplate, string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get current working directory: %w", err)
	}
	repoName, err := git.GetRepoName(currentDir)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get repo name: %w", err)
	}

	if repoName == repo {
		fmt.Println("Using templates from current repository: " + repoName)
		return getTemplateFromCurrentDirectory(ctx, owner, repo, issueType)
	}
	return nil, "", fmt.Errorf("failed to get template from local repository")
}

func getTemplateFromCurrentDirectory(ctx context.Context, owner, repo, issueType string) (*IssueTemplate, string, error) {
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
		ext := filepath.Ext(file.Name())
		if ext != ".yml" && ext != ".yaml" {
			continue
		}
		// TODO: Parse and match template with issueType
		fmt.Printf("Found template file: %s\n", file.Name())
	}

	return nil, "", fmt.Errorf("template not found in current directory")
}
