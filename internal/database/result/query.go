package result

import (
	"io"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// QueryResult represents a row-producing SQL execution result.
type QueryResult struct {
	rowStreamer
}

func NewQuery(rows pgx.Rows, duration time.Duration) *QueryResult {
	return &QueryResult{
		rowStreamer: rowStreamer{
			rows:     rows,
			duration: duration,
		},
	}
}

func (r *QueryResult) Type() Type {
	return ResultTypeQuery
}

func (r *QueryResult) Columns() []string {
	if r.columns == nil {
		fds := r.rows.FieldDescriptions()
		r.columns = make([]string, len(fds))
		for i, fd := range fds {
			r.columns[i] = fd.Name
		}
	}
	return r.columns
}

func (r *QueryResult) Rows() ([][]any, error) {
	collected := make([][]any, 0, 256)
	for {
		row, err := r.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}
		collected = append(collected, row)
	}
	r.Close()
	return collected, nil
}

func (r *QueryResult) Caption() string { return "" } // TODO: add execution time or row count

type rowStreamer struct {
	rows     pgx.Rows
	columns  []string
	closed   bool
	duration time.Duration
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

// CommandTag returns the PostgreSQL command tag for the streamed rows.
func (r *QueryResult) CommandTag() string {
	return r.rows.CommandTag().String()
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
