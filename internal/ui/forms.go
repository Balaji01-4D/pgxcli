package ui

import (
	"errors"
	"fmt"
	"image/color"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

const maxWidth = 80

type Styles struct {
	Base,
	HeaderText,
	Status,
	StatusHeader,
	Highlight,
	ErrorHeaderText,
	Help lipgloss.Style

	Red, Indigo, Green color.Color
}

func NewStyles(hasDarkBg bool) *Styles {
	var (
		s         = Styles{}
		lightDark = lipgloss.LightDark(hasDarkBg)
	)

	s.Red = lightDark(lipgloss.Color("#FE5F86"), lipgloss.Color("#FE5F86"))
	s.Indigo = lightDark(lipgloss.Color("#5A56E0"), lipgloss.Color("#7571F9"))
	s.Green = lightDark(lipgloss.Color("#02BA84"), lipgloss.Color("#02BF87"))
	s.Base = lipgloss.NewStyle().
		Padding(1, 4, 0, 1)
	s.HeaderText = lipgloss.NewStyle().
		Foreground(s.Indigo).
		Bold(true).
		Padding(0, 1, 0, 2)
	s.Status = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.Indigo).
		PaddingLeft(1).
		MarginTop(1)
	s.StatusHeader = lipgloss.NewStyle().
		Foreground(s.Green).
		Bold(true)
	s.Highlight = lipgloss.NewStyle().
		Foreground(lipgloss.Color("212"))
	s.ErrorHeaderText = s.HeaderText.
		Foreground(s.Red)
	s.Help = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
	return &s
}

type state int

const (
	statusNormal state = iota
	stateDone
)

var ErrFormAborted = errors.New("connection form aborted")

type ConnectionValues struct {
	Database string
	Username string
	Port     string
	Host     string
	Password string
}

type Model struct {
	state     state
	styles    func(bool) *Styles
	form      *huh.Form
	values    ConnectionValues
	hasDarkBg bool
	width     int
}

func NewModel(database, username, host, port string) *Model {
	port = strings.TrimSpace(port)
	if port == "" || port == "0" {
		port = "5432"
	}

	m := &Model{
		width:  maxWidth,
		styles: NewStyles,
		values: ConnectionValues{
			Database: database,
			Username: username,
			Host:     host,
			Port:     port,
		},
	}

	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("database").
				Title("Database").
				Placeholder("postgres").
				Validate(huh.ValidateNotEmpty()).
				Value(&m.values.Database),
			huh.NewInput().
				Key("username").
				Title("Username").
				Validate(huh.ValidateNotEmpty()).
				Value(&m.values.Username),
			huh.NewInput().
				Key("port").
				Title("Port").
				Placeholder("5432").
				Validate(validatePort).
				Value(&m.values.Port),
			huh.NewInput().
				Key("host").
				Title("Host").
				Description("Database host address, leave blank for unix socket.").
				Placeholder("localhost").
				Value(&m.values.Host),
			huh.NewInput().
				Key("password").
				Title("Password").
				Description("leave blank to use default").
				EchoMode(huh.EchoModePassword).
				Value(&m.values.Password),
		),
	).
		WithWidth(45).
		WithShowHelp(false).
		WithShowErrors(false)
	return m
}

func (m *Model) Init() tea.Cmd {
	return m.form.Init()
}

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	styles := m.styles(m.hasDarkBg)
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		m.hasDarkBg = msg.IsDark()
	case tea.WindowSizeMsg:
		m.width = min(msg.Width, maxWidth) - styles.Base.GetHorizontalFrameSize()
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Interrupt
		case "esc", "q":
			return m, tea.Quit
		}
	}

	var cmds []tea.Cmd

	// Process the form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	if m.form.State == huh.StateCompleted {
		// Quit when the form is done.
		cmds = append(cmds, tea.Quit)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() tea.View {
	s := m.styles(m.hasDarkBg)

	switch m.form.State {
	case huh.StateCompleted:
		var b strings.Builder
		fmt.Fprintf(&b, "Connection details captured.\n\n")
		fmt.Fprintf(&b, "Database: %s\n", fallback(m.values.Database, "(None)"))
		fmt.Fprintf(&b, "Username: %s\n", fallback(m.values.Username, "(None)"))
		fmt.Fprintf(&b, "Port: %s\n", fallback(m.values.Port, "(None)"))
		fmt.Fprintf(&b, "Host: %s\n", fallback(m.values.Host, "(None)"))
		fmt.Fprintf(&b, "Password: %s", maskedPassword(m.values.Password))
		return tea.NewView(s.Status.Margin(0, 1).Padding(1, 2).Width(48).Render(b.String()) + "\n\n")
	default:
		// Form (left side)
		v := strings.TrimSuffix(m.form.View(), "\n\n")
		form := lipgloss.NewStyle().Margin(1, 0).Render(v)

		// Status (right side)
		var status string
		{
			buildInfo := strings.Join([]string{
				"Database: " + fallback(m.form.GetString("database"), "(None)"),
				"Username: " + fallback(m.form.GetString("username"), "(None)"),
				"Port: " + fallback(m.form.GetString("port"), "(None)"),
				"Host: " + fallback(m.form.GetString("host"), "(None)"),
				"Password: " + maskedPassword(m.form.GetString("password")),
			}, "\n")

			const statusWidth = 28
			statusMarginLeft := m.width - statusWidth - lipgloss.Width(form) - s.Status.GetMarginRight()
			if statusMarginLeft < 0 {
				statusMarginLeft = 0
			}
			status = s.Status.
				Height(lipgloss.Height(form)).
				Width(statusWidth).
				MarginLeft(statusMarginLeft).
				Render(s.StatusHeader.Render("Current Input") + "\n" + buildInfo)
		}

		errors := m.form.Errors()
		header := m.appBoundaryView("PostgreSQL Connection Form")
		if len(errors) > 0 {
			header = m.appErrorBoundaryView(m.errorView())
		}
		body := lipgloss.JoinHorizontal(lipgloss.Left, form, status)

		footer := m.appBoundaryView(m.form.Help().ShortHelpView(m.form.KeyBinds()))
		if len(errors) > 0 {
			footer = m.appErrorBoundaryView("")
		}

		return tea.NewView(s.Base.Render(header + "\n" + body + "\n\n" + footer))
	}
}

func (m *Model) errorView() string {
	var s strings.Builder
	for _, err := range m.form.Errors() {
		s.WriteString(err.Error())
	}
	return s.String()
}

func (m *Model) appBoundaryView(text string) string {
	s := m.styles(m.hasDarkBg)
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		s.HeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Foreground(s.Indigo)),
	)
}

func (m *Model) appErrorBoundaryView(text string) string {
	s := m.styles(m.hasDarkBg)
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		s.ErrorHeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Foreground(s.Red)),
	)
}

func (m *Model) Result() (ConnectionValues, error) {
	if m.form.State != huh.StateCompleted {
		return ConnectionValues{}, ErrFormAborted
	}

	port, err := strconv.Atoi(strings.TrimSpace(m.values.Port))
	if err != nil || port < 1 || port > 65535 {
		return ConnectionValues{}, fmt.Errorf("invalid port: %q", m.values.Port)
	}

	values := m.values
	return values, nil
}

func RunConnectionForm(database, username, host, port string) (ConnectionValues, error) {
	finalModel, err := tea.NewProgram(NewModel(database, username, host, port)).Run()
	if err != nil {
		return ConnectionValues{}, err
	}

	m, ok := finalModel.(*Model)
	if !ok {
		return ConnectionValues{}, fmt.Errorf("unexpected model type: %T", finalModel)
	}

	return m.Result()
}

func fallback(value string, empty string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return empty
	}
	return value
}

func maskedPassword(password string) string {
	if password == "" {
		return "(None)"
	}
	return strings.Repeat("•", len(password))
}

func validatePort(v string) error {
	v = strings.TrimSpace(v)
	if v == "" {
		return errors.New("port is required")
	}
	p, err := strconv.Atoi(v)
	if err != nil {
		return errors.New("port must be a number")
	}
	if p < 1 || p > 65535 {
		return errors.New("port must be between 1 and 65535")
	}
	return nil
}
