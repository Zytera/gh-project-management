package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/api"
)

// ListRecentIssues lists recent open issues from a repository
func ListRecentIssues(ctx context.Context, owner, repo string, limit int) ([]Issue, error) {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	query := `
		query($owner: String!, $repo: String!, $limit: Int!) {
			repository(owner: $owner, name: $repo) {
				issues(first: $limit, states: OPEN, orderBy: {field: CREATED_AT, direction: DESC}) {
					nodes {
						id
						number
						title
						url
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"owner": owner,
		"repo":  repo,
		"limit": limit,
	}

	var response struct {
		Repository struct {
			Issues struct {
				Nodes []Issue `json:"nodes"`
			} `json:"issues"`
		} `json:"repository"`
	}

	err = client.DoWithContext(ctx, query, variables, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list issues: %w", err)
	}

	return response.Repository.Issues.Nodes, nil
}
