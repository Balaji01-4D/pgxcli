package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const filename = "config.toml"

type Config struct {
	Prompt string `toml:"prompt"`
}

var DefaultConfig = Config{
	Prompt: `\u@\h:\d> `,
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

func CheckConfigExists(configDir string) (string, bool){
	path := filepath.Join(configDir, filename)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return path, false
	}
	return path, true
}

func SaveConfig(path string, cfg Config) error {
	dir := filepath.Dir(path)
	os.MkdirAll(dir, os.ModePerm)
	f, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		return err
	}
	defer f.Close()

	enc := toml.NewEncoder(f)
	return enc.Encode(cfg)
}
