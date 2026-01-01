package parser

import (
	_ "embed"
	"encoding/json"
)

//go:embed pgliterals.json
var PGLiterals []byte

var literals = parsePGLiterals(PGLiterals)

func parsePGLiterals(data []byte) map[string]any {
	var result map[string]any
	err := json.Unmarshal(data, &result)
	if err != nil {
		panic("failed to parse PGLiterals: " + err.Error())
	}
	return result
}

func GetPGLiterals() map[string]any {
	return literals
}
