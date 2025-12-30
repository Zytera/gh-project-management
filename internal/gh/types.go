package gh

// Organization represents a GitHub organization
type Organization struct {
	Login string `json:"login"`
	Name  string `json:"name"`
}

// Project represents a GitHub Project V2
type Project struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
	Title  string `json:"title"`
	Owner  string `json:"-"` // Set manually, not from JSON
}

// Repository represents a GitHub repository
type Repository struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Field represents a GitHub Project V2 field
type Field struct {
	ID      string        `json:"id"`
	Name    string        `json:"name"`
	Options []FieldOption `json:"options,omitempty"`
}

// FieldOption represents an option in a single-select field
type FieldOption struct {
	ID    string     `json:"id"`
	Name  string     `json:"name"`
	Color FieldColor `json:"color,omitempty"`
}

// FieldColor represents available colors for single-select field options
type FieldColor string

const (
	ColorBlue   FieldColor = "BLUE"
	ColorGray   FieldColor = "GRAY"
	ColorGreen  FieldColor = "GREEN"
	ColorOrange FieldColor = "ORANGE"
	ColorPink   FieldColor = "PINK"
	ColorPurple FieldColor = "PURPLE"
	ColorRed    FieldColor = "RED"
	ColorYellow FieldColor = "YELLOW"
)

// DefaultTeamColors provides a default color mapping for teams
var DefaultTeamColors = []FieldColor{
	ColorBlue,
	ColorGreen,
	ColorOrange,
	ColorPurple,
}

// PriorityLevels defines the priority levels and their colors
var PriorityLevels = map[string]FieldColor{
	"Critical": ColorRed,
	"High":     ColorOrange,
	"Medium":   ColorYellow,
	"Low":      ColorGray,
}
