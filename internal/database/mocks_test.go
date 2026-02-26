package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"
)

type MockConn struct {
	mock.Mock
}

func (mc *MockConn) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) { 
	argsMocks := mc.Called(ctx, sql)
	return argsMocks.Get(0).(pgx.Rows), argsMocks.Error(1)
}

func (mc *MockConn) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	argsMocks := mc.Called(ctx, sql)
	return argsMocks.Get(0).(pgconn.CommandTag), argsMocks.Error(1)
}

func (mc *MockConn) Ping(ctx context.Context) error { 
	argMocks := mc.Called(ctx)
	return argMocks.Error(0)
}

func (mc *MockConn) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row { return nil }
func (mc *MockConn) Close(ctx context.Context) error { return nil }
func (mc *MockConn) Config() *pgx.ConnConfig { return nil }


type MockRows struct {
	mock.Mock
	data [][]any
	fields []pgconn.FieldDescription
	index int
}

func (m *MockRows) FieldDescriptions() []pgconn.FieldDescription {
	return m.fields
}

func (m *MockRows) Next() bool {
	if m.index < len(m.data) {
		m.index++
		return true
	}
	return false
}

func (m *MockRows) Scan(dest ...any) error {
	row := m.data[m.index-1]
	for i := range dest {
		ptr := dest[i].(*any)
		*ptr = row[i]
	}
	return nil
}

func (m *MockRows) Conn() *pgx.Conn { return &pgx.Conn{} }
func (m *MockRows) Close() {}
func (m *MockRows) Err() error                    { return nil }
func (m *MockRows) CommandTag() pgconn.CommandTag { return pgconn.CommandTag{} }
func (m *MockRows) Values() ([]any, error) {
	if m.index == 0 || m.index > len(m.data) {
		return nil, nil
	}
	row := m.data[m.index-1]
	vals := make([]any, len(row))
	copy(vals, row)
	return vals, nil
}
func (m *MockRows) RawValues() [][]byte           { return nil }
