package cmd

import (
	"log"

	createEpic "github.com/Zytera/gh-project-managment/internal/create-epic"

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

	model := createEpic.NewModel(6)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
	return nil
}
