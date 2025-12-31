package templates

// IssueTemplate represents a GitHub Issue Form template
type IssueTemplate struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"description"`
	Title       string      `yaml:"title"`
	Type        string      `yaml:"type"`
	Labels      []string    `yaml:"labels"`
	Body        []BodyField `yaml:"body"`
	LastUpdated string      // Git commit SHA or timestamp
}

// BodyField represents a field in the issue template body
type BodyField struct {
	Type        string          `yaml:"type"`
	ID          string          `yaml:"id"`
	Attributes  FieldAttributes `yaml:"attributes"`
	Validations Validations     `yaml:"validations"`
}

// FieldAttributes contains the configuration for a field
type FieldAttributes struct {
	Label       string   `yaml:"label"`
	Description string   `yaml:"description"`
	Placeholder string   `yaml:"placeholder"`
	Value       string   `yaml:"value"`    // For markdown type
	Options     []string `yaml:"options"`  // For dropdown type
	Multiple    bool     `yaml:"multiple"` // For checkboxes/dropdown
}

// Validations contains validation rules for a field
type Validations struct {
	Required bool `yaml:"required"`
}

// FieldType constants
const (
	FieldTypeMarkdown   = "markdown"
	FieldTypeTextarea   = "textarea"
	FieldTypeInput      = "input"
	FieldTypeDropdown   = "dropdown"
	FieldTypeCheckboxes = "checkboxes"
)

// IssueTypeConfig represents a GitHub Issue Type
type IssueTypeConfig struct {
	ID          string
	Name        string
	Description string
	IsEnabled   bool
	Color       string
}

// TemplateFieldValue represents a field value provided via CLI
type TemplateFieldValue struct {
	FieldID string
	Value   string
}
