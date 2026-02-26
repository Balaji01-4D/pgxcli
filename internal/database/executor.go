package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	dbresult "github.com/balaji01-4d/pgxcli/internal/database/db_result"
	"github.com/balaji01-4d/pgxcli/internal/parser"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Conn interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Config() *pgx.ConnConfig
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
}

// executor struct to execute queries
type Executor struct {
	Host     string
	Port     uint16
	Database string
	Schema   string
	User     string
	Password string
	URI      string
	Conn     Conn

	Logger *slog.Logger
}

func NewExecutor(ctx context.Context, c Connector, logger *slog.Logger) (*Executor, error) {
	conn, err := c.Connect(ctx)
	if err != nil {
		return nil, err
	}

	err = conn.Ping(ctx)
	if err != nil {
		logger.Error("Connection ping failed", "error", err)
		return nil, err
	}

	return &Executor{
		Host:     conn.Config().Host,
		Port:     conn.Config().Port,
		Database: conn.Config().Database,
		User:     conn.Config().User,
		Password: conn.Config().Password,
		URI:      conn.Config().ConnString(),
		Conn:     conn,
		Logger:   logger,
	}, nil
}

// For executing queries like SELECT, SHOW etc.
func (e *Executor) query(ctx context.Context, sql string, args ...any) (dbresult.Result, error) {
	e.Logger.Debug("Executing query", "sql", sql)
	start := time.Now()
	rows, err := e.Conn.Query(ctx, sql, args...)
	if err != nil {
		e.Logger.Error("Query failed", "error", err, "sql", sql)
		return nil, err
	}
	e.Logger.Info("Query completed", "sql", sql, "duration_ms", time.Since(start).Milliseconds())
	return dbresult.NewQueryResult(
		rows,
		start,
	), nil
}

// For executing commands like INSERT, UPDATE, DELETE etc.
func (e *Executor) exec(ctx context.Context, sql string, args ...any) (dbresult.Result, error) {
	e.Logger.Debug("Executing command", "sql", sql)
	start := time.Now()
	tag, err := e.Conn.Exec(ctx, sql, args...)
	if err != nil {
		e.Logger.Error("Command failed", "error", err, "sql", sql)
		return nil, err
	}
	dur := time.Since(start)
	e.Logger.Info("Command completed", "duration_ms", dur.Milliseconds(), "rows_affected", tag.RowsAffected(), "status", tag.String())
	return dbresult.NewExecResult(tag, dur), nil
}

// Execute method to determine whether to run query or exec based on SQL type
func (e *Executor) Execute(ctx context.Context, sql string, args ...any) (dbresult.Result, error) {
	if parser.IsQuery(sql) {
		return e.query(ctx, sql, args...)
	}
	return e.exec(ctx, sql, args...)
}

func (e *Executor) Close(ctx context.Context) error {
	if e.Conn != nil {
		return e.Conn.Close(ctx)
	}
	return nil
}

func (e *Executor) Ping(ctx context.Context) error {
	if e.Conn == nil {
		return fmt.Errorf("database not connected")
	}
	return e.Conn.Ping(ctx)
}

func (e *Executor) IsConnected() bool {
	return e.Conn != nil
}
