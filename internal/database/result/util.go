package result

import "github.com/jackc/pgx/v5"

func columnsFromRows(rows pgx.Rows) []string {
	fds := rows.FieldDescriptions()
	columns := make([]string, len(fds))
	for i, fd := range fds {
		columns[i] = fd.Name
	}
	return columns
}
