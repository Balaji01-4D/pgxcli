package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setIsolatedUserConfigEnv(t *testing.T) {
	t.Helper()

	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("USERPROFILE", tempHome)
	t.Setenv("XDG_CONFIG_HOME", tempHome)
	t.Setenv("APPDATA", tempHome)
}

func TestLoad_Success(t *testing.T) {
	setIsolatedUserConfigEnv(t)

	cfg, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "\\u@\\h:\\d> ", cfg.Main.Prompt)
	assert.Equal(t, "monokai", cfg.Main.Style)
	assert.Equal(t, "default", cfg.Main.HistoryFile)
	assert.Equal(t, "default", cfg.Main.LogFile)
	assert.Equal(t, "auto", cfg.Main.Pager)
	assert.Equal(t, OnErrorStop, cfg.Main.OnError)
}

func TestLoad_UserConfigOverridesDefaults(t *testing.T) {
	setIsolatedUserConfigEnv(t)

	userConfigPath, err := UserConfigPath()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(userConfigPath), 0o700))

	userConfig := `[main]
prompt = "custom> "
style = "dracula"
history_file = "/custom/history.txt"
log_file = "/custom/log.txt"
pager = "never"
on_error = "RESUME"
`
	require.NoError(t, os.WriteFile(userConfigPath, []byte(userConfig), 0o644))

	cfg, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "custom> ", cfg.Main.Prompt)
	assert.Equal(t, "dracula", cfg.Main.Style)
	assert.Equal(t, "/custom/history.txt", cfg.Main.HistoryFile)
	assert.Equal(t, "/custom/log.txt", cfg.Main.LogFile)
	assert.Equal(t, "never", cfg.Main.Pager)
	assert.Equal(t, OnErrorResume, cfg.Main.OnError)
}

func TestLoad_PartialUserConfigMergesWithDefaults(t *testing.T) {
	setIsolatedUserConfigEnv(t)

	userConfigPath, err := UserConfigPath()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(userConfigPath), 0o700))

	userConfig := `[main]
prompt = "custom> "
`
	require.NoError(t, os.WriteFile(userConfigPath, []byte(userConfig), 0o644))

	cfg, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "custom> ", cfg.Main.Prompt)
	assert.Equal(t, "monokai", cfg.Main.Style)
	assert.Equal(t, "default", cfg.Main.HistoryFile)
	assert.Equal(t, "default", cfg.Main.LogFile)
	assert.Equal(t, "auto", cfg.Main.Pager)
	assert.Equal(t, OnErrorStop, cfg.Main.OnError)
}

func TestLoad_InvalidUserConfig(t *testing.T) {
	setIsolatedUserConfigEnv(t)

	userConfigPath, err := UserConfigPath()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(userConfigPath), 0o700))

	invalidConfig := `this is not valid toml [[[`
	require.NoError(t, os.WriteFile(userConfigPath, []byte(invalidConfig), 0o644))

	_, err = Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read user config")
}

func TestUserConfigPath(t *testing.T) {
	path, err := UserConfigPath()
	require.NoError(t, err)
	assert.NotEmpty(t, path)
	assert.Contains(t, path, appName)
}

func TestEnsureUserConfig_CreatesConfigOnFirstRun(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	err := ensureUserConfig(configPath)
	require.NoError(t, err)

	assert.FileExists(t, configPath)

	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "[main]")
	assert.Contains(t, string(content), "prompt")
	assert.Contains(t, string(content), "style")
}

func TestEnsureUserConfig_DoesNotOverwriteExisting(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	customContent := `[main]
prompt = "custom> "
style = "dracula"
history_file = "custom"
log_file = "custom"
pager = "always"
on_error = "STOP"
`
	require.NoError(t, os.WriteFile(configPath, []byte(customContent), 0o644))

	err := ensureUserConfig(configPath)
	require.NoError(t, err)

	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, customContent, string(content))
}

func TestEnsureUserConfig_CreatesDirectoryWithRestrictivePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory permission bits are not reliable on Windows")
	}

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nested", "deep", "config.toml")

	err := ensureUserConfig(configPath)
	require.NoError(t, err)

	parentDir := filepath.Dir(configPath)
	info, err := os.Stat(parentDir)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o700), info.Mode().Perm())
}
