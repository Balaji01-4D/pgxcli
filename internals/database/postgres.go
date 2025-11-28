package database

import (
	"context"
	"fmt"
	"os"
	"pgcli/internals/logger"
	"pgcli/internals/repl"
	"strings"

	osUser "os/user"

	"github.com/jackc/pgx/v5"
)

type Postgres struct {
	CurrentBD           string
	Executor            *Executor
	ForcePasswordPrompt bool
	NeverPasswordPrompt bool
	ctx                 context.Context
}

func NewPostgres(neverPasswordPrompt, forcePasswordPrompt bool, ctx context.Context) *Postgres {
	return &Postgres{
		NeverPasswordPrompt: neverPasswordPrompt,
		ForcePasswordPrompt: forcePasswordPrompt,
		ctx:                 ctx,
	}
}

func (p *Postgres) Connect(host, user, password, database, dsn string, port uint16) error {

	if user == "" {
		currentUser, err := osUser.Current()
		if err != nil {
			return fmt.Errorf("failed to get current user: %w", err)
		}
		user = currentUser.Username
	}

	if database == "" {
		database = user
	}

	if p.NeverPasswordPrompt && password == "" {
		password = os.Getenv("PGPASSWORD")
	}

	if p.ForcePasswordPrompt && password == "" {
		fmt.Print("Password: ")
		var pwd string
		fmt.Scanln(&pwd)
		password = strings.TrimSpace(pwd)
	}

	if dsn != "" {
		parsedDsn, err := pgx.ParseConfig(dsn)
		if err != nil {
			return fmt.Errorf("failed to parse DSN: %w", err)
		}

		host = parsedDsn.Host
		port = parsedDsn.Port
	}

	exec, err := NewExecutor(host, database, user, password, port, dsn, p.ctx)
	if err != nil {
		return err
	}
	p.Executor = exec
	p.CurrentBD = database
	logger.Log.Info("Database connection established", "database", database, "user", user)

	return nil

}

func (p *Postgres) ConnectDSN(dsn string) error {
	return p.Connect("", "", "", "", dsn, 0)
}

func (p *Postgres) ConnectURI(uri string) error {
	parsedURI, err := pgx.ParseConfig(uri)
	if err != nil {
		return fmt.Errorf("failed to parse URI: %w", err)
	}
	return p.Connect(parsedURI.Host, parsedURI.User, parsedURI.Password, parsedURI.Database, "", parsedURI.Port)
}

func (p *Postgres) Close() {
	if p.Executor != nil {
		p.Executor.Close()
	}
}

func (p *Postgres) IsConnected() bool {
	return p.Executor != nil && p.Executor.IsConnected()
}

func (p *Postgres) GetConnectionInfo() {
	logger.Log.Debug("Connection information",
		"connection string", p.Executor.Pool.Config().ConnString(),
		"host", p.Executor.Host,
		"Port", p.Executor.Port,
		"Database", p.Executor.Database,
		"User", p.Executor.User,
		"URI", p.Executor.URI,
	)
}

func (p *Postgres) RunCli() error {
	if !p.IsConnected() {
		return fmt.Errorf("not connected to any database")
	}

	for {
		query, err := repl.RunRepl(p.CurrentBD)
		if err != nil {
			return err
		}
		if strings.TrimSpace(query) == "exit" || strings.TrimSpace(query) == "quit" {
			break
		}

		result, err := p.Executor.Execute(p.ctx, query)
		if err != nil {
			logger.Log.Error("Query execution failed", "error", err)
			fmt.Fprintf(os.Stderr, "Error executing query: %v\n", err)
			continue
		}
		HandleResult(result)
	}

	return nil
}

func HandleResult(result Result) {
	switch res := result.(type) {
	case *QueryResult:
		fmt.Println(res.Render())
	case *ExecResult:
		fmt.Println(res.Render())
	default:
		fmt.Println("Unknown result type")
	}
}
