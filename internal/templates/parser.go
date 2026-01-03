package templates

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// ParseTemplate parses a YAML template file
func ParseTemplate(content []byte) (*IssueTemplate, error) {
	var template IssueTemplate
	if err := yaml.Unmarshal(content, &template); err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}
	return &template, nil
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
