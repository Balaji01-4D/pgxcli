package result

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type mockRows struct {
	pgx.Rows
	tag pgconn.CommandTag
}

func (m *mockRows) Next() bool { return false }
func (m *mockRows) Close()     {}
func (m *mockRows) Err() error { return nil }
func (m *mockRows) CommandTag() pgconn.CommandTag { return m.tag }

func TestMultiQueryResult_Columns_Nil(t *testing.T) {
	r := NewMultiQuery(nil, time.Now())
	if cols := r.Columns(); cols != nil {
		t.Errorf("expected nil columns, got %v", cols)
	}
}

func TestMultiQueryResult_Rows_KeepsCurrRow(t *testing.T) {
	tag := pgconn.NewCommandTag("SELECT 1")
	r := &MultiQueryResult{
		currRow: &rowStreamer{
			columns: []string{"id", "name"},
			rows:    &mockRows{tag: tag},
		},
	}

	rows, err := r.Rows()
	if err != nil {
		t.Fatalf("Rows() failed: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}

	if r.currRow == nil {
		t.Fatal("expected currRow to be kept after Rows() call, but it was nil")
	}

	cols := r.Columns()
	if len(cols) != 2 || cols[0] != "id" || cols[1] != "name" {
		t.Errorf("expected columns to be available after Rows(), got %v", cols)
	}

	if r.CommandTag() != "SELECT 1" {
		t.Errorf("expected command tag 'SELECT 1', got %q", r.CommandTag())
	}
}
