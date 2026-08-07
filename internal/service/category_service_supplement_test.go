package service

import (
	"testing"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func dropAllCategorySupplementTables(db *gorm.DB) {
	tables := []string{"users", "categories", "products", "cards", "orders", "payment_channels", "email_configs", "nodes", "operation_logs", "order_logs", "email_logs", "upgrade_logs", "file_uploads"}
	for _, tb := range tables {
		db.Migrator().DropTable(tb)
	}
}

func setupCategorySupplementTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_fk=1"), &gorm.Config{})
	require.NoError(t, err)
	dropAllCategorySupplementTables(db)
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

func TestCategoryService_CreateDBNil(t *testing.T) {
	originalDB := database.DB
	database.DB = nil
	defer func() { database.DB = originalDB }()

	svc := NewCategoryService()
	cat, err := svc.Create("测试", "", 1)
	assert.Error(t, err)
	assert.Nil(t, cat)
	assert.Contains(t, err.Error(), "数据库未连接")
}

func TestCategoryService_UpdateDBNil(t *testing.T) {
	originalDB := database.DB
	database.DB = nil
	defer func() { database.DB = originalDB }()

	svc := NewCategoryService()
	cat, err := svc.Update(1, "test", "", 1, 1)
	assert.Error(t, err)
	assert.Nil(t, cat)
	assert.Contains(t, err.Error(), "数据库未连接")
}

func TestCategoryService_DeleteDBNil(t *testing.T) {
	originalDB := database.DB
	database.DB = nil
	defer func() { database.DB = originalDB }()

	svc := NewCategoryService()
	err := svc.Delete(1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "数据库未连接")
}

func TestCategoryService_GetAllDBNil(t *testing.T) {
	originalDB := database.DB
	database.DB = nil
	defer func() { database.DB = originalDB }()

	svc := NewCategoryService()
	cats, err := svc.GetAll()
	require.NoError(t, err)
	assert.NotNil(t, cats)
	assert.Len(t, cats, 0)
}

func TestCategoryService_ListDBNil(t *testing.T) {
	originalDB := database.DB
	database.DB = nil
	defer func() { database.DB = originalDB }()

	svc := NewCategoryService()
	result, err := svc.List(1, 10, "")
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(0), result.Total)
	assert.Len(t, result.Items, 0)
}

func TestCategoryService_DeleteNonexistent(t *testing.T) {
	setupCategorySupplementTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCategoryService()
	err := svc.Delete(9999)
	assert.NoError(t, err)
}

func TestCategoryService_CreateWithZeroSort(t *testing.T) {
	setupCategorySupplementTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCategoryService()
	cat, err := svc.Create("零排序分类", "", 0)
	require.NoError(t, err)
	assert.Equal(t, 0, cat.Sort)
	assert.Equal(t, model.CategoryEnabled, cat.Status)
}

func TestCategoryService_GetAllExcludesDisabled(t *testing.T) {
	setupCategorySupplementTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCategoryService()

	_, err := svc.Create("启用分类", "", 1)
	require.NoError(t, err)
	disabled, err := svc.Create("禁用分类", "", 2)
	require.NoError(t, err)

	_, err = svc.Update(disabled.ID, "禁用分类", "", 2, model.CategoryDisabled)
	require.NoError(t, err)

	cats, err := svc.GetAll()
	require.NoError(t, err)
	assert.Len(t, cats, 1)
	assert.Equal(t, "启用分类", cats[0].Name)
}

func TestCategoryService_ListWithComplexFilter(t *testing.T) {
	setupCategorySupplementTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCategoryService()

	_, err := svc.Create("电子设备", "", 1)
	require.NoError(t, err)
	_, err = svc.Create("服装鞋帽", "", 2)
	require.NoError(t, err)
	_, err = svc.Create("电子产品", "", 3)
	require.NoError(t, err)

	result, err := svc.List(1, 10, "电子")
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.Total)
	assert.Len(t, result.Items, 2)
}

func TestCategoryService_UpdateSortAndIcon(t *testing.T) {
	setupCategorySupplementTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCategoryService()

	cat, err := svc.Create("原分类", "icon1", 5)
	require.NoError(t, err)

	updated, err := svc.Update(cat.ID, "原分类", "icon2", 10, model.CategoryEnabled)
	require.NoError(t, err)
	assert.Equal(t, "icon2", updated.Icon)
	assert.Equal(t, 10, updated.Sort)
}

func TestCategoryService_DeleteAndRecreate(t *testing.T) {
	setupCategorySupplementTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCategoryService()

	cat, err := svc.Create("待删分类", "", 1)
	require.NoError(t, err)
	require.NoError(t, svc.Delete(cat.ID))

	_, err = svc.Create("新分类", "", 2)
	require.NoError(t, err)

	result, err := svc.List(1, 10, "")
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)
	assert.Equal(t, "新分类", result.Items[0].Name)
}