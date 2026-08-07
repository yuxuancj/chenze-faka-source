package main

import (
	"os"
	"path/filepath"
	"testing"

	"chenze-faka/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEnv(t *testing.T) {
	t.Run("env var set", func(t *testing.T) {
		os.Setenv("TEST_MAIN_KEY", "test_value")
		defer os.Unsetenv("TEST_MAIN_KEY")
		result := getEnv("TEST_MAIN_KEY", "default")
		assert.Equal(t, "test_value", result)
	})

	t.Run("default fallback when not set", func(t *testing.T) {
		os.Unsetenv("TEST_MAIN_NONEXISTENT")
		result := getEnv("TEST_MAIN_NONEXISTENT", "default_value")
		assert.Equal(t, "default_value", result)
	})

	t.Run("empty env var returns default", func(t *testing.T) {
		os.Setenv("TEST_MAIN_EMPTY", "")
		defer os.Unsetenv("TEST_MAIN_EMPTY")
		result := getEnv("TEST_MAIN_EMPTY", "default_for_empty")
		assert.Equal(t, "default_for_empty", result)
	})
}

func TestGetEnvInt(t *testing.T) {
	t.Run("valid int env var", func(t *testing.T) {
		os.Setenv("TEST_MAIN_INT", "8080")
		defer os.Unsetenv("TEST_MAIN_INT")
		result := getEnvInt("TEST_MAIN_INT", 3000)
		assert.Equal(t, 8080, result)
	})

	t.Run("invalid value falls back to default", func(t *testing.T) {
		os.Setenv("TEST_MAIN_INT_BAD", "not_a_number")
		defer os.Unsetenv("TEST_MAIN_INT_BAD")
		result := getEnvInt("TEST_MAIN_INT_BAD", 3000)
		assert.Equal(t, 3000, result)
	})

	t.Run("default when not set", func(t *testing.T) {
		os.Unsetenv("TEST_MAIN_INT_NONEXISTENT")
		result := getEnvInt("TEST_MAIN_INT_NONEXISTENT", 9999)
		assert.Equal(t, 9999, result)
	})

	t.Run("negative int value", func(t *testing.T) {
		os.Setenv("TEST_MAIN_INT_NEG", "-1")
		defer os.Unsetenv("TEST_MAIN_INT_NEG")
		result := getEnvInt("TEST_MAIN_INT_NEG", 0)
		assert.Equal(t, -1, result)
	})
}

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()
	require.NotNil(t, cfg)
	assert.Equal(t, "晨泽发卡", cfg.System.SiteName)
	assert.Equal(t, "debug", cfg.System.Mode)
	assert.Equal(t, "mysql", cfg.Database.Driver)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, "root", cfg.Database.User)
	assert.Equal(t, "", cfg.Database.Password)
	assert.Equal(t, "chenze_faka", cfg.Database.DBName)
	assert.Equal(t, "", cfg.Database.SQLite)
	assert.Equal(t, "chenze-faka-secret", cfg.JWT.Secret)
	assert.Equal(t, 72, cfg.JWT.ExpireTime)
	assert.Equal(t, 3306, cfg.Database.Port)
	assert.True(t, cfg.License.Enabled)
	assert.Equal(t, "https://auth.seanld.com", cfg.License.BaseURL)
	assert.Equal(t, "app_1c1467945bb2_3105", cfg.License.AppKey)
	assert.Equal(t, 3600, cfg.License.Interval)
	assert.Equal(t, 86400, cfg.License.GracePeriod)
	assert.Equal(t, "license.cache", cfg.License.CacheFile)
}

func TestLoadConfig(t *testing.T) {
	t.Run("loads config from valid yaml file", func(t *testing.T) {
		dir := t.TempDir()
		configPath := filepath.Join(dir, "config.yaml")
		configContent := "system:\n  site_name: \"测试站点\"\n  port: 9999\n  mode: \"release\"\ndatabase:\n  driver: \"sqlite\"\n  host: \"db.example.com\"\n  port: 3307\n  user: \"admin\"\n  password: \"secret\"\n  dbname: \"test_db\"\njwt:\n  secret: \"my-secret\"\n  expire_hours: 48\nlicense:\n  enabled: false\n"
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		cfg, err := loadConfig(configPath)
		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Equal(t, "测试站点", cfg.System.SiteName)
		assert.Equal(t, 9999, cfg.System.Port)
		assert.Equal(t, "release", cfg.System.Mode)
		assert.Equal(t, "sqlite", cfg.Database.Driver)
		assert.Equal(t, "db.example.com", cfg.Database.Host)
		assert.Equal(t, 3307, cfg.Database.Port)
		assert.Equal(t, "admin", cfg.Database.User)
		assert.Equal(t, "secret", cfg.Database.Password)
		assert.Equal(t, "test_db", cfg.Database.DBName)
		assert.Equal(t, "my-secret", cfg.JWT.Secret)
		assert.Equal(t, 48, cfg.JWT.ExpireTime)
		assert.False(t, cfg.License.Enabled)
	})

	t.Run("fills default values for missing fields", func(t *testing.T) {
		dir := t.TempDir()
		configPath := filepath.Join(dir, "config.yaml")
		configContent := "system:\n  site_name: \"部分配置\"\ndatabase:\n  driver: \"sqlite\"\n"
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		cfg, err := loadConfig(configPath)
		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Equal(t, "部分配置", cfg.System.SiteName)
		assert.Equal(t, 12398, cfg.System.Port)
		assert.Equal(t, "debug", cfg.System.Mode)
		assert.Equal(t, "sqlite", cfg.Database.Driver)
		assert.Equal(t, 3306, cfg.Database.Port)
		assert.Equal(t, "chenze-faka-secret", cfg.JWT.Secret)
		assert.Equal(t, 72, cfg.JWT.ExpireTime)
	})

	t.Run("falls back to default when file not found", func(t *testing.T) {
		cfg, err := loadConfig(filepath.Join(t.TempDir(), "nonexistent.yaml"))
		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Equal(t, "晨泽发卡", cfg.System.SiteName)
	})

	t.Run("falls back to default when yaml is invalid", func(t *testing.T) {
		dir := t.TempDir()
		configPath := filepath.Join(dir, "bad.yaml")
		err := os.WriteFile(configPath, []byte("key: [unclosed"), 0644)
		require.NoError(t, err)

		cfg, err := loadConfig(configPath)
		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Equal(t, "晨泽发卡", cfg.System.SiteName)
	})
}

func TestCheckInstalled(t *testing.T) {
	lockFile := "install.lock"
	t.Cleanup(func() { os.Remove(lockFile) })

	t.Run("returns true when install.lock exists", func(t *testing.T) {
		err := os.WriteFile(lockFile, []byte("installed"), 0644)
		require.NoError(t, err)
		assert.True(t, checkInstalled())
	})

	t.Run("returns false when install.lock does not exist", func(t *testing.T) {
		os.Remove(lockFile)
		assert.False(t, checkInstalled())
	})
}

func TestInitLicenseService(t *testing.T) {
	t.Run("returns service when license disabled", func(t *testing.T) {
		cfg := &model.Config{
			License: model.LicenseConfig{
				Enabled:    false,
				BaseURL:    "https://example.com",
				AppKey:     "test-key",
				AppSecret:  "test-secret",
				LicenseKey: "test-license",
				Domain:     "test.example.com",
				ServerIP:   "127.0.0.1",
				CacheFile:  filepath.Join(t.TempDir(), "license.cache"),
				Interval:   3600,
				GracePeriod: 86400,
			},
		}
		svc := initLicenseService(cfg, true)
		require.NotNil(t, svc)
	})

	t.Run("returns service when not installed", func(t *testing.T) {
		cfg := &model.Config{
			License: model.LicenseConfig{
				Enabled:    true,
				BaseURL:    "https://example.com",
				AppKey:     "test-key",
				AppSecret:  "test-secret",
				LicenseKey: "test-license",
				Domain:     "test.example.com",
				ServerIP:   "127.0.0.1",
				CacheFile:  filepath.Join(t.TempDir(), "license.cache"),
				Interval:   3600,
				GracePeriod: 86400,
			},
		}
		svc := initLicenseService(cfg, false)
		require.NotNil(t, svc)
	})

	t.Run("handles enabled and installed without panic", func(t *testing.T) {
		cfg := &model.Config{
			License: model.LicenseConfig{
				Enabled:    true,
				BaseURL:    "https://192.0.2.1",
				AppKey:     "test-key",
				AppSecret:  "test-secret",
				LicenseKey: "test-license",
				Domain:     "test.example.com",
				ServerIP:   "127.0.0.1",
				CacheFile:  filepath.Join(t.TempDir(), "license.cache"),
				Interval:   3600,
				GracePeriod: 86400,
			},
		}
		assert.NotPanics(t, func() {
			svc := initLicenseService(cfg, true)
			require.NotNil(t, svc)
			svc.StopPeriodicVerify()
		})
	})
}

func TestStartCleanupTask(t *testing.T) {
	assert.NotPanics(t, func() {
		startCleanupTask()
	})
}