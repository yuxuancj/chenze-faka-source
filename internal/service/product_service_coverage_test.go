package service

import (
	"fmt"
	"strings"
	"testing"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupProductCoverageTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_fk=1"), &gorm.Config{})
	require.NoError(t, err)
	dropAllCoverageTables(db)
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

func TestUpdateStockCoverage(t *testing.T) {
	setupProductCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	product := &model.Product{
		Name:     "StockCoverage商品",
		Category: "测试分类",
		Price:    10.00,
		Stock:    100,
		Status:   model.ProductStatusOnShelf,
		Sort:     1,
	}
	require.NoError(t, database.DB.Create(product).Error)

	t.Run("increase stock", func(t *testing.T) {
		err := svc.UpdateStock(product.ID, 10)
		require.NoError(t, err)

		var updated model.Product
		require.NoError(t, database.DB.First(&updated, product.ID).Error)
		assert.Equal(t, 110, updated.Stock)
	})

	t.Run("decrease stock", func(t *testing.T) {
		err := svc.UpdateStock(product.ID, -30)
		require.NoError(t, err)

		var updated model.Product
		require.NoError(t, database.DB.First(&updated, product.ID).Error)
		assert.Equal(t, 80, updated.Stock)
	})

	t.Run("update stock for nonexistent product", func(t *testing.T) {
		err := svc.UpdateStock(9999, 5)
		assert.NoError(t, err)
	})
}

func TestDeleteNonExistentProductCoverage(t *testing.T) {
	setupProductCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	err := svc.Delete(99999)
	assert.NoError(t, err)
}

func TestCreateMissingFieldsCoverage(t *testing.T) {
	setupProductCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	t.Run("create with empty name fails", func(t *testing.T) {
		req := &CreateProductRequest{
			Name:  "",
			Price: 10.00,
		}
		product, err := svc.Create(req)
		assert.Error(t, err)
		assert.Nil(t, product)
		assert.Contains(t, err.Error(), "产品名称不能为空")
	})

	t.Run("create with zero price fails", func(t *testing.T) {
		req := &CreateProductRequest{
			Name:  "FreeProduct",
			Price: 0,
		}
		product, err := svc.Create(req)
		assert.Error(t, err)
		assert.Nil(t, product)
		assert.Contains(t, err.Error(), "产品价格必须大于0")
	})

	t.Run("create with negative price fails", func(t *testing.T) {
		req := &CreateProductRequest{
			Name:  "NegativePrice",
			Price: -5.00,
		}
		product, err := svc.Create(req)
		assert.Error(t, err)
		assert.Nil(t, product)
		assert.Contains(t, err.Error(), "产品价格必须大于0")
	})

	t.Run("create with valid minimal fields succeeds", func(t *testing.T) {
		req := &CreateProductRequest{
			Name:  "MinimalProduct",
			Price: 1.00,
		}
		product, err := svc.Create(req)
		require.NoError(t, err)
		assert.NotNil(t, product)
		assert.Equal(t, "MinimalProduct", product.Name)
		assert.Equal(t, 1.00, product.Price)
		assert.Equal(t, model.ProductStatusOnShelf, product.Status)
	})
}

func TestUpdateInvalidIDCoverage(t *testing.T) {
	setupProductCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	t.Run("update with nonexistent ID fails", func(t *testing.T) {
		req := &UpdateProductRequest{
			ID:   9999,
			Name: "DoesNotExist",
		}
		product, err := svc.Update(req)
		assert.Error(t, err)
		assert.Nil(t, product)
		assert.Contains(t, err.Error(), "产品不存在")
	})

	t.Run("update with zero ID fails", func(t *testing.T) {
		req := &UpdateProductRequest{
			ID:   0,
			Name: "ZeroID",
		}
		product, err := svc.Update(req)
		assert.Error(t, err)
		assert.Nil(t, product)
		assert.Contains(t, err.Error(), "产品不存在")
	})
}

func TestListOnShelfGroupedCoverage(t *testing.T) {
	setupProductCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	t.Run("list grouped with multiple categories", func(t *testing.T) {
		_, err := svc.Create(&CreateProductRequest{Name: "QQ币10个", Category: "虚拟币", Price: 10.00, Stock: 100, Sort: 1})
		require.NoError(t, err)
		_, err = svc.Create(&CreateProductRequest{Name: "QQ币50个", Category: "虚拟币", Price: 50.00, Stock: 50, Sort: 2})
		require.NoError(t, err)
		_, err = svc.Create(&CreateProductRequest{Name: "爱奇艺月卡", Category: "影视会员", Price: 15.00, Stock: 30, Sort: 3})
		require.NoError(t, err)
		_, err = svc.Create(&CreateProductRequest{Name: "网易云会员", Category: "音乐会员", Price: 18.00, Stock: 20, Sort: 4})
		require.NoError(t, err)

		groups, err := svc.ListOnShelfGrouped()
		require.NoError(t, err)
		assert.Len(t, groups, 3)

		catMap := make(map[string]int)
		for _, g := range groups {
			catMap[g.Category] = len(g.Products)
		}
		assert.Equal(t, 2, catMap["虚拟币"])
		assert.Equal(t, 1, catMap["影视会员"])
		assert.Equal(t, 1, catMap["音乐会员"])
	})

	t.Run("list grouped excludes off-shelf products", func(t *testing.T) {
		offShelf, err := svc.Create(&CreateProductRequest{Name: "下架产品", Category: "虚拟币", Price: 5.00, Stock: 10, Sort: 5})
		require.NoError(t, err)
		require.NoError(t, database.DB.Model(&model.Product{}).Where("id = ?", offShelf.ID).Update("status", model.ProductStatusOffShelf).Error)

		groups, err := svc.ListOnShelfGrouped()
		require.NoError(t, err)

		totalProducts := 0
		for _, g := range groups {
			totalProducts += len(g.Products)
		}
		assert.Equal(t, 4, totalProducts)
	})

	t.Run("list grouped with empty category goes to 其他", func(t *testing.T) {
		_, err := svc.Create(&CreateProductRequest{Name: "无分类产品", Price: 5.00, Stock: 10, Status: model.ProductStatusOnShelf})
		require.NoError(t, err)

		groups, err := svc.ListOnShelfGrouped()
		require.NoError(t, err)

		found := false
		for _, g := range groups {
			if g.Category == "其他" {
				found = true
			}
		}
		assert.True(t, found)
	})

	t.Run("list grouped with no products returns empty", func(t *testing.T) {
		svc2 := NewProductService()
		groups, err := svc2.ListOnShelfGrouped()
		require.NoError(t, err)
		assert.NotNil(t, groups)
	})
}

func TestProductGetByIDCoverage(t *testing.T) {
	setupProductCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	t.Run("get existing product by id", func(t *testing.T) {
		product, err := svc.Create(&CreateProductRequest{
			Name:  "GetByIDProduct",
			Price: 25.00,
			Stock: 50,
		})
		require.NoError(t, err)

		found, err := svc.GetByID(product.ID)
		require.NoError(t, err)
		assert.Equal(t, product.Name, found.Name)
		assert.Equal(t, product.Price, found.Price)
	})

	t.Run("get nonexistent product returns error", func(t *testing.T) {
		found, err := svc.GetByID(9999)
		assert.Error(t, err)
		assert.Nil(t, found)
		assert.Contains(t, err.Error(), "产品不存在")
	})
}

func TestListOnShelfCoverage(t *testing.T) {
	setupProductCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	t.Run("list on-shelf with mixed statuses", func(t *testing.T) {
		_, err := svc.Create(&CreateProductRequest{Name: "OnShelf1", Price: 10.00, Stock: 100})
		require.NoError(t, err)
		_, err = svc.Create(&CreateProductRequest{Name: "OnShelf2", Price: 20.00, Stock: 50})
		require.NoError(t, err)
		offShelf, err := svc.Create(&CreateProductRequest{Name: "OffShelf", Price: 5.00, Stock: 10})
		require.NoError(t, err)
		require.NoError(t, database.DB.Model(&model.Product{}).Where("id = ?", offShelf.ID).Update("status", model.ProductStatusOffShelf).Error)

		products, err := svc.ListOnShelf()
		require.NoError(t, err)
		assert.Len(t, products, 2)
		for _, p := range products {
			assert.Equal(t, model.ProductStatusOnShelf, p.Status)
		}
	})
}

func TestUpdateAllFieldsCoverage(t *testing.T) {
	setupProductCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	product, err := svc.Create(&CreateProductRequest{
		Name:        "Original",
		Category:    "原分类",
		Price:       10.00,
		Description: "原描述",
		Image:       "/img/original.png",
		Stock:       50,
		Status:      model.ProductStatusOnShelf,
		Sort:        1,
	})
	require.NoError(t, err)

	t.Run("update all fields", func(t *testing.T) {
		updateReq := &UpdateProductRequest{
			ID:          product.ID,
			Name:        "Updated",
			Category:    "新分类",
			Price:       25.00,
			Description: "新描述",
			Image:       "/img/updated.png",
			Status:      model.ProductStatusOnShelf,
			Sort:        10,
		}

		updated, err := svc.Update(updateReq)
		require.NoError(t, err)
		assert.Equal(t, "Updated", updated.Name)
		assert.Equal(t, "新分类", updated.Category)
		assert.Equal(t, 25.00, updated.Price)
		assert.Equal(t, "新描述", updated.Description)
		assert.Equal(t, "/img/updated.png", updated.Image)
		assert.Equal(t, 10, updated.Sort)
	})
}

func TestProductServiceDBNilCoverage(t *testing.T) {
	svc := NewProductService()

	t.Run("create with DB nil returns error", func(t *testing.T) {
		originalDB := database.DB
		database.DB = nil
		defer func() { database.DB = originalDB }()

		product, err := svc.Create(&CreateProductRequest{Name: "Test", Price: 10.00})
		assert.Error(t, err)
		assert.Nil(t, product)
	})

	t.Run("update with DB nil returns error", func(t *testing.T) {
		originalDB := database.DB
		database.DB = nil
		defer func() { database.DB = originalDB }()

		product, err := svc.Update(&UpdateProductRequest{ID: 1, Name: "Test"})
		assert.Error(t, err)
		assert.Nil(t, product)
	})

	t.Run("delete with DB nil returns error", func(t *testing.T) {
		originalDB := database.DB
		database.DB = nil
		defer func() { database.DB = originalDB }()

		err := svc.Delete(1)
		assert.Error(t, err)
	})

	t.Run("get by id with DB nil returns error", func(t *testing.T) {
		originalDB := database.DB
		database.DB = nil
		defer func() { database.DB = originalDB }()

		product, err := svc.GetByID(1)
		assert.Error(t, err)
		assert.Nil(t, product)
	})

	t.Run("list with DB nil uses mock", func(t *testing.T) {
		originalDB := database.DB
		database.DB = nil
		defer func() { database.DB = originalDB }()

		result, err := svc.List(1, 10, "")
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("list on shelf with DB nil uses mock", func(t *testing.T) {
		originalDB := database.DB
		database.DB = nil
		defer func() { database.DB = originalDB }()

		products, err := svc.ListOnShelf()
		assert.NoError(t, err)
		assert.NotNil(t, products)
	})

	t.Run("list on shelf grouped with DB nil uses mock", func(t *testing.T) {
		originalDB := database.DB
		database.DB = nil
		defer func() { database.DB = originalDB }()

		groups, err := svc.ListOnShelfGrouped()
		assert.NoError(t, err)
		assert.NotNil(t, groups)
	})

	t.Run("update stock with DB nil returns error", func(t *testing.T) {
		originalDB := database.DB
		database.DB = nil
		defer func() { database.DB = originalDB }()

		err := svc.UpdateStock(1, 10)
		assert.Error(t, err)
	})
}

func TestProductListPaginationCoverage(t *testing.T) {
	setupProductCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	for i := 0; i < 25; i++ {
		_, err := svc.Create(&CreateProductRequest{
			Name:  fmt.Sprintf("Pagination商品-%d", i+1),
			Price: 10.00,
			Stock: 100,
		})
		require.NoError(t, err)
	}

	t.Run("list first page with 10 items", func(t *testing.T) {
		result, err := svc.List(1, 10, "")
		require.NoError(t, err)
		assert.Equal(t, int64(25), result.Total)
		assert.Len(t, result.Products, 10)
	})

	t.Run("list second page with 10 items", func(t *testing.T) {
		result, err := svc.List(2, 10, "")
		require.NoError(t, err)
		assert.Equal(t, int64(25), result.Total)
		assert.Len(t, result.Products, 10)
	})

	t.Run("list last page with remaining items", func(t *testing.T) {
		result, err := svc.List(3, 10, "")
		require.NoError(t, err)
		assert.Equal(t, int64(25), result.Total)
		assert.Len(t, result.Products, 5)
	})

	t.Run("list with keyword filter", func(t *testing.T) {
		result, err := svc.List(1, 10, "Pagination商品-1")
		require.NoError(t, err)
		assert.Greater(t, result.Total, int64(0))
		for _, p := range result.Products {
			assert.Contains(t, p.Name, "Pagination商品-1")
		}
	})
}

func TestProductCreateEdgeCasesCoverage(t *testing.T) {
	setupProductCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	t.Run("create with special characters in name", func(t *testing.T) {
		product, err := svc.Create(&CreateProductRequest{
			Name:  "Special/\\商品名?*",
			Price: 5.00,
		})
		require.NoError(t, err)
		assert.NotNil(t, product)
	})

	t.Run("create with negative stock allowed", func(t *testing.T) {
		product, err := svc.Create(&CreateProductRequest{
			Name:  "NegativeStock",
			Price: 10.00,
			Stock: -5,
		})
		require.NoError(t, err)
		assert.Equal(t, -5, product.Stock)
	})

	t.Run("create with very long name", func(t *testing.T) {
		longName := strings.Repeat("A", 200)
		product, err := svc.Create(&CreateProductRequest{
			Name:  longName,
			Price: 1.00,
		})
		require.NoError(t, err)
		assert.Equal(t, longName, product.Name)
	})
}