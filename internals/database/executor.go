package database

import (
	"context"
	"pgcli/internals/parser"
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
	Pool *pgxpool.Pool
}

// constructor function to create new executor
func NewExecutor(pool *pgxpool.Pool) Executor {
	return Executor{
		Pool: pool,
	}
}

// For executing queries like SELECT, SHOW etc.
func (e *Executor) query(ctx context.Context, sql string, args ...interface{}) (*QueryResult, error) {
	start := time.Now()
	rows, err := e.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	dur := time.Since(start)
	fds := rows.FieldDescriptions()
	columns := make([]string, len(fds))
	for i, fd := range fds {
		columns[i] = fd.Name
	}
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
	start := time.Now()
	tag, err := e.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	dur := time.Since(start)
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
