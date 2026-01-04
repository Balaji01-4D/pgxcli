package database

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/balaji01-4d/pgxcli/internal/config"
	"github.com/balaji01-4d/pgxcli/internal/logger"

	osUser "os/user"

	"github.com/jackc/pgx/v5"
)

const (
	DefaultPrompt = `\u@\h:\d> `
	MaxLenPrompt  = 30
)

type Client struct {
	CurrentDB           string
	Executor            *Executor
	ForcePasswordPrompt bool
	NeverPasswordPrompt bool
	ctx                 context.Context
	Config              config.Config

	now time.Time
}

func New(neverPasswordPrompt, forcePasswordPrompt bool, ctx context.Context, cfg config.Config) *Client {
	postgres := &Client{
		NeverPasswordPrompt: neverPasswordPrompt,
		ForcePasswordPrompt: forcePasswordPrompt,
		ctx:                 ctx,
		Config:              cfg,
		now:                 time.Now(),
	}
	return postgres
}

func (p *Client) Connect(host, user, password, database, dsn string, port uint16) error {
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
	p.CurrentDB = database
	logger.Log.Info("Database connection established", "database", database, "user", user)

	return nil
}

func (p *Client) ConnectDSN(dsn string) error {
	return p.Connect("", "", "", "", dsn, 0)
}

func (p *Client) ConnectURI(uri string) error {
	parsedURI, err := pgx.ParseConfig(uri)
	if err != nil {
		return fmt.Errorf("failed to parse URI: %w", err)
	}
	return p.Connect(parsedURI.Host, parsedURI.User, parsedURI.Password, parsedURI.Database, "", parsedURI.Port)
}

func (p *Client) Close() {
	if p.Executor != nil {
		p.Executor.Close(p.ctx)
	}
}

func (p *Client) IsConnected() bool {
	return p.Executor != nil && p.Executor.IsConnected()
}

func (p *Client) GetConnectionInfo() {
	logger.Log.Debug("Connection information",
		"connection string", p.Executor.Conn.Config().ConnString(),
		"host", p.Executor.Host,
		"Port", p.Executor.Port,
		"Database", p.Executor.Database,
		"User", p.Executor.User,
		"URI", p.Executor.URI,
	)
}

func (p *Client) ChangeDatabase(dbName string) error {
	if !p.IsConnected() {
		return fmt.Errorf("not connected to any database")
	}

	exec, err := NewExecutor(
		p.Executor.Host,
		dbName,
		p.Executor.User,
		p.Executor.Password,
		p.Executor.Port,
		"",
		p.ctx,
	)
	if err != nil {
		return err
	}
	p.Executor = exec
	p.CurrentDB = dbName
	logger.Log.Info("Database changed", "database", dbName)

	return nil
}
