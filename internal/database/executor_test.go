package database

import (
	"context"
	"log/slog"
	"testing"

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
		Conn: conn,
		Logger: slog.Default(),
	}

	result, err := executor.query(ctx, "select * from users")
	assert.NoError(t, err)

	assert.Equal(t, []string{"id", "name", "age"}, result.columns)

	var id, name, age any

	assert.True(t, result.rows.Next())
	assert.NoError(t, result.rows.Scan(&id, &name, &age))

	assert.Equal(t, records[0][0], id)
	assert.Equal(t, records[0][1], name)
	assert.Equal(t, records[0][2], age)

	assert.True(t, result.rows.Next())
	assert.NoError(t, result.rows.Scan(&id, &name, &age))

	assert.Equal(t, records[1][0], id)
	assert.Equal(t, records[1][1], name)
	assert.Equal(t, records[1][2], age)

	assert.False(t, result.rows.Next())
	conn.AssertExpectations(t)
}

func TestExecutor_query_Error(t *testing.T) {
	ctx := context.Background()

	conn := new(MockConn)
	conn.On("Query", ctx, "select * from users").Return(&MockRows{}, assert.AnError)

	executor := &Executor{
		Conn: conn,
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
		Conn: conn,
		Logger: slog.Default(),
	}

	result, err := executor.query(ctx, "select * from users")
	assert.NoError(t, err)

	assert.Equal(t, []string{"id", "name", "age"}, result.columns)
	assert.False(t, result.rows.Next())
	conn.AssertExpectations(t)
}

func TestExecutor_query_RelationNotFound(t *testing.T) {
	ctx := context.Background()

	conn := new(MockConn)
	conn.On("Query", ctx, "select * from users").Return(&MockRows{}, &pgconn.PgError{
		Code: "42P01",
	})

	executor := &Executor{
		Conn: conn,
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
		Conn: conn,
		Logger: slog.Default(),
	}

	result, err := executor.exec(ctx, "delete from users where id = 1")
	assert.Equal(t, result.RowsAffected, int64(1))
	assert.Equal(t, result.Status, "DELETE 1")
	assert.NoError(t, err)
	conn.AssertExpectations(t)
}

func TestExecutor_Execute_Insert(t *testing.T) {
	ctx := context.Background()

	conn := new(MockConn)
	tag := pgconn.NewCommandTag("INSERT 0 1")

	conn.On("Exec", ctx, "insert into users (name) values ('name1')").Return(tag, nil)

	executor := &Executor{
		Conn: conn,
		Logger: slog.Default(),
	}

	result, err := executor.exec(ctx, "insert into users (name) values ('name1')")
	assert.Equal(t, result.RowsAffected, int64(1))
	assert.Equal(t, result.Status, "INSERT 0 1")
	assert.NoError(t, err)
	conn.AssertExpectations(t)
}

func TestExecutor_Execute_Error(t *testing.T) {
	ctx := context.Background()

	conn := new(MockConn)
	conn.On("Exec", ctx, "delete from users where id = 1").Return(pgconn.NewCommandTag(""), assert.AnError)

	executor := &Executor{
		Conn: conn,
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
		Conn: conn,
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
		Conn: conn,
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
