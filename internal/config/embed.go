package config

import _ "embed"

//go:embed config.toml
var defaultConfigFile []byte
