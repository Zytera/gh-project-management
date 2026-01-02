package templates

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Zytera/gh-project-management/internal/git"
)

//go:embed default/*.yml
var templatesFS embed.FS

// GetDefaultTemplate returns a default template for a given type by reading from embedded files
func GetDefaultTemplate(issueType string) (*IssueTemplate, error) {
	// Map issue type to template file name
	templateFile := GetTemplateFileName(issueType)

	// Read from embedded FS (relative to this package)
	templatePath := filepath.Join("default", templateFile)

	content, err := templatesFS.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("default template not found for type %s: %w", issueType, err)
	}

	return ParseTemplate(content)
}

// GetTemplateFileName returns the template filename for a given issue type
func GetTemplateFileName(issueType string) string {
	switch issueType {
	case "epic", "Epic":
		return "epic.yml"
	case "story", "Story", "user story", "User Story":
		return "user_story.yml"
	case "task", "Task":
		return "task.yml"
	case "bug", "Bug":
		return "bug.yml"
	case "feature", "Feature":
		return "feature.yml"
	default:
		// Normalize custom types: lowercase and replace spaces with underscores
		normalized := strings.ToLower(strings.ReplaceAll(issueType, " ", "_"))
		return normalized + ".yml"
	}
}

func GetTemplateFromLocalRepo(ctx context.Context, owner, repo, issueType string) (*IssueTemplate, string, error) {
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

		fmt.Printf("Found template file: %s\n", file.Name())

		content, err := templatesFS.ReadFile(file.Name())
		if err != nil {
			return nil, "", fmt.Errorf("default template not found for type %s: %w", issueType, err)
		}

		template, err := ParseTemplate(content)
		if err != nil {
			return nil, "", fmt.Errorf("failed to parse template: %w", err)
		}

		return template, file.Name(), nil
	}

	return nil, "", fmt.Errorf("template not found in current directory")
}
