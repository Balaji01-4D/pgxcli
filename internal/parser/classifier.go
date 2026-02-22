package parser

import pg_query "github.com/pganalyze/pg_query_go/v6"

func CommandType(sql string) string {
	tree, err := pg_query.Parse(sql)
	if err != nil {
		return "INVALID"
	}

	if len(tree.Stmts) == 0 {
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

		case *pg_query.Node_InsertStmt:
			if len(node.InsertStmt.ReturningList) == 0 {
				hasWrite = true
			}

		case *pg_query.Node_UpdateStmt:
			if len(node.UpdateStmt.ReturningList) == 0 {
				hasWrite = true
			}

		case *pg_query.Node_DeleteStmt:
			if len(node.DeleteStmt.ReturningList) == 0 {
				hasWrite = true
			}

		case *pg_query.Node_VariableShowStmt,
			*pg_query.Node_ExplainStmt,
			*pg_query.Node_ExecuteStmt:
			continue // safe

		case *pg_query.Node_CreateStmt,
			*pg_query.Node_AlterTableStmt,
			*pg_query.Node_DropStmt,
			*pg_query.Node_TruncateStmt,
			*pg_query.Node_RenameStmt:
			hasWrite = true

		case *pg_query.Node_CopyStmt:
			if !node.CopyStmt.IsFrom {
				hasWrite = false
			} else {
				hasWrite = true
			}

		case *pg_query.Node_VariableSetStmt:
			hasWrite = true

		default:
			hasWrite = true
		}
	}

	if hasWrite {
		return "EXECUTE"
	}
	return "QUERY"
}

// IsQuery returns true if the SQL statement is a read-only query.
func IsQuery(sql string) bool {
	return CommandType(sql) == "QUERY"
}

// IsExecute returns true if the SQL statement modifies data.
func IsExecute(sql string) bool {
	return CommandType(sql) == "EXECUTE"
}

// IsValid returns true if the SQL statement can be parsed successfully.
func IsValid(sql string) bool {
	return CommandType(sql) != "INVALID"
}
