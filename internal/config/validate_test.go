package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate_Success(t *testing.T) {
	cfg := Config{
		Main: main{
			Prompt:      "test> ",
			Style:       "monokai",
			HistoryFile: "default",
			LogFile:     "default",
			Pager:       "auto",
			OnError:     OnErrorStop,
		},
	}

	err := validate(cfg)
	assert.NoError(t, err)
}

func TestValidate_MultipleErrors(t *testing.T) {
	cfg := Config{
		Main: main{
			Prompt:      "",
			Style:       "",
			HistoryFile: "",
			LogFile:     "",
			Pager:       "",
			OnError:     "",
		},
	}

	err := validate(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prompt must not be empty")
	assert.Contains(t, err.Error(), "style must not be empty")
	assert.Contains(t, err.Error(), "history file path must not be empty")
	assert.Contains(t, err.Error(), "log file path must not be empty")
	assert.Contains(t, err.Error(), "pager mode must not be empty")
	assert.Contains(t, err.Error(), "on_error action must not be empty")
}

func TestLoad_ValidationFailsOnEmptyPrompt(t *testing.T) {
	setIsolatedUserConfigEnv(t)

	userConfigPath, err := UserConfigPath()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(userConfigPath), 0o700))

	userConfig := `[main]
prompt = ""
style = "monokai"
history_file = "default"
log_file = "default"
pager = "auto"
on_error = "STOP"
`
	require.NoError(t, os.WriteFile(userConfigPath, []byte(userConfig), 0o644))

	_, err = Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validate config")
	assert.Contains(t, err.Error(), "prompt must not be empty")
}

func TestLoad_ValidationFailsOnEmptyStyle(t *testing.T) {
	setIsolatedUserConfigEnv(t)

	userConfigPath, err := UserConfigPath()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(userConfigPath), 0o700))

	userConfig := `[main]
prompt = "test> "
style = ""
history_file = "default"
log_file = "default"
pager = "auto"
on_error = "STOP"
`
	require.NoError(t, os.WriteFile(userConfigPath, []byte(userConfig), 0o644))

	_, err = Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validate config")
	assert.Contains(t, err.Error(), "style must not be empty")
}

func TestLoad_ValidationFailsOnInvalidPagerMode(t *testing.T) {
	setIsolatedUserConfigEnv(t)

	userConfigPath, err := UserConfigPath()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(userConfigPath), 0o700))

	userConfig := `[main]
prompt = "test> "
style = "monokai"
history_file = "default"
log_file = "default"
pager = "sometimes"
on_error = "STOP"
`
	require.NoError(t, os.WriteFile(userConfigPath, []byte(userConfig), 0o644))

	_, err = Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validate config")
	assert.Contains(t, err.Error(), "pager mode must be one of: auto, always, never")
}

func TestLoad_ValidationFailsOnInvalidOnErrorAction(t *testing.T) {
	setIsolatedUserConfigEnv(t)

	userConfigPath, err := UserConfigPath()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(userConfigPath), 0o700))

	userConfig := `[main]
prompt = "test> "
style = "monokai"
history_file = "default"
log_file = "default"
pager = "auto"
on_error = "continue"
`
	require.NoError(t, os.WriteFile(userConfigPath, []byte(userConfig), 0o644))

	_, err = Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validate config")
	assert.Contains(t, err.Error(), "on_error action must be one of: STOP, RESUME")
}

func TestLoad_ValidationAllowsTrimmedPagerMode(t *testing.T) {
	setIsolatedUserConfigEnv(t)

	userConfigPath, err := UserConfigPath()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(userConfigPath), 0o700))

	userConfig := `[main]
prompt = "test> "
style = "monokai"
history_file = "default"
log_file = "default"
pager = " auto "
on_error = "STOP"
`
	require.NoError(t, os.WriteFile(userConfigPath, []byte(userConfig), 0o644))

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, " auto ", cfg.Main.Pager)
	assert.Equal(t, OnErrorStop, cfg.Main.OnError)
}
