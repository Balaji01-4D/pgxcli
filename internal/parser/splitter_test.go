package parser_test

import (
	"testing"

	"github.com/balaji01-4d/pgxcli/internal/parser"
	"github.com/stretchr/testify/assert"
)

func TestSplitSQLStatements(t *testing.T) {
	tests := []struct {
		name string
		sql  string
		want []string
	}{
		{"Empty", "", []string{""}},
		{"WhitespaceOnly", "   \n\t  ", []string{"   \n\t  "}},
		{"SingleStatement", "SELECT 1;", []string{"SELECT 1;"}},
		{"NoTrailingSemicolon", "SELECT 1", []string{"SELECT 1"}},
		{"MultipleStatements", "SELECT 1; SELECT 2; SELECT 3;", []string{"SELECT 1;", "SELECT 2;", "SELECT 3;"}},
		{"LeadingAndTrailingSemicolons", ";SELECT 1;;", []string{";", "SELECT 1;", ";"}},
		{"EmptyStatementsBetweenValids", "SELECT 1;;SELECT 2;", []string{"SELECT 1;", ";", "SELECT 2;"}},
		{"MultilineStatements", "SELECT\n1;\nSELECT\n2;", []string{"SELECT\n1;", "SELECT\n2;"}},
		{"SemicolonInStringLiteral", "SELECT 'value;with;semicolons';", []string{"SELECT 'value;with;semicolons';"}},
		{"SemicolonInDoubleQuotedIdentifier", `SELECT "col;name" FROM t;`, []string{`SELECT "col;name" FROM t;`}},
		{"DoubleSingleQuoteEscaping", "SELECT 'it''s;fine';", []string{"SELECT 'it''s;fine';"}},
		{"MultilineComment", "SELECT 1; /* comment; inside */ SELECT 2;", []string{"SELECT 1;", "/* comment; inside */ SELECT 2;"}},
		{"NestedCommentLikeInput", "SELECT 1 /* outer /* inner */ still */; SELECT 2;", []string{"SELECT 1 /* outer /* inner */ still */;", "SELECT 2;"}},
		{"LineCommentAtEndNoSemicolon", "SELECT 1 -- comment", []string{"SELECT 1 -- comment"}},
		{"MixedCommentStyles", "SELECT 1; -- c1\n/* c2; */ SELECT 2;", []string{"SELECT 1;", "-- c1\n/* c2; */ SELECT 2;"}},
		{"SemicolonInDollarQuotedString", "SELECT $$abc;def$$;", []string{"SELECT $$abc;def$$;"}},
		{"Unicode", "SELECT '你好;世界';", []string{"SELECT '你好;世界';"}},
		{"UnclosedString", "SELECT 'abc;", []string{"SELECT 'abc;"}},
		{"UnclosedComment", "SELECT 1; /* unclosed", []string{"SELECT 1;", "/* unclosed"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.SplitSQLStatements(tt.sql)
			assert.Equal(t, tt.want, got)
		})
	}
}
