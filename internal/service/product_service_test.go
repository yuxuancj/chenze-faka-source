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

func dropAllProductTables(db *gorm.DB) {
	tables := []string{"users", "categories", "products", "cards", "orders", "payment_channels", "email_configs", "nodes", "operation_logs", "order_logs", "email_logs", "upgrade_logs", "file_uploads"}
	for _, t := range tables {
		db.Migrator().DropTable(t)
	}
}

func setupProductTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	dropAllProductTables(db)
	database.DB = db
	require.NoError(t, db.AutoMigrate(
		&model.User{}, &model.Category{}, &model.Product{},
		&model.Card{}, &model.Order{}, &model.PaymentChannel{},
		&model.EmailConfig{}, &model.Node{},
		&model.OperationLog{}, &model.OrderLog{}, &model.EmailLog{},
		&model.UpgradeLog{}, &model.FileUpload{},
	))
}

func TestProductService_Create(t *testing.T) {
	setupProductTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	t.Run("create product with valid data", func(t *testing.T) {
		req := &CreateProductRequest{
			Name:        "测试产品",
			Category:    "虚拟币",
			Price:       10.00,
			Description: "测试描述",
			Image:       "/img/test.png",
			Stock:       100,
			Status:      0,
			Sort:        1,
		}

		product, err := svc.Create(req)
		require.NoError(t, err)
		assert.NotNil(t, product)
		assert.Equal(t, "测试产品", product.Name)
		assert.Equal(t, "虚拟币", product.Category)
		assert.Equal(t, 10.00, product.Price)
		assert.Equal(t, 100, product.Stock)
		assert.Equal(t, model.ProductStatusOnShelf, product.Status)
	})

	t.Run("create product with empty name", func(t *testing.T) {
		req := &CreateProductRequest{
			Name:  "",
			Price: 10.00,
		}

		product, err := svc.Create(req)
		assert.Error(t, err)
		assert.Nil(t, product)
		assert.Contains(t, err.Error(), "产品名称不能为空")
	})

	t.Run("create product with zero price", func(t *testing.T) {
		req := &CreateProductRequest{
			Name:  "免费产品",
			Price: 0,
		}

		product, err := svc.Create(req)
		assert.Error(t, err)
		assert.Nil(t, product)
		assert.Contains(t, err.Error(), "产品价格必须大于0")
	})

	t.Run("create product with negative price", func(t *testing.T) {
		req := &CreateProductRequest{
			Name:  "负价产品",
			Price: -5.00,
		}

		product, err := svc.Create(req)
		assert.Error(t, err)
		assert.Nil(t, product)
		assert.Contains(t, err.Error(), "产品价格必须大于0")
	})

	t.Run("create product with off-shelf status gets converted to on-shelf", func(t *testing.T) {
		req := &CreateProductRequest{
			Name:   "下架产品",
			Price:  10.00,
			Status: model.ProductStatusOffShelf,
		}

		product, err := svc.Create(req)
		require.NoError(t, err)
		assert.Equal(t, model.ProductStatusOnShelf, product.Status)
	})
}

func TestProductService_Update(t *testing.T) {
	setupProductTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	createReq := &CreateProductRequest{
		Name:     "原始产品",
		Category: "分类A",
		Price:    10.00,
		Stock:    50,
		Status:   model.ProductStatusOnShelf,
		Sort:     1,
	}
	product, err := svc.Create(createReq)
	require.NoError(t, err)

	t.Run("update product name and price", func(t *testing.T) {
		updateReq := &UpdateProductRequest{
			ID:    product.ID,
			Name:  "更新后产品",
			Price: 25.00,
		}

		updated, err := svc.Update(updateReq)
		require.NoError(t, err)
		assert.Equal(t, "更新后产品", updated.Name)
		assert.Equal(t, 25.00, updated.Price)
	})

	t.Run("update product category and sort", func(t *testing.T) {
		updateReq := &UpdateProductRequest{
			ID:       product.ID,
			Category: "新分类",
			Sort:     10,
		}

		updated, err := svc.Update(updateReq)
		require.NoError(t, err)
		assert.Equal(t, "新分类", updated.Category)
		assert.Equal(t, 10, updated.Sort)
	})

	t.Run("update product description and image", func(t *testing.T) {
		updateReq := &UpdateProductRequest{
			ID:          product.ID,
			Description: "新描述",
			Image:       "/img/new.png",
		}

		updated, err := svc.Update(updateReq)
		require.NoError(t, err)
		assert.Equal(t, "新描述", updated.Description)
		assert.Equal(t, "/img/new.png", updated.Image)
	})

	t.Run("update product status - zero value is ignored", func(t *testing.T) {
		updateReq := &UpdateProductRequest{
			ID:     product.ID,
			Status: model.ProductStatusOffShelf,
		}

		updated, err := svc.Update(updateReq)
		require.NoError(t, err)
		assert.Equal(t, model.ProductStatusOnShelf, updated.Status)
	})

	t.Run("update nonexistent product", func(t *testing.T) {
		updateReq := &UpdateProductRequest{
			ID:   9999,
			Name: "不存在的产品",
		}

		updated, err := svc.Update(updateReq)
		assert.Error(t, err)
		assert.Nil(t, updated)
		assert.Contains(t, err.Error(), "产品不存在")
	})
}

func TestProductService_Delete(t *testing.T) {
	setupProductTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	createReq := &CreateProductRequest{
		Name:  "待删除产品",
		Price: 10.00,
		Stock: 50,
	}
	product, err := svc.Create(createReq)
	require.NoError(t, err)

	err = svc.Delete(product.ID)
	require.NoError(t, err)

	_, err = svc.GetByID(product.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "产品不存在")
}

func TestProductService_GetByID(t *testing.T) {
	setupProductTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	createReq := &CreateProductRequest{
		Name:  "查询产品",
		Price: 15.00,
		Stock: 30,
	}
	product, err := svc.Create(createReq)
	require.NoError(t, err)

	t.Run("get product by valid id", func(t *testing.T) {
		found, err := svc.GetByID(product.ID)
		require.NoError(t, err)
		assert.Equal(t, product.Name, found.Name)
		assert.Equal(t, product.Price, found.Price)
	})

	t.Run("get product by invalid id", func(t *testing.T) {
		found, err := svc.GetByID(9999)
		assert.Error(t, err)
		assert.Nil(t, found)
		assert.Contains(t, err.Error(), "产品不存在")
	})
}

func TestProductService_List(t *testing.T) {
	setupProductTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	products := []*CreateProductRequest{
		{Name: "产品Alpha", Price: 10.00, Stock: 100, Sort: 1},
		{Name: "产品Beta", Price: 20.00, Stock: 50, Sort: 2},
		{Name: "产品Gamma", Price: 30.00, Stock: 30, Sort: 3},
		{Name: "产品Delta", Price: 40.00, Stock: 20, Sort: 4},
		{Name: "产品Epsilon", Price: 50.00, Stock: 10, Sort: 5},
	}
	for _, req := range products {
		_, err := svc.Create(req)
		require.NoError(t, err)
	}

	t.Run("list with default params", func(t *testing.T) {
		result, err := svc.List(0, 0, "")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(5), result.Total)
		assert.Len(t, result.Products, 5)
	})

	t.Run("list with pagination", func(t *testing.T) {
		result, err := svc.List(1, 3, "")
		require.NoError(t, err)
		assert.Len(t, result.Products, 3)

		result2, err := svc.List(2, 3, "")
		require.NoError(t, err)
		assert.Len(t, result2.Products, 2)
	})

	t.Run("list with keyword filter", func(t *testing.T) {
		result, err := svc.List(1, 10, "Alpha")
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
		assert.Len(t, result.Products, 1)
		assert.Equal(t, "产品Alpha", result.Products[0].Name)
	})

	t.Run("list with empty keyword returns all", func(t *testing.T) {
		result, err := svc.List(1, 10, "")
		require.NoError(t, err)
		assert.Equal(t, int64(5), result.Total)
	})
}

func TestProductService_ListOnShelf(t *testing.T) {
	setupProductTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	_, err := svc.Create(&CreateProductRequest{Name: "上架产品A", Price: 10.00, Stock: 100})
	require.NoError(t, err)
	_, err = svc.Create(&CreateProductRequest{Name: "上架产品B", Price: 20.00, Stock: 50})
	require.NoError(t, err)
	offShelf, err := svc.Create(&CreateProductRequest{Name: "下架产品", Price: 30.00, Stock: 30})
	require.NoError(t, err)
	require.NoError(t, database.DB.Model(&model.Product{}).Where("id = ?", offShelf.ID).Update("status", model.ProductStatusOffShelf).Error)

	t.Run("list on-shelf products only", func(t *testing.T) {
		products, err := svc.ListOnShelf()
		require.NoError(t, err)
		assert.Len(t, products, 2)
		for _, p := range products {
			assert.Equal(t, model.ProductStatusOnShelf, p.Status)
		}
	})
}

func TestProductService_ListOnShelfGrouped(t *testing.T) {
	setupProductTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	_, err := svc.Create(&CreateProductRequest{Name: "QQ币10个", Category: "虚拟币", Price: 10.00, Stock: 100, Sort: 1})
	require.NoError(t, err)
	_, err = svc.Create(&CreateProductRequest{Name: "QQ币50个", Category: "虚拟币", Price: 50.00, Stock: 50, Sort: 2})
	require.NoError(t, err)
	_, err = svc.Create(&CreateProductRequest{Name: "爱奇艺月卡", Category: "影视会员", Price: 15.00, Stock: 30, Sort: 3})
	require.NoError(t, err)
	offShelf, err := svc.Create(&CreateProductRequest{Name: "下架产品", Category: "虚拟币", Price: 5.00, Stock: 10, Sort: 4})
	require.NoError(t, err)
	require.NoError(t, database.DB.Model(&model.Product{}).Where("id = ?", offShelf.ID).Update("status", model.ProductStatusOffShelf).Error)

	t.Run("list on-shelf grouped by category", func(t *testing.T) {
		groups, err := svc.ListOnShelfGrouped()
		require.NoError(t, err)
		assert.Len(t, groups, 2)

		catMap := make(map[string]int)
		for _, g := range groups {
			catMap[g.Category] = len(g.Products)
		}
		assert.Equal(t, 2, catMap["虚拟币"])
		assert.Equal(t, 1, catMap["影视会员"])
	})

	t.Run("products without category go to 其他", func(t *testing.T) {
		_, err := svc.Create(&CreateProductRequest{Name: "无分类产品", Price: 5.00, Stock: 10, Status: model.ProductStatusOnShelf})
		require.NoError(t, err)

		groups, err := svc.ListOnShelfGrouped()
		require.NoError(t, err)

		found := false
		for _, g := range groups {
			if g.Category == "其他" {
				found = true
				assert.Len(t, g.Products, 1)
			}
		}
		assert.True(t, found)
	})
}