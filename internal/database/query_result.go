package database

import (
	"fmt"
	"io"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jedib0t/go-pretty/v6/table"
)

// batchSize controls how many rows are buffered per rendered table chunk,
// preventing OOM when querying very large tables.
const batchSize = 500

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

// RenderTo streams query results to w in batches of batchSize rows to avoid
// loading the entire result set into memory at once.
func (r *QueryResult) RenderTo(w io.Writer) error {
	header := make(table.Row, len(r.columns))
	for i, col := range r.columns {
		header[i] = col
	}

	wroteAny := false
	for {
		var batch []table.Row
		for i := 0; i < batchSize; i++ {
			values, err := r.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			row := make(table.Row, len(values))
			copy(row, values)
			batch = append(batch, row)
		}
		if len(batch) == 0 {
			break
		}
		wroteAny = true
		tw := table.NewWriter()
		tw.AppendHeader(header)
		for _, row := range batch {
			tw.AppendRow(row)
		}
		if _, err := fmt.Fprintln(w, tw.Render()); err != nil {
			return err
		}
	}

	// Always render at least a header-only table so callers can see column names
	// even when the query returns zero rows.
	if !wroteAny {
		tw := table.NewWriter()
		tw.AppendHeader(header)
		_, err := fmt.Fprintln(w, tw.Render())
		return err
	}

	return nil
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
