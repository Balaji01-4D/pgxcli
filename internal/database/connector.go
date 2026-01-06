package database

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Connector interface {
	Connect(ctx context.Context) (*pgx.Conn, error)
	UpdatePassword(password string)
}

type ConnStringConnector struct {
	ConnString string
}

func NewConnStringConnector(connString string) *ConnStringConnector {
	return &ConnStringConnector{
		ConnString: connString,
	}
}

func (c *ConnStringConnector) UpdatePassword(newPassword string) {
	cfg, err := pgx.ParseConfig(c.ConnString)
	if err != nil {
		return
	}
	cfg.Password = newPassword
	c.ConnString = cfg.ConnString()
}

func (c *ConnStringConnector) Connect(ctx context.Context) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, c.ConnString)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

type ConfigConnector struct {
	Host     string
	Database string
	User     string
	Password string
	Port     uint16
}

func NewConfigConnector(host, database, user, password string, port uint16) *ConfigConnector {
	return &ConfigConnector{
		Host:     host,
		Database: database,
		User:     user,
		Password: password,
		Port:     port,
	}
}

func (c *ConfigConnector) UpdatePassword(newPassword string) {
	c.Password = newPassword
}

func (c *ConfigConnector) Connect(ctx context.Context) (*pgx.Conn, error) {
	connConfig, err := pgx.ParseConfig("")
	if err != nil {
		return nil, err
	}

	if c.Host != "" {
		connConfig.Host = c.Host
	}
	if c.Database != "" {
		connConfig.Database = c.Database
	}
	if c.User != "" {
		connConfig.User = c.User
	}
	if c.Password != "" {
		connConfig.Password = c.Password
	}
	if c.Port != 0 {
		connConfig.Port = uint16(c.Port)
	}

	conn, err := pgx.ConnectConfig(ctx, connConfig)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
