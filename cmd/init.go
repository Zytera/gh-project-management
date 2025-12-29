package cmd

import (
	"errors"
	"fmt"

	"github.com/Zytera/gh-project-management/internal/config"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize gh-project-management configuration",
	Long:  `Interactive setup to configure your first project context.`,
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	// Check if config already exists
	globalConfig, err := config.LoadGlobal()
	if err != nil {
		return err
	}

	if len(globalConfig.Contexts) > 0 {
		fmt.Println("⚠️  Configuration already exists. Use 'gh project-management context add' to add a new project.")
		return nil
	}

	fmt.Println("⚙️  No configuration found. Let's set up your first project!")
	fmt.Println()

	// Step 1: Context name
	var contextName string
	contextForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Context name").
				Description("A short name for this project configuration (e.g., 'project-management')").
				Value(&contextName).
				Validate(func(s string) error {
					if len(s) == 0 {
						return errors.New("context name cannot be empty")
					}
					return nil
				}),
		),
	)

	if err := contextForm.Run(); err != nil {
		return fmt.Errorf("error running form: %w", err)
	}

	// Collect configuration via the shared interactive form
	ctx, err := collectContextConfiguration()
	if err != nil {
		return err
	}

	globalConfig.Contexts[contextName] = *ctx
	globalConfig.CurrentContext = contextName

	if err := config.Save(globalConfig); err != nil {
		return fmt.Errorf("error saving configuration: %w", err)
	}

	configPath, _ := config.GetConfigPath()
	fmt.Printf("\n✓ Configuration saved to %s\n", configPath)
	fmt.Printf("✓ Context '%s' set as current\n\n", contextName)
	fmt.Println("You can now use commands like:")
	fmt.Println("  gh project-management create-epic")
	fmt.Println("  gh project-management create-user-story")
	fmt.Println("  gh project-management create-task")

	return nil
}
