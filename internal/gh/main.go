package gh

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

// Organization represents a GitHub organization
type Organization struct {
	Login string `json:"login"`
	Name  string `json:"name"`
}

// Project represents a GitHub Project
type Project struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	Owner  string `json:"-"` // Set manually, not from JSON
}

// ListOrganizations lists all organizations the user belongs to using GraphQL
func ListOrganizations() ([]Organization, error) {
	query := `query { viewer { organizations(first: 100) { nodes { login name } } } }`

	cmd := exec.Command("gh", "api", "graphql", "-f", fmt.Sprintf("query=%s", query))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error listing organizations: %w\nOutput: %s", err, string(output))
	}

	var response struct {
		Data struct {
			Viewer struct {
				Organizations struct {
					Nodes []Organization `json:"nodes"`
				} `json:"organizations"`
			} `json:"viewer"`
		} `json:"data"`
	}

	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("error parsing organizations: %w", err)
	}

	return response.Data.Viewer.Organizations.Nodes, nil
}

// ListOrgProjects lists projects owned by an organization using GraphQL
func ListOrgProjects(org string) ([]Project, error) {
	query := `query($owner: String!) { organization(login: $owner) { projectsV2(first: 100) { nodes { id number title } } } }`

	cmd := exec.Command("gh", "api", "graphql",
		"-f", fmt.Sprintf("query=%s", query),
		"-f", fmt.Sprintf("owner=%s", org))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error listing projects for org %s: %w\nOutput: %s", org, err, string(output))
	}

	var response struct {
		Data struct {
			Organization struct {
				ProjectsV2 struct {
					Nodes []Project `json:"nodes"`
				} `json:"projectsV2"`
			} `json:"organization"`
		} `json:"data"`
	}

	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("error parsing projects: %w", err)
	}

	projects := response.Data.Organization.ProjectsV2.Nodes

	// Mark owner
	for i := range projects {
		projects[i].Owner = org
	}

	return projects, nil
}

// ListUserProjects lists projects owned by the authenticated user using GraphQL
func ListUserProjects() ([]Project, error) {
	query := `query { viewer { projectsV2(first: 100) { nodes { id number title } } } }`

	cmd := exec.Command("gh", "api", "graphql", "-f", fmt.Sprintf("query=%s", query))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error listing user projects: %w\nOutput: %s", err, string(output))
	}

	var response struct {
		Data struct {
			Viewer struct {
				ProjectsV2 struct {
					Nodes []Project `json:"nodes"`
				} `json:"projectsV2"`
			} `json:"viewer"`
		} `json:"data"`
	}

	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("error parsing projects: %w", err)
	}

	projects := response.Data.Viewer.ProjectsV2.Nodes

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

// Repository represents a GitHub repository
type Repository struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ListOrgRepositories lists repositories for an organization using GraphQL
func ListOrgRepositories(org string) ([]Repository, error) {
	query := `query($owner: String!) { organization(login: $owner) { repositories(first: 100, orderBy: {field: UPDATED_AT, direction: DESC}) { nodes { name description } } } }`

	cmd := exec.Command("gh", "api", "graphql",
		"-f", fmt.Sprintf("query=%s", query),
		"-f", fmt.Sprintf("owner=%s", org))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error listing repositories for org %s: %w\nOutput: %s", org, err, string(output))
	}

	var response struct {
		Data struct {
			Organization struct {
				Repositories struct {
					Nodes []Repository `json:"nodes"`
				} `json:"repositories"`
			} `json:"organization"`
		} `json:"data"`
	}

	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("error parsing repositories: %w", err)
	}

	return response.Data.Organization.Repositories.Nodes, nil
}

// GetRepositoryNames returns just the names of repositories as a slice
func GetRepositoryNames(repos []Repository) []string {
	names := make([]string, len(repos))
	for i, repo := range repos {
		names[i] = repo.Name
	}
	return names
}
