package cmd

import (
	"log"

	createIssue "github.com/Zytera/gh-project-managment/internal/create-issue"
	"github.com/Zytera/gh-project-managment/internal/step"

	"github.com/spf13/cobra"

	tea "github.com/charmbracelet/bubbletea"
)

var createEpicCmd = &cobra.Command{
	Use:   "create-epic",
	Short: "TODO",
	Long:  `TODO`,
	Args:  cobra.ExactArgs(0),
	RunE:  runCreateEpic,
}

func init() {
	// Add command to root
	rootCmd.AddCommand(createEpicCmd)
}

func runCreateEpic(cmd *cobra.Command, args []string) error {

	steps := []step.Step{
		{
			Title: "validate",
			Model: createIssue.NewModel(),
		},
	}

	main := step.New(steps)

	p := tea.NewProgram(*main, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
	return nil
}
