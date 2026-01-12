package database

import (
	"context"

	"github.com/balaji01-4d/pgxspecial"
	"github.com/balaji01-4d/pgxspecial/database"
	_ "github.com/balaji01-4d/pgxspecial/dbcommands"
)

const (
	Exit pgxspecial.SpecialResultKind = 100 + iota
	ChangeDB
	Conninfo
)

func init() {
	registerSpecialCommands()
}

func registerSpecialCommands() {
	pgxspecial.RegisterCommand(pgxspecial.SpecialCommandRegistry{
		Cmd:         "\\q",
		Syntax:      "\\q",
		Description: "Quit Pgxcli",
		Handler: func(_ context.Context, _ database.Queryer, _ string, _ bool) (pgxspecial.SpecialCommandResult, error) {
			return ExitAction{}, nil
		},
		CaseSensitive: true,
	})

	pgxspecial.RegisterCommand(pgxspecial.SpecialCommandRegistry{
		Cmd:         "\\c",
		Syntax:      "\\c database_name",
		Description: "Change a new database",
		Handler: func(_ context.Context, _ database.Queryer, s string, _ bool) (pgxspecial.SpecialCommandResult, error) {
			return ChangeDbAction{Name: s}, nil
		},
		CaseSensitive: true,
		Alias:         []string{"\\connect"},
	})

	pgxspecial.RegisterCommand(pgxspecial.SpecialCommandRegistry{
		Cmd:         "\\conninfo",
		Syntax:      "\\conninfo",
		Description: "Get connection details",
		Handler: func(ctx context.Context, db database.Queryer, args string, verbose bool) (pgxspecial.SpecialCommandResult, error) {
			return ConnInfoAction{}, nil
		},
		CaseSensitive: false,
	})
}

type ExitAction struct{}

func (e ExitAction) ResultKind() pgxspecial.SpecialResultKind {
	return Exit
}

type ChangeDbAction struct {
	Name string
}

func (c ChangeDbAction) ResultKind() pgxspecial.SpecialResultKind {
	return ChangeDB
}

type ConnInfoAction struct{}

func (g ConnInfoAction) ResultKind() pgxspecial.SpecialResultKind {
	return Conninfo
}
