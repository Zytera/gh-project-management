package templates

import (
	"context"
	"fmt"
)

func GetTemplate(ctx context.Context, owner string, repo string, ssueType string) (*templates.IssueTemplate, string) {
	// Try to get template from repo first
	repoTemplate, _, repoErr := template.GetTemplate(ctx, cfg.Owner, cfg.DefaultRepo, issueType)
	if repoErr == nil && repoTemplate != nil {
		template = repoTemplate
		templateSource = fmt.Sprintf("repository (.github/ISSUE_TEMPLATE/%s)", templates.GetTemplateFileName(issueType))
	} else {
		// Fall back to default template
		template, err = templates.GetDefaultTemplate(issueType)
		if err != nil {
			return fmt.Errorf("failed to get template for type '%s': %w\n\nAvailable default types: epic, user_story, task, bug, feature", issueType, err)
		}
		templateSource = "default embedded template"
	}

}
