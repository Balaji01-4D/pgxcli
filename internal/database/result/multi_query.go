package result

import (
	"io"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type MultiQueryResult struct {
	mrr *pgconn.MultiResultReader
	currRow *rowStreamer
	start time.Time
	duration time.Duration
}

func NewMultiQuery(mrr *pgconn.MultiResultReader, startedAt time.Time) *MultiQueryResult {
	return &MultiQueryResult{
		mrr: mrr,
		start: startedAt,
	}
}

func (r *MultiQueryResult) Type() Type {
	return ResultTypeMultiQuery
}

// Duration must be called after close.
func (r *MultiQueryResult) Duration() time.Duration {
	return r.duration
}

func (r *MultiQueryResult) Columns() []string {
	if r.currRow == nil {
		return nil
	}
	return r.currRow.columns
}

func (r *MultiQueryResult) Rows() ([][]any, error) {
	if r.currRow == nil {
		return [][]any{}, nil
	}

	collected := make([][]any, 0, 256)
	for {
		row, err := r.currRow.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}
		collected = append(collected, row)
	}

	r.currRow.Close()
	return collected, nil
}

// Next advances to the next result.
// It returns false when there are no more result.
// use the Rows(), Columns() and CommandTag() to get the current result.
func (r *MultiQueryResult) Next() bool {
	if r.currRow != nil {
		r.currRow.Close()
		r.currRow = nil
	}

	if r.mrr.NextResult() {
		rr := r.mrr.ResultReader()
		rows := pgx.RowsFromResultReader(pgtype.NewMap(), rr)
		cols := columnsFromRows(rows)

		r.currRow = &rowStreamer{
			rows:    rows,
			columns: cols,
		}
		return true
	}
	return false
}

func (r *MultiQueryResult) CommandTag() string {
	if r.currRow == nil {
		return ""
	}
	return r.currRow.CommandTag()
}

func (r *MultiQueryResult) Close() error {
	err := r.mrr.Close()

	r.duration = time.Since(r.start)
	return err
}
