package database

import (
	"context"
	"fmt"
	"pgcli/internals/logger"
	"pgcli/internals/parser"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// skip these two interfaces for now
type Tx interface {
	Query(ctx context.Context, sql string, args ...interface{}) (*QueryResult, error)
	Exec(ctx context.Context, sql string, args ...interface{}) (*ExecResult, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type Result interface {
	GetType() string
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
	Pool     *pgxpool.Pool
}

// constructor function to create new executor
func NewExecutor(host, database string, user string, password string,
	port uint16, dsn string, ctx context.Context) (*Executor, error) {

	if dsn == "" {
		var dsnParts []string
		if host != "" {
			dsnParts = append(dsnParts, fmt.Sprintf("host=%s", host))
		}
		if port != 0 {
			dsnParts = append(dsnParts, fmt.Sprintf("port=%d", port))
		}
		if user != "" {
			dsnParts = append(dsnParts, fmt.Sprintf("user=%s", user))
		}
		if database != "" {
			dsnParts = append(dsnParts, fmt.Sprintf("dbname=%s", database))
		}
		if password != "" {
			dsnParts = append(dsnParts, fmt.Sprintf("password=%s", password))
		}
		dsn = strings.Join(dsnParts, " ")
		logger.Log.Debug("Constructed DSN", "dsn", dsn)
	}

	// create a new connection pool
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		logger.Log.Error("Failed to create connection pool", "error", err)
		return nil, err
	}

	// test the connection
	err = pool.Ping(ctx)
	if err != nil {
		logger.Log.Error("Connection ping failed", "error", err)
		return nil, err
	}
	return &Executor{
		Host:     host,
		Port:     port,
		Database: database,
		User:     user,
		Password: password,
		URI:      dsn,
		Pool:     pool,
	}, nil
}

// For executing queries like SELECT, SHOW etc.
func (e *Executor) query(ctx context.Context, sql string, args ...interface{}) (*QueryResult, error) {
	logger.Log.Debug("Executing query", "sql", sql)
	start := time.Now()
	rows, err := e.Pool.Query(ctx, sql, args...)
	if err != nil {
		logger.Log.Error("Query failed", "error", err, "sql", sql)
		return nil, err
	}
	dur := time.Since(start)
	fds := rows.FieldDescriptions()
	columns := make([]string, len(fds))
	for i, fd := range fds {
		columns[i] = fd.Name
	}
	logger.Log.Info("Query completed", "duration_ms", dur.Milliseconds(), "columns", len(columns))
	return &QueryResult{
		rowStreamer: rowStreamer{
			rows:     rows,
			columns:  columns,
			duration: dur,
		},
	}, nil
}

// For executing commands like INSERT, UPDATE, DELETE etc.
func (e *Executor) exec(ctx context.Context, sql string, args ...interface{}) (*ExecResult, error) {
	logger.Log.Debug("Executing command", "sql", sql)
	start := time.Now()
	tag, err := e.Pool.Exec(ctx, sql, args...)
	if err != nil {
		logger.Log.Error("Command failed", "error", err, "sql", sql)
		return nil, err
	}
	dur := time.Since(start)
	logger.Log.Info("Command completed", "duration_ms", dur.Milliseconds(), "rows_affected", tag.RowsAffected(), "status", tag.String())
	return &ExecResult{
		RowsAffected: tag.RowsAffected(),
		Status:       tag.String(),
		Duration:     dur,
	}, nil

}

// Execute method to determine whether to run query or exec based on SQL type
func (e *Executor) Execute(ctx context.Context, sql string, args ...interface{}) (Result, error) {
	if parser.IsQuery(sql) {
		return e.query(ctx, sql, args...)
	}
	return e.exec(ctx, sql, args...)
}

func (e *Executor) Close() {
	if e.Pool != nil {
		e.Pool.Close()
	}
}

func (e *Executor) Ping(ctx context.Context) error {
	if e.Pool == nil {
		return fmt.Errorf("database not connected")
	}
	return e.Pool.Ping(ctx)
}

func (e *Executor) IsConnected() bool {
	return e.Pool != nil
}

func (e *Executor) GetConnectionInfo() {
	cfg := e.Pool.Config()
	logger.Log.Debug("Connection information",
		"host", cfg.ConnConfig.Host,
		"port", cfg.ConnConfig.Port,
		"database", cfg.ConnConfig.Database,
		"user", cfg.ConnConfig.User,
	)
}
