package repository

import (
	"chenze-faka/internal/model"
	"chenze-faka/pkg/database"

	"gorm.io/gorm"
)

type CardRepository struct {
	db *gorm.DB
}

func NewCardRepository() *CardRepository {
	return &CardRepository{db: database.DB}
}

func (r *CardRepository) Create(card *model.Card) error {
	return r.db.Create(card).Error
}

func (r *CardRepository) BatchCreate(cards []*model.Card) error {
	return r.db.CreateInBatches(cards, 100).Error
}

func (r *CardRepository) Update(card *model.Card) error {
	return r.db.Save(card).Error
}

func (r *CardRepository) GetByID(id uint) (*model.Card, error) {
	var card model.Card
	err := r.db.First(&card, id).Error
	if err != nil {
		return nil, err
	}
	return &card, nil
}

func (r *CardRepository) GetByCardNo(cardNo string) (*model.Card, error) {
	var card model.Card
	err := r.db.Where("card_no = ?", cardNo).First(&card).Error
	if err != nil {
		return nil, err
	}
	return &card, nil
}

func (r *CardRepository) List(page, pageSize int, productID uint, status int) ([]model.Card, int64, error) {
	var cards []model.Card
	var total int64

	query := r.db.Model(&model.Card{})
	if productID > 0 {
		query = query.Where("product_id = ?", productID)
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&cards).Error; err != nil {
		return nil, 0, err
	}

	return cards, total, nil
}

func (r *CardRepository) GetAvailableCards(productID uint, limit int) ([]*model.Card, error) {
	var cards []*model.Card
	err := r.db.Where("product_id = ? AND status = ?", productID, model.CardStatusUnused).
		Limit(limit).Find(&cards).Error
	return cards, err
}

func (r *CardRepository) MarkAsSold(id uint, orderNo string) error {
	return r.db.Model(&model.Card{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":   model.CardStatusSold,
			"order_no": orderNo,
			"used_at":  gorm.Expr("NOW()"),
		}).Error
}

func (r *CardRepository) CountByProduct(productID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Card{}).Where("product_id = ? AND status = ?", productID, model.CardStatusUnused).Count(&count).Error
	return count, err
}

func (r *CardRepository) Delete(id uint) error {
	return r.db.Delete(&model.Card{}, id).Error
}
