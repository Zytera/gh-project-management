package templates

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed default/*.yml
var templatesFS embed.FS

// ParseTemplate parses a YAML template file
func ParseTemplate(content []byte) (*IssueTemplate, error) {
	var template IssueTemplate
	if err := yaml.Unmarshal(content, &template); err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}
	return &template, nil
}

// GetDefaultTemplate returns a default template for a given type by reading from embedded files
func GetDefaultTemplate(issueType string) (*IssueTemplate, error) {
	// Map issue type to template file name
	templateFile := GetTemplateFileName(issueType)

	// Read from embedded FS (relative to this package)
	templatePath := filepath.Join("default", templateFile)

	content, err := templatesFS.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("default template not found for type %s: %w", issueType, err)
	}

	return ParseTemplate(content)
}

// GetTemplateFileName returns the template filename for a given issue type
func GetTemplateFileName(issueType string) string {
	switch issueType {
	case "epic", "Epic":
		return "epic.yml"
	case "story", "Story", "user story", "User Story":
		return "user_story.yml"
	case "task", "Task":
		return "task.yml"
	case "bug", "Bug":
		return "bug.yml"
	case "feature", "Feature":
		return "feature.yml"
	default:
		// Normalize custom types: lowercase and replace spaces with underscores
		normalized := strings.ToLower(strings.ReplaceAll(issueType, " ", "_"))
		return normalized + ".yml"
	}
}

// GetRequiredFields returns all required fields from a template
func (t *IssueTemplate) GetRequiredFields() []BodyField {
	var required []BodyField
	for _, field := range t.Body {
		// Skip markdown fields as they're not input fields
		if field.Type == FieldTypeMarkdown {
			continue
		}
		if field.Validations.Required {
			required = append(required, field)
		}
	}
	return required
}

// GetAllInputFields returns all input fields (excluding markdown)
func (t *IssueTemplate) GetAllInputFields() []BodyField {
	var inputs []BodyField
	for _, field := range t.Body {
		if field.Type != FieldTypeMarkdown {
			inputs = append(inputs, field)
		}
	}
	return inputs
}

// ValidateFieldValue validates a field value based on field type and validations
func (f *BodyField) ValidateFieldValue(value string) error {
	// Check required
	if f.Validations.Required && value == "" {
		return fmt.Errorf("field %s is required", f.ID)
	}

	// Type-specific validation
	switch f.Type {
	case FieldTypeDropdown:
		if value != "" && len(f.Attributes.Options) > 0 {
			// Validate value is in options
			found := false
			for _, option := range f.Attributes.Options {
				if option == value {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("invalid value for field %s: must be one of %v", f.ID, f.Attributes.Options)
			}
		}
	}

	return nil
}
