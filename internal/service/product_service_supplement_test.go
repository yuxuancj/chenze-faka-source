package service

import (
	"fmt"
	"testing"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductDBNilSupplement(t *testing.T) {
	svc := NewProductService()

	originalDB := database.DB
	database.DB = nil
	defer func() { database.DB = originalDB }()

	t.Run("get by id with nil db", func(t *testing.T) {
		p, err := svc.GetByID(1)
		assert.Error(t, err)
		assert.Nil(t, p)
	})

	t.Run("update with nil db", func(t *testing.T) {
		_, err := svc.Update(&UpdateProductRequest{ID: 1, Name: "Test"})
		assert.Error(t, err)
	})

	t.Run("delete with nil db", func(t *testing.T) {
		err := svc.Delete(1)
		assert.Error(t, err)
	})

	t.Run("list with nil db returns mock", func(t *testing.T) {
		result, err := svc.List(1, 10, "")
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("list on shelf grouped with nil db returns mock", func(t *testing.T) {
		result, err := svc.ListOnShelfGrouped()
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("create with nil db", func(t *testing.T) {
		p, err := svc.Create(&CreateProductRequest{Name: "Test", Price: 10, Stock: 5})
		assert.Error(t, err)
		assert.Nil(t, p)
	})

	t.Run("update stock with nil db", func(t *testing.T) {
		err := svc.UpdateStock(1, 10)
		assert.Error(t, err)
	})
}

func TestProductUpdateStockNegative(t *testing.T) {
	setupProductCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	product := &model.Product{
		Name:     "NegativeStock Product",
		Category: "测试分类",
		Price:    10.00,
		Stock:    50,
		Status:   model.ProductStatusOnShelf,
		Sort:     1,
	}
	require.NoError(t, database.DB.Create(product).Error)

	t.Run("decrease stock below 0 results in negative stock", func(t *testing.T) {
		err := svc.UpdateStock(product.ID, -100)
		require.NoError(t, err)

		var updated model.Product
		require.NoError(t, database.DB.First(&updated, product.ID).Error)
		assert.Equal(t, -50, updated.Stock)
	})
}

func TestProductUpdateNonExistent(t *testing.T) {
	setupProductCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	t.Run("update non-existent product returns error", func(t *testing.T) {
		_, err := svc.Update(&UpdateProductRequest{ID: 99999, Name: "Ghost"})
		assert.Error(t, err)
	})
}

func TestProductDeleteNonExistent(t *testing.T) {
	setupProductCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	t.Run("delete non-existent product returns nil", func(t *testing.T) {
		err := svc.Delete(99999)
		assert.NoError(t, err)
	})
}

func TestProductGetByIDNonExistent(t *testing.T) {
	setupProductCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	t.Run("get by non-existent id returns error", func(t *testing.T) {
		p, err := svc.GetByID(99999)
		assert.Error(t, err)
		assert.Nil(t, p)
	})
}

func TestProductCreateDuplicateName(t *testing.T) {
	setupProductCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	_, err := svc.Create(&CreateProductRequest{
		Name:  "DuplicateName Product",
		Price: 10.00,
		Stock: 50,
	})
	require.NoError(t, err)

	t.Run("create with same name may still work", func(t *testing.T) {
		p, err := svc.Create(&CreateProductRequest{
			Name:  "DuplicateName Product",
			Price: 20.00,
			Stock: 30,
		})
		require.NoError(t, err)
		assert.NotNil(t, p)
	})
}

func TestProductListOnShelfMultipleCategories(t *testing.T) {
	setupProductCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	for i := 0; i < 5; i++ {
		_, err := svc.Create(&CreateProductRequest{
			Name:     fmt.Sprintf("Cat1-Product-%d", i),
			Price:    10.00,
			Stock:    100,
			Category: "分类A",
		})
		require.NoError(t, err)
	}

	for i := 0; i < 3; i++ {
		_, err := svc.Create(&CreateProductRequest{
			Name:     fmt.Sprintf("Cat2-Product-%d", i),
			Price:    20.00,
			Stock:    50,
			Category: "分类B",
		})
		require.NoError(t, err)
	}

	t.Run("list on shelf grouped returns 2 categories", func(t *testing.T) {
		result, err := svc.ListOnShelfGrouped()
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("list on shelf with off-shelf product not included", func(t *testing.T) {
		p, err := svc.Create(&CreateProductRequest{
			Name:     "OffShelf Product",
			Price:    30.00,
			Stock:    10,
			Category: "分类C",
		})
		require.NoError(t, err)

		require.NoError(t, database.DB.Model(&model.Product{}).Where("id = ?", p.ID).Update("status", model.ProductStatusOffShelf).Error)

		result, err := svc.ListOnShelfGrouped()
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})
}

func TestProductUpdateStatusViaUpdate(t *testing.T) {
	setupProductCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	p, err := svc.Create(&CreateProductRequest{
		Name:  "Status Product",
		Price: 10.00,
		Stock: 50,
	})
	require.NoError(t, err)

	t.Run("update status to off-shelf via Update (0 is treated as not-set)", func(t *testing.T) {
		_, err := svc.Update(&UpdateProductRequest{
			ID:     p.ID,
			Status: model.ProductStatusOffShelf,
		})
		require.NoError(t, err)

		updated, err := svc.GetByID(p.ID)
		require.NoError(t, err)
		assert.Equal(t, model.ProductStatusOnShelf, updated.Status)
	})

	t.Run("update status back to on-shelf via Update", func(t *testing.T) {
		_, err := svc.Update(&UpdateProductRequest{
			ID:     p.ID,
			Status: model.ProductStatusOnShelf,
		})
		require.NoError(t, err)

		updated, err := svc.GetByID(p.ID)
		require.NoError(t, err)
		assert.Equal(t, model.ProductStatusOnShelf, updated.Status)
	})
}

func TestProductListEdgeCases(t *testing.T) {
	setupProductCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	for i := 0; i < 15; i++ {
		_, err := svc.Create(&CreateProductRequest{
			Name:  fmt.Sprintf("ListProduct-%d", i+1),
			Price: 10.00,
			Stock: 100,
		})
		require.NoError(t, err)
	}

	t.Run("list with negative page defaults to 1", func(t *testing.T) {
		result, err := svc.List(-1, 10, "")
		require.NoError(t, err)
		assert.Equal(t, int64(15), result.Total)
		assert.Equal(t, 1, result.Page)
	})

	t.Run("list with negative page size defaults to 10", func(t *testing.T) {
		result, err := svc.List(1, -5, "")
		require.NoError(t, err)
		assert.Equal(t, int64(15), result.Total)
		assert.Equal(t, 10, result.PageSize)
	})

	t.Run("list with keyword filter", func(t *testing.T) {
		result, err := svc.List(1, 10, "ListProduct-5")
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
	})

	t.Run("list with keyword no match", func(t *testing.T) {
		result, err := svc.List(1, 10, "NONEXISTENT-KEYWORD")
		require.NoError(t, err)
		assert.Equal(t, int64(0), result.Total)
	})
}

func TestProductListOnShelfNoProducts(t *testing.T) {
	setupProductCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	t.Run("list on shelf with no products returns empty", func(t *testing.T) {
		result, err := svc.ListOnShelf()
		require.NoError(t, err)
		assert.Len(t, result, 0)
	})

	t.Run("list on shelf grouped with no products returns empty", func(t *testing.T) {
		result, err := svc.ListOnShelfGrouped()
		require.NoError(t, err)
		assert.Len(t, result, 0)
	})
}

func TestProductCreateMissingFields(t *testing.T) {
	setupProductCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	t.Run("create with empty name fails", func(t *testing.T) {
		p, err := svc.Create(&CreateProductRequest{Name: "", Price: 10, Stock: 5})
		assert.Error(t, err)
		assert.Nil(t, p)
	})

	t.Run("create with zero price fails", func(t *testing.T) {
		p, err := svc.Create(&CreateProductRequest{Name: "Test", Price: 0, Stock: 5})
		assert.Error(t, err)
		assert.Nil(t, p)
	})

	t.Run("create with negative price fails", func(t *testing.T) {
		p, err := svc.Create(&CreateProductRequest{Name: "Test", Price: -5, Stock: 5})
		assert.Error(t, err)
		assert.Nil(t, p)
	})
}

func TestContainsStrAndSearchStr(t *testing.T) {
	t.Run("containsStr finds substring", func(t *testing.T) {
		assert.True(t, containsStr("hello world", "world"))
		assert.True(t, containsStr("hello world", "hello"))
		assert.True(t, containsStr("hello world", "o w"))
	})

	t.Run("containsStr returns false for non-substring", func(t *testing.T) {
		assert.False(t, containsStr("hello world", "xyz"))
		assert.False(t, containsStr("hello", "hello world"))
	})

	t.Run("containsStr with empty substring", func(t *testing.T) {
		assert.True(t, containsStr("hello", ""))
	})

	t.Run("searchStr finds substring at start", func(t *testing.T) {
		assert.True(t, searchStr("hello", "hel"))
	})

	t.Run("searchStr finds substring at end", func(t *testing.T) {
		assert.True(t, searchStr("hello", "llo"))
	})

	t.Run("searchStr finds substring in middle", func(t *testing.T) {
		assert.True(t, searchStr("hello", "ell"))
	})

	t.Run("searchStr returns false for non-match", func(t *testing.T) {
		assert.False(t, searchStr("hello", "xyz"))
	})

	t.Run("searchStr with single character", func(t *testing.T) {
		assert.True(t, searchStr("hello", "h"))
		assert.False(t, searchStr("hello", "z"))
	})

	t.Run("searchStr with matching single char", func(t *testing.T) {
		assert.True(t, searchStr("a", "a"))
	})

	t.Run("searchStr with non-matching single char", func(t *testing.T) {
		assert.False(t, searchStr("a", "b"))
	})
}

func TestProductFullWorkflow(t *testing.T) {
	setupProductCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	t.Run("full CRUD workflow", func(t *testing.T) {
		p, err := svc.Create(&CreateProductRequest{
			Name:     "Workflow Product",
			Category: "Workflow",
			Price:    99.99,
			Stock:    100,
		})
		require.NoError(t, err)
		require.NotNil(t, p)
		assert.Equal(t, "Workflow Product", p.Name)
		assert.Equal(t, 99.99, p.Price)

		updated, err := svc.Update(&UpdateProductRequest{
			ID:   p.ID,
			Name: "Updated Product",
		})
		require.NoError(t, err)
		assert.Equal(t, "Updated Product", updated.Name)

		fetched, err := svc.GetByID(p.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Product", fetched.Name)

		err = svc.Delete(p.ID)
		require.NoError(t, err)

		_, err = svc.GetByID(p.ID)
		assert.Error(t, err)
	})
}

func TestProductUpdateStockCoverage(t *testing.T) {
	setupProductCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewProductService()

	p, err := svc.Create(&CreateProductRequest{
		Name:  "Stock Product",
		Price: 10.00,
		Stock: 50,
	})
	require.NoError(t, err)

	t.Run("increase stock", func(t *testing.T) {
		err := svc.UpdateStock(p.ID, 10)
		require.NoError(t, err)

		updated, err := svc.GetByID(p.ID)
		require.NoError(t, err)
		assert.Equal(t, 60, updated.Stock)
	})

	t.Run("decrease stock", func(t *testing.T) {
		err := svc.UpdateStock(p.ID, -20)
		require.NoError(t, err)

		updated, err := svc.GetByID(p.ID)
		require.NoError(t, err)
		assert.Equal(t, 40, updated.Stock)
	})

	t.Run("update non-existent product stock", func(t *testing.T) {
		err := svc.UpdateStock(99999, 5)
		require.NoError(t, err)
	})
}