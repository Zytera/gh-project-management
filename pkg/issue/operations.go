package issue

import (
	"context"
	"fmt"

	"github.com/Zytera/gh-project-management/internal/config"
	"github.com/Zytera/gh-project-management/internal/gh"
)

// CreateIssueParams contains parameters for creating an issue
type CreateIssueParams struct {
	Owner    string
	Repo     string
	Title    string
	Template string // epic, user-story, task
}

// CreateEpic creates an epic issue
func CreateEpic(ctx context.Context, cfg *config.Config, title string) (*gh.Issue, error) {
	return gh.CreateIssueWithTemplate(ctx, cfg.Owner, cfg.DefaultRepo, title, gh.TemplateEpic)
}

// CreateUserStory creates a user story issue
func CreateUserStory(ctx context.Context, cfg *config.Config, title string) (*gh.Issue, error) {
	return gh.CreateIssueWithTemplate(ctx, cfg.Owner, cfg.DefaultRepo, title, gh.TemplateUserStory)
}

// CreateTask creates a task issue
func CreateTask(ctx context.Context, cfg *config.Config, title string) (*gh.Issue, error) {
	return gh.CreateIssueWithTemplate(ctx, cfg.Owner, cfg.DefaultRepo, title, gh.TemplateTask)
}

// CreateIssue creates an issue with custom parameters
func CreateIssue(ctx context.Context, params CreateIssueParams) (*gh.Issue, error) {
	if params.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	if params.Owner == "" || params.Repo == "" {
		return nil, fmt.Errorf("owner and repo are required")
	}

	templateName := params.Template
	if templateName == "" {
		templateName = gh.TemplateTask
	}

	return gh.CreateIssueWithTemplate(ctx, params.Owner, params.Repo, params.Title, templateName)
}
