package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Zytera/gh-project-management/internal/config"
	"github.com/Zytera/gh-project-management/internal/gh"
	"github.com/Zytera/gh-project-management/internal/templates"
	"github.com/Zytera/gh-project-management/pkg/issue"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var (
	issueType   string
	issueTitle  string
	issueFields []string // Format: "fieldname=value"
	showFields  bool     // Flag to show available fields for a type

	// Custom fields
	createTeam       string
	createPriority   string
	createNoTransfer bool // Disable automatic transfer when Team is set

	// Dependencies and linking
	createDependsOn []string // Issues that block this issue
	createParent    string   // Parent issue to link to
)

var issueCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new issue",
	Long: `Create a new issue using templates (from repository or defaults).

The command uses the --type flag to specify the issue type and --field flags
to provide field values. You can see available fields for a type using --show-fields.

You can also set custom fields, dependencies, and parent links in a single command.

Examples:
  # Show available fields for a type
  gh project-management issue create --type epic --show-fields

  # Create an epic
  gh project-management issue create --type epic \
    --title "User Management" \
    --field description="Complete user management system" \
    --field objective="Enable user registration" \
    --field stories="Story 1, Story 2" \
    --field acceptance_criteria="Users can register" \
    --field teams="Backend, Frontend"

  # Create a task with custom fields and dependencies
  # Automatically transfers to Backend repo when --team is set
  gh project-management issue create --type task \
    --title "Implement API" \
    --field description="Create REST endpoint" \
    --field checklist="- [ ] Create endpoint\n- [ ] Add tests" \
    --team Backend \
    --priority High \
    --depends-on 45 \
    --depends-on 46 \
    --parent 44

  # Create with team but prevent automatic transfer
  gh project-management issue create --type task \
    --title "Fix bug" \
    --field description="Fix login bug" \
    --team Backend \
    --priority Critical \
    --no-transfer

Available default types: epic, user_story, task, bug, feature
Custom types will be loaded from .github/ISSUE_TEMPLATE/ in your repository.`,
	RunE: runIssueCreate,
}

func runIssueCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Prompt for issue type if not provided
	if issueType == "" {
		issueType, err = promptForIssueType()
		if err != nil {
			return fmt.Errorf("failed to get issue type: %w", err)
		}
	}

	// Get template (from repo or default)
	var template *templates.IssueTemplate
	var templateSource string

	template, templateSource = templates.GetTemplate(ctx, cfg.Owner, cfg.DefaultRepo, issueType)

	// If --show-fields is set, display fields and exit
	if showFields {
		return displayTemplateFields(template, templateSource)
	}

	// Prompt for title if not provided
	if issueTitle == "" {
		issueTitle, err = promptForTitle()
		if err != nil {
			return fmt.Errorf("failed to get issue title: %w", err)
		}
	}

	// Parse field values
	fields := make(map[string]string)
	for _, fieldStr := range issueFields {
		parts := strings.SplitN(fieldStr, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid field format '%s', expected 'fieldname=value'", fieldStr)
		}
		fieldName := strings.TrimSpace(parts[0])
		fieldValue := strings.TrimSpace(parts[1])
		fields[fieldName] = fieldValue
	}

	// Prompt for template fields interactively if none provided
	if len(issueFields) == 0 {
		fields = promptForTemplateFields(template, fields)
	}

	// Check for missing required fields
	missingFields := []string{}
	for _, field := range template.Body {
		if field.Type == templates.FieldTypeMarkdown {
			continue
		}
		if field.Validations.Required {
			if _, exists := fields[field.ID]; !exists {
				label := field.Attributes.Label
				if label == "" {
					label = field.ID
				}
				missingFields = append(missingFields, fmt.Sprintf("  --field %s=\"...\"  # %s (required)", field.ID, label))
			}
		}
	}

	if len(missingFields) > 0 {
		fmt.Printf("Missing required fields for type '%s':\n\n", issueType)
		for _, f := range missingFields {
			fmt.Println(f)
		}
		fmt.Printf("\nUse --show-fields to see all available fields.\n")
		return fmt.Errorf("missing required fields")
	}

	// Prompt for custom fields if not provided
	if createTeam == "" {
		team, err := promptForTeam(cfg)
		if err != nil {
			fmt.Printf("⚠️  Warning: Failed to prompt for team: %v\n", err)
		} else if team != "" {
			createTeam = team
		}
	}

	if createPriority == "" {
		priority, err := promptForPriority()
		if err != nil {
			fmt.Printf("⚠️  Warning: Failed to prompt for priority: %v\n", err)
		} else if priority != "" {
			createPriority = priority
		}
	}

	// Prompt for parent if not provided
	if createParent == "" {
		parent, err := promptForParent(ctx, cfg)
		if err != nil {
			fmt.Printf("⚠️  Warning: Failed to prompt for parent: %v\n", err)
		} else if parent != "" {
			createParent = parent
		}
	}

	// Prompt for dependencies if not provided
	if len(createDependsOn) == 0 {
		deps, err := promptForDependencies(ctx, cfg)
		if err != nil {
			fmt.Printf("⚠️  Warning: Failed to prompt for dependencies: %v\n", err)
		} else if len(deps) > 0 {
			createDependsOn = deps
		}
	}

	fmt.Printf("Creating %s issue using %s...\n", issueType, templateSource)

	// Create the issue
	params := issue.CreateDynamicIssueParams{
		Config:    cfg,
		IssueType: issueType,
		Title:     issueTitle,
		Fields:    fields,
	}

	result, err := issue.CreateDynamicIssue(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create issue: %w", err)
	}

	createdIssue := result.Issue
	projectItemID := result.ProjectItemID

	fmt.Printf("\n✓ Successfully created issue #%d: %s\n", createdIssue.Number, createdIssue.Title)
	fmt.Printf("  URL: %s\n", createdIssue.URL)

	// Link to parent if specified
	if createParent != "" {
		_, _, parentNumber, err := gh.ParseIssueReference(createParent, cfg.Owner, cfg.DefaultRepo)
		if err != nil {
			fmt.Printf("\n⚠️  Warning: Invalid parent reference '%s': %v\n", createParent, err)
		} else {
			fmt.Printf("\nLinking to parent issue #%d...\n", parentNumber)
			err = gh.AddSubIssue(ctx, cfg.Owner, cfg.DefaultRepo, parentNumber, createdIssue.Number)
			if err != nil {
				fmt.Printf("⚠️  Warning: Failed to link to parent: %v\n", err)
			} else {
				fmt.Printf("✓ Linked to parent issue #%d\n", parentNumber)
			}
		}
	}

	// Set custom fields if any are specified
	if createTeam != "" || createPriority != "" {
		fmt.Printf("\nSetting custom fields...\n")

		// Get project number
		projectNumber, err := strconv.Atoi(cfg.ProjectID)
		if err != nil {
			fmt.Printf("⚠️  Warning: Invalid project ID: %v\n", err)
		} else {
			// Get project node ID
			var projectNodeID string
			if cfg.OwnerType == config.OwnerTypeOrg {
				projectNodeID, err = gh.GetProjectNodeID(ctx, cfg.Owner, projectNumber)
			} else {
				projectNodeID, err = gh.GetUserProjectNodeID(ctx, projectNumber)
			}
			if err != nil {
				fmt.Printf("⚠️  Warning: Failed to get project: %v\n", err)
			} else {
				// Get project fields
				fields, err := gh.GetProjectFields(ctx, projectNodeID)
				if err != nil {
					fmt.Printf("⚠️  Warning: Failed to get project fields: %v\n", err)
				} else {
					// Set Team field
					if createTeam != "" {
						if err := setFieldValue(ctx, projectNodeID, projectItemID, fields, "Team", createTeam); err != nil {
							fmt.Printf("⚠️  Warning: Failed to set Team: %v\n", err)
						} else {
							fmt.Printf("  ✓ Team: %s\n", createTeam)
						}
					}

					// Set Priority field
					if createPriority != "" {
						if err := setFieldValue(ctx, projectNodeID, projectItemID, fields, "Priority", createPriority); err != nil {
							fmt.Printf("⚠️  Warning: Failed to set Priority: %v\n", err)
						} else {
							fmt.Printf("  ✓ Priority: %s\n", createPriority)
						}
					}
				}
			}
		}
	}

	// Add dependencies if specified
	if len(createDependsOn) > 0 {
		fmt.Printf("\nAdding dependencies...\n")
		for _, depRef := range createDependsOn {
			_, _, depNumber, err := gh.ParseIssueReference(depRef, cfg.Owner, cfg.DefaultRepo)
			if err != nil {
				fmt.Printf("⚠️  Warning: Invalid dependency reference '%s': %v\n", depRef, err)
				continue
			}

			// This issue is blocked by depNumber
			err = gh.AddBlockedBy(ctx, cfg.Owner, cfg.DefaultRepo, createdIssue.Number, depNumber)
			if err != nil {
				fmt.Printf("⚠️  Warning: Failed to add dependency on #%d: %v\n", depNumber, err)
			} else {
				fmt.Printf("  ✓ Blocked by issue #%d\n", depNumber)
			}
		}
	}

	// Auto-transfer if Team is set and not disabled
	if createTeam != "" && !createNoTransfer {
		targetRepo, exists := teamRepoMapping[createTeam]
		if !exists {
			fmt.Printf("\n⚠️  Warning: No repository mapping found for team '%s', skipping transfer\n", createTeam)
			fmt.Printf("💡 Available teams: Backend, App, Web, Auth\n")
		} else {
			fmt.Printf("\n🚀 Auto-transferring to %s/%s based on Team field...\n", cfg.Owner, targetRepo)

			sourceRepo := fmt.Sprintf("%s/%s", cfg.Owner, cfg.DefaultRepo)
			_, err = gh.TransferIssue(ctx, createdIssue.Number, cfg.Owner, targetRepo, sourceRepo)
			if err != nil {
				fmt.Printf("⚠️  Warning: Failed to transfer: %v\n", err)
			} else {
				fmt.Printf("\n✓ Successfully transferred to %s/%s\n", cfg.Owner, targetRepo)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("  1. Note the new issue number from the output above\n")
				fmt.Printf("  2. Update parent issue body with cross-repo reference: ### %s/%s#<new-number>\n", cfg.Owner, targetRepo)
			}
		}
	} else {
		// Show next steps if not auto-transferred
		fmt.Printf("\nNext steps:\n")
		if createParent == "" {
			fmt.Printf("  1. Link to parent: gh project-management link add <parent> %d\n", createdIssue.Number)
		}
		if createTeam == "" && createPriority == "" {
			fmt.Printf("  2. Set custom fields: gh project-management field set %d --team <team> --priority <priority>\n", createdIssue.Number)
		}
		if len(createDependsOn) == 0 {
			fmt.Printf("  3. Add dependencies: gh project-management dependency add %d <blocking-issue>\n", createdIssue.Number)
		}
		if createTeam != "" && createNoTransfer {
			fmt.Printf("  4. Transfer manually: gh project-management transfer issue %d --target <repo>\n", createdIssue.Number)
		} else if createTeam == "" {
			fmt.Printf("  4. Set team and transfer: gh project-management field set %d --team <team> --transfer\n", createdIssue.Number)
		}
	}

	return nil
}

// setFieldValue sets a single-select field value in the project
func setFieldValue(ctx context.Context, projectNodeID, projectItemID string, fields []gh.Field, fieldName, value string) error {
	field := gh.FindFieldByName(fields, fieldName)
	if field == nil {
		return fmt.Errorf("%s field not found in project", fieldName)
	}

	// Find the option ID for the value
	var optionID string
	for _, opt := range field.Options {
		if opt.Name == value {
			optionID = opt.ID
			break
		}
	}
	if optionID == "" {
		return fmt.Errorf("value '%s' not found in %s field options", value, fieldName)
	}

	return gh.UpdateProjectItemField(ctx, projectNodeID, projectItemID, field.ID, optionID)
}

func displayTemplateFields(template *templates.IssueTemplate, source string) error {
	fmt.Printf("Template: %s\n", template.Name)
	fmt.Printf("Type: %s\n", template.Type)
	fmt.Printf("Source: %s\n", source)
	if template.Description != "" {
		fmt.Printf("Description: %s\n", template.Description)
	}
	fmt.Printf("\n")

	// Display required fields
	requiredFields := []templates.BodyField{}
	optionalFields := []templates.BodyField{}

	for _, field := range template.Body {
		if field.Type == templates.FieldTypeMarkdown {
			continue
		}
		if field.Validations.Required {
			requiredFields = append(requiredFields, field)
		} else {
			optionalFields = append(optionalFields, field)
		}
	}

	if len(requiredFields) > 0 {
		fmt.Println("Required fields:")
		for _, field := range requiredFields {
			label := field.Attributes.Label
			if label == "" {
				label = field.ID
			}
			description := field.Attributes.Description
			fmt.Printf("  --field %s=\"...\"", field.ID)
			if label != field.ID {
				fmt.Printf("  # %s", label)
			}
			if description != "" {
				fmt.Printf("\n      %s", description)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	if len(optionalFields) > 0 {
		fmt.Println("Optional fields:")
		for _, field := range optionalFields {
			label := field.Attributes.Label
			if label == "" {
				label = field.ID
			}
			description := field.Attributes.Description
			fmt.Printf("  --field %s=\"...\"", field.ID)
			if label != field.ID {
				fmt.Printf("  # %s", label)
			}
			if description != "" {
				fmt.Printf("\n      %s", description)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	// Show example usage
	fmt.Println("Example usage:")
	fmt.Printf("  gh project-management issue create --type %s \\\n", template.Type)
	fmt.Printf("    --title \"Issue Title\"")
	if len(requiredFields) > 0 {
		fmt.Printf(" \\\n")
		for i, field := range requiredFields {
			label := field.Attributes.Label
			if label == "" {
				label = field.ID
			}
			fmt.Printf("    --field %s=\"%s value\"", field.ID, label)
			if i < len(requiredFields)-1 {
				fmt.Printf(" \\")
			}
			fmt.Println()
		}
	} else {
		fmt.Println()
	}

	return nil
}

// promptForParent asks the user if they want to link to a parent issue
func promptForParent(ctx context.Context, cfg *config.Config) (string, error) {
	fmt.Println()

	var wantsParent bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("🔗 Parent Issue").
				Description("Do you want to link this issue to a parent?").
				Value(&wantsParent),
		),
	)

	if err := confirmForm.Run(); err != nil {
		return "", fmt.Errorf("error confirming parent: %w", err)
	}

	if !wantsParent {
		return "", nil
	}

	// List recent issues
	issues, err := gh.ListRecentIssues(ctx, cfg.Owner, cfg.DefaultRepo, 20)
	if err != nil {
		return "", err
	}

	if len(issues) == 0 {
		fmt.Println("No open issues found in the repository.")
		return "", nil
	}

	// Build issue options
	issueOptions := make([]huh.Option[int], len(issues))
	for i, iss := range issues {
		label := fmt.Sprintf("#%d - %s", iss.Number, iss.Title)
		issueOptions[i] = huh.NewOption(label, i)
	}

	var selectedIndex int
	parentForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Select Parent Issue").
				Description("Choose the parent issue to link to").
				Options(issueOptions...).
				Value(&selectedIndex),
		),
	)

	if err := parentForm.Run(); err != nil {
		return "", fmt.Errorf("error selecting parent: %w", err)
	}

	return fmt.Sprintf("%d", issues[selectedIndex].Number), nil
}

// promptForDependencies asks the user if they want to add dependencies
func promptForDependencies(ctx context.Context, cfg *config.Config) ([]string, error) {
	fmt.Println()

	var wantsDependencies bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("🔒 Dependencies").
				Description("Do you want to add dependencies (issues that block this one)?").
				Value(&wantsDependencies),
		),
	)

	if err := confirmForm.Run(); err != nil {
		return nil, fmt.Errorf("error confirming dependencies: %w", err)
	}

	if !wantsDependencies {
		return nil, nil
	}

	// List recent issues
	issues, err := gh.ListRecentIssues(ctx, cfg.Owner, cfg.DefaultRepo, 20)
	if err != nil {
		return nil, err
	}

	if len(issues) == 0 {
		fmt.Println("No open issues found in the repository.")
		return nil, nil
	}

	// Build issue options
	issueOptions := make([]huh.Option[int], len(issues))
	for i, iss := range issues {
		label := fmt.Sprintf("#%d - %s", iss.Number, iss.Title)
		issueOptions[i] = huh.NewOption(label, i)
	}

	dependencies := []string{}
	addMore := true

	for addMore {
		var selectedIndex int
		var continueAdding bool

		dependencyForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[int]().
					Title("Select Blocking Issue").
					Description(fmt.Sprintf("Choose an issue that blocks this one (%d dependencies added)", len(dependencies))).
					Options(issueOptions...).
					Value(&selectedIndex),

				huh.NewConfirm().
					Title("Add another dependency?").
					Value(&continueAdding),
			),
		)

		if err := dependencyForm.Run(); err != nil {
			return dependencies, fmt.Errorf("error selecting dependency: %w", err)
		}

		issueNum := fmt.Sprintf("%d", issues[selectedIndex].Number)
		dependencies = append(dependencies, issueNum)
		fmt.Printf("✓ Added dependency on issue #%s\n", issueNum)

		addMore = continueAdding
	}

	return dependencies, nil
}

// promptForIssueType asks the user to select an issue type
func promptForIssueType() (string, error) {
	fmt.Println()

	var selectedType string
	typeForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("📝 Issue Type").
				Description("Select the type of issue to create").
				Options(
					huh.NewOption("Epic - Project epics", "epic"),
					huh.NewOption("User Story - User stories", "user_story"),
					huh.NewOption("Task - Technical tasks", "task"),
					huh.NewOption("Bug - Bug reports", "bug"),
					huh.NewOption("Feature - Feature requests", "feature"),
				).
				Value(&selectedType),
		),
	)

	if err := typeForm.Run(); err != nil {
		return "", fmt.Errorf("error selecting issue type: %w", err)
	}

	// If custom is selected, prompt for custom type name
	if selectedType == "custom" {
		var customType string
		customForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Custom Type Name").
					Description("Enter the name of your custom issue type").
					Value(&customType),
			),
		)

		if err := customForm.Run(); err != nil {
			return "", fmt.Errorf("error getting custom type: %w", err)
		}

		if customType == "" {
			return "", fmt.Errorf("custom type cannot be empty")
		}

		return customType, nil
	}

	return selectedType, nil
}

// promptForTitle asks the user for an issue title
func promptForTitle() (string, error) {
	fmt.Println()

	var title string
	titleForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("📌 Issue Title").
				Description("Enter the title for your issue").
				Value(&title).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("title cannot be empty")
					}
					return nil
				}),
		),
	)

	if err := titleForm.Run(); err != nil {
		return "", fmt.Errorf("error getting issue title: %w", err)
	}

	return strings.TrimSpace(title), nil
}

// promptForTeam asks the user to select a team
func promptForTeam(cfg *config.Config) (string, error) {
	fmt.Println()

	var wantsTeam bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("👥 Team Assignment").
				Description("Do you want to assign this issue to a team?").
				Value(&wantsTeam),
		),
	)

	if err := confirmForm.Run(); err != nil {
		return "", fmt.Errorf("error confirming team assignment: %w", err)
	}

	if !wantsTeam {
		return "", nil
	}

	// Build team options
	teamOptions := make([]huh.Option[string], 0, len(cfg.TeamRepos))
	for team := range cfg.TeamRepos {
		teamOptions = append(teamOptions, huh.NewOption(team, team))
	}

	var selectedTeam string
	teamForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select Team").
				Description("Choose the team to assign this issue to (will auto-transfer to team repo)").
				Options(teamOptions...).
				Value(&selectedTeam),
		),
	)

	if err := teamForm.Run(); err != nil {
		return "", fmt.Errorf("error selecting team: %w", err)
	}

	return selectedTeam, nil
}

// promptForPriority asks the user to select a priority
func promptForPriority() (string, error) {
	fmt.Println()

	var wantsPriority bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("⚡ Priority").
				Description("Do you want to set a priority for this issue?").
				Value(&wantsPriority),
		),
	)

	if err := confirmForm.Run(); err != nil {
		return "", fmt.Errorf("error confirming priority: %w", err)
	}

	if !wantsPriority {
		return "", nil
	}

	var selectedPriority string
	priorityForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select Priority").
				Description("Choose the priority level").
				Options(
					huh.NewOption("🔴 Critical", "Critical"),
					huh.NewOption("🟠 High", "High"),
					huh.NewOption("🟡 Medium", "Medium"),
					huh.NewOption("⚪ Low", "Low"),
				).
				Value(&selectedPriority),
		),
	)

	if err := priorityForm.Run(); err != nil {
		return "", fmt.Errorf("error selecting priority: %w", err)
	}

	return selectedPriority, nil
}

// promptForTemplateFields asks the user to fill template fields interactively
func promptForTemplateFields(template *templates.IssueTemplate, existingFields map[string]string) map[string]string {
	fields := make(map[string]string)

	// Copy existing fields
	for k, v := range existingFields {
		fields[k] = v
	}

	fmt.Println()
	fmt.Println("📋 Template Fields")
	fmt.Println()

	// Process each field in the template
	for _, field := range template.Body {
		// Skip markdown fields
		if field.Type == "markdown" {
			continue
		}

		// Only process input fields (textarea, input, dropdown)
		if field.Type != "textarea" && field.Type != "input" && field.Type != "dropdown" {
			continue
		}

		// Skip if already provided
		if _, exists := fields[field.ID]; exists {
			continue
		}

		required := isFieldRequired(field, template)
		var value string

		// Build title and description
		title := field.Attributes.Label
		description := field.Attributes.Description
		if required {
			title += " (required)"
		}

		if field.Type == "dropdown" {
			// Handle dropdown fields
			if len(field.Attributes.Options) == 0 {
				// No options available, skip
				continue
			}

			// Build options for the select
			selectOptions := make([]huh.Option[string], len(field.Attributes.Options))
			for i, option := range field.Attributes.Options {
				selectOptions[i] = huh.NewOption(option, option)
			}

			dropdownForm := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title(title).
						Description(description).
						Options(selectOptions...).
						Value(&value),
				),
			)

			if err := dropdownForm.Run(); err != nil {
				// User cancelled or error
				if required {
					fmt.Printf("⚠️  Warning: Skipping required field '%s'\n", field.ID)
				}
				continue
			}
		} else if field.Type == "textarea" {
			// Use Text for multiline input
			textForm := huh.NewForm(
				huh.NewGroup(
					huh.NewText().
						Title(title).
						Description(description).
						Value(&value).
						Validate(func(s string) error {
							if required && strings.TrimSpace(s) == "" {
								return fmt.Errorf("this field is required")
							}
							return nil
						}),
				),
			)

			if err := textForm.Run(); err != nil {
				// User cancelled or error, continue to next field
				continue
			}
		} else {
			// Use Input for single-line input
			inputForm := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title(title).
						Description(description).
						Value(&value).
						Validate(func(s string) error {
							if required && strings.TrimSpace(s) == "" {
								return fmt.Errorf("this field is required")
							}
							return nil
						}),
				),
			)

			if err := inputForm.Run(); err != nil {
				// User cancelled or error, continue to next field
				continue
			}
		}

		// Store the value if not empty
		value = strings.TrimSpace(value)
		if value != "" {
			fields[field.ID] = value
		}
	}

	return fields
}

// isFieldRequired checks if a field is required in the template
func isFieldRequired(field templates.BodyField, template *templates.IssueTemplate) bool {
	for _, reqField := range template.GetRequiredFields() {
		if reqField.ID == field.ID {
			return true
		}
	}
	return false
}

func init() {
	// Template flags
	issueCreateCmd.Flags().StringVar(&issueType, "type", "", "Issue type (epic, user_story, task, bug, feature, or custom)")
	issueCreateCmd.Flags().StringVar(&issueTitle, "title", "", "Issue title (required)")
	issueCreateCmd.Flags().StringSliceVar(&issueFields, "field", []string{}, "Field values in format 'fieldname=value' (can be repeated)")
	issueCreateCmd.Flags().BoolVar(&showFields, "show-fields", false, "Show available fields for the specified type")

	// Custom fields
	issueCreateCmd.Flags().StringVar(&createTeam, "team", "", "Team value (Backend, App, Web, Auth) - automatically transfers to team repo")
	issueCreateCmd.Flags().StringVar(&createPriority, "priority", "", "Priority value (Critical, High, Medium, Low)")
	issueCreateCmd.Flags().BoolVar(&createNoTransfer, "no-transfer", false, "Prevent automatic transfer when Team field is set")

	// Dependencies and linking
	issueCreateCmd.Flags().StringSliceVar(&createDependsOn, "depends-on", []string{}, "Issues that block this issue (can be repeated)")
	issueCreateCmd.Flags().StringVar(&createParent, "parent", "", "Parent issue to link to")
}
