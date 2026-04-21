package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	// default stores the default value
	Default  = "default"
	filename = "config.toml"
	appName  = "pgxcli"
)

// config represents the high level configuration
type Config struct {
	Main main `mapstructure:"main" toml:"main"`
}

type main struct {
	Prompt      string        `mapstructure:"prompt" toml:"prompt"`
	Style       string        `mapstructure:"style" toml:"style"`
	HistoryFile string        `mapstructure:"history_file" toml:"history_file"`
	LogFile     string        `mapstructure:"log_file" toml:"log_file"`
	Pager       string        `mapstructure:"pager" toml:"pager"`
	OnError     OnErrorAction `mapstructure:"on_error" toml:"on_error"`
}

// Load reads the embedded default configuration and merges with user configuration.
func Load() (*Config, error) {
	userPath, err := UserConfigPath()
	if err != nil {
		return nil, err
	}
	if err := ensureUserConfig(userPath); err != nil {
		return nil, err
	}

	defaultV := viper.New()
	defaultV.SetConfigType("toml")
	if err := defaultV.ReadConfig(bytes.NewReader(defaultConfigFile)); err != nil {
		return nil, fmt.Errorf("read default config: %w", err)
	}

	userV := viper.New()
	userV.SetConfigFile(userPath)
	if err := userV.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read user config: %w", err)
	}

	// user settings land on top of default settings
	if err := defaultV.MergeConfigMap(userV.AllSettings()); err != nil {
		return nil, fmt.Errorf("merge configs: %w", err)
	}

	var cfg Config
	if err := defaultV.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}
	return &cfg, nil
}

func GetDefaultConfig() (*Config, error) {
	defaultV := viper.New()
	defaultV.SetConfigType("toml")
	if err := defaultV.ReadConfig(bytes.NewReader(defaultConfigFile)); err != nil {
		return nil, fmt.Errorf("read default config: %w", err)
	}

	var cfg Config
	if err := defaultV.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &cfg, nil
}

// returns the configuration file path or error, example: ~/.config/pgxcli/config.toml
func UserConfigPath() (string, error) {
	userdir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userdir, appName, filename), nil
}

// ensureUserConfig write embed on firt run
func ensureUserConfig(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	if err := os.WriteFile(path, defaultConfigFile, 0o644); err != nil {
		return fmt.Errorf("write default config: %w", err)
	}
	return nil
}
