package database

import (
	"bytes"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRowsValues is a pgx.Rows implementation whose Values() returns real data,
// used to exercise the RenderTo streaming path.
type mockRowsValues struct {
	fields []pgconn.FieldDescription
	data   [][]any
	index  int
	err    error
}

func (m *mockRowsValues) FieldDescriptions() []pgconn.FieldDescription { return m.fields }
func (m *mockRowsValues) Next() bool {
	if m.err != nil {
		return false
	}
	m.index++
	return m.index <= len(m.data)
}
func (m *mockRowsValues) Values() ([]any, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.data[m.index-1], nil
}
func (m *mockRowsValues) Scan(dest ...any) error          { return nil }
func (m *mockRowsValues) Conn() *pgx.Conn                 { return nil }
func (m *mockRowsValues) Close()                          {}
func (m *mockRowsValues) Err() error                      { return m.err }
func (m *mockRowsValues) CommandTag() pgconn.CommandTag   { return pgconn.CommandTag{} }
func (m *mockRowsValues) RawValues() [][]byte             { return nil }

func newQueryResult(rows *mockRowsValues, columns []string) *QueryResult {
	return &QueryResult{
		rowStreamer: rowStreamer{
			rows:    rows,
			columns: columns,
		},
	}
}

func TestQueryResult_RenderTo(t *testing.T) {
	rows := &mockRowsValues{
		fields: []pgconn.FieldDescription{{Name: "id"}, {Name: "name"}},
		data: [][]any{
			{1, "alice"},
			{2, "bob"},
		},
	}
	qr := newQueryResult(rows, []string{"id", "name"})

	var buf bytes.Buffer
	err := qr.RenderTo(&buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "alice")
	assert.Contains(t, output, "bob")
}

func TestQueryResult_RenderTo_Empty(t *testing.T) {
	rows := &mockRowsValues{
		fields: []pgconn.FieldDescription{{Name: "id"}, {Name: "name"}},
		data:   [][]any{},
	}
	qr := newQueryResult(rows, []string{"id", "name"})

	var buf bytes.Buffer
	err := qr.RenderTo(&buf)
	require.NoError(t, err)

	// Headers should still appear even with zero rows.
	output := buf.String()
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "NAME")
}

func TestQueryResult_RenderTo_ErrorPropagated(t *testing.T) {
	scanErr := errors.New("scan error")
	rows := &mockRowsValues{
		fields: []pgconn.FieldDescription{{Name: "id"}},
		data:   [][]any{{1}},
		err:    scanErr,
	}
	qr := newQueryResult(rows, []string{"id"})

	var buf bytes.Buffer
	err := qr.RenderTo(&buf)
	assert.Error(t, err)
}
