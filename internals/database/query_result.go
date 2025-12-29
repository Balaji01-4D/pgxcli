package database

import (
	"io"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jedib0t/go-pretty/v6/table"
)

// query → returns rows (SELECT, SHOW, etc.)

type QueryResult struct {
	rowStreamer
}

type rowStreamer struct {
	rows     pgx.Rows
	columns  []string
	closed   bool
	duration time.Duration
}

func (r *rowStreamer) Columns() []string {
	return r.columns
}

// Next returns the next row as []any or io.EOF when done.
func (r *rowStreamer) Next() ([]any, error) {
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
		// convert []any as-is; nil for NULLs
		return vals, nil
	}
	if err := r.rows.Err(); err != nil {
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

func (r *rowStreamer) GetType() string {
	return "QUERY"
}


func (r *QueryResult) Render() (table.Writer, error) {
	tw := table.NewWriter()

	row := make(table.Row, len(r.columns))
	for i, col := range r.columns {
		row[i] = col
	}
	tw.AppendHeader(row)

	for {
		values, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		row := make(table.Row, len(values))
		copy(row, values)
		tw.AppendRow(row)
	}
	return tw, nil
}
