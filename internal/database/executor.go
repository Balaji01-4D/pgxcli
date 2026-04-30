package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/balaji01-4d/pgxcli/internal/database/result"
	"github.com/balaji01-4d/pgxspecial"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type conn interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Config() *pgx.ConnConfig
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
}

type executor struct {
	Host     string
	Port     uint16
	Database string
	Schema   string
	User     string
	Password string
	URI      string
	Conn     conn

	Logger *slog.Logger
}

func newExecutor(ctx context.Context, c Connector, logger *slog.Logger) (*executor, error) {
	conn, err := c.Connect(ctx)
	if err != nil {
		return nil, err
	}

	err = conn.Ping(ctx)
	if err != nil {
		logger.Error("Connection ping failed", "error", err)
		if err := conn.Close(ctx); err != nil {
			logger.Error("Failed to close connection", "error", err)
		}

		return nil, err
	}

	return &executor{
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
func (e *executor) query(ctx context.Context, sql string, args ...any) (result.Result, error) {
	e.Logger.Debug("Executing query", "sql", sql)
	start := time.Now()
	rows, err := e.Conn.Query(ctx, sql, args...)
	if err != nil {
		e.Logger.Error("Query failed", "error", err, "sql", sql)
		return nil, err
	}
	dur := time.Since(start)

	e.Logger.Info("Query completed", "duration_ms", dur.Milliseconds())
	return result.NewQuery(rows, dur), nil
}

// execute determines whether to run query or exec based on SQL type.
func (e *executor) execute(ctx context.Context, sql string, args ...any) (result.Result, error) {
	return e.query(ctx, sql, args...)
}

func (e *executor) executeSpecial(ctx context.Context, cmd string) (pgxspecial.SpecialCommandResult, bool, error) {
	specialResult, ok, err := pgxspecial.ExecuteSpecialCommand(ctx, e.Conn, cmd)
	if err != nil {
		e.Logger.Error("Special command execution failed", "error", err, "command", cmd)
		return nil, ok, err
	}

	if !ok || specialResult == nil || specialResult.ResultKind() != pgxspecial.ResultKindRows {
		return specialResult, ok, nil
	}

	rowResult, isRowResult := specialResult.(pgxspecial.RowResult)
	if !isRowResult {
		return nil, ok, fmt.Errorf("invalid row special result type: %T", specialResult)
	}

	normalizedRows, err := result.NewSpecialRow(rowResult.Rows)
	if err != nil {
		e.Logger.Error("Failed to materialize special command rows", "error", err, "command", cmd)
		return nil, ok, err
	}

	return normalizedRows, ok, nil
}

func (e *executor) close(ctx context.Context) error {
	if e.Conn != nil {
		return e.Conn.Close(ctx)
	}
	return nil
}

func (e *executor) ping(ctx context.Context) error {
	if e.Conn == nil {
		return fmt.Errorf("database not connected")
	}
	return e.Conn.Ping(ctx)
}

func (e *executor) isConnected() bool {
	return e.Conn != nil
}
