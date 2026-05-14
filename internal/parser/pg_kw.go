package parser

import (
	_ "embed"
	"strings"
)

//go:embed pg_kw.txt
var pgKeywords string

// LoadPgKeywords returns PostgreSQL keyword suggestions used by autocompletion.
func LoadPgKeywords() []string {
	return suggestionsFromFile(pgKeywords)
}

func suggestionsFromFile(contents string) []string {
	var suggestions []string
	for _, line := range strings.Split(contents, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			suggestions = append(suggestions, strings.ToUpper(line))
		}
	}

	return suggestions
}
