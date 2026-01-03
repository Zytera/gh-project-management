package gh

import (
	"context"
	"fmt"
	"time"

	"github.com/Zytera/gh-project-management/internal/templates"
	"github.com/cli/go-gh/v2/pkg/api"
)

// newIssueTypesClient creates a GraphQL client with the required headers for issue types API
func newIssueTypesClient() (*api.GraphQLClient, error) {
	opts := api.ClientOptions{
		Headers: map[string]string{
			"GraphQL-Features": "issue_types",
		},
		Timeout: 30 * time.Second,
	}

	client, err := api.NewGraphQLClient(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client with issue types headers: %w", err)
	}

	return client, nil
}

// ListOrgIssueTypes lists all issue types for an organization
func ListOrgIssueTypes(ctx context.Context, org string) ([]templates.IssueTypeConfig, error) {
	client, err := newIssueTypesClient()
	if err != nil {
		return nil, err
	}

	query := `
		query($org: String!) {
			organization(login: $org) {
				issueTypes(first: 25) {
					edges {
						node {
							id
							name
							description
							isEnabled
						}
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"org": org,
	}

	var response struct {
		Organization struct {
			IssueTypes struct {
				Edges []struct {
					Node templates.IssueTypeConfig `json:"node"`
				} `json:"edges"`
			} `json:"issueTypes"`
		} `json:"organization"`
	}

	err = client.DoWithContext(ctx, query, variables, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list issue types: %w", err)
	}

	var issueTypes []templates.IssueTypeConfig
	for _, edge := range response.Organization.IssueTypes.Edges {
		issueTypes = append(issueTypes, edge.Node)
	}

	return issueTypes, nil
}

// CreateIssueType creates a new issue type in the organization
func CreateIssueType(ctx context.Context, orgID, name, description string) (*templates.IssueTypeConfig, error) {
	client, err := newIssueTypesClient()
	if err != nil {
		return nil, err
	}

	mutation := `
		mutation($orgId: ID!, $name: String!, $description: String) {
			createIssueType(input: {
				organizationId: $orgId
				name: $name
				description: $description
			}) {
				issueType {
					id
					name
					description
					isEnabled
				}
			}
		}
	`

	variables := map[string]interface{}{
		"orgId":       orgID,
		"name":        name,
		"description": description,
	}

	var response struct {
		CreateIssueType struct {
			IssueType templates.IssueTypeConfig `json:"issueType"`
		} `json:"createIssueType"`
	}

	err = client.DoWithContext(ctx, mutation, variables, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue type: %w", err)
	}

	return &response.CreateIssueType.IssueType, nil
}

// GetOrgNodeID gets the organization's node ID
func GetOrgNodeID(ctx context.Context, org string) (string, error) {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return "", fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	query := `
		query($org: String!) {
			organization(login: $org) {
				id
			}
		}
	`

	variables := map[string]interface{}{
		"org": org,
	}

	var response struct {
		Organization struct {
			ID string `json:"id"`
		} `json:"organization"`
	}

	err = client.DoWithContext(ctx, query, variables, &response)
	if err != nil {
		return "", fmt.Errorf("failed to get organization node ID: %w", err)
	}

	return response.Organization.ID, nil
}

// EnsureIssueType ensures an issue type exists, creates it if it doesn't
func EnsureIssueType(ctx context.Context, org, issueTypeName, description string) (*templates.IssueTypeConfig, error) {
	// First, try to list existing issue types
	existingTypes, err := ListOrgIssueTypes(ctx, org)
	if err != nil {
		// If listing fails, issue types might not be available for this org
		// Return nil without error (we'll skip issue type assignment)
		return nil, nil
	}

	// Check if type already exists
	for _, issueType := range existingTypes {
		if issueType.Name == issueTypeName {
			return &issueType, nil
		}
	}

	// Type doesn't exist, create it
	orgID, err := GetOrgNodeID(ctx, org)
	if err != nil {
		return nil, err
	}

	return CreateIssueType(ctx, orgID, issueTypeName, description)
}

// SyncIssueTypesWithTemplates ensures issue types exist for all templates
func SyncIssueTypesWithTemplates(ctx context.Context, org string, templatesMap map[string]*templates.IssueTemplate) error {
	for typeName, template := range templatesMap {
		_, err := EnsureIssueType(ctx, org, typeName, template.Description)
		if err != nil {
			return fmt.Errorf("failed to ensure issue type %s: %w", typeName, err)
		}
	}
	return nil
}

// GetIssueTypeIDByName gets the ID of an issue type by its name
// Returns empty string if not found or if issue types are not available
func GetIssueTypeIDByName(ctx context.Context, org, typeName string) (string, error) {
	// Try to list existing issue types
	existingTypes, err := ListOrgIssueTypes(ctx, org)
	if err != nil {
		// If listing fails, issue types might not be available for this org
		return "", nil
	}

	// Search for the type by name
	for _, issueType := range existingTypes {
		if issueType.Name == typeName && issueType.IsEnabled {
			return issueType.ID, nil
		}
	}

	// Type not found
	return "", nil
}
