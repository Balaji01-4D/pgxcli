package parser

import pg_query "github.com/pganalyze/pg_query_go/v6"

func CommandType(sql string) string {
	tree, err := pg_query.Parse(sql)
	if err != nil {
		return "INVALID"
	}

	hasWrite := false

	for _, stmt := range tree.Stmts {
		switch node := stmt.Stmt.Node.(type) {

		case *pg_query.Node_SelectStmt:
			// Detect SELECT INTO (writes data)
			if node.SelectStmt.IntoClause != nil {
				hasWrite = true
			}

		case *pg_query.Node_InsertStmt,
			*pg_query.Node_UpdateStmt,
			*pg_query.Node_DeleteStmt,
			*pg_query.Node_CreateStmt,
			*pg_query.Node_AlterTableStmt,
			*pg_query.Node_DropStmt,
			*pg_query.Node_TruncateStmt,
			*pg_query.Node_CopyStmt,
			*pg_query.Node_RenameStmt:
			hasWrite = true

		case *pg_query.Node_VariableSetStmt:
			hasWrite = true

		case *pg_query.Node_VariableShowStmt:
			continue // safe

		default:
			hasWrite = true
		}
	}

	if hasWrite {
		return "EXECUTE"
	}
	return "QUERY"
}

func IsQuery(sql string) bool {
	return CommandType(sql) == "QUERY"
}

func IsExecute(sql string) bool {
	return CommandType(sql) == "EXECUTE"
}
