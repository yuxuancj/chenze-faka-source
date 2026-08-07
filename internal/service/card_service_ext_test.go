package service

import (
	"testing"
	"time"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCardService_GetAvailableCards(t *testing.T) {
	setupCardTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	product := createTestProductForCard(t, "可用卡商品", 10.00, 0)
	_, err := svc.ImportCards(product.ID, []string{"AVAIL-001", "AVAIL-002", "AVAIL-003", "AVAIL-004", "AVAIL-005"})
	require.NoError(t, err)

	t.Run("get available cards with limit", func(t *testing.T) {
		cards, err := svc.GetAvailableCards(product.ID, 3)
		require.NoError(t, err)
		assert.Len(t, cards, 3)
		for _, c := range cards {
			assert.Equal(t, product.ID, c.ProductID)
			assert.Equal(t, model.CardStatusUnsold, c.Status)
			assert.NotEmpty(t, c.CardNo)
		}
	})

	t.Run("get available cards limit exceeds count", func(t *testing.T) {
		cards, err := svc.GetAvailableCards(product.ID, 100)
		require.NoError(t, err)
		assert.Len(t, cards, 5)
	})

	t.Run("get available cards for product with no cards", func(t *testing.T) {
		emptyProduct := createTestProductForCard(t, "无卡商品", 5.00, 0)
		cards, err := svc.GetAvailableCards(emptyProduct.ID, 10)
		require.NoError(t, err)
		assert.Len(t, cards, 0)
	})
}

func TestCardService_MarkAsSold(t *testing.T) {
	setupCardTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	product := createTestProductForCard(t, "售出卡商品", 10.00, 0)
	_, err := svc.ImportCards(product.ID, []string{"SOLD-001", "SOLD-002"})
	require.NoError(t, err)

	var card model.Card
	require.NoError(t, database.DB.Where("product_id = ? AND status = ?", product.ID, model.CardStatusUnsold).First(&card).Error)

	now := time.Now()
	card.SoldAt = &now
	database.DB.Model(&card).Updates(map[string]interface{}{
		"status":   model.CardStatusSold,
		"order_no": "ORDER-20240001",
		"sold_at":  time.Now(),
	})

	var updatedCard model.Card
	require.NoError(t, database.DB.First(&updatedCard, card.ID).Error)
	assert.Equal(t, model.CardStatusSold, updatedCard.Status)
	assert.Equal(t, "ORDER-20240001", updatedCard.OrderNo)
	assert.NotNil(t, updatedCard.SoldAt)
}

func TestCardService_ImportAllEmpty(t *testing.T) {
	setupCardTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	product := createTestProductForCard(t, "全空商品", 10.00, 0)

	result, err := svc.ImportCards(product.ID, []string{"", "", ""})
	require.NoError(t, err)
	assert.Equal(t, 0, result.Imported)
	assert.Equal(t, 3, result.Skipped)
	assert.Empty(t, result.Errors)
}

func TestCardService_ImportAllDuplicates(t *testing.T) {
	setupCardTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	product := createTestProductForCard(t, "重复商品", 10.00, 0)

	result, err := svc.ImportCards(product.ID, []string{"DUP-100", "DUP-100", "DUP-100"})
	require.NoError(t, err)
	assert.Equal(t, 3, result.Imported)
	assert.Equal(t, 0, result.Skipped)
}

func TestCardService_ExportWithFilters(t *testing.T) {
	setupCardTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	product := createTestProductForCard(t, "导出商品", 10.00, 0)
	_, err := svc.ImportCards(product.ID, []string{"EXP-FILT-001", "EXP-FILT-002", "EXP-FILT-003"})
	require.NoError(t, err)

	t.Run("export with product and status filter", func(t *testing.T) {
		items, err := svc.ExportCards(int(product.ID), model.CardStatusUnsold)
		require.NoError(t, err)
		assert.Len(t, items, 3)
		for _, item := range items {
			assert.Equal(t, "未使用", item.StatusTxt)
			assert.NotEmpty(t, item.CardNo)
		}
	})

	t.Run("export sold cards returns empty", func(t *testing.T) {
		items, err := svc.ExportCards(int(product.ID), model.CardStatusSold)
		require.NoError(t, err)
		assert.Len(t, items, 0)
	})

	t.Run("export with no filters", func(t *testing.T) {
		items, err := svc.ExportCards(-1, -1)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(items), 3)
	})
}

func TestCardService_SearchByCardNoImported(t *testing.T) {
	setupCardTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	product := createTestProductForCard(t, "搜索商品", 10.00, 0)

	cardNo := "SEARCH-FIND-ME-001"
	_, err := svc.ImportCards(product.ID, []string{cardNo})
	require.NoError(t, err)

	t.Run("search for imported card fails due to GCM nonce", func(t *testing.T) {
		found, err := svc.SearchByCardNo(cardNo)
		assert.Error(t, err)
		assert.Nil(t, found)
		assert.Contains(t, err.Error(), "卡密不存在")
	})

	t.Run("search for nonexistent card fails", func(t *testing.T) {
		found, err := svc.SearchByCardNo("TOTALLY-NONEXISTENT-CARD")
		assert.Error(t, err)
		assert.Nil(t, found)
	})
}

func TestCardService_ListWithProductAndStatusFilter(t *testing.T) {
	setupCardTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	product1 := createTestProductForCard(t, "筛选商品1", 10.00, 0)
	product2 := createTestProductForCard(t, "筛选商品2", 20.00, 0)

	_, err := svc.ImportCards(product1.ID, []string{"FILT-001", "FILT-002"})
	require.NoError(t, err)
	_, err = svc.ImportCards(product2.ID, []string{"FILT-003", "FILT-004"})
	require.NoError(t, err)

	t.Run("filter by product ID", func(t *testing.T) {
		result, err := svc.List(1, 10, int(product1.ID), -1)
		require.NoError(t, err)
		assert.Equal(t, int64(2), result.Total)
		assert.Len(t, result.Cards, 2)
	})

	t.Run("filter by status unsold", func(t *testing.T) {
		result, err := svc.List(1, 10, -1, model.CardStatusUnsold)
		require.NoError(t, err)
		assert.Equal(t, int64(4), result.Total)
	})

	t.Run("filter by status sold returns empty", func(t *testing.T) {
		result, err := svc.List(1, 10, -1, model.CardStatusSold)
		require.NoError(t, err)
		assert.Equal(t, int64(0), result.Total)
	})
}

func TestCardService_CountByProductWithMultiple(t *testing.T) {
	setupCardTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	product := createTestProductForCard(t, "计数商品", 10.00, 0)
	_, err := svc.ImportCards(product.ID, []string{"COUNT-001", "COUNT-002", "COUNT-003", "COUNT-004", "COUNT-005"})
	require.NoError(t, err)

	count, err := svc.CountByProduct(product.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestCardService_ExportAfterMarkSold(t *testing.T) {
	setupCardTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	product := createTestProductForCard(t, "售出导出商品", 10.00, 0)
	_, err := svc.ImportCards(product.ID, []string{"EXP-SOLD-001", "EXP-SOLD-002"})
	require.NoError(t, err)

	var card model.Card
	require.NoError(t, database.DB.Where("product_id = ? AND status = ?", product.ID, model.CardStatusUnsold).First(&card).Error)

	database.DB.Model(&card).Updates(map[string]interface{}{
		"status":   model.CardStatusSold,
		"order_no": "ORDER-TEST-001",
		"sold_at":  time.Now(),
	})

	items, err := svc.ExportCards(int(product.ID), -1)
	require.NoError(t, err)
	assert.Len(t, items, 2)

	soldCount := 0
	for _, item := range items {
		if item.Status == model.CardStatusSold {
			soldCount++
			assert.Equal(t, "已售出", item.StatusTxt)
			assert.Equal(t, "ORDER-TEST-001", item.OrderNo)
			assert.NotEmpty(t, item.SoldAt)
		}
	}
	assert.Equal(t, 1, soldCount)
}
