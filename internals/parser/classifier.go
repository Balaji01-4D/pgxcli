package parser

import pg_query "github.com/pganalyze/pg_query_go/v6"

func CommandType(sql string) string {
	tree, err := pg_query.Parse(sql)
	if err != nil {
		return "INVALID"
	}

	for _, stmt := range tree.Stmts {
		node := stmt.Stmt.Node
		switch node.(type) {
		case *pg_query.Node_SelectStmt:
			return "QUERY"
		case *pg_query.Node_InsertStmt,
			*pg_query.Node_UpdateStmt,
			*pg_query.Node_DeleteStmt:
			return "EXECUTE"
		case *pg_query.Node_CreateStmt,
			*pg_query.Node_AlterTableStmt,
			*pg_query.Node_DropStmt:
			return "EXECUTE"
		default:
			return "EXECUTE"
		}
	}
	return "EXECUTE"
}
