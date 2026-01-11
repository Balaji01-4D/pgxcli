package config

import (
	_ "embed"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	Default  = "default"
	filename = "config.toml"
)

//go:embed config.toml
var config []byte

type Config struct {
	Main main `toml:"main"`
}

type main struct {
	Prompt      string `toml:"prompt"`
	HistoryFile string `toml:"history_file"`
	LogFile string `toml:"log_file"`
}

var DefaultConfig = Config{
	Main: main{
		Prompt:      `\u@\h:\d> `,
		HistoryFile: "default",
		LogFile: "default", 
	},
}

func GetConfigDir() (string, error) {
	userdir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userdir, "pgxcli"), nil
}

func LoadConfig(path string) (config Config, err error) {
	var cfg Config
	_, err = toml.DecodeFile(path, &cfg)
	return cfg, err
}

func CheckConfigExists(configDir string) (string, bool) {
	path := filepath.Join(configDir, filename)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return path, false
	}
	return path, true
}

func SaveConfig(path string) error {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	return os.WriteFile(path, config, 0644)
}