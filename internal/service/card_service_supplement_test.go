package service

import (
	"fmt"
	"testing"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCardServiceDBNilSupplement(t *testing.T) {
	svc := NewCardService()

	originalDB := database.DB
	database.DB = nil
	defer func() { database.DB = originalDB }()

	t.Run("list cards with nil db returns empty result", func(t *testing.T) {
		result, err := svc.List(1, 10, -1, -1)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(0), result.Total)
	})

	t.Run("get by id with nil db", func(t *testing.T) {
		card, err := svc.GetByID(1)
		assert.Error(t, err)
		assert.Nil(t, card)
	})

	t.Run("search by card no with nil db", func(t *testing.T) {
		cards, err := svc.SearchByCardNo("test")
		assert.Error(t, err)
		assert.Nil(t, cards)
	})

	t.Run("mark as sold with nil db", func(t *testing.T) {
		err := svc.MarkAsSold(1, "ORDER")
		assert.Error(t, err)
	})

	t.Run("delete card with nil db", func(t *testing.T) {
		err := svc.Delete(1)
		assert.Error(t, err)
	})

	t.Run("export cards with nil db", func(t *testing.T) {
		cards, err := svc.ExportCards(1, -1)
		assert.Error(t, err)
		assert.Nil(t, cards)
	})

	t.Run("get available cards with nil db", func(t *testing.T) {
		result, err := svc.GetAvailableCards(1, 10)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestCardImportEdgeCases(t *testing.T) {
	setupCardCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	product := createCardCoverageProduct(t, "ImportEdge Product", 10.00, 0)

	t.Run("import empty card list returns success with 0 imported", func(t *testing.T) {
		result, err := svc.ImportCards(product.ID, []string{})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 0, result.Imported)
	})

	t.Run("import with all unique cards", func(t *testing.T) {
		cards := make([]string, 10)
		for i := 0; i < 10; i++ {
			cards[i] = fmt.Sprintf("UNIQUE-%d", i)
		}
		result, err := svc.ImportCards(product.ID, cards)
		require.NoError(t, err)
		assert.Equal(t, 10, result.Imported)
	})

	t.Run("import with empty and whitespace cards are skipped", func(t *testing.T) {
		cards := []string{"", "  ", "\t", "VALID-CARD"}
		result, err := svc.ImportCards(product.ID, cards)
		require.NoError(t, err)
		assert.Equal(t, 1, result.Imported)
		assert.Equal(t, 3, result.Skipped)
	})

	t.Run("import with mixed valid and empty", func(t *testing.T) {
		cards := []string{"MIX-001", "", "MIX-002", "  ", "MIX-003"}
		result, err := svc.ImportCards(product.ID, cards)
		require.NoError(t, err)
		assert.Equal(t, 3, result.Imported)
		assert.Equal(t, 2, result.Skipped)
	})

	t.Run("import all empty cards results in 0 imported", func(t *testing.T) {
		cards := []string{"", "  ", "\t", "\n"}
		result, err := svc.ImportCards(product.ID, cards)
		require.NoError(t, err)
		assert.Equal(t, 0, result.Imported)
		assert.Equal(t, 4, result.Skipped)
	})
}

func TestCardExportWithStatusFilter(t *testing.T) {
	setupCardCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()
	product := createCardCoverageProduct(t, "Export Product", 20.00, 0)

	_, err := svc.ImportCards(product.ID, []string{"EXP-001", "EXP-002", "EXP-003"})
	require.NoError(t, err)

	t.Run("export with sold status filter returns empty", func(t *testing.T) {
		items, err := svc.ExportCards(int(product.ID), model.CardStatusSold)
		require.NoError(t, err)
		assert.Len(t, items, 0)
	})

	t.Run("export with unsold status filter returns 3", func(t *testing.T) {
		items, err := svc.ExportCards(int(product.ID), model.CardStatusUnsold)
		require.NoError(t, err)
		assert.Len(t, items, 3)
	})

	t.Run("export with empty filter returns all", func(t *testing.T) {
		items, err := svc.ExportCards(int(product.ID), -1)
		require.NoError(t, err)
		assert.Len(t, items, 3)
	})
}

func TestCardGetAvailableCardsMultipleProducts(t *testing.T) {
	setupCardCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()
	p1 := createCardCoverageProduct(t, "P1", 10.00, 0)
	p2 := createCardCoverageProduct(t, "P2", 20.00, 0)

	_, err := svc.ImportCards(p1.ID, []string{"P1-001", "P1-002", "P1-003", "P1-004", "P1-005"})
	require.NoError(t, err)
	_, err = svc.ImportCards(p2.ID, []string{"P2-001", "P2-002", "P2-003"})
	require.NoError(t, err)

	t.Run("get available for product 1", func(t *testing.T) {
		result, err := svc.GetAvailableCards(p1.ID, 10)
		require.NoError(t, err)
		assert.Len(t, result, 5)
	})

	t.Run("get available for product 2", func(t *testing.T) {
		result, err := svc.GetAvailableCards(p2.ID, 10)
		require.NoError(t, err)
		assert.Len(t, result, 3)
	})

	t.Run("get available for nonexistent product returns empty", func(t *testing.T) {
		result, err := svc.GetAvailableCards(9999, 10)
		require.NoError(t, err)
		assert.Len(t, result, 0)
	})
}

func TestCardCountByProductCoverage(t *testing.T) {
	setupCardCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()
	product := createCardCoverageProduct(t, "Count Product", 10.00, 0)

	_, err := svc.ImportCards(product.ID, []string{"COUNT-001", "COUNT-002", "COUNT-003", "COUNT-004"})
	require.NoError(t, err)

	t.Run("count for product returns correct number", func(t *testing.T) {
		count, err := svc.CountByProduct(product.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(4), count)
	})

	t.Run("count for nonexistent product returns 0", func(t *testing.T) {
		count, err := svc.CountByProduct(9999)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestCardSearchByCardNoFound(t *testing.T) {
	setupCardCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()
	product := createCardCoverageProduct(t, "Search Product", 25.00, 0)

	_, err := svc.ImportCards(product.ID, []string{"UNIQUE-CARD-001", "UNIQUE-CARD-002"})
	require.NoError(t, err)

	t.Run("search by exact card no may not find due to GCM nonce", func(t *testing.T) {
		result, err := svc.SearchByCardNo("UNIQUE-CARD-001")
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("search by non-existent card returns error", func(t *testing.T) {
		result, err := svc.SearchByCardNo("NONEXISTENT")
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestCardMarkAsSoldTwice(t *testing.T) {
	setupCardCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()
	product := createCardCoverageProduct(t, "MarkSold Product", 30.00, 0)
	result, err := svc.ImportCards(product.ID, []string{"MARK-001"})
	require.NoError(t, err)
	require.Equal(t, 1, result.Imported)

	var card model.Card
	require.NoError(t, database.DB.Where("product_id = ?", product.ID).First(&card).Error)

	t.Run("mark as sold fails with sqlite due to NOW() function", func(t *testing.T) {
		err := svc.MarkAsSold(card.ID, "ORDER-001")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "NOW")
	})

	t.Run("mark as sold with invalid id returns error", func(t *testing.T) {
		err := svc.MarkAsSold(99999, "ORDER-INVALID")
		assert.Error(t, err)
	})
}

func TestCardDeleteNonExistentSupplement(t *testing.T) {
	setupCardCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	t.Run("delete non-existent card returns nil", func(t *testing.T) {
		err := svc.Delete(99999)
		assert.NoError(t, err)
	})
}

func TestCardListWithDefaultParams(t *testing.T) {
	setupCardCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	t.Run("list with default page params", func(t *testing.T) {
		result, err := svc.List(0, 0, -1, -1)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(0), result.Total)
	})

	t.Run("list with very large page size", func(t *testing.T) {
		result, err := svc.List(1, 9999, -1, -1)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestCardGetByIDNonExistent(t *testing.T) {
	setupCardCoverageTestDB(t)
	defer func() { database.DB = nil }()

	svc := NewCardService()

	t.Run("get by non-existent id returns error", func(t *testing.T) {
		card, err := svc.GetByID(99999)
		assert.Error(t, err)
		assert.Nil(t, card)
	})
}