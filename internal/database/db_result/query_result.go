package dbresult

import (
	"fmt"
	"io"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// query → returns rows (SELECT, SHOW, etc.)
type QueryResult struct {
	rows          pgx.Rows
	columns       []string
	closed        bool
	RowStreamed   int
	duration      ResultDuration
	queryStartsAt time.Time

	firstRowAt time.Time
	streamDone bool
}

func NewQueryResult(rows pgx.Rows, queryStartsAt time.Time) *QueryResult {
	columns := rows.FieldDescriptions()
	colNames := make([]string, len(columns))
	for i, col := range columns {
		colNames[i] = string(col.Name)
	}

	return &QueryResult{
		rows:          rows,
		columns:       colNames,
		queryStartsAt: queryStartsAt,
	}
}

func (q *QueryResult) Columns() []string {
	return q.columns
}

// Next returns the next row as []any or io.EOF when done.
func (q *QueryResult) Next() ([]any, error) {
	if q.closed {
		return nil, io.EOF
	}

	// Attempt to read next row
	if q.rows.Next() {
		vals, err := q.rows.Values()
		if err != nil {
			q.rows.Close()
			q.closed = true
			return nil, err
		}

		q.RowStreamed++

		// First row timing (critical moment)
		if q.RowStreamed == 1 {
			now := time.Now()
			q.firstRowAt = now

			elapsed := now.Sub(q.queryStartsAt)

			q.duration.TTFR = elapsed
			q.duration.ExecutionTime = elapsed
		}

		// Convert pgtype values → native Go
		for i, v := range vals {
			vals[i] = convertValue(v)
		}

		return vals, nil
	}

	// Handle iteration error
	if err := q.rows.Err(); err != nil {
		q.closed = true
		return nil, err
	}

	// EOF reached — finalize stream duration once
	q.rows.Close()
	q.closed = true

	if !q.streamDone {
		q.streamDone = true

		end := time.Now()

		// If no rows were returned
		if q.RowStreamed == 0 {
			elapsed := end.Sub(q.queryStartsAt)
			q.duration.TTFR = elapsed
			q.duration.ExecutionTime = elapsed
			q.duration.StreamDuration = 0
		} else {
			q.duration.StreamDuration = end.Sub(q.firstRowAt)
		}
	}

	return nil, io.EOF
}

func (q *QueryResult) Duration() ResultDuration {
	return q.duration
}

func (q *QueryResult) CommandTag() pgconn.CommandTag {
	return q.rows.CommandTag()
}

func (q *QueryResult) DurationString() string {
	return fmt.Sprintf("Execution: %s, TTFR: %s, Stream Duration: %s", q.duration.ExecutionTime, q.duration.TTFR.String(), q.duration.StreamDuration.String())
}

func (q *QueryResult) Close() error {
	if q.closed {
		return nil
	}
	q.rows.Close()
	q.closed = true
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
