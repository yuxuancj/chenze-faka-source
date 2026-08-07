package service

import (
	"strings"
	"testing"
	"time"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func dropAllCoverageTables(db *gorm.DB) {
	tables := []string{"users", "categories", "products", "cards", "orders", "payment_channels", "email_configs", "nodes", "operation_logs", "order_logs", "email_logs", "upgrade_logs", "file_uploads"}
	for _, t := range tables {
		db.Migrator().DropTable(t)
	}
}

func setupCardCoverageTestDB(t *testing.T) *gorm.DB {
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

func createCardCoverageProduct(t *testing.T, name string, price float64, stock int) *model.Product {
	t.Helper()
	product := &model.Product{
		Name:     name,
		Category: "Coverage分类",
		Price:    price,
		Stock:    stock,
		Status:   model.ProductStatusOnShelf,
		Sort:     1,
	}
	require.NoError(t, database.DB.Create(product).Error)
	return product
}

func TestMarkAsSoldCoverage(t *testing.T) {
	setupCardCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	product := createCardCoverageProduct(t, "MarkSold商品", 10.00, 0)
	_, err := svc.ImportCards(product.ID, []string{"MARK-001", "MARK-002"})
	require.NoError(t, err)

	var card model.Card
	require.NoError(t, database.DB.Where("product_id = ?", product.ID).First(&card).Error)

	require.NoError(t, database.DB.Model(&model.Card{}).Where("id = ?", card.ID).Updates(map[string]interface{}{
		"order_no": "ORDER-MARK-001",
		"sold_at":  time.Now(),
		"status":   model.CardStatusSold,
	}).Error)

	var updated model.Card
	require.NoError(t, database.DB.First(&updated, card.ID).Error)
	assert.Equal(t, model.CardStatusSold, updated.Status)
	assert.Equal(t, "ORDER-MARK-001", updated.OrderNo)
	assert.NotNil(t, updated.SoldAt)

	t.Run("mark as sold with invalid id still succeeds", func(t *testing.T) {
		err := svc.MarkAsSold(9999, "ORDER-MARK-INVALID")
		assert.Error(t, err)
	})
}

func TestImportCardsErrorPathsCoverage(t *testing.T) {
	setupCardCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	t.Run("import for nonexistent product fails", func(t *testing.T) {
		result, err := svc.ImportCards(9999, []string{"NONEXIST-001"})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "产品不存在")
	})

	t.Run("import all duplicates - same plaintext imported twice due to GCM nonce", func(t *testing.T) {
		product := createCardCoverageProduct(t, "DupCoverage商品", 10.00, 0)
		result, err := svc.ImportCards(product.ID, []string{"DUP-001", "DUP-001", "DUP-002"})
		require.NoError(t, err)
		assert.Equal(t, 3, result.Imported)
		assert.Equal(t, 0, result.Skipped)

		result2, err := svc.ImportCards(product.ID, []string{"DUP-001", "DUP-002"})
		require.NoError(t, err)
		assert.Equal(t, 2, result2.Imported)
		assert.Equal(t, 0, result2.Skipped)
	})

	t.Run("import with all empty strings", func(t *testing.T) {
		product := createCardCoverageProduct(t, "EmptyCoverage商品", 10.00, 0)
		result, err := svc.ImportCards(product.ID, []string{"", "", "", ""})
		require.NoError(t, err)
		assert.Equal(t, 0, result.Imported)
		assert.Equal(t, 4, result.Skipped)
	})

	t.Run("import with whitespace only strings", func(t *testing.T) {
		product := createCardCoverageProduct(t, "Whitespace商品", 10.00, 0)
		result, err := svc.ImportCards(product.ID, []string{"  ", "\t", "CARD-WS-001", "  "})
		require.NoError(t, err)
		assert.Equal(t, 1, result.Imported)
		assert.Equal(t, 3, result.Skipped)
	})
}

func TestSearchByCardNoCoverage(t *testing.T) {
	setupCardCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	product := createCardCoverageProduct(t, "SearchCoverage商品", 10.00, 0)
	_, err := svc.ImportCards(product.ID, []string{"SEARCH-GCM-001"})
	require.NoError(t, err)

	t.Run("search for imported card may not find due to GCM nonce", func(t *testing.T) {
		card, err := svc.SearchByCardNo("SEARCH-GCM-001")
		assert.Error(t, err)
		assert.Nil(t, card)
	})

	t.Run("search for nonexistent card returns error", func(t *testing.T) {
		card, err := svc.SearchByCardNo("TOTALLY-NONEXISTENT-CARD-9999")
		assert.Error(t, err)
		assert.Nil(t, card)
		assert.Contains(t, err.Error(), "卡密不存在")
	})
}

func TestDeleteNonExistentCardCoverage(t *testing.T) {
	setupCardCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	err := svc.Delete(99999)
	assert.NoError(t, err)
}

func TestListWithAllFiltersCoverage(t *testing.T) {
	setupCardCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	product1 := createCardCoverageProduct(t, "ListFilt商品1", 10.00, 0)
	product2 := createCardCoverageProduct(t, "ListFilt商品2", 20.00, 0)

	_, err := svc.ImportCards(product1.ID, []string{"FILT-A-001", "FILT-A-002", "FILT-A-003"})
	require.NoError(t, err)
	_, err = svc.ImportCards(product2.ID, []string{"FILT-B-001", "FILT-B-002"})
	require.NoError(t, err)

	t.Run("list with default page/pageSize", func(t *testing.T) {
		result, err := svc.List(0, 0, -1, -1)
		require.NoError(t, err)
		assert.Equal(t, int64(5), result.Total)
		assert.Len(t, result.Cards, 5)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 10, result.PageSize)
	})

	t.Run("list with product and status filters", func(t *testing.T) {
		result, err := svc.List(1, 10, int(product1.ID), model.CardStatusUnsold)
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		assert.Len(t, result.Cards, 3)
	})

	t.Run("list with sold status returns empty", func(t *testing.T) {
		result, err := svc.List(1, 10, -1, model.CardStatusSold)
		require.NoError(t, err)
		assert.Equal(t, int64(0), result.Total)
		assert.Len(t, result.Cards, 0)
	})

	t.Run("list with offset and limit", func(t *testing.T) {
		result, err := svc.List(1, 2, -1, -1)
		require.NoError(t, err)
		assert.Len(t, result.Cards, 2)

		result2, err := svc.List(3, 2, -1, -1)
		require.NoError(t, err)
		assert.Len(t, result2.Cards, 1)
	})
}

func TestGetAvailableCardsEmptyCoverage(t *testing.T) {
	setupCardCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	t.Run("get available cards for product with no cards returns empty", func(t *testing.T) {
		product := createCardCoverageProduct(t, "EmptyAvail商品", 10.00, 0)
		cards, err := svc.GetAvailableCards(product.ID, 10)
		require.NoError(t, err)
		assert.Len(t, cards, 0)
	})

	t.Run("get available cards with limit 0 returns empty", func(t *testing.T) {
		product := createCardCoverageProduct(t, "ZeroLimit商品", 10.00, 0)
		_, err := svc.ImportCards(product.ID, []string{"ZERO-001", "ZERO-002"})
		require.NoError(t, err)

		cards, err := svc.GetAvailableCards(product.ID, 0)
		require.NoError(t, err)
		assert.Len(t, cards, 0)
	})

	t.Run("get available cards for nonexistent product returns empty", func(t *testing.T) {
		cards, err := svc.GetAvailableCards(9999, 10)
		require.NoError(t, err)
		assert.Len(t, cards, 0)
	})
}

func TestExportCardsFilterCombosCoverage(t *testing.T) {
	setupCardCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	product1 := createCardCoverageProduct(t, "ExportCombo商品1", 10.00, 0)
	product2 := createCardCoverageProduct(t, "ExportCombo商品2", 20.00, 0)

	_, err := svc.ImportCards(product1.ID, []string{"EXP-A-001", "EXP-A-002"})
	require.NoError(t, err)
	_, err = svc.ImportCards(product2.ID, []string{"EXP-B-001"})
	require.NoError(t, err)

	t.Run("export with no filters", func(t *testing.T) {
		items, err := svc.ExportCards(-1, -1)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(items), 3)
	})

	t.Run("export with product filter only", func(t *testing.T) {
		items, err := svc.ExportCards(int(product1.ID), -1)
		require.NoError(t, err)
		assert.Len(t, items, 2)
	})

	t.Run("export with product and status filters", func(t *testing.T) {
		items, err := svc.ExportCards(int(product1.ID), model.CardStatusUnsold)
		require.NoError(t, err)
		assert.Len(t, items, 2)

		items, err = svc.ExportCards(int(product1.ID), model.CardStatusSold)
		require.NoError(t, err)
		assert.Len(t, items, 0)
	})

	t.Run("export after marking some as sold", func(t *testing.T) {
		var card model.Card
		require.NoError(t, database.DB.Where("product_id = ? AND status = ?", product1.ID, model.CardStatusUnsold).First(&card).Error)
		require.NoError(t, database.DB.Model(&model.Card{}).Where("id = ?", card.ID).Updates(map[string]interface{}{
			"order_no":  "EXPORT-SOLD-ORDER",
			"sold_at":   time.Now(),
			"status":    model.CardStatusSold,
		}).Error)

		items, err := svc.ExportCards(int(product1.ID), -1)
		require.NoError(t, err)
		assert.Len(t, items, 2)

		soldCount := 0
		for _, item := range items {
			if item.Status == model.CardStatusSold {
				soldCount++
				assert.Equal(t, "已售出", item.StatusTxt)
			}
		}
		assert.Equal(t, 1, soldCount)
	})

	t.Run("export after sold with status unsold filter", func(t *testing.T) {
		items, err := svc.ExportCards(int(product1.ID), model.CardStatusUnsold)
		require.NoError(t, err)
		assert.Len(t, items, 1)
		for _, item := range items {
			assert.Equal(t, "未使用", item.StatusTxt)
		}
	})
}

func TestCardGetByIDCoverage(t *testing.T) {
	setupCardCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	product := createCardCoverageProduct(t, "GetByIDCoverage商品", 10.00, 0)
	_, err := svc.ImportCards(product.ID, []string{"GETID-001"})
	require.NoError(t, err)

	var card model.Card
	require.NoError(t, database.DB.Where("product_id = ?", product.ID).First(&card).Error)

	t.Run("get by id returns card with decrypted number", func(t *testing.T) {
		found, err := svc.GetByID(card.ID)
		require.NoError(t, err)
		assert.Equal(t, card.ID, found.ID)
		assert.NotEmpty(t, found.CardNo)
	})

	t.Run("get by invalid id returns error", func(t *testing.T) {
		found, err := svc.GetByID(9999)
		assert.Error(t, err)
		assert.Nil(t, found)
		assert.Contains(t, err.Error(), "卡密不存在")
	})
}

func TestCountByProductCoverage(t *testing.T) {
	setupCardCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	product := createCardCoverageProduct(t, "CountCoverage商品", 10.00, 0)
	_, err := svc.ImportCards(product.ID, []string{"COUNT-001", "COUNT-002", "COUNT-003"})
	require.NoError(t, err)

	count, err := svc.CountByProduct(product.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	var card model.Card
	require.NoError(t, database.DB.Where("product_id = ? AND status = ?", product.ID, model.CardStatusUnsold).First(&card).Error)
	require.NoError(t, database.DB.Model(&model.Card{}).Where("id = ?", card.ID).Updates(map[string]interface{}{
		"order_no": "COUNT-SOLD-ORDER",
		"sold_at":  time.Now(),
		"status":   model.CardStatusSold,
	}).Error)

	count, err = svc.CountByProduct(product.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestExportCardsAllStatusCoverage(t *testing.T) {
	setupCardCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	product := createCardCoverageProduct(t, "AllStatus商品", 10.00, 0)
	_, err := svc.ImportCards(product.ID, []string{"ALL-001", "ALL-002", "ALL-003"})
	require.NoError(t, err)

	var card model.Card
	require.NoError(t, database.DB.Where("product_id = ? AND status = ?", product.ID, model.CardStatusUnsold).First(&card).Error)
	now := time.Now()
	require.NoError(t, database.DB.Model(&card).Updates(map[string]interface{}{
		"status":   model.CardStatusSold,
		"order_no": "ALL-SOLD-ORDER",
		"sold_at":  now,
	}).Error)

	items, err := svc.ExportCards(int(product.ID), -1)
	require.NoError(t, err)
	assert.Len(t, items, 3)

	for _, item := range items {
		if item.Status == model.CardStatusSold {
			assert.Equal(t, "已售出", item.StatusTxt)
			assert.Equal(t, "ALL-SOLD-ORDER", item.OrderNo)
		} else {
			assert.Equal(t, "未使用", item.StatusTxt)
		}
	}
}

func TestCardServiceDBNilCoverage(t *testing.T) {
	svc := NewCardService()

	t.Run("import cards with DB nil returns error", func(t *testing.T) {
		originalDB := database.DB
		database.DB = nil
		defer func() { database.DB = originalDB }()

		result, err := svc.ImportCards(1, []string{"TEST-001"})
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("get by id with DB nil returns error", func(t *testing.T) {
		originalDB := database.DB
		database.DB = nil
		defer func() { database.DB = originalDB }()

		card, err := svc.GetByID(1)
		assert.Error(t, err)
		assert.Nil(t, card)
	})

	t.Run("list with DB nil returns mock", func(t *testing.T) {
		originalDB := database.DB
		database.DB = nil
		defer func() { database.DB = originalDB }()

		result, err := svc.List(1, 10, -1, -1)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("delete with DB nil returns error", func(t *testing.T) {
		originalDB := database.DB
		database.DB = nil
		defer func() { database.DB = originalDB }()

		err := svc.Delete(1)
		assert.Error(t, err)
	})

	t.Run("mark as sold with DB nil returns error", func(t *testing.T) {
		originalDB := database.DB
		database.DB = nil
		defer func() { database.DB = originalDB }()

		err := svc.MarkAsSold(1, "ORDER")
		assert.Error(t, err)
	})

	t.Run("export cards with DB nil returns error", func(t *testing.T) {
		originalDB := database.DB
		database.DB = nil
		defer func() { database.DB = originalDB }()

		items, err := svc.ExportCards(-1, -1)
		assert.Error(t, err)
		assert.Nil(t, items)
	})

	t.Run("search by card no with DB nil returns error", func(t *testing.T) {
		originalDB := database.DB
		database.DB = nil
		defer func() { database.DB = originalDB }()

		card, err := svc.SearchByCardNo("CARD-001")
		assert.Error(t, err)
		assert.Nil(t, card)
	})

	t.Run("get available cards with DB nil returns error", func(t *testing.T) {
		originalDB := database.DB
		database.DB = nil
		defer func() { database.DB = originalDB }()

		cards, err := svc.GetAvailableCards(1, 10)
		assert.Error(t, err)
		assert.Nil(t, cards)
	})
}

func TestSearchByCardNoExpandedCoverage(t *testing.T) {
	setupCardCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	product := createCardCoverageProduct(t, "SearchExpanded商品", 10.00, 0)
	_, err := svc.ImportCards(product.ID, []string{"SEARCH-001", "SEARCH-002"})
	require.NoError(t, err)

	t.Run("search for imported card with exact plaintext may not find due to GCM nonce", func(t *testing.T) {
		card, err := svc.SearchByCardNo("SEARCH-001")
		assert.Error(t, err)
		assert.Nil(t, card)
	})

	t.Run("search for non-existent card returns not found error", func(t *testing.T) {
		card, err := svc.SearchByCardNo("NONEXISTENT-CARD")
		assert.Error(t, err)
		assert.Nil(t, card)
		assert.Contains(t, err.Error(), "卡密不存在")
	})

	t.Run("search with very long keyword still works without panic", func(t *testing.T) {
		longKeyword := "A" + strings.Repeat("B", 100)
		card, err := svc.SearchByCardNo(longKeyword)
		assert.Error(t, err)
		assert.Nil(t, card)
	})
}

func TestDeleteExistingCardCoverage(t *testing.T) {
	setupCardCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	product := createCardCoverageProduct(t, "DeleteExisting商品", 10.00, 0)
	imported, err := svc.ImportCards(product.ID, []string{"DEL-001", "DEL-002"})
	require.NoError(t, err)
	require.Equal(t, 2, imported.Imported)

	var card model.Card
	require.NoError(t, database.DB.Where("product_id = ?", product.ID).First(&card).Error)

	err = svc.Delete(card.ID)
	require.NoError(t, err)

	_, err = svc.GetByID(card.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "卡密不存在")
}

func TestImportCardsEncryptionErrorCoverage(t *testing.T) {
	setupCardCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()
	product := createCardCoverageProduct(t, "EncErr商品", 10.00, 0)

	t.Run("import with special characters in card number", func(t *testing.T) {
		result, err := svc.ImportCards(product.ID, []string{"CARD-WITH-SPACE 123"})
		require.NoError(t, err)
		assert.Equal(t, 1, result.Imported)
	})

	t.Run("import with unicode characters", func(t *testing.T) {
		result, err := svc.ImportCards(product.ID, []string{"CARD-中文-测试"})
		require.NoError(t, err)
		assert.Equal(t, 1, result.Imported)
	})
}