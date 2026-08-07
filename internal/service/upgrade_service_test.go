package service

import (
	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
	t.Helper()

	originalDB := database.DB
	t.Cleanup(func() {
		if database.DB != nil {
			if sqlDB, err := database.DB.DB(); err == nil {
				sqlDB.Close()
			}
		}
		database.DB = originalDB
	})

	tmpFile, err := os.CreateTemp("", "test-upgrade-*.db")
	require.NoError(t, err)
	tmpFile.Close()
	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})

	db, err := gorm.Open(sqlite.Open(tmpFile.Name()), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&model.UpgradeLog{})
	require.NoError(t, err)

	database.DB = db
}

func TestNewUpgradeService(t *testing.T) {
	svc := NewUpgradeService()
	require.NotNil(t, svc)
	assert.IsType(t, &UpgradeService{}, svc)
}

func TestGetVersion(t *testing.T) {
	svc := NewUpgradeService()
	result := svc.GetVersion()

	require.NotNil(t, result)
	assert.Equal(t, "1.0.0", result["version"])
	assert.Equal(t, "Chenze Faka", result["name"])
	assert.Equal(t, "自动发卡系统", result["description"])
	assert.Equal(t, "https://example.com/api/upgrade/check", result["check_url"])
}

func TestCheckUpdate(t *testing.T) {
	svc := NewUpgradeService()
	result := svc.CheckUpdate()

	require.NotNil(t, result)
	assert.Equal(t, false, result["has_update"])
	assert.Equal(t, "1.0.0", result["current"])
	assert.Equal(t, "1.0.0", result["latest"])
	assert.Equal(t, "", result["download_url"])
	assert.Equal(t, "", result["release_notes"])
}

func TestUploadPackage_NoDB(t *testing.T) {
	originalDB := database.DB
	t.Cleanup(func() {
		database.DB = originalDB
	})
	database.DB = nil

	svc := NewUpgradeService()
	err := svc.UploadPackage("test.zip", []byte("data"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "数据库未连接")
}

func TestUploadPackage_EmptyFilename(t *testing.T) {
	setupTestDB(t)
	svc := NewUpgradeService()
	err := svc.UploadPackage("", []byte("data"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "文件名不能为空")
}

func TestUploadPackage_Success(t *testing.T) {
	setupTestDB(t)
	svc := NewUpgradeService()
	err := svc.UploadPackage("test-package.zip", []byte("package-content"))
	require.NoError(t, err)
}

func TestApplyUpgrade_NoDB(t *testing.T) {
	originalDB := database.DB
	t.Cleanup(func() {
		database.DB = originalDB
	})
	database.DB = nil

	svc := NewUpgradeService()
	err := svc.ApplyUpgrade()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "数据库未连接")
}

func TestApplyUpgrade_NoPending(t *testing.T) {
	setupTestDB(t)
	svc := NewUpgradeService()
	err := svc.ApplyUpgrade()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "没有待升级的包")
}

func TestApplyUpgrade_WithPending(t *testing.T) {
	setupTestDB(t)
	svc := NewUpgradeService()

	err := svc.UploadPackage("upgrade-v2.zip", []byte("data"))
	require.NoError(t, err)

	err = svc.ApplyUpgrade()
	require.NoError(t, err)
}

func TestListUpgradeLogs_NoDB(t *testing.T) {
	originalDB := database.DB
	t.Cleanup(func() {
		database.DB = originalDB
	})
	database.DB = nil

	svc := NewUpgradeService()
	result, err := svc.ListUpgradeLogs(1, 10)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "数据库未连接")
}

func TestListUpgradeLogs_Empty(t *testing.T) {
	setupTestDB(t)
	svc := NewUpgradeService()

	result, err := svc.ListUpgradeLogs(1, 10)
	require.NoError(t, err)
	require.NotNil(t, result)

	logs := result["logs"].([]model.UpgradeLog)
	assert.Empty(t, logs)
	assert.Equal(t, int64(0), result["total"])
	assert.Equal(t, 1, result["page"])
	assert.Equal(t, 10, result["page_size"])
}

func TestListUpgradeLogs_WithLogs(t *testing.T) {
	setupTestDB(t)
	svc := NewUpgradeService()

	err := svc.UploadPackage("pkg1.zip", []byte("aaa"))
	require.NoError(t, err)
	err = svc.UploadPackage("pkg2.zip", []byte("bbb"))
	require.NoError(t, err)

	result, err := svc.ListUpgradeLogs(1, 10)
	require.NoError(t, err)
	require.NotNil(t, result)

	logs := result["logs"].([]model.UpgradeLog)
	assert.Len(t, logs, 2)
	assert.Equal(t, int64(2), result["total"])
}

func TestListUpgradeLogs_DefaultPagination(t *testing.T) {
	setupTestDB(t)
	svc := NewUpgradeService()

	result, err := svc.ListUpgradeLogs(0, 0)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 1, result["page"])
	assert.Equal(t, 10, result["page_size"])
}

func TestListUpgradeLogs_CustomPagination(t *testing.T) {
	setupTestDB(t)
	svc := NewUpgradeService()

	for i := 0; i < 25; i++ {
		err := svc.UploadPackage("pkg.zip", []byte("data"))
		require.NoError(t, err)
	}

	result, err := svc.ListUpgradeLogs(2, 10)
	require.NoError(t, err)

	logs := result["logs"].([]model.UpgradeLog)
	assert.Len(t, logs, 10)
	assert.Equal(t, int64(25), result["total"])
	assert.Equal(t, 2, result["page"])
	assert.Equal(t, 10, result["page_size"])
}