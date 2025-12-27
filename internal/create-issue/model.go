package createissue

import (
	"github.com/Zytera/gh-project-managment/internal/step"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

type model struct {
	happy bool
	form  *huh.Form
}

func NewModel() *model {

	m := model{}
	m.happy = false
	confirm := huh.NewConfirm().
		Title("Are you sure? ").
		Description("Please confirm. ").
		Affirmative("Yes!").
		Negative("No.").
		Value(&m.happy)
	m.form = huh.NewForm(huh.NewGroup(confirm))
	return &m
}

func (m *model) Init() tea.Cmd {
	return m.form.Init()
}

func (m *model) GetTitle() string {
	return "Create Issue"
}

func (m *model) Update(msg tea.Msg) (step.StepModel, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}

	// Process the form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		if m.form.State == huh.StateCompleted {
			return m, func() tea.Msg {
				return step.NextStepMsg{}
			}
		}
	}

	// Asegurar que siempre retornamos un comando válido
	if cmd == nil {
		cmd = tea.Batch()
	}

	return m, cmd
}

func (m *model) View() string {
	return m.form.View()
}
