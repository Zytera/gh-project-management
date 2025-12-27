package step

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Zytera/gh-project-managment/internal/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// NextStepMsg signals that we should move to the next step
type NextStepMsg struct{}

func (m Main) Init() tea.Cmd {
	// Inicializar todos los step models
	var cmds []tea.Cmd
	for _, step := range m.steps {
		if cmd := step.Model.Init(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}

func (m Main) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	current := m.steps[m.index]

	var cmd tea.Cmd
	switch msg := msg.(type) {
	case NextStepMsg:
		// Auto-advance to next step
		if m.index < len(m.steps)-1 {
			m.index++
			// Initialize the new step
			if nextStep := m.steps[m.index].Model.Init(); nextStep != nil {
				return m, nextStep
			}
		}
		return m, nil
	case tea.WindowSizeMsg:
		// Comprobamos el tamaño del terminal
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "pgup":
			if m.index > 0 {
				m.index--
			}
		case "pgdown":
			if m.index < len(m.steps)-1 {
				m.index++
			}
		default:
			// Delegar el mensaje al step model actual
			updatedModel, stepCmd := current.Model.Update(msg)
			m.steps[m.index].Model = updatedModel
			cmd = stepCmd
		}
	default:
		// Delegar otros tipos de mensajes al step model actual
		current := m.steps[m.index]
		updatedModel, stepCmd := current.Model.Update(msg)
		m.steps[m.index].Model = updatedModel
		cmd = stepCmd
	}
	return m, cmd
}

func (m Main) View() string {
	const minWidth = 80
	const minHeight = 40

	if m.width < minWidth || m.height < minHeight {
		return "Minimum size required: width >= " + strconv.Itoa(minWidth) + ", height >= " + strconv.Itoa(minHeight)
	}
	if len(m.steps) == 0 {
		return "No steps available"
	}

	current := m.steps[m.index]

	// Current step header
	stepHeader := styles.HeaderStyle.Render(fmt.Sprintf("Step %d: %s", m.index+1, current.Model.GetTitle()))

	// Current step content
	stepContent := styles.ContentStyle.Render(current.Model.View())

	// Help text
	helpText := m.renderHelpText()

	// Combine all elements with proper spacing
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		stepHeader,
		"",
		stepContent,
		"",
		helpText,
	)

	return styles.ContainerStyle.Render(content)
}

func (m Main) renderHelpText() string {
	var helpItems []string

	helpItems = append(helpItems, "pgdown: prev step")
	helpItems = append(helpItems, "pgup: next step")

	helpItems = append(helpItems, "ctrl+c/q: quit")

	helpText := strings.Join(helpItems, " • ")
	return styles.HelpStyle.Render(helpText)
}
