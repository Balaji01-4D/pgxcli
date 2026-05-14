package parser_test

import (
	"testing"

	"github.com/balaji01-4d/pgxcli/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadPgKeywords(t *testing.T) {
	keywords := parser.LoadPgKeywords()
	require.NotEmpty(t, keywords)
	require.Len(t, keywords, 511)
	assert.Contains(t, keywords, "SELECT")
	assert.Contains(t, keywords, "INSERT")
	assert.Contains(t, keywords, "UPDATE")
	assert.Contains(t, keywords, "DELETE")
}
