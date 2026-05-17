package database

import (
	"context"
	"net"
	"time"

	"github.com/jackc/pgx/v5"
)

// Connector describes how the client obtains and updates a database connection.
type Connector interface {
	Connect(ctx context.Context) (*pgx.Conn, error)
	UpdatePassword(password string)
	Password() string
}

// pgConnector holds pgx connection configuration and creates database connections.
type pgConnector struct {
	cfg *pgx.ConnConfig
}

// NewPGConnectorFromConnString builds a connector from a PostgreSQL connection string.
func NewPGConnectorFromConnString(connString string) (Connector, error) {
	cfg, err := pgx.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	return &pgConnector{cfg: cfg}, nil
}

// NewPGConnectorFromFields builds a connector from individual connection fields.
func NewPGConnectorFromFields(host, database, user, password string, port uint16) (Connector, error) {
	cfg, err := pgx.ParseConfig("")
	if err != nil {
		return nil, err
	}

	checkAndSet := func(field *string, value string) {
		if value != "" {
			*field = value
		}
	}
	checkAndSet(&cfg.Host, host)
	checkAndSet(&cfg.Database, database)
	checkAndSet(&cfg.User, user)
	checkAndSet(&cfg.Password, password)

	if port != 0 {
		cfg.Port = port
	}
	return &pgConnector{cfg: cfg}, nil
}

// UpdatePassword updates the password on the underlying connection config.
func (c *pgConnector) UpdatePassword(newPassword string) {
	c.cfg.Password = newPassword
}

// Password returns the password from the underlying connection config.
func (c *pgConnector) Password() string {
	return c.cfg.Password
}

// Connect opens a new pgx connection using the connector configuration.
func (c *pgConnector) Connect(ctx context.Context) (*pgx.Conn, error) {
	c.cfg.DefaultQueryExecMode = pgx.QueryExecModeExec

	dialer := &net.Dialer{}
	dialer.Timeout = 5 * time.Second
	if c.cfg.ConnectTimeout > 0 {
		dialer.Timeout = c.cfg.ConnectTimeout
	}
	c.cfg.DialFunc = dialer.DialContext

	conn, err := pgx.ConnectConfig(ctx, c.cfg)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
