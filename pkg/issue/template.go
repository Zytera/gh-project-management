package issue

import (
	"context"
	"fmt"

	"github.com/Zytera/gh-project-management/internal/gh"
	"github.com/Zytera/gh-project-management/internal/templates"
)

func GetTemplate(ctx context.Context, owner string, repo string, issueType string) (*templates.IssueTemplate, string, error) {

	var template *templates.IssueTemplate
	var templateSource string
	var err error

	template, templateSource, err = templates.GetTemplateFromLocalRepo(ctx, owner, repo, issueType)
	if err == nil {
		fmt.Println("Template found in local repository")
		return template, templateSource, nil
	}

	template, _, err = gh.GetTemplateFromRepo(ctx, owner, repo, issueType)
	if err == nil && template != nil {
		return template, fmt.Sprintf("repository (.github/ISSUE_TEMPLATE/%s)", templates.GetTemplateFileName(issueType)), nil
	} else {
		// Fall back to default template
		template, err = templates.GetDefaultTemplate(issueType)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get template for type '%s': %w\n\nAvailable default types: epic, user_story, task, bug, feature", issueType, err)
		}
		templateSource = "default embedded template"
	}

	return template, templateSource, nil
}
