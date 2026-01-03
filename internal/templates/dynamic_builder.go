package templates

import (
	"fmt"
	"strings"
)

// FieldValue represents a field name-value pair
type FieldValue struct {
	Name  string
	Value string
}

// BuildBodyFromTemplate builds an issue body dynamically from a template and field values
func BuildBodyFromTemplate(template *IssueTemplate, fields map[string]string) (string, error) {
	var builder strings.Builder

	for _, field := range template.Body {
		// Skip markdown fields (they're just informational text)
		if field.Type == FieldTypeMarkdown {
			continue
		}

		// Get field value
		value, hasValue := fields[field.ID]

		// Validate required fields
		if field.Validations.Required && !hasValue {
			return "", fmt.Errorf("required field '%s' is missing", field.ID)
		}

		// Validate field value if present
		if hasValue && value != "" {
			if err := field.ValidateFieldValue(value); err != nil {
				return "", err
			}
		}

		// Only include fields that have values
		if !hasValue || value == "" {
			continue
		}

		// Build field section
		label := field.Attributes.Label
		if label == "" {
			label = field.ID
		}

		builder.WriteString(fmt.Sprintf("### %s\n\n", label))
		builder.WriteString(value)
		builder.WriteString("\n\n")
	}

	return builder.String(), nil
}

// ValidateFields validates that all required fields are present
func ValidateFields(template *IssueTemplate, fields map[string]string) error {
	requiredFields := template.GetRequiredFields()

	for _, field := range requiredFields {
		value, exists := fields[field.ID]
		if !exists || value == "" {
			return fmt.Errorf("required field '%s' (%s) is missing", field.ID, field.Attributes.Label)
		}

		// Validate field value
		if err := field.ValidateFieldValue(value); err != nil {
			return err
		}
	}

	return nil
}
