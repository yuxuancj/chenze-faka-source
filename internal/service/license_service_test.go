package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"chenze-faka/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLicenseService_Md5Hex(t *testing.T) {
	t.Run("md5 hex produces consistent results", func(t *testing.T) {
		result1 := md5Hex("hello")
		result2 := md5Hex("hello")
		assert.Equal(t, result1, result2)
		assert.Len(t, result1, 32)
	})

	t.Run("md5 hex different inputs produce different outputs", func(t *testing.T) {
		result1 := md5Hex("hello")
		result2 := md5Hex("world")
		assert.NotEqual(t, result1, result2)
	})

	t.Run("md5 hex empty string", func(t *testing.T) {
		result := md5Hex("")
		assert.Len(t, result, 32)
		assert.Equal(t, "d41d8cd98f00b204e9800998ecf8427e", result)
	})
}

func TestLicenseService_HmacSHA256(t *testing.T) {
	t.Run("hmac sha256 produces consistent results", func(t *testing.T) {
		result1 := hmacSHA256("key", "message")
		result2 := hmacSHA256("key", "message")
		assert.Equal(t, result1, result2)
		assert.Len(t, result1, 64)
	})

	t.Run("hmac sha256 different keys produce different outputs", func(t *testing.T) {
		result1 := hmacSHA256("key1", "message")
		result2 := hmacSHA256("key2", "message")
		assert.NotEqual(t, result1, result2)
	})
}

func TestLicenseService_QuickVerify(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req verifyRequest
		json.NewDecoder(r.Body).Decode(&req)

		resp := verifyResponse{}
		resp.Code = 200
		resp.Msg = "ok"
		resp.Data.Result = "pass"
		resp.Data.AppName = "TestApp"
		resp.Data.ExpireAt = "2099-12-31"

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &model.LicenseConfig{
		Enabled:     true,
		BaseURL:     server.URL,
		AppKey:      "test_app_key",
		AppSecret:   "test_app_secret",
		LicenseKey:  "test-license-key",
		Domain:      "test.example.com",
		ServerIP:    "127.0.0.1",
		CacheFile:   filepath.Join(t.TempDir(), "test_license.cache"),
		Interval:    3600,
		GracePeriod: 86400,
	}

	svc := NewLicenseService(cfg)

	t.Run("quick verify with valid license", func(t *testing.T) {
		result, err := svc.QuickVerify("test-license-key", "test.example.com", "127.0.0.1")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.Verified)
		assert.Equal(t, "TestApp", result.AppName)
		assert.Equal(t, "2099-12-31", result.ExpireAt)
		assert.Equal(t, "test-license-key", result.LicenseKey)
	})

	t.Run("isVerified and gracePeriod on original svc are unaffected", func(t *testing.T) {
		assert.False(t, svc.IsVerified())
		assert.False(t, svc.IsGracePeriodValid())
	})
}

func TestLicenseService_QuickVerifyFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := verifyResponse{}
		resp.Code = 403
		resp.Msg = "license expired"
		resp.Data.Result = "fail"
		resp.Data.Reason = "License expired"

		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &model.LicenseConfig{
		Enabled:     true,
		BaseURL:     server.URL,
		AppKey:      "test_app_key",
		AppSecret:   "test_app_secret",
		LicenseKey:  "bad-license",
		Domain:      "test.example.com",
		ServerIP:    "127.0.0.1",
		CacheFile:   filepath.Join(t.TempDir(), "test_license.cache"),
		Interval:    3600,
		GracePeriod: 86400,
	}

	svc := NewLicenseService(cfg)

	t.Run("quick verify with invalid license", func(t *testing.T) {
		result, err := svc.QuickVerify("bad-license", "test.example.com", "127.0.0.1")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.Verified)
	})

	t.Run("isVerified remains false after failure", func(t *testing.T) {
		assert.False(t, svc.IsVerified())
	})
}

func TestLicenseService_QuickVerifyNetworkError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	server.Close()

	cfg := &model.LicenseConfig{
		Enabled:     true,
		BaseURL:     server.URL,
		AppKey:      "test_app_key",
		AppSecret:   "test_app_secret",
		LicenseKey:  "license-key",
		Domain:      "test.example.com",
		ServerIP:    "127.0.0.1",
		CacheFile:   filepath.Join(t.TempDir(), "test_license.cache"),
		Interval:    3600,
		GracePeriod: 86400,
	}

	svc := NewLicenseService(cfg)

	t.Run("quick verify with network error uses grace period", func(t *testing.T) {
		result, err := svc.QuickVerify("license-key", "test.example.com", "127.0.0.1")
		require.NoError(t, err)
		require.NotNil(t, result)
	})
}

func TestLicenseService_IsVerified(t *testing.T) {
	cfg := &model.LicenseConfig{
		Enabled:     false,
		BaseURL:     "https://example.com",
		AppKey:      "key",
		AppSecret:   "secret",
		CacheFile:   filepath.Join(t.TempDir(), "test_license.cache"),
		Interval:    3600,
		GracePeriod: 86400,
	}

	svc := NewLicenseService(cfg)

	t.Run("initially not verified", func(t *testing.T) {
		assert.False(t, svc.IsVerified())
	})

	t.Run("verify sets verified to true when disabled", func(t *testing.T) {
		ok, err := svc.Verify()
		require.NoError(t, err)
		assert.True(t, ok)
		assert.True(t, svc.IsVerified())
	})
}

func TestLicenseService_IsGracePeriodValid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := verifyResponse{}
		resp.Code = 200
		resp.Msg = "ok"
		resp.Data.Result = "pass"
		resp.Data.AppName = "TestApp"
		resp.Data.ExpireAt = "2099-12-31"
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &model.LicenseConfig{
		Enabled:     true,
		BaseURL:     server.URL,
		AppKey:      "key",
		AppSecret:   "secret",
		LicenseKey:  "license-key",
		Domain:      "test.example.com",
		ServerIP:    "127.0.0.1",
		CacheFile:   filepath.Join(t.TempDir(), "test_license.cache"),
		Interval:    3600,
		GracePeriod: 86400,
	}

	svc := NewLicenseService(cfg)

	t.Run("grace period invalid when no last success", func(t *testing.T) {
		cache := svc.GetCache()
		assert.Equal(t, int64(0), cache.LastSuccessTime)
		assert.False(t, svc.IsGracePeriodValid())
	})

	t.Run("grace period valid after successful verify", func(t *testing.T) {
		_, err := svc.Verify()
		require.NoError(t, err)
		assert.True(t, svc.IsGracePeriodValid())
	})
}

func TestLicenseService_GetCache(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := verifyResponse{}
		resp.Code = 200
		resp.Msg = "ok"
		resp.Data.Result = "pass"
		resp.Data.AppName = "TestApp"
		resp.Data.ExpireAt = "2099-12-31"
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := &model.LicenseConfig{
		Enabled:     true,
		BaseURL:     server.URL,
		AppKey:      "key",
		AppSecret:   "secret",
		LicenseKey:  "license-key",
		Domain:      "test.example.com",
		ServerIP:    "127.0.0.1",
		CacheFile:   filepath.Join(t.TempDir(), "test_license.cache"),
		Interval:    3600,
		GracePeriod: 86400,
	}

	svc := NewLicenseService(cfg)

	t.Run("get cache after verify", func(t *testing.T) {
		svc.Verify()
		cache := svc.GetCache()
		require.NotNil(t, cache)
		assert.True(t, cache.LastVerifyTime > 0)
		assert.True(t, cache.LastSuccessTime > 0)
		assert.Equal(t, "pass", cache.LastResult)
	})
}

func TestLicenseService_LoadAndSaveCache(t *testing.T) {
	cacheFile := filepath.Join(t.TempDir(), "test_save_cache.cache")

	t.Run("load cache from nonexistent file", func(t *testing.T) {
		cfg := &model.LicenseConfig{
			Enabled:     false,
			BaseURL:     "https://example.com",
			AppKey:      "key",
			AppSecret:   "secret",
			CacheFile:   cacheFile,
			Interval:    3600,
			GracePeriod: 86400,
		}
		svc := NewLicenseService(cfg)
		cache := svc.GetCache()
		assert.NotNil(t, cache)
	})

	t.Run("save and reload cache via disabled verify", func(t *testing.T) {
		cfg := &model.LicenseConfig{
			Enabled:     false,
			BaseURL:     "https://example.com",
			AppKey:      "key",
			AppSecret:   "secret",
			CacheFile:   cacheFile,
			Interval:    3600,
			GracePeriod: 86400,
		}
		svc := NewLicenseService(cfg)
		svc.Verify()
		svc.handleError("test error")

		cache1 := svc.GetCache()
		require.True(t, cache1.LastVerifyTime > 0)

		cfg2 := &model.LicenseConfig{
			Enabled:     false,
			BaseURL:     "https://example.com",
			AppKey:      "key",
			AppSecret:   "secret",
			CacheFile:   cacheFile,
			Interval:    3600,
			GracePeriod: 86400,
		}
		svc2 := NewLicenseService(cfg2)
		cache2 := svc2.GetCache()
		assert.Equal(t, cache1.LastVerifyTime, cache2.LastVerifyTime)
		assert.Equal(t, cache1.LastMessage, cache2.LastMessage)
	})

	t.Run("load cache from valid JSON file", func(t *testing.T) {
		cache := &LicenseCache{
			LastVerifyTime:  time.Now().Unix(),
			LastSuccessTime: time.Now().Unix(),
			LastResult:      "pass",
			LastMessage:     "ok",
			ExpireAt:        "2099-12-31",
			AppName:         "TestApp",
			UsedServer:      "https://example.com",
		}
		data, _ := json.Marshal(cache)
		os.WriteFile(cacheFile, data, 0644)

		cfg := &model.LicenseConfig{
			Enabled:     false,
			BaseURL:     "https://example.com",
			AppKey:      "key",
			AppSecret:   "secret",
			CacheFile:   cacheFile,
			Interval:    3600,
			GracePeriod: 86400,
		}
		svc := NewLicenseService(cfg)
		loaded := svc.GetCache()
		assert.Equal(t, cache.LastResult, loaded.LastResult)
		assert.Equal(t, cache.AppName, loaded.AppName)
	})
}

func TestLicenseService_HandleError(t *testing.T) {
	cfg := &model.LicenseConfig{
		Enabled:     false,
		BaseURL:     "https://example.com",
		AppKey:      "key",
		AppSecret:   "secret",
		CacheFile:   filepath.Join(t.TempDir(), "test_handle_error.cache"),
		Interval:    3600,
		GracePeriod: 86400,
	}

	svc := NewLicenseService(cfg)
	svc.handleError("test error message")

	cache := svc.GetCache()
	assert.True(t, cache.LastVerifyTime > 0)
	assert.Equal(t, "test error message", cache.LastMessage)
}

func TestLicenseService_VerifyDisabled(t *testing.T) {
	cfg := &model.LicenseConfig{
		Enabled:     false,
		BaseURL:     "https://example.com",
		AppKey:      "key",
		AppSecret:   "secret",
		CacheFile:   filepath.Join(t.TempDir(), "test_verify_disabled.cache"),
		Interval:    3600,
		GracePeriod: 86400,
	}

	svc := NewLicenseService(cfg)
	ok, err := svc.Verify()
	require.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, svc.IsVerified())
}

func TestLicenseService_DefaultConfig(t *testing.T) {
	cfg := &model.LicenseConfig{}
	svc := NewLicenseService(cfg)

	assert.NotNil(t, svc.cfg)
	assert.NotEmpty(t, svc.cfg.BaseURL)
	assert.NotEmpty(t, svc.cfg.AppKey)
	assert.NotEmpty(t, svc.cfg.AppSecret)
	assert.True(t, svc.cfg.Interval > 0)
	assert.True(t, svc.cfg.GracePeriod > 0)
	assert.NotEmpty(t, svc.cfg.CacheFile)
}