package pg

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RowStreamer interface {
	Columns() []string
	Next() ([]interface{}, error) // returns io.EOF when done
	Close() error
	Duration() time.Duration
}

type Tx interface {
	Query(ctx context.Context, sql string, args ...interface{}) (RowStreamer, error)
	Exec(ctx context.Context, sql string, args ...interface{}) (*ExecResult, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type ExecResult struct {
	RowsAffected int64
	Status       string
	Duration     time.Duration
}

type Excutor struct {
	Pool *pgxpool.Pool
}

func NewExecutor(pool *pgxpool.Pool) *Excutor {
	return &Excutor{
		Pool: pool,
	}
}

func (e *Excutor) Query(ctx context.Context, sql string, args ...interface{}) (RowStreamer, error) {
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
	return &rowStreamer{
		rows:    rows,
		columns: columns,
		duration: dur,
	}, nil
}

func (e *Excutor) Exec(ctx context.Context, sql string, args ...interface{}) (*ExecResult, error) {
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
