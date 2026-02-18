package config

import (
	_ "embed"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	// default stores the default value
	Default  = "default"
	filename = "config.toml"
)

//go:embed config.toml
var config []byte

// config represents the high level configuration
type Config struct {
	Main main `toml:"main"`
}

type main struct {
	Prompt      string `toml:"prompt"`
	HistoryFile string `toml:"history_file"`
	LogFile string `toml:"log_file"`
	Style   string `toml:"style"`
}

// default configuration
var DefaultConfig = Config{
	Main: main{
		Prompt:      `\u@\h:\d> `,
		HistoryFile: Default,
		LogFile: Default, 
	},
}

// returns the configuration directory or error, example: ~/.config/pgxcli
func GetConfigDir() (string, error) {
	userdir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userdir, "pgxcli"), nil
}

// loads the configuration in the provided path
func LoadConfig(path string) (config Config, err error) {
	var cfg Config
	_, err = toml.DecodeFile(path, &cfg)
	return cfg, err
}

// ensures the configuration exists in the given directory
func CheckConfigExists(configDir string) (string, bool) {
	path := filepath.Join(configDir, filename)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return path, false
	}
	return path, true
}

// used to save the default configuration when default configuration doesn't exist
func SaveConfig(path string) error {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	return os.WriteFile(path, config, 0644)
}