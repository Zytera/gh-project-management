package createissue

import (
	"context"
	"errors"
	"fmt"

	"github.com/Zytera/gh-project-management/internal/config"
	"github.com/Zytera/gh-project-management/internal/gh"
	"github.com/charmbracelet/huh"
)

// CreateEpicForm creates and runs an interactive form to create an epic issue
func CreateEpicForm(ctx context.Context) error {
	cfg := ctx.Value(config.ConfigKey{}).(*config.Config)

	var (
		name string
		url  string
		err  error
	)

	// Create the form
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Epic Name").
				Description("Enter the name of the epic").
				Value(&name).
				Placeholder("Epic Name").
				Validate(func(s string) error {
					if len(s) == 0 {
						return errors.New("name cannot be empty")
					}
					return nil
				}),
		),
	).WithTheme(huh.ThemeCharm())

	// Run the form
	if err := form.Run(); err != nil {
		return err
	}

	// Create the issue after form completion
	fmt.Println("\n🔧 Creating epic...")
	issue, err := gh.CreateIssueWithTemplate(ctx, cfg.Org, cfg.DefaultRepo, name, gh.TemplateEpic)
	if err != nil {
		return fmt.Errorf("failed to create epic: %w", err)
	}

	url = issue.URL

	// Show success message
	fmt.Printf("\n✅ Epic created successfully!\n")
	fmt.Printf("📋 Issue #%d: %s\n", issue.Number, issue.Title)
	fmt.Printf("🔗 URL: %s\n\n", url)

	return nil
}

// CreateUserStoryForm creates and runs an interactive form to create a user story
func CreateUserStoryForm(ctx context.Context) error {
	cfg := ctx.Value(config.ConfigKey{}).(*config.Config)

	var (
		name string
	)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("User Story Name").
				Description("Enter the name of the user story").
				Value(&name).
				Placeholder("User Story Name").
				Validate(func(s string) error {
					if len(s) == 0 {
						return errors.New("name cannot be empty")
					}
					return nil
				}),
		),
	).WithTheme(huh.ThemeCharm())

	if err := form.Run(); err != nil {
		return err
	}

	fmt.Println("\n🔧 Creating user story...")
	issue, err := gh.CreateIssueWithTemplate(ctx, cfg.Org, cfg.DefaultRepo, name, gh.TemplateUserStory)
	if err != nil {
		return fmt.Errorf("failed to create user story: %w", err)
	}

	fmt.Printf("\n✅ User story created successfully!\n")
	fmt.Printf("📋 Issue #%d: %s\n", issue.Number, issue.Title)
	fmt.Printf("🔗 URL: %s\n\n", issue.URL)

	return nil
}

// CreateTaskForm creates and runs an interactive form to create a task
func CreateTaskForm(ctx context.Context) error {
	cfg := ctx.Value(config.ConfigKey{}).(*config.Config)

	var (
		name string
	)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Task Name").
				Description("Enter the name of the task").
				Value(&name).
				Placeholder("Task Name").
				Validate(func(s string) error {
					if len(s) == 0 {
						return errors.New("name cannot be empty")
					}
					return nil
				}),
		),
	).WithTheme(huh.ThemeCharm())

	if err := form.Run(); err != nil {
		return err
	}

	fmt.Println("\n🔧 Creating task...")
	issue, err := gh.CreateIssueWithTemplate(ctx, cfg.Org, cfg.DefaultRepo, name, gh.TemplateTask)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	fmt.Printf("\n✅ Task created successfully!\n")
	fmt.Printf("📋 Issue #%d: %s\n", issue.Number, issue.Title)
	fmt.Printf("🔗 URL: %s\n\n", issue.URL)

	return nil
}
