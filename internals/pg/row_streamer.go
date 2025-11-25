package pg

import (
	"io"
	"time"

	"github.com/jackc/pgx/v5"
)


type rowStreamer struct {
	rows    pgx.Rows
	columns []string
	closed  bool
	duration time.Duration
}

func (r *rowStreamer) Columns() []string {
	return r.columns
}

// Next returns the next row as []interface{} or io.EOF when done.
func (r *rowStreamer) Next() ([]interface{}, error) {
	if r.closed {
		return nil, io.EOF
	}
	if r.rows.Next() {
		vals, err := r.rows.Values()
		if err != nil {
			r.rows.Close()
			r.closed = true
			return nil, err
		}
		// convert []interface{} as-is; nil for NULLs
		return vals, nil
	}
	if err := r.rows.Err(); err != nil {
		r.rows.Close()
		r.closed = true
		return nil, err
	}
	// no more rows
	r.rows.Close()
	r.closed = true
	return nil, io.EOF
}

func (r *rowStreamer) Close() error {
	if r.closed {
		return nil
	}
	r.rows.Close()
	r.closed = true
	return nil
}

func (r *rowStreamer) Duration() time.Duration {
	return r.duration
}