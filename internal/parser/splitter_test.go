package parser_test

import (
	"testing"

	"github.com/balaji01-4d/pgxcli/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitSqlStatement_Empty(t *testing.T) {
	assertSplit(t, "", nil)
}

func TestSplitSqlStatement_WhitespaceOnly(t *testing.T) {
	assertSplit(t, "   \n\t  ", nil)
}

func TestSplitSqlStatement_SingleStatement(t *testing.T) {
	assertSplit(t, "SELECT 1;", []string{"SELECT 1"})
}

func TestSplitSqlStatement_NoTrailingSemicolon(t *testing.T) {
	assertSplit(t, "SELECT 1", []string{"SELECT 1"})
}

func TestSplitSqlStatement_MultipleStatements(t *testing.T) {
	assertSplit(t, "SELECT 1; SELECT 2; SELECT 3;", []string{"SELECT 1", "SELECT 2", "SELECT 3"})
}

func TestSplitSqlStatement_LeadingAndTrailingSemicolons(t *testing.T) {
	assertSplit(t, ";SELECT 1;;", []string{"SELECT 1"})
}

func TestSplitSqlStatement_EmptyStatementsBetweenValids(t *testing.T) {
	assertSplit(t, "SELECT 1;;SELECT 2;", []string{"SELECT 1", "SELECT 2"})
}

func TestSplitSqlStatement_MultilineStatements(t *testing.T) {
	assertSplit(t, "SELECT\n1;\nSELECT\n2;", []string{"SELECT\n1", "SELECT\n2"})
}

func TestSplitSqlStatement_SemicolonInStringLiteral(t *testing.T) {
	assertSplit(t, "SELECT 'value;with;semicolons';", []string{"SELECT 'value;with;semicolons'"})
}

func TestSplitSqlStatement_SemicolonInDoubleQuotedIdentifier(t *testing.T) {
	assertSplit(t, `SELECT "col;name" FROM t;`, []string{`SELECT "col;name" FROM t`})
}

func TestSplitSqlStatement_DoubleSingleQuoteEscaping(t *testing.T) {
	assertSplit(t, "SELECT 'it''s;fine';", []string{"SELECT 'it''s;fine'"})
}

func TestSplitSqlStatement_BackslashEscapedQuoteErrors(t *testing.T) {
	assertSplitErr(t, "SELECT 'it\\'s;fine';")
}

func TestSplitSqlStatement_MultilineComment(t *testing.T) {
	assertSplit(t, "SELECT 1; /* comment; inside */ SELECT 2;", []string{"SELECT 1", "/* comment; inside */ SELECT 2"})
}

func TestSplitSqlStatement_NestedCommentLikeInput(t *testing.T) {
	assertSplit(t, "SELECT 1 /* outer /* inner */ still */; SELECT 2;", []string{"SELECT 1 /* outer /* inner */ still */", "SELECT 2"})
}

func TestSplitSqlStatement_LineCommentAtEndNoSemicolon(t *testing.T) {
	assertSplit(t, "SELECT 1 -- comment", []string{"SELECT 1 -- comment"})
}

func TestSplitSqlStatement_MixedCommentStyles(t *testing.T) {
	assertSplit(t, "SELECT 1; -- c1\n/* c2; */ SELECT 2;", []string{"SELECT 1", "-- c1\n/* c2; */ SELECT 2"})
}

func TestSplitSqlStatement_SemicolonInDollarQuotedString(t *testing.T) {
	assertSplit(t, "SELECT $$abc;def$$;", []string{"SELECT $$abc;def$$"})
}

func TestSplitSqlStatement_BeginEndBlockErrors(t *testing.T) {
	assertSplitErr(t, "BEGIN SELECT 1; SELECT 2; END;")
}

func TestSplitSqlStatement_UnclosedStringErrors(t *testing.T) {
	assertSplitErr(t, "SELECT 'abc;")
}

func TestSplitSqlStatement_UnclosedCommentErrors(t *testing.T) {
	assertSplitErr(t, "SELECT 1; /* unclosed")
}

func TestSplitSqlStatement_Unicode(t *testing.T) {
	assertSplit(t, "SELECT '你好;世界';", []string{"SELECT '你好;世界'"})
}

func assertSplit(t *testing.T, sql string, want []string) {
	t.Helper()
	got, err := parser.SplitSqlStatement(sql)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func assertSplitErr(t *testing.T, sql string) {
	t.Helper()
	_, err := parser.SplitSqlStatement(sql)
	require.Error(t, err)
}
