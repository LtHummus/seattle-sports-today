package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	noStyle      = lipgloss.NewStyle()

	helpStyle           = blurredStyle
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	focusedButton = focusedStyle.Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))

	cursorStyle = focusedStyle
)

type model struct {
	focusIndex int
	inputs     []textinput.Model
	cursorMode cursor.Mode

	processing bool
	quitting   bool
	spinner    spinner.Model
}

func initialModel() model {
	m := model{
		inputs: make([]textinput.Model, 4),
	}

	m.spinner = spinner.New()
	m.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	m.spinner.Spinner = spinner.Line

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 60
		t.PromptStyle = focusedStyle
		t.TextStyle = focusedStyle

		switch i {
		case 0:
			t.Placeholder = "Date (yyyy-mm-dd)"
			t.Focus()
			t.CharLimit = 10
		case 1:
			t.Placeholder = "Clock Time"
			t.CharLimit = 8
		case 2:
			t.Placeholder = "Team Name"
		case 3:
			t.Placeholder = "Venue"
		}

		m.inputs[i] = t
	}

	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m model) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" && m.focusIndex == len(m.inputs) {
				m.processing = true
			}

			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
				} else {
					m.inputs[i].Blur()
					m.inputs[i].PromptStyle = noStyle
					m.inputs[i].TextStyle = noStyle
				}
			}

			return m, tea.Batch(cmds...)
		}
	}

	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m model) updateProcessing(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		default:
			return m, nil
		}

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.processing {
		return m.updateForm(msg)
	} else {
		return m.updateProcessing(msg)
	}
}

func (m model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m model) View() string {
	var b strings.Builder

	if !m.processing {
		for i := range m.inputs {
			b.WriteString(m.inputs[i].View())
			if i < len(m.inputs)-1 {
				b.WriteRune('\n')
			}
		}

		button := &blurredButton
		if m.focusIndex == len(m.inputs) {
			button = &focusedButton
		}
		fmt.Fprintf(&b, "\n\n%s\n\n", *button)
	} else {
		fmt.Fprintf(&b, "\n\n %s Writing....\n\n", m.spinner.View())
		if m.quitting {
			fmt.Fprintf(&b, "\n")
		}
	}

	return b.String()
}

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Printf("could not initialize: %s\n", err)
		os.Exit(1)
	}
}
