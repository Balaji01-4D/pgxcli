package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandType(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected string
	}{
		// Queries - Basic
		{"Select", "SELECT * FROM users", "QUERY"},
		{"Select with CTE", "WITH input AS (SELECT 1) SELECT * FROM input", "QUERY"},
		{"Explain", "EXPLAIN SELECT * FROM users", "QUERY"},
		{"Show", "SHOW client_encoding", "QUERY"},
		{"Values", "VALUES (1), (2)", "QUERY"},

		// Queries - Advanced
		{"Execute Prepared", "EXECUTE myplan('abc')", "QUERY"},
		{"CTE Only Select", "WITH t AS (SELECT 1) SELECT * FROM t", "QUERY"},
		{"Nested CTE", "WITH a AS (WITH b AS (SELECT 1) SELECT * FROM b) SELECT * FROM a", "QUERY"},
		{"Select Function", "SELECT now()", "QUERY"},
		{"Select Generate Series", "SELECT * FROM generate_series(1,10)", "QUERY"},
		{"Select Literal", "SELECT 1", "QUERY"},
		{"Table Keyword", "TABLE users", "QUERY"},
		// {"Values Alias", "VALUES (1,2) AS t(a,b)", "QUERY"},
		{"Explain Analyze", "EXPLAIN ANALYZE SELECT * FROM users", "QUERY"},
		{"Explain Verbose", "EXPLAIN (VERBOSE) SELECT * FROM users", "QUERY"},
		{"Show All", "SHOW ALL", "QUERY"},
		{"Show Quoted", "SHOW \"server_version\"", "QUERY"},

		// Insert/Update/Delete with Returning
		{"Insert Returning", "INSERT INTO users (name) VALUES ('bob') RETURNING id", "QUERY"},
		{"Update Returning", "UPDATE users SET name = 'alice' RETURNING id, name", "QUERY"},
		{"Delete Returning", "DELETE FROM users RETURNING id", "QUERY"},
		{"Insert With CTE", "WITH x AS (SELECT 1) INSERT INTO users(name) VALUES('a') RETURNING id", "QUERY"},
		{"Insert Default Returning", "INSERT INTO users DEFAULT VALUES RETURNING id", "QUERY"},
		{"Update From", "UPDATE users u SET name='a' FROM accounts a WHERE u.id=a.uid RETURNING u.id", "QUERY"},
		{"Delete Using", "DELETE FROM users u USING accounts a WHERE u.id=a.uid RETURNING u.id", "QUERY"},
		{"Insert On Conflict Returning", "INSERT INTO users(id,name) VALUES (1,'a') ON CONFLICT (id) DO UPDATE SET name='b' RETURNING id", "QUERY"},

		// Executions - DML & DDL
		{"Insert", "INSERT INTO users (name) VALUES ('bob')", "EXECUTE"},
		{"Update", "UPDATE users SET name = 'alice' WHERE id = 1", "EXECUTE"},
		{"Delete", "DELETE FROM users WHERE id = 1", "EXECUTE"},
		{"Create Table", "CREATE TABLE foo (id int)", "EXECUTE"},
		{"Drop Table", "DROP TABLE foo", "EXECUTE"},
		{"Alter Table", "ALTER TABLE foo ADD COLUMN bar int", "EXECUTE"},
		{"Truncate", "TRUNCATE TABLE foo", "EXECUTE"},
		{"Set", "SET client_encoding = 'UTF8'", "EXECUTE"},

		// Executions - Maintenance & Setup
		{"Insert Simple", "INSERT INTO users(name) VALUES('bob')", "EXECUTE"},
		{"Vacuum", "VACUUM users", "EXECUTE"},
		{"Analyze", "ANALYZE users", "EXECUTE"},
		{"Reindex", "REINDEX TABLE users", "EXECUTE"},
		{"Grant", "GRANT SELECT ON users TO public", "EXECUTE"},
		{"Revoke", "REVOKE SELECT ON users FROM public", "EXECUTE"},

		// Transaction Control
		{"Begin", "BEGIN", "EXECUTE"},
		{"Commit", "COMMIT", "EXECUTE"},
		{"Rollback", "ROLLBACK", "EXECUTE"},
		{"Savepoint", "SAVEPOINT s1", "EXECUTE"},
		{"Release Savepoint", "RELEASE SAVEPOINT s1", "EXECUTE"},

		// COPY
		{"Copy To Stdout", "COPY (SELECT 1) TO STDOUT", "QUERY"},
		{"Copy From Stdin", "COPY users FROM STDIN", "EXECUTE"},

		// Robustness - Whitespace & Case
		{"Lowercase Select", "select * from users", "QUERY"},
		{"Mixed Case", "SeLeCt * FrOm users", "QUERY"},
		{"Leading Spaces", "   SELECT * FROM users", "QUERY"},
		{"Leading Newline", "\nSELECT * FROM users", "QUERY"},
		{"Trailing Semicolon", "SELECT * FROM users;", "QUERY"},
		{"Multiple Semicolons", "SELECT * FROM users;;", "QUERY"},

		// Robustness - Comments
		{"Line Comment Before", "-- comment\nSELECT * FROM users", "QUERY"},
		{"Block Comment Before", "/* comment */ SELECT * FROM users", "QUERY"},
		{"Inline Comment", "SELECT * FROM users -- hello", "QUERY"},
		{"Comment Only", "-- just comment", "INVALID"},
		{"Block Comment Only", "/* only comment */", "INVALID"},

		// Regressions
		{"Execute Prepared Regression", "EXECUTE invq('x')", "QUERY"},
		{"Returning Expression Regression", "INSERT INTO users(name) VALUES('a') RETURNING 1+1", "QUERY"},
		{"Explain Analyze Regression", "EXPLAIN ANALYZE SELECT 1", "QUERY"},
		{"Values Regression", "VALUES (1)", "QUERY"},

		// Invalid / Unknown
		{"Invalid", "NOT A VALID SQL", "INVALID"},
		{"Empty String", "", "INVALID"},
		{"Whitespace Only", "   \n\t  ", "INVALID"},
		{"Garbage", "asdfghjkl", "INVALID"},
		{"Partial Keyword", "SELEC * FROM users", "INVALID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, CommandType(tt.sql))
		})
	}
}

func TestIsQuery(t *testing.T) {
	assert.True(t, IsQuery("SELECT 1"))
	assert.False(t, IsQuery("INSERT INTO foo VALUES(1)"))
}

func TestIsExecute(t *testing.T) {
	assert.True(t, IsExecute("INSERT INTO foo VALUES(1)"))
	assert.False(t, IsExecute("SELECT 1"))
}

func TestIsValid(t *testing.T) {
	assert.True(t, IsValid("SELECT 1"))
	assert.False(t, IsValid("foo bar"))
}
