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

	rows := &MockRows{
		fields: []pgconn.FieldDescription{
			{Name: "id"}, {Name: "name"},
		},
		data: [][]any{
			{1, "name1"},
			{2, "name2"},
		},
	}

	conn := new(MockConn)
	conn.On("Query", ctx, "select id, name from users").Return(rows, nil)

	executor := &Executor{
		Conn: conn,
		Logger: slog.Default(),
	}

	result, err := executor.query(ctx, "select id, name from users")
	assert.NoError(t, err)

	assert.Equal(t, []string{"id", "name"}, result.columns)

	var id, name any

	assert.True(t, result.rows.Next())
	assert.NoError(t, result.rows.Scan(&id, &name))

	assert.Equal(t, 1, id)
	assert.Equal(t, "name1", name)

	assert.True(t, result.rows.Next())
	assert.NoError(t, result.rows.Scan(&id, &name))

	assert.Equal(t, 2, id)
	assert.Equal(t, "name2", name)
	

	assert.False(t, result.rows.Next())

	conn.AssertExpectations(t)
}
