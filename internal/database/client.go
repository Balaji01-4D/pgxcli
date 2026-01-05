package database

import (
	"context"
	"fmt"
	"os"
	osUser "os/user"
	"strings"
	"time"

	"github.com/balaji01-4d/pgxcli/internal/config"
	"github.com/balaji01-4d/pgxcli/internal/logger"
	"github.com/balaji01-4d/pgxspecial"

	"github.com/jackc/pgx/v5"
)

type Client struct {
	CurrentDB           string
	Executor            *Executor
	ForcePasswordPrompt bool
	NeverPasswordPrompt bool

	now time.Time
}

func New(neverPasswordPrompt, forcePasswordPrompt bool, ctx context.Context, cfg config.Config) *Client {
	postgres := &Client{
		NeverPasswordPrompt: neverPasswordPrompt,
		ForcePasswordPrompt: forcePasswordPrompt,
		now:                 time.Now(),
	}
	return postgres
}

func (c *Client) Connect(ctx context.Context, host, user, password, database, dsn string, port uint16) error {
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

	if c.NeverPasswordPrompt && password == "" {
		password = os.Getenv("PGPASSWORD")
	}

	if c.ForcePasswordPrompt && password == "" {
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

	exec, err := NewExecutor(host, database, user, password, port, dsn, ctx)
	if err != nil {
		return err
	}
	c.Executor = exec
	c.CurrentDB = database
	logger.Log.Info("Database connection established", "database", database, "user", user)

	return nil
}

func (c *Client) ConnectDSN(ctx context.Context, dsn string) error {
	return c.Connect(ctx, "", "", "", "", dsn, 0)
}

func (c *Client) ConnectURI(ctx context.Context, uri string) error {
	parsedURI, err := pgx.ParseConfig(uri)
	if err != nil {
		return fmt.Errorf("failed to parse URI: %w", err)
	}
	return c.Connect(ctx, parsedURI.Host, parsedURI.User, parsedURI.Password, parsedURI.Database, "", parsedURI.Port)
}

func (c *Client) ExecuteSpecial(ctx context.Context,
	command string) (pgxspecial.SpecialCommandResult, bool, error) {
	result, okay, err := pgxspecial.ExecuteSpecialCommand(ctx, c.Executor.Conn, command)
	logger.Log.Info("Executed special command", "command", command, "result", result, "okay", okay, "err", err)
	return result, okay, err
}

func (c *Client) ExecuteQuery(ctx context.Context, query string) (Result, error) {
	return c.Executor.Execute(ctx, query)
}

func (c *Client) IsConnected() bool {
	return c.Executor != nil && c.Executor.IsConnected()
}

func (c *Client) ChangeDatabase(ctx context.Context, dbName string) error {
	if !c.IsConnected() {
		return fmt.Errorf("not connected to any database")
	}

	exec, err := NewExecutor(
		c.Executor.Host,
		dbName,
		c.Executor.User,
		c.Executor.Password,
		c.Executor.Port,
		"",
		ctx,
	)
	if err != nil {
		return err
	}
	c.Executor = exec
	c.CurrentDB = dbName
	logger.Log.Info("Database changed", "database", dbName)

	return nil
}

func (c *Client) ParsePrompt(str string) string {
	str = strings.ReplaceAll(str, "\\t", c.now.Format("02/06/2006 15:04:05"))
	if c.Executor.User != "" {
		str = strings.ReplaceAll(str, "\\u", c.Executor.User)
	} else {
		str = strings.ReplaceAll(str, "\\u", "(nil)")
	}

	if c.Executor.Host != "" {
		str = strings.ReplaceAll(str, "\\H", c.Executor.Host)
		str = strings.ReplaceAll(str, "\\h", func() string {
			return strings.Split(c.Executor.Host, ".")[0]
		}())
	} else {
		str = strings.ReplaceAll(str, "\\H", "(nil)")
		str = strings.ReplaceAll(str, "\\h", "(nil)")
	}

	if c.CurrentDB != "" {
		str = strings.ReplaceAll(str, "\\d", c.CurrentDB)
	} else {
		str = strings.ReplaceAll(str, "\\d", "(nil)")
	}
	if c.Executor.Port != 0 {
		str = strings.ReplaceAll(str, "\\p", fmt.Sprintf("%d", c.Executor.Port))
	} else {
		str = strings.ReplaceAll(str, "\\p", "5432")
	}

	str = strings.ReplaceAll(str, "\\n", "\n")

	return str
}

func (c *Client) GetUser() string {
	return c.Executor.User
}

func (c *Client) GetDatabase() string {
	return c.Executor.Database
}

func (c *Client) GetPort() uint16 {
	return c.Executor.Port
}

func (c *Client) GetHost() string {
	return c.Executor.Host
}

func (c *Client) GetConnectionInfo() {
	logger.Log.Debug("Connection information",
		"connection string", c.Executor.Conn.Config().ConnString(),
		"host", c.Executor.Host,
		"Port", c.Executor.Port,
		"Database", c.Executor.Database,
		"User", c.Executor.User,
		"URI", c.Executor.URI,
	)
}

func (c *Client) Close(ctx context.Context) {
	if c.Executor != nil {
		c.Executor.Close(ctx)
	}
}
