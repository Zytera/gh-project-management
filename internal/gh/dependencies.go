package gh

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cli/go-gh/v2/pkg/api"
)

// AddBlockedBy establishes a dependency where blockedIssue is blocked by blockingIssue
func AddBlockedBy(ctx context.Context, owner, repo string, blockedIssueNumber, blockingIssueNumber int) error {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	// Get issue node IDs for both issues
	blockedIssueNodeID, err := GetIssueNodeID(ctx, *client, owner, repo, blockedIssueNumber)
	if err != nil {
		return fmt.Errorf("failed to get blocked issue node ID: %w", err)
	}

	blockingIssueNodeID, err := GetIssueNodeID(ctx, *client, owner, repo, blockingIssueNumber)
	if err != nil {
		return fmt.Errorf("failed to get blocking issue node ID: %w", err)
	}

	// Add the blocked-by relationship
	mutation := `
		mutation($issueId: ID!, $blockingIssueId: ID!) {
			addBlockedBy(input: {
				issueId: $issueId,
				blockingIssueId: $blockingIssueId
			}) {
				issue {
					id
					number
					title
				}
			}
		}
	`

	variables := map[string]interface{}{
		"issueId":         blockedIssueNodeID,
		"blockingIssueId": blockingIssueNodeID,
	}

	var response struct {
		AddBlockedBy struct {
			Issue struct {
				ID     string `json:"id"`
				Number int    `json:"number"`
				Title  string `json:"title"`
			} `json:"issue"`
		} `json:"addBlockedBy"`
	}

	err = client.DoWithContext(ctx, mutation, variables, &response)
	if err != nil {
		return fmt.Errorf("failed to add blocked-by relationship: %w", err)
	}

	return nil
}

// GetIssueNodeID retrieves the GraphQL node ID for an issue by its number
func GetIssueNodeID(ctx context.Context, client api.GraphQLClient, owner, repo string, issueNumber int) (string, error) {
	query := `
		query($owner: String!, $repo: String!, $number: Int!) {
			repository(owner: $owner, name: $repo) {
				issue(number: $number) {
					id
				}
			}
		}
	`

	variables := map[string]interface{}{
		"owner":  owner,
		"repo":   repo,
		"number": issueNumber,
	}

	var response struct {
		Repository struct {
			Issue struct {
				ID string `json:"id"`
			} `json:"issue"`
		} `json:"repository"`
	}

	err := client.DoWithContext(ctx, query, variables, &response)
	if err != nil {
		return "", fmt.Errorf("failed to query issue #%d: %w", issueNumber, err)
	}

	if response.Repository.Issue.ID == "" {
		return "", fmt.Errorf("issue #%d not found in %s/%s", issueNumber, owner, repo)
	}

	return response.Repository.Issue.ID, nil
}

// ParseIssueReference parses an issue reference in format "owner/repo#number" or just "#number"
// Returns owner, repo, and issue number
func ParseIssueReference(ref string, defaultOwner, defaultRepo string) (string, string, int, error) {
	// Handle format: "#123" or "123"
	if len(ref) > 0 && ref[0] == '#' {
		num, err := strconv.Atoi(ref[1:])
		if err != nil {
			return "", "", 0, fmt.Errorf("invalid issue number in '%s': %w", ref, err)
		}
		return defaultOwner, defaultRepo, num, nil
	}

	// Handle format: "owner/repo#123"
	// Find the '#' separator
	parts := splitOnce(ref, '#')
	if len(parts) != 2 {
		// Try parsing as just a number
		num, err := strconv.Atoi(ref)
		if err != nil {
			return "", "", 0, fmt.Errorf("invalid issue reference '%s': expected format 'owner/repo#number' or '#number'", ref)
		}
		return defaultOwner, defaultRepo, num, nil
	}

	// Split owner/repo
	repoPath := parts[0]
	numStr := parts[1]

	repoParts := splitOnce(repoPath, '/')
	if len(repoParts) != 2 {
		return "", "", 0, fmt.Errorf("invalid repository path in '%s': expected 'owner/repo'", ref)
	}

	owner := repoParts[0]
	repo := repoParts[1]

	num, err := strconv.Atoi(numStr)
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid issue number in '%s': %w", ref, err)
	}

	return owner, repo, num, nil
}

// splitOnce splits a string on the first occurrence of sep
func splitOnce(s string, sep rune) []string {
	for i, c := range s {
		if c == sep {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}
