// Package database manages PostgreSQL connections, execution, and special commands.
package database

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/balaji01-4d/pgxcli/internal/database/result"
	"github.com/balaji01-4d/pgxspecial"
)

const nilPlaceholder = "(nil)"

// Client provides high-level connection management and SQL execution operations.
type Client struct {
	currentDB string
	executor  *executor

	now time.Time

	logger *slog.Logger
}

// New creates a database client with logger-backed connection lifecycle reporting.
func New(logger *slog.Logger) *Client {
	postgres := &Client{
		now:    time.Now(),
		logger: logger,
	}
	return postgres
}

// Connect opens a database connection using the provided connector.
func (c *Client) Connect(ctx context.Context, connector Connector) error {
	exec, err := newExecutor(ctx, connector, c.logger)
	if err != nil {
		return err
	}
	c.executor = exec
	c.currentDB = exec.Database
	c.logger.Info("Database connection established", "database", exec.Database, "user", exec.User)

	return nil
}

// ExecuteSpecial executes a pgxspecial command (for example: \q, \c, \conninfo).
func (c *Client) ExecuteSpecial(ctx context.Context, command string) (pgxspecial.SpecialCommandResult, bool, error) {
	return c.executor.executeSpecial(ctx, command)
}

// QueryOne runs SQL through the underlying executor and returns typed results.
func (c *Client) QueryOne(ctx context.Context, query string) (result.Result, error) {
	return c.executor.queryOne(ctx, query)
}

// Query runs multiple SQL queries in a single string.
// by sending them together, we can save round trips to the database.
func (c *Client) Query(ctx context.Context, query string) (result.Result, error) {
	return c.executor.query(ctx, query)
}

// IsConnected reports whether the client currently has an active connection.
func (c *Client) IsConnected() bool {
	return c.executor != nil && c.executor.isConnected()
}

// ChangeDatabase reconnects to the same server with a different database name.
func (c *Client) ChangeDatabase(ctx context.Context, dbName string) error {
	if !c.IsConnected() {
		return fmt.Errorf("not connected to any database")
	}

	dbName = strings.TrimSpace(dbName)
	if dbName == "" {
		return fmt.Errorf("database name is required")
	}

	connConfig := c.executor.Conn.Config().Copy()
	connConfig.Database = dbName

	connector := &pgConnector{cfg: connConfig}
	oldExecutor := c.executor

	exec, err := newExecutor(
		ctx,
		connector,
		c.logger,
	)
	if err != nil {
		return err
	}

	c.executor = exec
	c.currentDB = exec.Database

	if oldExecutor != nil {
		if err := oldExecutor.close(ctx); err != nil {
			c.logger.Error("Failed to close previous connection after database switch", "error", err)
			return fmt.Errorf("database changed to %s but failed to close previous connection: %w", exec.Database, err)
		}
	}

	c.logger.Info("Database changed", "database", exec.Database)

	return nil
}

// ParsePrompt resolves prompt placeholders using current connection metadata.
func (c *Client) ParsePrompt(str string) string {
	var user, host, shortHost, db, port string

	t := c.now.Format("02/06/2006 15:04:05")

	if c.executor == nil {
		user, host, shortHost, db, port = nilPlaceholder, nilPlaceholder, nilPlaceholder, nilPlaceholder, "5432"
	} else {
		if c.executor.User != "" {
			user = c.executor.User
		} else {
			user = nilPlaceholder
		}

		if c.executor.Host != "" {
			host = c.executor.Host
			idx := strings.IndexByte(host, '.')
			if idx == -1 {
				shortHost = host
			} else {
				shortHost = host[:idx]
			}
		} else {
			host, shortHost = nilPlaceholder, nilPlaceholder
		}

		if c.currentDB != "" {
			db = c.currentDB
		} else {
			db = nilPlaceholder
		}

		if c.executor.Port != 0 {
			port = strconv.FormatUint(uint64(c.executor.Port), 10)
		} else {
			port = "5432"
		}
	}

	return strings.NewReplacer(
		`\t`, t,
		`\u`, user,
		`\H`, host,
		`\h`, shortHost,
		`\d`, db,
		`\p`, port,
		`\n`, "\n",
	).Replace(str)
}

// GetUser returns the current connection user name.
func (c *Client) GetUser() string {
	if c.executor == nil {
		return ""
	}
	return c.executor.User
}

// GetDatabase returns the current database name.
func (c *Client) GetDatabase() string {
	if c.executor == nil {
		return ""
	}
	return c.executor.Database
}

// GetPort returns the current connection port.
func (c *Client) GetPort() uint16 {
	if c.executor == nil {
		return 0
	}
	return c.executor.Port
}

// GetHost returns the current connection host.
func (c *Client) GetHost() string {
	if c.executor == nil {
		return ""
	}
	return c.executor.Host
}

// Ping verifies connectivity to the current database.
func (c *Client) Ping(ctx context.Context) error {
	if !c.IsConnected() {
		return fmt.Errorf("not connected to any database")
	}
	return c.executor.ping(ctx)
}

// Close closes the current database connection if one exists.
func (c *Client) Close(ctx context.Context) error {
	if c.executor != nil {
		return c.executor.close(ctx)
	}
	return nil
}
