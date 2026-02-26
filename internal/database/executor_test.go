package database

import (
	"context"
	"io"
	"log/slog"
	"testing"

	dbresult "github.com/balaji01-4d/pgxcli/internal/database/db_result"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestExecutor_query(t *testing.T) {
	ctx := context.Background()

	columns := []pgconn.FieldDescription{
		{Name: "id"}, {Name: "name"}, {Name: "age"},
	}
	records := [][]any{
		{1, "name1", 30},
		{2, "name2", 25},
	}

	rows := &MockRows{
		fields: columns,
		data:   records,
	}

	conn := new(MockConn)
	conn.On("Query", ctx, "select * from users").Return(rows, nil)

	executor := &Executor{
		Conn:   conn,
		Logger: slog.Default(),
	}

	result, err := executor.query(ctx, "select * from users")
	assert.NoError(t, err)

	qr := result.(*dbresult.QueryResult)
	assert.Equal(t, []string{"id", "name", "age"}, qr.Columns())

	row1, err := qr.Next()
	assert.NoError(t, err)
	assert.Equal(t, records[0][0], row1[0])
	assert.Equal(t, records[0][1], row1[1])
	assert.Equal(t, records[0][2], row1[2])

	row2, err := qr.Next()
	assert.NoError(t, err)
	assert.Equal(t, records[1][0], row2[0])
	assert.Equal(t, records[1][1], row2[1])
	assert.Equal(t, records[1][2], row2[2])

	_, err = qr.Next()
	assert.ErrorIs(t, err, io.EOF)
	conn.AssertExpectations(t)
}

func TestExecutor_query_Error(t *testing.T) {
	ctx := context.Background()

	conn := new(MockConn)
	conn.On("Query", ctx, "select * from users").Return(&MockRows{}, assert.AnError)

	executor := &Executor{
		Conn:   conn,
		Logger: slog.Default(),
	}

	result, err := executor.query(ctx, "select * from users")
	assert.Nil(t, result)
	assert.Error(t, err)
	conn.AssertExpectations(t)
}

func TestExecutor_query_Empty(t *testing.T) {
	ctx := context.Background()

	columns := []pgconn.FieldDescription{
		{Name: "id"}, {Name: "name"}, {Name: "age"},
	}
	records := [][]any{}

	rows := &MockRows{
		fields: columns,
		data:   records,
	}

	conn := new(MockConn)
	conn.On("Query", ctx, "select * from users").Return(rows, nil)

	executor := &Executor{
		Conn:   conn,
		Logger: slog.Default(),
	}

	result, err := executor.query(ctx, "select * from users")
	assert.NoError(t, err)

	qr := result.(*dbresult.QueryResult)
	assert.Equal(t, []string{"id", "name", "age"}, qr.Columns())
	_, err = qr.Next()
	assert.ErrorIs(t, err, io.EOF)
	conn.AssertExpectations(t)
}

func TestExecutor_query_RelationNotFound(t *testing.T) {
	ctx := context.Background()

	conn := new(MockConn)
	conn.On("Query", ctx, "select * from users").Return(&MockRows{}, &pgconn.PgError{
		Code: "42P01",
	})

	executor := &Executor{
		Conn:   conn,
		Logger: slog.Default(),
	}

	result, err := executor.query(ctx, "select * from users")
	assert.Nil(t, result)
	assert.Error(t, err)
	conn.AssertExpectations(t)
}

func TestExecutor_Execute(t *testing.T) {
	ctx := context.Background()

	conn := new(MockConn)
	tag := pgconn.NewCommandTag("DELETE 1")

	conn.On("Exec", ctx, "delete from users where id = 1").Return(tag, nil)

	executor := &Executor{
		Conn:   conn,
		Logger: slog.Default(),
	}

	result, err := executor.exec(ctx, "delete from users where id = 1")
	assert.NoError(t, err)
	er := result.(*dbresult.ExecResult)
	assert.Equal(t, int64(1), er.RowsAffected())
	assert.Equal(t, "DELETE 1", er.String())
	conn.AssertExpectations(t)
}

func TestExecutor_Execute_Insert(t *testing.T) {
	ctx := context.Background()

	conn := new(MockConn)
	tag := pgconn.NewCommandTag("INSERT 0 1")

	conn.On("Exec", ctx, "insert into users (name) values ('name1')").Return(tag, nil)

	executor := &Executor{
		Conn:   conn,
		Logger: slog.Default(),
	}

	result, err := executor.exec(ctx, "insert into users (name) values ('name1')")
	assert.NoError(t, err)
	er := result.(*dbresult.ExecResult)
	assert.Equal(t, int64(1), er.RowsAffected())
	assert.Equal(t, "INSERT 0 1", er.String())
	conn.AssertExpectations(t)
}

func TestExecutor_Execute_Error(t *testing.T) {
	ctx := context.Background()

	conn := new(MockConn)
	conn.On("Exec", ctx, "delete from users where id = 1").Return(pgconn.NewCommandTag(""), assert.AnError)

	executor := &Executor{
		Conn:   conn,
		Logger: slog.Default(),
	}

	result, err := executor.exec(ctx, "delete from users where id = 1")
	assert.Nil(t, result)
	assert.Error(t, err)
	conn.AssertExpectations(t)
}

func TestExecutor_Execute_RelationNotFound(t *testing.T) {
	ctx := context.Background()

	conn := new(MockConn)
	relationNotFoundErr := &pgconn.PgError{
		Code: "42P01",
	}
	conn.On("Exec", ctx, "delete from users where id = 1").Return(pgconn.NewCommandTag(""), relationNotFoundErr)

	executor := &Executor{
		Conn:   conn,
		Logger: slog.Default(),
	}

	result, err := executor.exec(ctx, "delete from users where id = 1")
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.ErrorIs(t, err, relationNotFoundErr)
	conn.AssertExpectations(t)
}

func TestExecutor_ping(t *testing.T) {
	ctx := context.Background()

	conn := new(MockConn)
	conn.On("Ping", ctx).Return(nil)

	executor := &Executor{
		Conn:   conn,
		Logger: slog.Default(),
	}

	err := executor.Ping(ctx)
	assert.NoError(t, err)
	conn.AssertExpectations(t)
}

func TestExecutor_ping_Error(t *testing.T) {
	ctx := context.Background()

	// no connection setup for ping to simulate database not connected error
	executor := &Executor{
		Logger: slog.Default(),
	}

	err := executor.Ping(ctx)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "database not connected")
}

func TestExecutor_IsConnected(t *testing.T) {
	conn := new(MockConn)

	executor := &Executor{
		Conn:   conn,
		Logger: slog.Default(),
	}

	assert.True(t, executor.IsConnected())
}
