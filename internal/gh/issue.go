package gh

import (
	"context"
	"errors"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/api"
)

// Issue represents a GitHub issue
type Issue struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
	URL    string `json:"url"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

// CreateIssue creates an issue in the specified repository and returns the issue URL
func CreateIssue(ctx context.Context, owner, repo, title, body string) (*Issue, error) {
	if owner == "" || repo == "" {
		return nil, errors.New("owner and repo cannot be empty")
	}
	if title == "" {
		return nil, errors.New("title cannot be empty")
	}

	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	// 1) Get repository ID
	var repoQuery struct {
		Repository struct {
			ID string `json:"id"`
		} `json:"repository"`
	}

	repoQueryQL := `
		query RepoID($owner: String!, $name: String!) {
			repository(owner: $owner, name: $name) {
				id
			}
		}
	`

	vars := map[string]interface{}{
		"owner": owner,
		"name":  repo,
	}

	if err := client.DoWithContext(ctx, repoQueryQL, vars, &repoQuery); err != nil {
		return nil, fmt.Errorf("failed to query repository id: %w", err)
	}

	if repoQuery.Repository.ID == "" {
		return nil, errors.New("repository not found or has no id")
	}

	repoID := repoQuery.Repository.ID

	// 2) Create the issue
	var createResp struct {
		CreateIssue struct {
			Issue Issue `json:"issue"`
		} `json:"createIssue"`
	}

	createIssueQL := `
		mutation CreateIssue($input: CreateIssueInput!) {
			createIssue(input: $input) {
				issue {
					id
					number
					url
					title
					body
				}
			}
		}
	`

	vars = map[string]interface{}{
		"input": map[string]interface{}{
			"repositoryId": repoID,
			"title":        title,
			"body":         body,
		},
	}

	if err := client.DoWithContext(ctx, createIssueQL, vars, &createResp); err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	if createResp.CreateIssue.Issue.URL == "" {
		return nil, errors.New("issue created but response missing URL")
	}

	return &createResp.CreateIssue.Issue, nil
}
