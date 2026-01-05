package config

import (
	"os"
	path "path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_ValidConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := path.Join(tempDir, "config.toml")
	configContent := `prompt = "\\u@\\h:\\d> "`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)

	cfg, err := LoadConfig(configPath)
	assert.NoError(t, err)
	assert.Equal(t, "\\u@\\h:\\d> ", cfg.Main.Prompt)
}

func TestLoadConfig_InvalidConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := path.Join(tempDir, "config.toml")
	invalidContent := `prompt = "\\u@\\h:\\d> ` // Missing closing quote

	err := os.WriteFile(configPath, []byte(invalidContent), 0644)
	assert.NoError(t, err)

	_, err = LoadConfig(configPath)
	assert.Error(t, err)
}

func TestLoadConfig_MissingFile(t *testing.T) {
	_, err := LoadConfig("non_existent_config.toml")
	assert.Error(t, err)
}

func TestSaveConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := path.Join(tempDir, "config.toml")

	cfg := Config{
		Main: main{
			Prompt:      "\\u@\\h:\\d> ",
			HistoryFile: "default",
		},
	}

	err := SaveConfig(configPath, cfg)
	assert.NoError(t, err)
	loadedCfg, err := LoadConfig(configPath)
	assert.NoError(t, err)
	assert.Equal(t, cfg.Main.Prompt, loadedCfg.Main.Prompt)
}
