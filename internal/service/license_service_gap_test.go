package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"chenze-faka/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetHostname(t *testing.T) {
	t.Run("returns hostname matching os.Hostname", func(t *testing.T) {
		result := getHostname()
		assert.NotEmpty(t, result)

		expected, err := os.Hostname()
		if err == nil {
			assert.Equal(t, expected, result)
		}
	})

	t.Run("fallback returns non-empty string", func(t *testing.T) {
		result := getHostname()
		assert.NotEmpty(t, result)
	})
}

func TestStartPeriodicVerify(t *testing.T) {
	t.Run("starts and stops without panic when enabled", func(t *testing.T) {
		cfg := &model.LicenseConfig{
			Enabled:     true,
			BaseURL:     "https://example.com",
			AppKey:      "test-key",
			AppSecret:   "test-secret",
			LicenseKey:  "test-license",
			Domain:      "test.example.com",
			ServerIP:    "127.0.0.1",
			CacheFile:   filepath.Join(t.TempDir(), "test_start.cache"),
			Interval:    3600,
			GracePeriod: 86400,
		}
		svc := NewLicenseService(cfg)

		assert.NotPanics(t, func() {
			svc.StartPeriodicVerify()
		})

		time.Sleep(50 * time.Millisecond)

		assert.NotPanics(t, func() {
			svc.StopPeriodicVerify()
		})
	})

	t.Run("does not start when disabled", func(t *testing.T) {
		cfg := &model.LicenseConfig{
			Enabled:     false,
			BaseURL:     "https://example.com",
			AppKey:      "test-key",
			AppSecret:   "test-secret",
			CacheFile:   filepath.Join(t.TempDir(), "test_disabled.cache"),
			Interval:    3600,
			GracePeriod: 86400,
		}
		svc := NewLicenseService(cfg)

		assert.NotPanics(t, func() {
			svc.StartPeriodicVerify()
		})

		assert.NotPanics(t, func() {
			svc.StopPeriodicVerify()
		})
	})

	t.Run("multiple start and stop cycles without panic", func(t *testing.T) {
		cfg := &model.LicenseConfig{
			Enabled:     true,
			BaseURL:     "https://example.com",
			AppKey:      "test-key",
			AppSecret:   "test-secret",
			LicenseKey:  "test-license",
			Domain:      "test.example.com",
			ServerIP:    "127.0.0.1",
			CacheFile:   filepath.Join(t.TempDir(), "test_multi.cache"),
			Interval:    3600,
			GracePeriod: 86400,
		}
		svc := NewLicenseService(cfg)

		for i := 0; i < 3; i++ {
			assert.NotPanics(t, func() {
				svc.StartPeriodicVerify()
			})
			time.Sleep(10 * time.Millisecond)
			assert.NotPanics(t, func() {
				svc.StopPeriodicVerify()
			})
		}
	})
}

func TestStopPeriodicVerifyWithoutStart(t *testing.T) {
	t.Run("calling stop without start does not panic", func(t *testing.T) {
		cfg := &model.LicenseConfig{
			Enabled:     true,
			BaseURL:     "https://example.com",
			AppKey:      "test-key",
			AppSecret:   "test-secret",
			CacheFile:   filepath.Join(t.TempDir(), "test_stop.cache"),
			Interval:    3600,
			GracePeriod: 86400,
		}
		svc := NewLicenseService(cfg)

		assert.NotPanics(t, func() {
			svc.StopPeriodicVerify()
		})
	})

	t.Run("stop on nil stopCh does not panic", func(t *testing.T) {
		svc := &LicenseService{
			stopCh: nil,
		}

		assert.NotPanics(t, func() {
			svc.StopPeriodicVerify()
		})
	})
}

func TestStartPeriodicVerifyStopChCleanup(t *testing.T) {
	cfg := &model.LicenseConfig{
		Enabled:     true,
		BaseURL:     "https://example.com",
		AppKey:      "test-key",
		AppSecret:   "test-secret",
		LicenseKey:  "test-license",
		Domain:      "test.example.com",
		ServerIP:    "127.0.0.1",
		CacheFile:   filepath.Join(t.TempDir(), "test_cleanup.cache"),
		Interval:    3600,
		GracePeriod: 86400,
	}
	svc := NewLicenseService(cfg)

	svc.StartPeriodicVerify()
	time.Sleep(50 * time.Millisecond)
	svc.StopPeriodicVerify()

	assert.NotPanics(t, func() {
		svc.StopPeriodicVerify()
	})
}

func TestStartPeriodicVerifyVerifyCalled(t *testing.T) {
	cfg := &model.LicenseConfig{
		Enabled:     true,
		BaseURL:     "https://192.0.2.1",
		AppKey:      "test-key",
		AppSecret:   "test-secret",
		LicenseKey:  "test-license",
		Domain:      "test.example.com",
		ServerIP:    "127.0.0.1",
		CacheFile:   filepath.Join(t.TempDir(), "test_verify.cache"),
		Interval:    3600,
		GracePeriod: 86400,
	}
	svc := NewLicenseService(cfg)

	assert.NotPanics(t, func() {
		svc.StartPeriodicVerify()
	})

	time.Sleep(100 * time.Millisecond)
	svc.StopPeriodicVerify()

	cache := svc.GetCache()
	require.NotNil(t, cache)
}