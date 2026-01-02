package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/api"
)

// AddSubIssue adds a child issue to a parent issue using GitHub's sub-issues GraphQL API
// Requires the GraphQL-Features: sub_issues header
func AddSubIssue(ctx context.Context, owner, repo string, parentNumber, childNumber int) error {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	// Get issue node IDs for both issues
	parentNodeID, err := GetIssueNodeID(ctx, *client, owner, repo, parentNumber)
	if err != nil {
		return fmt.Errorf("failed to get parent issue node ID: %w", err)
	}

	childNodeID, err := GetIssueNodeID(ctx, *client, owner, repo, childNumber)
	if err != nil {
		return fmt.Errorf("failed to get child issue node ID: %w", err)
	}

	// Add the parent-child relationship using addSubIssue mutation
	mutation := `
		mutation($issueId: ID!, $subIssueId: ID!) {
			addSubIssue(input: {
				issueId: $issueId,
				subIssueId: $subIssueId
			}) {
				issue {
					id
					number
					title
				}
				subIssue {
					id
					number
					title
				}
			}
		}
	`

	variables := map[string]interface{}{
		"issueId":    parentNodeID,
		"subIssueId": childNodeID,
	}

	var response struct {
		AddSubIssue struct {
			Issue struct {
				ID     string `json:"id"`
				Number int    `json:"number"`
				Title  string `json:"title"`
			} `json:"issue"`
			SubIssue struct {
				ID     string `json:"id"`
				Number int    `json:"number"`
				Title  string `json:"title"`
			} `json:"subIssue"`
		} `json:"addSubIssue"`
	}

	// Execute mutation with required header
	err = client.DoWithContext(ctx, mutation, variables, &response)
	if err != nil {
		return fmt.Errorf("failed to add sub-issue relationship: %w", err)
	}

	return nil
}

// RemoveSubIssue removes a child issue from a parent issue using GitHub's sub-issues GraphQL API
func RemoveSubIssue(ctx context.Context, owner, repo string, parentNumber, childNumber int) error {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	// Get issue node IDs for both issues
	parentNodeID, err := GetIssueNodeID(ctx, *client, owner, repo, parentNumber)
	if err != nil {
		return fmt.Errorf("failed to get parent issue node ID: %w", err)
	}

	childNodeID, err := GetIssueNodeID(ctx, *client, owner, repo, childNumber)
	if err != nil {
		return fmt.Errorf("failed to get child issue node ID: %w", err)
	}

	// Remove the parent-child relationship using removeSubIssue mutation
	mutation := `
		mutation($issueId: ID!, $subIssueId: ID!) {
			removeSubIssue(input: {
				issueId: $issueId,
				subIssueId: $subIssueId
			}) {
				issue {
					id
					number
					title
				}
				subIssue {
					id
					number
					title
				}
			}
		}
	`

	variables := map[string]interface{}{
		"issueId":    parentNodeID,
		"subIssueId": childNodeID,
	}

	var response struct {
		RemoveSubIssue struct {
			Issue struct {
				ID     string `json:"id"`
				Number int    `json:"number"`
				Title  string `json:"title"`
			} `json:"issue"`
			SubIssue struct {
				ID     string `json:"id"`
				Number int    `json:"number"`
				Title  string `json:"title"`
			} `json:"subIssue"`
		} `json:"removeSubIssue"`
	}

	// Execute mutation with required header
	err = client.DoWithContext(ctx, mutation, variables, &response)
	if err != nil {
		return fmt.Errorf("failed to remove sub-issue relationship: %w", err)
	}

	return nil
}
