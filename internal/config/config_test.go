package config

import (
	"os"
	path "path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_ValidConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := path.Join(tempDir, "config.toml")
	expectedCfg := Config{
		Main: main{
			Prompt:      "\\u@\\h:\\d> ",
			HistoryFile: "default",
		},
	}
	content, err := toml.Marshal(expectedCfg)
	assert.NoError(t, err)

	writeErr := os.WriteFile(configPath, content, 0644)
	assert.NoError(t, writeErr)

	actualCfg, err := LoadConfig(configPath)
	assert.NoError(t, err)
	assert.Equal(t, expectedCfg.Main.Prompt, actualCfg.Main.Prompt)
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
