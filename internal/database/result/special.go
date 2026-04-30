package result

import (
	"fmt"

	"github.com/balaji01-4d/pgxspecial"
	"github.com/jackc/pgx/v5"
)

// SpecialRow stores row-oriented special command output as plain values.
type SpecialRow struct {
	columns []string
	data    [][]any
}

func NewSpecialRow(rows pgx.Rows) (SpecialRow, error) {
	if rows == nil {
		return SpecialRow{}, fmt.Errorf("special command rows cannot be nil")
	}
	defer rows.Close()

	columns := make([]string, len(rows.FieldDescriptions()))
	for i, col := range rows.FieldDescriptions() {
		columns[i] = col.Name
	}

	data := make([][]any, 0)
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return SpecialRow{}, err
		}

		row := make([]any, len(values))
		copy(row, values)
		data = append(data, row)
	}

	if err := rows.Err(); err != nil {
		return SpecialRow{}, err
	}

	return SpecialRow{columns: columns, data: data}, nil
}

// ResultKind marks this result as a row-based special command output.
func (r SpecialRow) ResultKind() pgxspecial.SpecialResultKind {
	return pgxspecial.ResultKindRows
}

// Columns returns row header names.
func (r SpecialRow) Columns() []string {
	return r.columns
}

// Data returns row values.
func (r SpecialRow) Data() [][]any {
	return r.data
}
