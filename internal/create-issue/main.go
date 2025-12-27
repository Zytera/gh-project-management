package createissue

import (
	"context"
	"errors"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/api"
)

var templates = map[string]string{
	"epic": epicTemplate,
}

// CreateIssue crea una issue en el repositorio indicado y devuelve la URL de la issue creada.
// Si se proporciona issueTemplate, se usará esa plantilla para asignar labels y assignees automáticamente.
func createIssue(owner, repo, title, issueTemplate string) (string, error) {
	if owner == "" || repo == "" {
		return "", errors.New("owner and repo cannot be empty")
	}
	if title == "" {
		return "", errors.New("title cannot be empty")
	}

	client, err := api.NewGraphQLClient(api.ClientOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create GitHub GraphQL client: %w", err)
	}

	// 1) Obtener repository ID
	var repoQuery struct {
		Repository struct {
			ID string `json:"id"`
		} `json:"repository"`
	}

	repoQueryQL := `
	query RepoID($owner: String!, $name: String!) {
		repository(owner: $owner, name: $name) {
			id
		}
	}`
	vars := map[string]interface{}{"owner": owner, "name": repo}

	if err := client.DoWithContext(context.Background(), repoQueryQL, vars, &repoQuery); err != nil {
		return "", fmt.Errorf("failed to query repository id: %w", err)
	}
	if repoQuery.Repository.ID == "" {
		return "", errors.New("repository not found or has no id")
	}
	repoID := repoQuery.Repository.ID

	// 1.5) Leer el contenido de la plantilla si se proporciona
	var body string
	if issueTemplate != "" {
		if tmpl, ok := templates[issueTemplate]; ok {
			body = tmpl
		}
	}

	// 2) Crear la issue con mutation createIssue
	var createResp struct {
		CreateIssue struct {
			Issue struct {
				ID     string `json:"id"`
				Number int    `json:"number"`
				URL    string `json:"url"`
			} `json:"issue"`
		} `json:"createIssue"`
	}

	createIssueQL := `
	mutation CreateIssue($input: CreateIssueInput!) {
		createIssue(input: $input) {
			issue {
				id
				number
				url
			}
		}
	}`

	vars = map[string]interface{}{
		"input": map[string]interface{}{
			"repositoryId": repoID,
			"title":        title,
			"body":         body,
		},
	}

	if err := client.DoWithContext(context.Background(), createIssueQL, vars, &createResp); err != nil {
		return "", fmt.Errorf("failed to create issue: %w", err)
	}

	if createResp.CreateIssue.Issue.URL == "" {
		return "", errors.New("issue created but response missing URL")
	}

	return createResp.CreateIssue.Issue.URL, nil
}
