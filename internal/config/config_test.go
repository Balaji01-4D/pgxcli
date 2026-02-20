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

	writeErr := os.WriteFile(configPath, content, 0o644)
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

	err := SaveConfig(configPath)
	assert.NoError(t, err)

	loadedCfg, err := LoadConfig(configPath)
	assert.NoError(t, err)

	assert.Equal(t, "default", loadedCfg.Main.HistoryFile)
}

func TestMergeConfig(t *testing.T) {

	testCase := []struct {
		name       string
		baseCfg    Config
		overrideCfg Config
		expectedCfg Config
	}{
		{
			name: "override history file",
			baseCfg: Config{
				Main: main{
					Prompt:      "\\u@\\h:\\d> ",
					HistoryFile: "default",
					LogFile:     "default",
				},
			},
			overrideCfg: Config{
				Main: main{
					Prompt:      "\\u@\\h:\\d> ",
					HistoryFile: "custom_history.txt",
				},
			},
			expectedCfg: Config{
				Main: main{
					Prompt:      "\\u@\\h:\\d> ",
					HistoryFile: "custom_history.txt",
					LogFile:     "default",
				},
			},
		},
		{
			name: "override log file",
			baseCfg: Config{
				Main: main{
					Prompt:      "\\u@\\h:\\d> ",
					HistoryFile: "default",
					LogFile:     "default",
				},
			},
			overrideCfg: Config{
				Main: main{
					Prompt:  "\\u@\\h:\\d> ",
					LogFile: "custom_log.txt",
				},
			},
			expectedCfg: Config{
				Main: main{
					Prompt:      "\\u@\\h:\\d> ",
					HistoryFile: "default",
					LogFile:     "custom_log.txt",
				},
			},
		},
		{
			name: "override prompt",
			baseCfg: Config{
				Main: main{
					Prompt:      "\\u@\\h:\\d> ",
					HistoryFile: "default",
					LogFile:     "default",
				},
			},
			overrideCfg: Config{
				Main: main{
					Prompt: "\\u@\\h:\\d$ ",
				},
			},
			expectedCfg: Config{
				Main: main{
					Prompt:      "\\u@\\h:\\d$ ",
					HistoryFile: "default",
					LogFile:     "default",
				},
			},
		},
		{
			name: "no override",
			baseCfg: Config{
				Main: main{
					Prompt:      "\\u@\\h:\\d> ",
					HistoryFile: "default",
					LogFile:     "default",
				},
			},
			overrideCfg: Config{},
			expectedCfg: Config{
				Main: main{
					Prompt:      "\\u@\\h:\\d> ",
					HistoryFile: "default",
					LogFile:     "default",
				},
			},	
		},
		{
			name: "override all fields",
			baseCfg: Config{
				Main: main{
					Prompt:      "\\u@\\h:\\d> ",
					HistoryFile: "default",
					LogFile:     "default",
				},
			},
			overrideCfg: Config{
				Main: main{
					Prompt:      "\\u@\\h:\\d$ ",
					HistoryFile: "custom_history.txt",
					LogFile:     "custom_log.txt",
				},
			},
			expectedCfg: Config{
				Main: main{
					Prompt:      "\\u@\\h:\\d$ ",
					HistoryFile: "custom_history.txt",
					LogFile:     "custom_log.txt",
				},
			},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			resultCfg := MergeConfig(tc.baseCfg, tc.overrideCfg)
			assert.Equal(t, tc.expectedCfg.Main.Prompt, resultCfg.Main.Prompt)
			assert.Equal(t, tc.expectedCfg.Main.HistoryFile, resultCfg.Main.HistoryFile)
			assert.Equal(t, tc.expectedCfg.Main.LogFile, resultCfg.Main.LogFile)
		})
	}
}
