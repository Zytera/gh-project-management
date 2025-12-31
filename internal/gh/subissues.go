package gh

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
)

// AddSubIssue adds a child issue to a parent issue using GitHub's tracked-by/tracks relationship
// This is done by updating the parent issue body to include a reference to the child issue
func AddSubIssue(ctx context.Context, owner, repo string, parentNumber, childNumber int) error {
	client, err := api.DefaultRESTClient()
	if err != nil {
		return fmt.Errorf("failed to create REST client: %w", err)
	}

	// Get the parent issue
	var parentIssue struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		Body   string `json:"body"`
	}

	err = client.Get(fmt.Sprintf("repos/%s/%s/issues/%d", owner, repo, parentNumber), &parentIssue)
	if err != nil {
		return fmt.Errorf("failed to get parent issue: %w", err)
	}

	// Get the child issue to get its title
	var childIssue struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
	}

	err = client.Get(fmt.Sprintf("repos/%s/%s/issues/%d", owner, repo, childNumber), &childIssue)
	if err != nil {
		return fmt.Errorf("failed to get child issue: %w", err)
	}

	// Update parent issue body to include the child issue reference
	newBody := addSubIssueToBody(parentIssue.Body, childNumber, childIssue.Title)

	// Update the parent issue
	updatePayload := map[string]interface{}{
		"body": newBody,
	}

	payloadBytes, err := json.Marshal(updatePayload)
	if err != nil {
		return fmt.Errorf("failed to marshal update payload: %w", err)
	}

	err = client.Patch(fmt.Sprintf("repos/%s/%s/issues/%d", owner, repo, parentNumber), bytes.NewReader(payloadBytes), nil)
	if err != nil {
		return fmt.Errorf("failed to update parent issue: %w", err)
	}

	return nil
}

// RemoveSubIssue removes a child issue from a parent issue
func RemoveSubIssue(ctx context.Context, owner, repo string, parentNumber, childNumber int) error {
	client, err := api.DefaultRESTClient()
	if err != nil {
		return fmt.Errorf("failed to create REST client: %w", err)
	}

	// Get the parent issue
	var parentIssue struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		Body   string `json:"body"`
	}

	err = client.Get(fmt.Sprintf("repos/%s/%s/issues/%d", owner, repo, parentNumber), &parentIssue)
	if err != nil {
		return fmt.Errorf("failed to get parent issue: %w", err)
	}

	// Remove the child issue reference from body
	newBody := removeSubIssueFromBody(parentIssue.Body, childNumber)

	// Update the parent issue
	updatePayload := map[string]interface{}{
		"body": newBody,
	}

	payloadBytes, err := json.Marshal(updatePayload)
	if err != nil {
		return fmt.Errorf("failed to marshal update payload: %w", err)
	}

	err = client.Patch(fmt.Sprintf("repos/%s/%s/issues/%d", owner, repo, parentNumber), bytes.NewReader(payloadBytes), nil)
	if err != nil {
		return fmt.Errorf("failed to update parent issue: %w", err)
	}

	return nil
}

// addSubIssueToBody adds a sub-issue reference to the issue body
// GitHub recognizes tasklist items with issue references as sub-issues
func addSubIssueToBody(body string, issueNumber int, issueTitle string) string {
	// Look for existing tasklist section
	tasklistPattern := regexp.MustCompile(`(?i)##\s*(?:tasks?|sub-?issues?|checklist)\s*\n`)

	issueRef := fmt.Sprintf("- [ ] #%d", issueNumber)

	// Check if the reference already exists
	if strings.Contains(body, issueRef) {
		return body
	}

	// If there's a tasklist section, add to it
	if tasklistPattern.MatchString(body) {
		// Find the section and add the new item
		lines := strings.Split(body, "\n")
		var newLines []string
		inTasklist := false
		added := false

		for _, line := range lines {
			newLines = append(newLines, line)

			if tasklistPattern.MatchString(line) {
				inTasklist = true
			} else if inTasklist && !added {
				// Check if this is still part of the tasklist
				if strings.HasPrefix(strings.TrimSpace(line), "- [ ]") || strings.HasPrefix(strings.TrimSpace(line), "- [x]") {
					continue // Keep looking for the end of the tasklist
				} else if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
					// End of tasklist, insert before this line
					newLines = append(newLines[:len(newLines)-1], issueRef, line)
					added = true
					inTasklist = false
				}
			}
		}

		if inTasklist && !added {
			// Tasklist was at the end, append there
			newLines = append(newLines, issueRef)
		}

		return strings.Join(newLines, "\n")
	}

	// No tasklist section exists, create one at the end
	if body == "" {
		return fmt.Sprintf("## Tasks\n\n%s\n", issueRef)
	}

	// Add tasklist section at the end
	separator := "\n\n"
	if strings.HasSuffix(body, "\n") {
		separator = "\n"
	}
	if strings.HasSuffix(body, "\n\n") {
		separator = ""
	}

	return fmt.Sprintf("%s%s## Tasks\n\n%s\n", body, separator, issueRef)
}

// removeSubIssueFromBody removes a sub-issue reference from the issue body
func removeSubIssueFromBody(body string, issueNumber int) string {
	issueRef := fmt.Sprintf("#%d", issueNumber)

	lines := strings.Split(body, "\n")
	var newLines []string

	for _, line := range lines {
		// Skip lines that reference this issue number in a tasklist item
		trimmed := strings.TrimSpace(line)
		if (strings.HasPrefix(trimmed, "- [ ]") || strings.HasPrefix(trimmed, "- [x]")) &&
			strings.Contains(trimmed, issueRef) {
			continue
		}
		newLines = append(newLines, line)
	}

	return strings.Join(newLines, "\n")
}
