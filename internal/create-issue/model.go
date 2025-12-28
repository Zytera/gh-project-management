package createissue

import (
	"context"
	"errors"

	"github.com/Zytera/gh-project-managment/internal/config"
	"github.com/Zytera/gh-project-managment/internal/step"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

type model struct {
	ctx   context.Context
	name  string
	form  *huh.Form
	error string
	url   string
}

func NewModel(ctx context.Context) *model {

	m := model{}
	m.ctx = ctx
	inputName := huh.NewInput().
		Title("Name").
		Description("Enter the name of the issue").
		Value(&m.name).
		Placeholder("Issue Name").
		Validate(func(s string) error {
			if len(s) == 0 {
				return errors.New("name cannot be empty")
			}
			return nil
		})
	m.form = huh.NewForm(huh.NewGroup(inputName))
	m.url = ""
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

	config := m.ctx.Value(config.ConfigKey{}).(*config.Config)

	// Process the form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		if m.form.State == huh.StateCompleted {
			if m.url == "" {
				url, err := createIssue(config.Org, config.ProjectManagementRepoName, m.name, "epic")
				if err != nil {
					m.error = err.Error()
					return m, cmd
				}
				m.url = url
			}

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
	return m.form.View() + "\n" + m.error + "\n" + m.url
}
