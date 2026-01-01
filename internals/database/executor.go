package database

import (
	"context"
	"fmt"
	"pgcli/internals/logger"
	"pgcli/internals/parser"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

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
	Conn     *pgx.Conn
}

func NewExec(ctx context.Context, c Connector) (*Executor, error) {
	conn, err := c.Connect(ctx)
	if err != nil {
		return nil, err
	}

	err = conn.Ping(ctx)
	if err != nil {
		logger.Log.Error("Connection ping failed", "error", err)
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
	}, nil
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

	// create a new connection conn
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		logger.Log.Error("Failed to create connection", "error", err)
		return nil, err
	}

	// test the connection
	err = conn.Ping(ctx)
	if err != nil {
		logger.Log.Error("Connection ping failed", "error", err)
		return nil, err
	}
	// reassigning values
	return &Executor{
		Host:     conn.Config().Host,
		Port:     conn.Config().Port,
		Database: conn.Config().Database,
		User:     conn.Config().User,
		Password: conn.Config().Password,
		URI:      conn.Config().ConnString(),
		Conn:     conn,
	}, nil
}

// For executing queries like SELECT, SHOW etc.
func (e *Executor) query(ctx context.Context, sql string, args ...interface{}) (*QueryResult, error) {
	logger.Log.Debug("Executing query", "sql", sql)
	start := time.Now()
	rows, err := e.Conn.Query(ctx, sql, args...)
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
	tag, err := e.Conn.Exec(ctx, sql, args...)
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

func (e *Executor) Close(ctx context.Context) {
	if e.Conn != nil {
		e.Conn.Close(ctx)
	}
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
