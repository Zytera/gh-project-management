package step

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Main struct {
	index  int
	steps  []Step
	width  int
	height int
}

type StepModel interface {
	Init() tea.Cmd
	Update(tea.Msg) (StepModel, tea.Cmd)
	View() string
	GetTitle() string
}

type Step struct {
	Title string
	Model StepModel
}

func New(steps []Step) *Main {
	return &Main{steps: steps}
}
