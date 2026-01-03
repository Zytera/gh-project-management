package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/api"
)

// AddIssueToProject adds an issue to a GitHub Project V2
func AddIssueToProject(ctx context.Context, projectID, issueNodeID string) (string, error) {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return "", fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	mutation := `
		mutation($projectId: ID!, $contentId: ID!) {
			addProjectV2ItemById(input: {
				projectId: $projectId
				contentId: $contentId
			}) {
				item {
					id
				}
			}
		}
	`

	variables := map[string]interface{}{
		"projectId": projectID,
		"contentId": issueNodeID,
	}

	var response struct {
		AddProjectV2ItemById struct {
			Item struct {
				ID string `json:"id"`
			} `json:"item"`
		} `json:"addProjectV2ItemById"`
	}

	err = client.DoWithContext(ctx, mutation, variables, &response)
	if err != nil {
		return "", fmt.Errorf("failed to add issue to project: %w", err)
	}

	return response.AddProjectV2ItemById.Item.ID, nil
}

// GetProjectNodeID gets the node ID of a project by organization and project number
func GetProjectNodeID(ctx context.Context, org string, projectNumber int) (string, error) {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return "", fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	query := `
		query($org: String!, $number: Int!) {
			organization(login: $org) {
				projectV2(number: $number) {
					id
				}
			}
		}
	`

	variables := map[string]interface{}{
		"org":    org,
		"number": projectNumber,
	}

	var response struct {
		Organization struct {
			ProjectV2 struct {
				ID string `json:"id"`
			} `json:"projectV2"`
		} `json:"organization"`
	}

	err = client.DoWithContext(ctx, query, variables, &response)
	if err != nil {
		return "", fmt.Errorf("failed to get project node ID: %w", err)
	}

	return response.Organization.ProjectV2.ID, nil
}

// GetUserProjectNodeID gets the node ID of a user project by project number
func GetUserProjectNodeID(ctx context.Context, projectNumber int) (string, error) {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return "", fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	query := `
		query($number: Int!) {
			viewer {
				projectV2(number: $number) {
					id
				}
			}
		}
	`

	variables := map[string]interface{}{
		"number": projectNumber,
	}

	var response struct {
		Viewer struct {
			ProjectV2 struct {
				ID string `json:"id"`
			} `json:"projectV2"`
		} `json:"viewer"`
	}

	err = client.DoWithContext(ctx, query, variables, &response)
	if err != nil {
		return "", fmt.Errorf("failed to get user project node ID: %w", err)
	}

	return response.Viewer.ProjectV2.ID, nil
}

// GetProjectItemID gets the project item ID for an issue in a project
func GetProjectItemID(ctx context.Context, projectID, issueNodeID string) (string, error) {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return "", fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	query := `
		query($projectId: ID!) {
			node(id: $projectId) {
				... on ProjectV2 {
					items(first: 100) {
						nodes {
							id
							content {
								... on Issue {
									id
								}
							}
						}
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"projectId": projectID,
	}

	var response struct {
		Node struct {
			Items struct {
				Nodes []struct {
					ID      string `json:"id"`
					Content struct {
						ID string `json:"id"`
					} `json:"content"`
				} `json:"nodes"`
			} `json:"items"`
		} `json:"node"`
	}

	err = client.DoWithContext(ctx, query, variables, &response)
	if err != nil {
		return "", fmt.Errorf("failed to query project items: %w", err)
	}

	// Find the item with matching issue ID
	for _, item := range response.Node.Items.Nodes {
		if item.Content.ID == issueNodeID {
			return item.ID, nil
		}
	}

	return "", fmt.Errorf("issue not found in project")
}

// UpdateProjectItemField updates a single-select field value for a project item
func UpdateProjectItemField(ctx context.Context, projectID, itemID, fieldID, optionID string) error {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	mutation := `
		mutation($projectId: ID!, $itemId: ID!, $fieldId: ID!, $optionId: String!) {
			updateProjectV2ItemFieldValue(input: {
				projectId: $projectId
				itemId: $itemId
				fieldId: $fieldId
				value: { singleSelectOptionId: $optionId }
			}) {
				projectV2Item {
					id
				}
			}
		}
	`

	variables := map[string]interface{}{
		"projectId": projectID,
		"itemId":    itemID,
		"fieldId":   fieldID,
		"optionId":  optionID,
	}

	var response struct {
		UpdateProjectV2ItemFieldValue struct {
			ProjectV2Item struct {
				ID string `json:"id"`
			} `json:"projectV2Item"`
		} `json:"updateProjectV2ItemFieldValue"`
	}

	err = client.DoWithContext(ctx, mutation, variables, &response)
	if err != nil {
		return fmt.Errorf("failed to update field value: %w", err)
	}

	return nil
}
