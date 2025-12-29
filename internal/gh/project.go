package gh

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/api"
)

// ListOrgProjects lists projects owned by an organization using GraphQL
func ListOrgProjects(org string) ([]Project, error) {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	query := `query($owner: String!) { organization(login: $owner) { projectsV2(first: 100) { nodes { id number title } } } }`

	variables := map[string]interface{}{
		"owner": org,
	}

	var response struct {
		Organization struct {
			ProjectsV2 struct {
				Nodes []Project `json:"nodes"`
			} `json:"projectsV2"`
		} `json:"organization"`
	}

	err = client.DoWithContext(context.Background(), query, variables, &response)
	if err != nil {
		return nil, fmt.Errorf("error listing projects for org %s: %w", org, err)
	}

	projects := response.Organization.ProjectsV2.Nodes

	// Mark owner
	for i := range projects {
		projects[i].Owner = org
	}

	return projects, nil
}

// ListUserProjects lists projects owned by the authenticated user using GraphQL
func ListUserProjects() ([]Project, error) {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	query := `query { viewer { projectsV2(first: 100) { nodes { id number title } } } }`

	var response struct {
		Viewer struct {
			ProjectsV2 struct {
				Nodes []Project `json:"nodes"`
			} `json:"projectsV2"`
		} `json:"viewer"`
	}

	err = client.DoWithContext(context.Background(), query, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("error listing user projects: %w", err)
	}

	projects := response.Viewer.ProjectsV2.Nodes

	// Mark as user projects
	for i := range projects {
		projects[i].Owner = "user"
	}

	return projects, nil
}

// FormatProjectDisplay returns a display string for a project
func FormatProjectDisplay(p Project) string {
	owner := p.Owner
	if owner == "user" {
		owner = "Personal"
	}
	return fmt.Sprintf("[%s] %s (#%d)", owner, p.Title, p.Number)
}

// GetProjectFields retrieves all fields for a project
func GetProjectFields(ctx context.Context, projectID string) ([]Field, error) {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	query := `
		query($projectId: ID!) {
			node(id: $projectId) {
				... on ProjectV2 {
					fields(first: 100) {
						nodes {
							... on ProjectV2FieldCommon {
								id
								name
							}
							... on ProjectV2SingleSelectField {
								id
								name
								options {
									id
									name
									color
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
			Fields struct {
				Nodes []json.RawMessage `json:"nodes"`
			} `json:"fields"`
		} `json:"node"`
	}

	err = client.DoWithContext(ctx, query, variables, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to query project fields: %w", err)
	}

	// Parse fields from raw messages
	fields := make([]Field, 0)
	for _, raw := range response.Node.Fields.Nodes {
		var field Field
		if err := json.Unmarshal(raw, &field); err != nil {
			continue // Skip fields we can't parse
		}
		fields = append(fields, field)
	}

	return fields, nil
}

// FindFieldByName finds a field by its name
func FindFieldByName(fields []Field, name string) *Field {
	for _, field := range fields {
		if field.Name == name {
			return &field
		}
	}
	return nil
}

// CreateSingleSelectField creates a new single-select field with options
func CreateSingleSelectField(ctx context.Context, projectID, fieldName string, options map[string]FieldColor) (*Field, error) {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	// Build options array
	singleSelectOptions := make([]map[string]string, 0, len(options))
	for name, color := range options {
		singleSelectOptions = append(singleSelectOptions, map[string]string{
			"name":  name,
			"color": string(color),
		})
	}

	mutation := `
		mutation($projectId: ID!, $name: String!, $dataType: ProjectV2CustomFieldType!, $singleSelectOptions: [ProjectV2SingleSelectFieldOptionInput!]) {
			createProjectV2Field(input: {
				projectId: $projectId,
				name: $name,
				dataType: $dataType,
				singleSelectOptions: $singleSelectOptions
			}) {
				projectV2Field {
					... on ProjectV2SingleSelectField {
						id
						name
						options {
							id
							name
							color
						}
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"projectId":           projectID,
		"name":                fieldName,
		"dataType":            "SINGLE_SELECT",
		"singleSelectOptions": singleSelectOptions,
	}

	var response struct {
		CreateProjectV2Field struct {
			ProjectV2Field Field `json:"projectV2Field"`
		} `json:"createProjectV2Field"`
	}

	err = client.DoWithContext(ctx, mutation, variables, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to create field: %w", err)
	}

	return &response.CreateProjectV2Field.ProjectV2Field, nil
}

// EnsureTeamField checks if the Team field exists, and creates it if it doesn't
func EnsureTeamField(ctx context.Context, projectID string, teams map[string]string) (*Field, error) {
	// First, check if field exists
	fields, err := GetProjectFields(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project fields: %w", err)
	}

	// Look for existing Team field
	existingField := FindFieldByName(fields, "Team")
	if existingField != nil {
		return existingField, nil
	}

	// Create Team field with team names as options
	teamOptions := make(map[string]FieldColor)
	colorIndex := 0
	for teamName := range teams {
		teamOptions[teamName] = DefaultTeamColors[colorIndex%len(DefaultTeamColors)]
		colorIndex++
	}

	newField, err := CreateSingleSelectField(ctx, projectID, "Team", teamOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create Team field: %w", err)
	}

	return newField, nil
}
