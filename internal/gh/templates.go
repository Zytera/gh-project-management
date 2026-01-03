package gh

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/Zytera/gh-project-management/internal/templates"
	"github.com/cli/go-gh/v2/pkg/api"
)

// GetTemplateFromRepo fetches a template file from the repository
func GetTemplateFromRepo(ctx context.Context, owner, repo, issueType string) (*templates.IssueTemplate, string, error) {
	client, err := api.DefaultRESTClient()
	if err != nil {
		return nil, "", fmt.Errorf("failed to create REST client: %w", err)
	}

	// Get template file name
	templateFile := templates.GetTemplateFileName(issueType)
	path := fmt.Sprintf(".github/ISSUE_TEMPLATE/%s", templateFile)

	// Fetch file content from repo
	var response struct {
		Content string `json:"content"`
		SHA     string `json:"sha"`
	}

	err = client.Get(fmt.Sprintf("repos/%s/%s/contents/%s", owner, repo, path), &response)
	if err != nil {
		// File doesn't exist, return nil (will use default template)
		return nil, "", nil
	}

	// Decode base64 content
	content, err := base64.StdEncoding.DecodeString(response.Content)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode template content: %w", err)
	}

	// Parse template
	template, err := templates.ParseTemplate(content)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse template from repo: %w", err)
	}

	template.LastUpdated = response.SHA
	return template, response.SHA, nil
}
