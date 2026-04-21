package parser

import pg_query "github.com/pganalyze/pg_query_go/v6"

func SplitSqlStatement(sql string) ([]string, error) {
	return pg_query.SplitWithParser(sql, true)
}