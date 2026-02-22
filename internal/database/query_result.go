package database

import (
	"io"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jedib0t/go-pretty/v6/table"
)

// query → returns rows (SELECT, SHOW, etc.)

type QueryResult struct {
	rowStreamer
}

func (r *QueryResult) isResult() {}

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

		// Convert pgtype values to native Go types for better formatting
		for i, v := range vals {
			vals[i] = convertValue(v)
		}

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

func (r *QueryResult) CommandTag() string {
	return r.rows.CommandTag().String()
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

func convertValue(v any) any {
	switch val := v.(type) {
	case pgtype.Numeric:
		d, err := val.Value()
		if err == nil {
			return d
		}
	}
	return v
}
