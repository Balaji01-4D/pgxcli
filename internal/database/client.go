package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/balaji01-4d/pgxcli/internal/logger"
	"github.com/balaji01-4d/pgxspecial"
)

type Client struct {
	CurrentDB string
	Executor  *Executor

	now time.Time
}

func New() *Client {
	postgres := &Client{
		now: time.Now(),
	}
	return postgres
}

func (c *Client) Connect(ctx context.Context, connector Connector) error {
	exec, err := NewExecutor(ctx, connector)
	if err != nil {
		return err
	}
	c.Executor = exec
	c.CurrentDB = exec.Database
	logger.Log.Info("Database connection established", "database", exec.Database, "user", exec.User)

	return nil
}

func (c *Client) ExecuteSpecial(ctx context.Context,
	command string,
) (pgxspecial.SpecialCommandResult, bool, error) {
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

	connConfig := c.Executor.Conn.Config().Copy()
	connConfig.Database = dbName

	connector, err := NewPGConnectorFromConnString(connConfig.ConnString())
	if err != nil {
		return err
	}

	exec, err := NewExecutor(
		ctx,
		connector,
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
