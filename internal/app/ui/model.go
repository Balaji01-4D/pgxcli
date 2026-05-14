// Package ui contains the BubbleTea model for pgxcli's interactive prompt.
package ui

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Balaji01-4D/bubbline/computil"
	"github.com/Balaji01-4D/bubbline/editline"
	"github.com/Balaji01-4D/bubbline/history"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/balaji01-4d/pgxcli/internal/config"
	"github.com/muesli/termenv"
)

var chromaFormatter = detectTerminalColorProfile()

var (
	userInputStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#908CAA"))
	appOutputStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#E0DEF4"))
	errorOutputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
	statusBarStyle   = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#908CAA")).
				Background(lipgloss.Color("#2A273F")).
				Padding(0, 1)
)

// ReadyMsg signals the ui that execution is done and it should prompt.
type ReadyMsg struct{ Prefix string }

// ExecCmdMsg is used to dispatch a batch/sequence of commands.
type ExecCmdMsg struct{ Cmd tea.Cmd }

type Model struct {
	input         *editline.Model
	width, height int
	executing     bool
	prevUserInput string
	historyFile   string

	// execute executes a query passed and return as ExecCmdMsg + ReadyMsg.
	execute func(string) tea.Cmd
}

func New(initialPrefix string, pgKeywords []string, historyFile string, executeFunc func(string) tea.Cmd) (*Model, error) {
	el := editline.New(0, 0)
	el.Prompt = initialPrefix
	if historyFile == "" || historyFile == config.Default {
		historyFile = getHistoryFilePath()
	}

	if err := applyEditlineConfig(el, historyFile, pgKeywords); err != nil {
		return nil, fmt.Errorf("applying input config: %w", err)
	}

	return &Model{
		input:       el,
		historyFile: historyFile,
		execute:     executeFunc,
	}, nil
}

func (m *Model) Init() tea.Cmd {
	return m.input.Focus()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case ReadyMsg:
		m.executing = false
		if msg.Prefix != "" {
			m.input.Prompt = msg.Prefix
		}
		m.input.Reset()
		return m, nil

	case ExecCmdMsg:
		return m, msg.Cmd

	case editline.InputCompleteMsg:
		return m.handleInput()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.SetSize(msg.Width, msg.Height-4)
		return m, nil

	case tea.KeyMsg:
		if m.executing {
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c":
			m.input.Reset()
			return m, nil
		}
	}

	var nextCmd tea.Cmd
	m.input, nextCmd = m.input.Update(msg)
	return m, nextCmd
}

func (m *Model) handleInput() (tea.Model, tea.Cmd) {
	input := m.input.Value()
	trimmed := strings.TrimSpace(input)

	if trimmed == "" {
		return m, tea.Sequence(
			m.printUserInput(userInputStyle.Render(m.input.Prompt), ""),
			func() tea.Msg {
				return ReadyMsg{Prefix: m.input.Prompt}
			},
		)
	}

	m.prevUserInput = input
	m.executing = true
	m.input.AddHistoryEntry(input)

	return m, tea.Sequence(
		m.printUserInput(userInputStyle.Render(m.input.Prompt), input),
		m.execute(trimmed),
	)
}

func (m *Model) printUserInput(prefix, input string) tea.Cmd {
	var highlightedInput string
	if input != "" {
		highlightedInput = postgresHighlighter("monokai")(input)
	}

	userContent := lipgloss.JoinHorizontal(lipgloss.Left, userInputStyle.Render(prefix), highlightedInput)
	return tea.Printf("%s", userContent)
}

func (m *Model) View() tea.View {
	statusStyle := statusBarStyle.Width(m.width)
	if m.executing {
		statusBar := statusStyle.AlignVertical(lipgloss.Bottom).Render("pgxcli")
		return tea.NewView(statusBar)
	}

	str := lipgloss.Sprintf("%s\n%s", m.input.View(), statusStyle.Render("pgxcli"))
	return tea.NewView(str)
}

func (m *Model) saveHistory() error {
	if m.historyFile == "" {
		return nil
	}
	return history.SaveHistory(m.input.GetHistory(), m.historyFile)
}

func (m *Model) Close() error {
	if err := m.saveHistory(); err != nil {
		return fmt.Errorf("saving history: %w", err)
	}
	return nil
}

// PrintCmd returns a command that prints formatted text.
func PrintCmd(text string) tea.Cmd {
	formattedInteraction := lipgloss.Sprintf(
		"%s\n",
		appOutputStyle.Render(text),
	)
	return tea.Printf("%s", formattedInteraction)
}

// PrintErrCmd returns a command that prints a formatted error.
func PrintErrCmd(err error) tea.Cmd {
	formattedError := lipgloss.Sprintf(
		"%s\n",
		errorOutputStyle.Render("✗ "+err.Error()),
	)
	return tea.Printf("%s", formattedError)
}

// ShowPagerCmd returns a command to execute a pager process.
func ShowPagerCmd(cmd *exec.Cmd) tea.Cmd {
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			return nil
		}
		return nil
	})
}

func postgresHighlighter(style string) func(string) string {
	return func(s string) string {
		var buf bytes.Buffer
		if err := quick.Highlight(&buf, s, "postgresql", chromaFormatter, style); err != nil {
			return s
		}
		return buf.String()
	}
}

func postgresAutocomplete(pgKeywords []string) func(v [][]rune, line, col int) (string, editline.Completions) {
	return func(v [][]rune, line, col int) (string, editline.Completions) {
		word, wstart, wend := computil.FindWord(v, line, col)
		if word == "" {
			return "", nil
		}
		upperWord := strings.ToUpper(word)
		var matches []string
		for _, kw := range pgKeywords {
			if strings.HasPrefix(kw, upperWord) {
				matches = append(matches, kw)
			}
		}
		if len(matches) == 0 {
			return "", nil
		}
		return "", editline.SimpleWordsCompletion(matches, "Keywords", col, wstart, wend)
	}
}

func detectTerminalColorProfile() string {
	switch termenv.ColorProfile() {
	case termenv.TrueColor:
		return "terminal16m"
	case termenv.ANSI256:
		return "terminal256"
	case termenv.ANSI:
		return "terminal16"
	default:
		return "noop"
	}
}

func applyEditlineConfig(el *editline.Model, historyFile string, pgKeywords []string) error {
	el.SetHelpDisabled(true)
	el.SetHighlighter(postgresHighlighter("monokai"))

	el.AutoComplete = postgresAutocomplete(pgKeywords)

	entries, err := history.LoadHistory(historyFile)
	if err != nil {
		return fmt.Errorf("loading history: %w", err)
	}

	el.SetHistory(entries)
	return nil
}

func getHistoryFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".pgxcli_history.jsonl")
}
