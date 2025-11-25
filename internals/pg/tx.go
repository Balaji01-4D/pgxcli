package pg

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)


type pgTransaction struct {
	conn *pgxpool.Conn // keep the connection so we can release on commit/rollback
	tx   pgx.Tx
}

func (e *Excutor) Begin(ctx context.Context) (Tx, error) {
	conn, err := e.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := conn.Conn().Begin(ctx)
	if err != nil {
		conn.Release()
		return nil, err
	}

	return &pgTransaction{
		conn: conn,
		tx:   tx,
	}, nil
}

func (t *pgTransaction) Query(ctx context.Context, sql string, args ...interface{}) (RowStreamer, error) {
	start := time.Now()
	rows, err := t.tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	dur := time.Since(start)
	fds := rows.FieldDescriptions()
	cols := make([]string, len(fds))
	for i := range fds {
		cols[i] = string(fds[i].Name)
	}
	return &rowStreamer{
		rows:    rows,
		columns: cols,
		closed:  false,
		duration: dur,
	}, nil
}

func (t *pgTransaction) Exec(ctx context.Context, sql string, args ...interface{}) (*ExecResult, error) {
	start := time.Now()
	tag, err := t.tx.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	dur := time.Since(start)
	return &ExecResult{
		RowsAffected: int64(tag.RowsAffected()),
		Status:       tag.String(),
		Duration:     dur,
	}, nil
}

func (t *pgTransaction) Commit(ctx context.Context) error {
	err := t.tx.Commit(ctx)
	// release connection regardless of commit error
	t.conn.Release()
	return err
}

func (t *pgTransaction) Rollback(ctx context.Context) error {
	err := t.tx.Rollback(ctx)
	// release connection regardless of rollback error
	t.conn.Release()
	// If rollback returns ErrTxClosed or no-op, ignore if commit already happened.
	if err == nil || err == context.Canceled || err == context.DeadlineExceeded {
		return nil
	}
	return err
}