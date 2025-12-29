package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/api"
)

// ListOrganizations lists all organizations the user belongs to using GraphQL
func ListOrganizations() ([]Organization, error) {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	query := `query { viewer { organizations(first: 100) { nodes { login name } } } }`

	var response struct {
		Viewer struct {
			Organizations struct {
				Nodes []Organization `json:"nodes"`
			} `json:"organizations"`
		} `json:"viewer"`
	}

	err = client.DoWithContext(context.Background(), query, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("error listing organizations: %w", err)
	}

	return response.Viewer.Organizations.Nodes, nil
}
