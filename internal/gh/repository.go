package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/api"
)

// ListOrgRepositories lists repositories for an organization using GraphQL
func ListOrgRepositories(org string) ([]Repository, error) {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	query := `query($owner: String!) { organization(login: $owner) { repositories(first: 100, orderBy: {field: UPDATED_AT, direction: DESC}) { nodes { name description } } } }`

	variables := map[string]interface{}{
		"owner": org,
	}

	var response struct {
		Organization struct {
			Repositories struct {
				Nodes []Repository `json:"nodes"`
			} `json:"repositories"`
		} `json:"organization"`
	}

	err = client.DoWithContext(context.Background(), query, variables, &response)
	if err != nil {
		return nil, fmt.Errorf("error listing repositories for org %s: %w", org, err)
	}

	return response.Organization.Repositories.Nodes, nil
}

// GetRepositoryNames returns just the names of repositories as a slice
func GetRepositoryNames(repos []Repository) []string {
	names := make([]string, len(repos))
	for i, repo := range repos {
		names[i] = repo.Name
	}
	return names
}
