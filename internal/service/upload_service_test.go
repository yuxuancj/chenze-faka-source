package service

import (
	"os"
	"path/filepath"
	"testing"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func dropAllUploadTables(db *gorm.DB) {
	tables := []string{"users", "categories", "products", "cards", "orders", "payment_channels", "email_configs", "nodes", "operation_logs", "order_logs", "email_logs", "upgrade_logs", "file_uploads"}
	for _, tb := range tables {
		db.Migrator().DropTable(tb)
	}
}

func setupUploadTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_fk=1"), &gorm.Config{})
	require.NoError(t, err)
	dropAllUploadTables(db)
	database.DB = db
	require.NoError(t, db.AutoMigrate(
		&model.User{}, &model.Category{}, &model.Product{},
		&model.Card{}, &model.Order{}, &model.PaymentChannel{},
		&model.EmailConfig{}, &model.Node{},
		&model.OperationLog{}, &model.OrderLog{}, &model.EmailLog{},
		&model.UpgradeLog{}, &model.FileUpload{},
	))
	return db
}

func cleanupUploadDir(t *testing.T) {
	t.Helper()
	os.RemoveAll(uploadDir)
}

func TestUploadService_SaveFile(t *testing.T) {
	setupUploadTestDB(t)
	defer func() { database.DB = nil }()
	defer cleanupUploadDir(t)

	svc := NewUploadService()

	t.Run("save image file successfully", func(t *testing.T) {
		data := []byte("\x89PNG\r\n\x1a\n" + string(make([]byte, 100)))
		result, err := svc.SaveFile("test.png", data, "image/png", 1)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test.png", result.OriginalName)
		assert.Equal(t, int64(len(data)), result.Size)
		assert.Equal(t, "image/png", result.MimeType)
		assert.Equal(t, model.FileTypeImage, result.Type)
		assert.NotEmpty(t, result.StoredName)
		assert.NotEmpty(t, result.Path)
		assert.FileExists(t, result.Path)
	})

	t.Run("save document file", func(t *testing.T) {
		data := []byte("Hello World")
		result, err := svc.SaveFile("readme.txt", data, "text/plain", 1)
		require.NoError(t, err)
		assert.Equal(t, model.FileTypeFile, result.Type)
	})

	t.Run("save file with unsupported type", func(t *testing.T) {
		data := []byte("test")
		result, err := svc.SaveFile("test.xyz", data, "application/octet-stream", 1)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "不支持的文件类型")
	})

	t.Run("save file exceeding size limit", func(t *testing.T) {
		data := make([]byte, maxFileLen+1)
		result, err := svc.SaveFile("big.jpg", data, "image/jpeg", 1)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "文件大小超出限制")
	})
}

func TestUploadService_GetFile(t *testing.T) {
	setupUploadTestDB(t)
	defer func() { database.DB = nil }()
	defer cleanupUploadDir(t)

	svc := NewUploadService()

	saved, err := svc.SaveFile("test.png", []byte("data"), "image/png", 1)
	require.NoError(t, err)

	t.Run("get existing file", func(t *testing.T) {
		result, err := svc.GetFile(saved.ID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, saved.ID, result.ID)
		assert.Equal(t, saved.OriginalName, result.OriginalName)
	})

	t.Run("get nonexistent file", func(t *testing.T) {
		result, err := svc.GetFile(9999)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "文件不存在")
	})
}

func TestUploadService_DeleteFile(t *testing.T) {
	setupUploadTestDB(t)
	defer func() { database.DB = nil }()
	defer cleanupUploadDir(t)

	svc := NewUploadService()

	saved, err := svc.SaveFile("test.png", []byte("data"), "image/png", 1)
	require.NoError(t, err)

	t.Run("delete existing file", func(t *testing.T) {
		path := saved.Path
		assert.FileExists(t, path)

		err := svc.DeleteFile(saved.ID)
		require.NoError(t, err)

		assert.NoFileExists(t, path)

		_, err = svc.GetFile(saved.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "文件不存在")
	})

	t.Run("delete nonexistent file", func(t *testing.T) {
		err := svc.DeleteFile(9999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "文件不存在")
	})
}

func TestUploadService_ListFiles(t *testing.T) {
	setupUploadTestDB(t)
	defer func() { database.DB = nil }()
	defer cleanupUploadDir(t)

	svc := NewUploadService()

	_, err := svc.SaveFile("img1.png", []byte("img"), "image/png", 1)
	require.NoError(t, err)
	_, err = svc.SaveFile("doc1.txt", []byte("doc"), "text/plain", 1)
	require.NoError(t, err)
	_, err = svc.SaveFile("img2.jpg", []byte("img2"), "image/jpeg", 1)
	require.NoError(t, err)

	t.Run("list all files", func(t *testing.T) {
		result, err := svc.ListFiles(1, 10, "")
		require.NoError(t, err)
		assert.Equal(t, int64(3), result["total"])
	})

	t.Run("list with file type filter", func(t *testing.T) {
		result, err := svc.ListFiles(1, 10, model.FileTypeImage)
		require.NoError(t, err)
		assert.Equal(t, int64(2), result["total"])
	})

	t.Run("list documents only", func(t *testing.T) {
		result, err := svc.ListFiles(1, 10, model.FileTypeFile)
		require.NoError(t, err)
		assert.Equal(t, int64(1), result["total"])
	})

	t.Run("list with pagination", func(t *testing.T) {
		result, err := svc.ListFiles(1, 2, "")
		require.NoError(t, err)
		assert.Equal(t, int64(3), result["total"])
		files := result["files"].([]model.FileUpload)
		assert.Len(t, files, 2)
	})
}

func TestUploadService_DetermineFileType(t *testing.T) {
	svc := NewUploadService()

	t.Run("image extensions", func(t *testing.T) {
		assert.Equal(t, model.FileTypeImage, svc.determineFileType(".jpg"))
		assert.Equal(t, model.FileTypeImage, svc.determineFileType(".jpeg"))
		assert.Equal(t, model.FileTypeImage, svc.determineFileType(".png"))
		assert.Equal(t, model.FileTypeImage, svc.determineFileType(".gif"))
		assert.Equal(t, model.FileTypeImage, svc.determineFileType(".webp"))
		assert.Equal(t, model.FileTypeImage, svc.determineFileType(".svg"))
	})

	t.Run("document extensions", func(t *testing.T) {
		assert.Equal(t, model.FileTypeFile, svc.determineFileType(".pdf"))
		assert.Equal(t, model.FileTypeFile, svc.determineFileType(".txt"))
	})

	t.Run("unsupported extensions", func(t *testing.T) {
		assert.Equal(t, "", svc.determineFileType(".exe"))
		assert.Equal(t, "", svc.determineFileType(".mp3"))
		assert.Equal(t, "", svc.determineFileType(".zip"))
		assert.Equal(t, "", svc.determineFileType(""))
	})
}

func TestUploadService_DBNil(t *testing.T) {
	originalDB := database.DB
	database.DB = nil
	defer func() { database.DB = originalDB }()

	svc := NewUploadService()

	t.Run("save file with nil DB", func(t *testing.T) {
		result, err := svc.SaveFile("test.png", []byte("data"), "image/png", 1)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "数据库未连接")
	})

	t.Run("get file with nil DB", func(t *testing.T) {
		result, err := svc.GetFile(1)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "数据库未连接")
	})

	t.Run("delete file with nil DB", func(t *testing.T) {
		err := svc.DeleteFile(1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "数据库未连接")
	})

	t.Run("list files with nil DB", func(t *testing.T) {
		result, err := svc.ListFiles(1, 10, "")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "数据库未连接")
	})
}

func TestUploadService_SaveFileWithFilePath(t *testing.T) {
	setupUploadTestDB(t)
	defer func() { database.DB = nil }()
	defer cleanupUploadDir(t)

	svc := NewUploadService()

	t.Run("saved file path is within upload dir", func(t *testing.T) {
		result, err := svc.SaveFile("photo.png", []byte("fake-png-data"), "image/png", 1)
		require.NoError(t, err)

		absUploadDir, _ := filepath.Abs(uploadDir)
		absResultDir, _ := filepath.Abs(filepath.Dir(result.Path))
		assert.Equal(t, absUploadDir, absResultDir)
	})
}