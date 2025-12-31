package gh

import (
	"context"
	"fmt"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
)

// TransferIssue transfers an issue to a different repository using GitHub's GraphQL API
func TransferIssue(ctx context.Context, issueNumber int, targetOwner, targetRepo, sourceRepo string) (int, error) {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return 0, fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	// Parse source repo (format: "owner/repo")
	parts := strings.Split(sourceRepo, "/")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid source repo format '%s', expected 'owner/repo'", sourceRepo)
	}
	sourceOwner := parts[0]
	sourceRepoName := parts[1]

	// Get issue node ID (using existing function from dependencies.go)
	issueNodeID, err := GetIssueNodeID(ctx, *client, sourceOwner, sourceRepoName, issueNumber)
	if err != nil {
		return 0, fmt.Errorf("failed to get issue node ID: %w", err)
	}

	// Get target repository node ID
	repoNodeID, err := getRepositoryNodeID(ctx, *client, targetOwner, targetRepo)
	if err != nil {
		return 0, fmt.Errorf("failed to get repository node ID: %w", err)
	}

	// Transfer the issue
	mutation := `
		mutation($issueId: ID!, $repositoryId: ID!) {
			transferIssue(input: {
				issueId: $issueId
				repositoryId: $repositoryId
			}) {
				issue {
					number
					url
				}
			}
		}
	`

	variables := map[string]interface{}{
		"issueId":      issueNodeID,
		"repositoryId": repoNodeID,
	}

	var response struct {
		TransferIssue struct {
			Issue struct {
				Number int    `json:"number"`
				URL    string `json:"url"`
			} `json:"issue"`
		} `json:"transferIssue"`
	}

	err = client.DoWithContext(ctx, mutation, variables, &response)
	if err != nil {
		return 0, fmt.Errorf("failed to transfer issue: %w", err)
	}

	return response.TransferIssue.Issue.Number, nil
}

// getRepositoryNodeID retrieves the GraphQL node ID for a repository
func getRepositoryNodeID(ctx context.Context, client api.GraphQLClient, owner, repo string) (string, error) {
	query := `
		query($owner: String!, $repo: String!) {
			repository(owner: $owner, name: $repo) {
				id
			}
		}
	`

	variables := map[string]interface{}{
		"owner": owner,
		"repo":  repo,
	}

	var response struct {
		Repository struct {
			ID string `json:"id"`
		} `json:"repository"`
	}

	err := client.DoWithContext(ctx, query, variables, &response)
	if err != nil {
		return "", fmt.Errorf("failed to query repository: %w", err)
	}

	if response.Repository.ID == "" {
		return "", fmt.Errorf("repository %s/%s not found", owner, repo)
	}

	return response.Repository.ID, nil
}
