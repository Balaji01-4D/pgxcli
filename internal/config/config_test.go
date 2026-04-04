package config

import (
	"os"
	path "path/filepath"
	"runtime"
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

func TestSaveConfig_CreatesDirWithRestrictivePermission(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory permission bits are not reliable on Windows")
	}

	tempDir := t.TempDir()
	configPath := path.Join(tempDir, "nested", "config.toml")

	err := SaveConfig(configPath)
	assert.NoError(t, err)

	info, err := os.Stat(path.Dir(configPath))
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0o700), info.Mode().Perm())
}

func TestMergeConfig(t *testing.T) {
	testCase := []struct {
		name        string
		baseCfg     Config
		overrideCfg Config
		expectedCfg Config
	}{
		{
			name: "override history file",
			baseCfg: Config{
				Main: main{
					Prompt:      "\\u@\\h:\\d> ",
					Style:       "monokai",
					HistoryFile: "default",
					LogFile:     "default",
				},
			},
			overrideCfg: Config{
				Main: main{
					Prompt:      "\\u@\\h:\\d> ",
					Style:       "monokai",
					HistoryFile: "custom_history.txt",
				},
			},
			expectedCfg: Config{
				Main: main{
					Prompt:      "\\u@\\h:\\d> ",
					Style:       "monokai",
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
					Style:       "monokai",
					HistoryFile: "default",
					LogFile:     "default",
				},
			},
			overrideCfg: Config{
				Main: main{
					Prompt:  "\\u@\\h:\\d> ",
					Style:   "monokai",
					LogFile: "custom_log.txt",
				},
			},
			expectedCfg: Config{
				Main: main{
					Prompt:      "\\u@\\h:\\d> ",
					Style:       "monokai",
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
					Style:       "monokai",
					HistoryFile: "default",
					LogFile:     "default",
				},
			},
			overrideCfg: Config{
				Main: main{
					Prompt: "\\u@\\h:\\d$ ",
					Style:  "monokai",
				},
			},
			expectedCfg: Config{
				Main: main{
					Prompt:      "\\u@\\h:\\d$ ",
					Style:       "monokai",
					HistoryFile: "default",
					LogFile:     "default",
				},
			},
		},
		{
			name: "override style",
			baseCfg: Config{
				Main: main{
					Prompt:      "\\u@\\h:\\d> ",
					Style:       "monokai",
					HistoryFile: "default",
					LogFile:     "default",
				},
			},
			overrideCfg: Config{
				Main: main{
					Style: "dracula",
				},
			},
			expectedCfg: Config{
				Main: main{
					Prompt:      "\\u@\\h:\\d> ",
					Style:       "dracula",
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
					Style:       "monokai",
					HistoryFile: "default",
					LogFile:     "default",
				},
			},
			overrideCfg: Config{},
			expectedCfg: Config{
				Main: main{
					Prompt:      "\\u@\\h:\\d> ",
					Style:       "monokai",
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
					Style:       "monokai",
					HistoryFile: "default",
					LogFile:     "default",
				},
			},
			overrideCfg: Config{
				Main: main{
					Prompt:      "\\u@\\h:\\d$ ",
					Style:       "dracula",
					HistoryFile: "custom_history.txt",
					LogFile:     "custom_log.txt",
				},
			},
			expectedCfg: Config{
				Main: main{
					Prompt:      "\\u@\\h:\\d$ ",
					Style:       "dracula",
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
			assert.Equal(t, tc.expectedCfg.Main.Style, resultCfg.Main.Style)
			assert.Equal(t, tc.expectedCfg.Main.HistoryFile, resultCfg.Main.HistoryFile)
			assert.Equal(t, tc.expectedCfg.Main.LogFile, resultCfg.Main.LogFile)
		})
	}
}
