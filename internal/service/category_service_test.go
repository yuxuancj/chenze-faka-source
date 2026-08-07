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

func setupCategoryTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	tables := []string{"users", "categories", "products", "cards", "orders", "payment_channels", "email_configs", "nodes", "operation_logs", "order_logs", "email_logs", "upgrade_logs", "file_uploads"}
	for _, tb := range tables {
		db.Migrator().DropTable(tb)
	}

	database.DB = db
	require.NoError(t, db.AutoMigrate(
		&model.User{}, &model.Category{}, &model.Product{},
		&model.Card{}, &model.Order{}, &model.PaymentChannel{},
		&model.EmailConfig{}, &model.Node{},
		&model.OperationLog{}, &model.OrderLog{}, &model.EmailLog{},
		&model.UpgradeLog{}, &model.FileUpload{},
	))
}

func TestCategoryService_Create(t *testing.T) {
	setupCategoryTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCategoryService()

	t.Run("create category successfully", func(t *testing.T) {
		cat, err := svc.Create("测试分类", "📁", 1)
		require.NoError(t, err)
		assert.NotNil(t, cat)
		assert.Equal(t, "测试分类", cat.Name)
		assert.Equal(t, "📁", cat.Icon)
		assert.Equal(t, 1, cat.Sort)
		assert.Equal(t, model.CategoryEnabled, cat.Status)
	})

	t.Run("create category with empty name fails", func(t *testing.T) {
		cat, err := svc.Create("", "📁", 1)
		assert.Error(t, err)
		assert.Nil(t, cat)
		assert.Contains(t, err.Error(), "分类名称不能为空")
	})

	t.Run("create category with empty icon succeeds", func(t *testing.T) {
		cat, err := svc.Create("无图标分类", "", 2)
		require.NoError(t, err)
		assert.NotNil(t, cat)
		assert.Equal(t, "", cat.Icon)
	})
}

func TestCategoryService_Update(t *testing.T) {
	setupCategoryTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCategoryService()

	cat, err := svc.Create("原始分类", "📁", 1)
	require.NoError(t, err)

	t.Run("update category name and sort", func(t *testing.T) {
		updated, err := svc.Update(cat.ID, "更新后分类", "📂", 5, model.CategoryEnabled)
		require.NoError(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, "更新后分类", updated.Name)
		assert.Equal(t, "📂", updated.Icon)
		assert.Equal(t, 5, updated.Sort)
	})

	t.Run("update category with empty name keeps old name", func(t *testing.T) {
		updated, err := svc.Update(cat.ID, "", "", 3, model.CategoryEnabled)
		require.NoError(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, "更新后分类", updated.Name)
	})

	t.Run("update nonexistent category", func(t *testing.T) {
		updated, err := svc.Update(9999, "不存在", "", 1, model.CategoryEnabled)
		assert.Error(t, err)
		assert.Nil(t, updated)
		assert.Contains(t, err.Error(), "分类不存在")
	})

	t.Run("update category status to disabled", func(t *testing.T) {
		updated, err := svc.Update(cat.ID, "禁用测试", "🚫", 1, model.CategoryDisabled)
		require.NoError(t, err)
		assert.Equal(t, model.CategoryDisabled, updated.Status)
	})
}

func TestCategoryService_Delete(t *testing.T) {
	setupCategoryTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCategoryService()

	cat, err := svc.Create("待删除分类", "🗑", 1)
	require.NoError(t, err)

	err = svc.Delete(cat.ID)
	require.NoError(t, err)

	_, err = svc.Update(cat.ID, "test", "", 1, model.CategoryEnabled)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "分类不存在")
}

func TestCategoryService_GetAll(t *testing.T) {
	setupCategoryTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCategoryService()

	_, err := svc.Create("分类C", "📑", 3)
	require.NoError(t, err)
	_, err = svc.Create("分类A", "📁", 1)
	require.NoError(t, err)
	_, err = svc.Create("分类B", "📂", 2)
	require.NoError(t, err)

	cats, err := svc.GetAll()
	require.NoError(t, err)
	assert.Len(t, cats, 3)
	assert.Equal(t, "分类A", cats[0].Name)
	assert.Equal(t, "分类B", cats[1].Name)
	assert.Equal(t, "分类C", cats[2].Name)
}

func TestCategoryService_List(t *testing.T) {
	setupCategoryTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCategoryService()

	_, err := svc.Create("列表分类A", "📁", 1)
	require.NoError(t, err)
	_, err = svc.Create("列表分类B", "📂", 2)
	require.NoError(t, err)
	_, err = svc.Create("列表分类C", "📃", 3)
	require.NoError(t, err)

	t.Run("list with pagination", func(t *testing.T) {
		result, err := svc.List(1, 2, "")
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		assert.Len(t, result.Items, 2)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 2, result.PageSize)
	})

	t.Run("list second page", func(t *testing.T) {
		result, err := svc.List(2, 2, "")
		require.NoError(t, err)
		assert.Len(t, result.Items, 1)
	})

	t.Run("list with keyword search", func(t *testing.T) {
		result, err := svc.List(1, 10, "分类A")
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
		assert.Len(t, result.Items, 1)
	})

	t.Run("list with empty keyword returns all", func(t *testing.T) {
		result, err := svc.List(1, 10, "")
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
	})

	t.Run("list with invalid page defaults to 1", func(t *testing.T) {
		result, err := svc.List(0, 10, "")
		require.NoError(t, err)
		assert.Equal(t, 1, result.Page)
	})

	t.Run("list with zero pageSize defaults to 10", func(t *testing.T) {
		result, err := svc.List(1, 0, "")
		require.NoError(t, err)
		assert.Equal(t, 10, result.PageSize)
	})
}

func TestCategoryService_UpdateDefaultStatus(t *testing.T) {
	setupCategoryTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCategoryService()

	cat, err := svc.Create("状态测试", "📁", 1)
	require.NoError(t, err)
	assert.Equal(t, model.CategoryEnabled, cat.Status)

	updated, err := svc.Update(cat.ID, "状态测试", "", 1, model.CategoryDisabled)
	require.NoError(t, err)
	assert.Equal(t, model.CategoryDisabled, updated.Status)
}
