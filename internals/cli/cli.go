package cli

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)


type CLI struct {
	CurrentDatabase string
	prompt		  string
	input		  textinput.Model
}


func InitModel(database string) *CLI {
	ti := textinput.New()
	ti.Placeholder = "Enter SQL command..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50

	cli := &CLI{
		CurrentDatabase: database,
		prompt:          fmt.Sprintf("(%s)> ", database),
		input:           ti,
	}
	return cli
}


func (c CLI) Init() tea.Cmd {
	return nil
}

func (m CLI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m CLI) View() string {
	return m.prompt + "\n" + m.input.View()
}