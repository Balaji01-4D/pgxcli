package database

import (
	"context"
	"io"
	"log/slog"
	"testing"

	dbresult "github.com/balaji01-4d/pgxcli/internal/database/result"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecutorQuery(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name        string
		query       string
		rows        *MockRows
		queryErr    error
		wantColumns []string
		wantRows    [][]any
		wantErr     bool
	}{
		{
			name:  "returns rows",
			query: "select * from users",
			rows: &MockRows{
				fields: []pgconn.FieldDescription{{Name: "id"}, {Name: "name"}, {Name: "age"}},
				data:   [][]any{{1, "name1", 30}, {2, "name2", 25}},
			},
			wantColumns: []string{"id", "name", "age"},
			wantRows:    [][]any{{1, "name1", 30}, {2, "name2", 25}},
		},
		{
			name:  "returns empty rows",
			query: "select * from users",
			rows: &MockRows{
				fields: []pgconn.FieldDescription{{Name: "id"}, {Name: "name"}, {Name: "age"}},
				data:   [][]any{},
			},
			wantColumns: []string{"id", "name", "age"},
			wantRows:    [][]any{},
		},
		{
			name:     "returns error",
			query:    "select * from users",
			rows:     &MockRows{},
			queryErr: assert.AnError,
			wantErr:  true,
		},
		{
			name:  "returns relation not found",
			query: "select * from users",
			rows:  &MockRows{},
			queryErr: &pgconn.PgError{
				Code: "42P01",
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			conn := new(MockConn)
			conn.On("Query", ctx, tc.query).Return(tc.rows, tc.queryErr)

			exec := &executor{Conn: conn, Logger: slog.Default()}
			result, err := exec.query(ctx, tc.query)

			if tc.wantErr {
				assert.Nil(t, result)
				assert.Error(t, err)
				conn.AssertExpectations(t)
				return
			}

			require.NoError(t, err)
			queryResult, ok := result.(*dbresult.QueryResult)
			require.True(t, ok)
			assert.Equal(t, tc.wantColumns, queryResult.Columns())

			for _, wantRow := range tc.wantRows {
				row, nextErr := queryResult.Next()
				require.NoError(t, nextErr)
				assert.Equal(t, wantRow, row)
			}

			row, nextErr := queryResult.Next()
			assert.Nil(t, row)
			assert.Equal(t, io.EOF, nextErr)
			conn.AssertExpectations(t)
		})
	}
}

func TestExecutorExecute(t *testing.T) {
	ctx := context.Background()

	relationNotFoundErr := &pgconn.PgError{Code: "42P01"}
	testCases := []struct {
		name       string
		query      string
		rows       *MockRows
		queryErr   error
		wantStatus string
		wantErr    bool
		wantErrIs  error
	}{
		{
			name:  "delete success",
			query: "delete from users where id = 1",
			rows: &MockRows{
				tag: pgconn.NewCommandTag("DELETE 1"),
			},
			wantStatus: "DELETE 1",
		},
		{
			name:  "insert success",
			query: "insert into users (name) values ('name1')",
			rows: &MockRows{
				tag: pgconn.NewCommandTag("INSERT 0 1"),
			},
			wantStatus: "INSERT 0 1",
		},
		{
			name:     "returns error",
			query:    "delete from users where id = 1",
			rows:     &MockRows{},
			queryErr: assert.AnError,
			wantErr:  true,
		},
		{
			name:      "returns relation not found",
			query:     "delete from users where id = 1",
			rows:      &MockRows{},
			queryErr:  relationNotFoundErr,
			wantErr:   true,
			wantErrIs: relationNotFoundErr,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			conn := new(MockConn)
			conn.On("Query", ctx, tc.query).Return(tc.rows, tc.queryErr)

			exec := &executor{Conn: conn, Logger: slog.Default()}
			result, err := exec.execute(ctx, tc.query)

			if tc.wantErr {
				assert.Nil(t, result)
				assert.Error(t, err)
				if tc.wantErrIs != nil {
					assert.ErrorIs(t, err, tc.wantErrIs)
				}
				conn.AssertExpectations(t)
				return
			}

			require.NoError(t, err)
			queryResult, ok := result.(*dbresult.QueryResult)
			require.True(t, ok)
			assert.Equal(t, tc.wantStatus, queryResult.CommandTag())
			conn.AssertExpectations(t)
		})
	}
}

func TestExecutorPing(t *testing.T) {
	ctx := context.Background()
	testCases := []struct {
		name       string
		withConn   bool
		pingErr    error
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:     "connected and ping succeeds",
			withConn: true,
		},
		{
			name:       "no connection",
			wantErr:    true,
			wantErrMsg: "database not connected",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			exec := &executor{Logger: slog.Default()}
			if tc.withConn {
				conn := new(MockConn)
				conn.On("Ping", ctx).Return(tc.pingErr)
				exec.Conn = conn
				defer conn.AssertExpectations(t)
			}

			err := exec.ping(ctx)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.wantErrMsg != "" {
					assert.ErrorContains(t, err, tc.wantErrMsg)
				}
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestExecutorIsConnected(t *testing.T) {
	testCases := []struct {
		name     string
		withConn bool
		want     bool
	}{
		{name: "connected", withConn: true, want: true},
		{name: "not connected", withConn: false, want: false},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			exec := &executor{Logger: slog.Default()}
			if tc.withConn {
				exec.Conn = new(MockConn)
			}
			assert.Equal(t, tc.want, exec.isConnected())
		})
	}
}
