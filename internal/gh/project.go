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

// AddOptionsToField adds new options to an existing single-select field
func AddOptionsToField(ctx context.Context, fieldID string, options map[string]FieldColor) error {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	// Build options array
	singleSelectOptions := make([]map[string]string, 0, len(options))
	for name, color := range options {
		singleSelectOptions = append(singleSelectOptions, map[string]string{
			"name":        name,
			"color":       string(color),
			"description": "",
		})
	}

	mutation := `
		mutation($fieldId: ID!, $singleSelectOptions: [ProjectV2SingleSelectFieldOptionInput!]) {
			updateProjectV2Field(input: {
				fieldId: $fieldId,
				singleSelectOptions: $singleSelectOptions
			}) {
				projectV2Field {
					... on ProjectV2SingleSelectField {
						id
						name
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"fieldId":             fieldID,
		"singleSelectOptions": singleSelectOptions,
	}

	var response struct {
		UpdateProjectV2Field struct {
			ProjectV2Field Field `json:"projectV2Field"`
		} `json:"updateProjectV2Field"`
	}

	err = client.DoWithContext(ctx, mutation, variables, &response)
	if err != nil {
		return fmt.Errorf("failed to update field options: %w", err)
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
			"name":        name,
			"color":       string(color),
			"description": "",
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

// EnsureTeamField checks if the Team field exists, creates it if it doesn't,
// and ensures all team options are present
func EnsureTeamField(ctx context.Context, projectID string, teams map[string]string) (*Field, error) {
	// First, check if field exists
	fields, err := GetProjectFields(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project fields: %w", err)
	}

	// Look for existing Team field
	existingField := FindFieldByName(fields, "Team")

	// Prepare team options
	teamOptions := make(map[string]FieldColor)
	colorIndex := 0
	for teamName := range teams {
		teamOptions[teamName] = DefaultTeamColors[colorIndex%len(DefaultTeamColors)]
		colorIndex++
	}

	// If field doesn't exist, create it
	if existingField == nil {
		newField, err := CreateSingleSelectField(ctx, projectID, "Team", teamOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to create Team field: %w", err)
		}
		return newField, nil
	}

	// Field exists, check if all team options are present
	existingOptions := make(map[string]bool)
	for _, opt := range existingField.Options {
		existingOptions[opt.Name] = true
	}

	// Find missing options
	missingOptions := make(map[string]FieldColor)
	for teamName, color := range teamOptions {
		if !existingOptions[teamName] {
			missingOptions[teamName] = color
		}
	}

	// If there are missing options, add them
	if len(missingOptions) > 0 {
		// We need to include ALL options (existing + new) when updating
		allOptions := make(map[string]FieldColor)

		// Add existing options (preserve their colors)
		for _, opt := range existingField.Options {
			if opt.Color != "" {
				allOptions[opt.Name] = opt.Color
			} else {
				allOptions[opt.Name] = ColorGray
			}
		}

		// Add new options
		for name, color := range missingOptions {
			allOptions[name] = color
		}

		if err := AddOptionsToField(ctx, existingField.ID, allOptions); err != nil {
			return nil, fmt.Errorf("failed to add missing team options: %w", err)
		}

		// Refresh field data
		fields, err = GetProjectFields(ctx, projectID)
		if err != nil {
			return nil, fmt.Errorf("failed to refresh project fields: %w", err)
		}
		existingField = FindFieldByName(fields, "Team")
	}

	return existingField, nil
}

// EnsurePriorityField checks if the Priority field exists, creates it if it doesn't,
// and ensures all priority options are present with correct values
func EnsurePriorityField(ctx context.Context, projectID string) (*Field, error) {
	// First, check if field exists
	fields, err := GetProjectFields(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project fields: %w", err)
	}

	// Look for existing Priority field
	existingField := FindFieldByName(fields, "Priority")

	// If field doesn't exist, create it
	if existingField == nil {
		newField, err := CreateSingleSelectField(ctx, projectID, "Priority", PriorityLevels)
		if err != nil {
			return nil, fmt.Errorf("failed to create Priority field: %w", err)
		}
		return newField, nil
	}

	// Field exists, check if all priority options are present
	existingOptions := make(map[string]bool)
	for _, opt := range existingField.Options {
		existingOptions[opt.Name] = true
	}

	// Find missing options
	missingOptions := make(map[string]FieldColor)
	for priorityName, color := range PriorityLevels {
		if !existingOptions[priorityName] {
			missingOptions[priorityName] = color
		}
	}

	// If there are missing options, add them
	if len(missingOptions) > 0 {
		// We need to include ALL options (existing + new) when updating
		allOptions := make(map[string]FieldColor)

		// Add existing options (preserve their colors if they match our expected priorities)
		for _, opt := range existingField.Options {
			if expectedColor, exists := PriorityLevels[opt.Name]; exists {
				// This is a known priority, use expected color
				allOptions[opt.Name] = expectedColor
			} else {
				// Unknown priority option, preserve it with its existing color
				if opt.Color != "" {
					allOptions[opt.Name] = opt.Color
				} else {
					allOptions[opt.Name] = ColorGray
				}
			}
		}

		// Add missing priority options
		for name, color := range missingOptions {
			allOptions[name] = color
		}

		if err := AddOptionsToField(ctx, existingField.ID, allOptions); err != nil {
			return nil, fmt.Errorf("failed to add missing priority options: %w", err)
		}

		// Refresh field data
		fields, err = GetProjectFields(ctx, projectID)
		if err != nil {
			return nil, fmt.Errorf("failed to refresh project fields: %w", err)
		}
		existingField = FindFieldByName(fields, "Priority")
	}

	return existingField, nil
}
