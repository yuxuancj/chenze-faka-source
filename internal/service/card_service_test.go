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

func dropAllCardTables(db *gorm.DB) {
	tables := []string{"users", "categories", "products", "cards", "orders", "payment_channels", "email_configs", "nodes", "operation_logs", "order_logs", "email_logs", "upgrade_logs", "file_uploads"}
	for _, t := range tables {
		db.Migrator().DropTable(t)
	}
}

func setupCardTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	dropAllCardTables(db)
	database.DB = db
	require.NoError(t, db.AutoMigrate(
		&model.User{}, &model.Category{}, &model.Product{},
		&model.Card{}, &model.Order{}, &model.PaymentChannel{},
		&model.EmailConfig{}, &model.Node{},
		&model.OperationLog{}, &model.OrderLog{}, &model.EmailLog{},
		&model.UpgradeLog{}, &model.FileUpload{},
	))
}

func createTestProductForCard(t *testing.T, name string, price float64, stock int) *model.Product {
	t.Helper()
	product := &model.Product{
		Name:        name,
		Category:    "测试分类",
		Price:       price,
		Description: "测试产品描述",
		Stock:       stock,
		Status:      model.ProductStatusOnShelf,
		Sort:        1,
	}
	require.NoError(t, database.DB.Create(product).Error)
	return product
}

func TestCardService_ImportCards(t *testing.T) {
	setupCardTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	t.Run("import cards with valid card numbers", func(t *testing.T) {
		product := createTestProductForCard(t, "卡密商品A", 10.00, 0)

		cardNos := []string{"CARD-001", "CARD-002", "CARD-003"}
		result, err := svc.ImportCards(product.ID, cardNos)
		require.NoError(t, err)
		assert.Equal(t, 3, result.Imported)
		assert.Equal(t, 3, result.TotalCount)
		assert.Equal(t, 0, result.Skipped)
		assert.Empty(t, result.Errors)

		var count int64
		require.NoError(t, database.DB.Model(&model.Card{}).Where("product_id = ?", product.ID).Count(&count).Error)
		assert.Equal(t, int64(3), count)

		var prod model.Product
		require.NoError(t, database.DB.First(&prod, product.ID).Error)
		assert.Equal(t, 3, prod.Stock)
	})

	t.Run("import cards with empty text skips them", func(t *testing.T) {
		product := createTestProductForCard(t, "卡密商品B", 10.00, 0)

		cardNos := []string{"", "CARD-101", "", "CARD-102"}
		result, err := svc.ImportCards(product.ID, cardNos)
		require.NoError(t, err)
		assert.Equal(t, 2, result.Imported)
		assert.Equal(t, 2, result.Skipped)
	})

	t.Run("import cards for nonexistent product", func(t *testing.T) {
		result, err := svc.ImportCards(9999, []string{"CARD-999"})
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "产品不存在")
	})

	t.Run("import duplicate cards", func(t *testing.T) {
		product := createTestProductForCard(t, "卡密商品C", 10.00, 0)

		cardNos := []string{"DUP-001", "DUP-001", "DUP-002"}
		result, err := svc.ImportCards(product.ID, cardNos)
		require.NoError(t, err)
		assert.Equal(t, 3, result.Imported)
		assert.Equal(t, 0, result.Skipped)
	})
}

func TestCardService_ParseCardText(t *testing.T) {
	svc := NewCardService()

	t.Run("parse multi-line text", func(t *testing.T) {
		text := "CARD-A\nCARD-B\nCARD-C\n"
		cards := svc.ParseCardText(text)
		assert.Len(t, cards, 3)
		assert.Equal(t, "CARD-A", cards[0])
		assert.Equal(t, "CARD-B", cards[1])
		assert.Equal(t, "CARD-C", cards[2])
	})

	t.Run("parse text with empty lines", func(t *testing.T) {
		text := "CARD-A\n\nCARD-B\n\n"
		cards := svc.ParseCardText(text)
		assert.Len(t, cards, 2)
	})

	t.Run("parse single line", func(t *testing.T) {
		text := "SINGLE-CARD"
		cards := svc.ParseCardText(text)
		assert.Len(t, cards, 1)
		assert.Equal(t, "SINGLE-CARD", cards[0])
	})
}

func TestCardService_GetByID(t *testing.T) {
	setupCardTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	product := createTestProductForCard(t, "GetByID商品", 10.00, 0)
	result, err := svc.ImportCards(product.ID, []string{"GET-001"})
	require.NoError(t, err)
	require.Equal(t, 1, result.Imported)

	var card model.Card
	require.NoError(t, database.DB.Where("product_id = ?", product.ID).First(&card).Error)

	t.Run("get by valid id", func(t *testing.T) {
		found, err := svc.GetByID(card.ID)
		require.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, card.ID, found.ID)
		assert.NotEmpty(t, found.CardNo)
	})

	t.Run("get by invalid id", func(t *testing.T) {
		found, err := svc.GetByID(9999)
		assert.Error(t, err)
		assert.Nil(t, found)
		assert.Contains(t, err.Error(), "卡密不存在")
	})
}

func TestCardService_List(t *testing.T) {
	setupCardTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	product := createTestProductForCard(t, "List商品", 10.00, 0)
	importResult, err := svc.ImportCards(product.ID, []string{"LIST-001", "LIST-002", "LIST-003"})
	require.NoError(t, err)
	require.Equal(t, 3, importResult.Imported)

	t.Run("list all cards", func(t *testing.T) {
		result, err := svc.List(1, 10, -1, -1)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(3), result.Total)
		assert.Len(t, result.Cards, 3)
	})

	t.Run("list with product filter", func(t *testing.T) {
		result, err := svc.List(1, 10, int(product.ID), -1)
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
	})

	t.Run("list with pagination", func(t *testing.T) {
		result, err := svc.List(1, 2, -1, -1)
		require.NoError(t, err)
		assert.Len(t, result.Cards, 2)

		result2, err := svc.List(2, 2, -1, -1)
		require.NoError(t, err)
		assert.Len(t, result2.Cards, 1)
	})
}

func TestCardService_Delete(t *testing.T) {
	setupCardTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	product := createTestProductForCard(t, "Delete商品", 10.00, 0)
	importResult, err := svc.ImportCards(product.ID, []string{"DEL-001"})
	require.NoError(t, err)
	require.Equal(t, 1, importResult.Imported)

	var card model.Card
	require.NoError(t, database.DB.Where("product_id = ?", product.ID).First(&card).Error)

	err = svc.Delete(card.ID)
	require.NoError(t, err)

	_, err = svc.GetByID(card.ID)
	assert.Error(t, err)
}

func TestCardService_CountByProduct(t *testing.T) {
	setupCardTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	product := createTestProductForCard(t, "Count商品", 10.00, 0)
	_, err := svc.ImportCards(product.ID, []string{"CNT-001", "CNT-002", "CNT-003"})
	require.NoError(t, err)

	count, err := svc.CountByProduct(product.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	count, err = svc.CountByProduct(9999)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestCardService_ExportCards(t *testing.T) {
	setupCardTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	product := createTestProductForCard(t, "Export商品", 10.00, 0)
	_, err := svc.ImportCards(product.ID, []string{"EXP-001", "EXP-002"})
	require.NoError(t, err)

	t.Run("export all cards", func(t *testing.T) {
		items, err := svc.ExportCards(-1, -1)
		require.NoError(t, err)
		assert.Len(t, items, 2)
		for _, item := range items {
			assert.NotEmpty(t, item.CardNo)
			assert.NotEmpty(t, item.StatusTxt)
		}
	})

	t.Run("export with product filter", func(t *testing.T) {
		items, err := svc.ExportCards(int(product.ID), -1)
		require.NoError(t, err)
		assert.Len(t, items, 2)
	})

	t.Run("export with status filter", func(t *testing.T) {
		items, err := svc.ExportCards(-1, model.CardStatusUnsold)
		require.NoError(t, err)
		assert.Len(t, items, 2)

		items, err = svc.ExportCards(-1, model.CardStatusSold)
		require.NoError(t, err)
		assert.Len(t, items, 0)
	})
}

func TestCardService_SearchByCardNo(t *testing.T) {
	setupCardTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	t.Run("search by nonexistent card no", func(t *testing.T) {
		card, err := svc.SearchByCardNo("NONEXISTENT-CARD-12345")
		assert.Error(t, err)
		assert.Nil(t, card)
	})

	t.Run("search by card number not found due to encryption", func(t *testing.T) {
		product := createTestProductForCard(t, "Search商品", 10.00, 0)
		_, err := svc.ImportCards(product.ID, []string{"SEARCH-KEY-001"})
		require.NoError(t, err)

		card, err := svc.SearchByCardNo("SEARCH-KEY-001")
		assert.Error(t, err)
		assert.Nil(t, card)
	})
}

func TestCardService_EncryptDecryptRoundtrip(t *testing.T) {
	svc := NewCardService()

	cardNo := "MY-SECRET-CARD-NUMBER-12345"

	encrypted, err := svc.encryptCardNo(cardNo)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)
	assert.NotEqual(t, cardNo, encrypted)

	decrypted, err := svc.decryptCardNo(encrypted)
	require.NoError(t, err)
	assert.Equal(t, cardNo, decrypted)
}

func TestCardService_DecryptInvalidData(t *testing.T) {
	svc := NewCardService()

	decrypted, err := svc.decryptCardNo("invalid-base64-data!!!")
	assert.Error(t, err)
	assert.Empty(t, decrypted)
}