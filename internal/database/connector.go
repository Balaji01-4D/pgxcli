package database

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Connector interface {
	Connect(ctx context.Context) (*pgx.Conn, error)
	UpdatePassword(password string)
}

type PGConnector struct {
	cfg *pgx.ConnConfig
}

func NewPGConnectorFromConnString(connString string) (*PGConnector, error) {
	cfg, err := pgx.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	return &PGConnector{cfg: cfg}, nil
}

func NewPGConnectorFromFields(host, database, user, password string, port uint16) (*PGConnector, error) {
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
	return &PGConnector{cfg: cfg}, nil
}

func (c *PGConnector) UpdatePassword(newPassword string) {
	c.cfg.Password = newPassword
}

func (c *PGConnector) Connect(ctx context.Context) (*pgx.Conn, error) {
	conn, err := pgx.ConnectConfig(ctx, c.cfg)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
