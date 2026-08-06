package service

import (
	"bufio"
	"errors"
	"fmt"
	"strings"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/crypto"
	"chenze-faka/internal/pkg/database"

	"gorm.io/gorm"
)

type CardService struct {
	encryptKey string
}

func NewCardService() *CardService {
	key := "chenze_faka_card_encrypt_key_2024!!"
	return &CardService{encryptKey: key}
}

type ImportCardsResult struct {
	TotalCount int      `json:"total_count"`
	Imported   int      `json:"imported"`
	Skipped    int      `json:"skipped"`
	Errors     []string `json:"errors,omitempty"`
}

type CardListResult struct {
	Cards    []model.Card `json:"cards"`
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
}

func (s *CardService) encryptCardNo(cardNo string) (string, error) {
	return crypto.AesEncrypt(cardNo, s.encryptKey)
}

func (s *CardService) decryptCardNo(encrypted string) (string, error) {
	return crypto.AesDecrypt(encrypted, s.encryptKey)
}

func (s *CardService) ImportCards(productID uint, cardNos []string) (*ImportCardsResult, error) {
	if database.DB == nil {
		return nil, errors.New("数据库未连接")
	}

	var product model.Product
	if err := database.DB.First(&product, productID).Error; err != nil {
		return nil, errors.New("产品不存在")
	}

	result := &ImportCardsResult{
		TotalCount: len(cardNos),
		Errors:     make([]string, 0),
	}

	for _, cardNo := range cardNos {
		cardNo = strings.TrimSpace(cardNo)
		if cardNo == "" {
			result.Skipped++
			continue
		}

		encrypted, err := s.encryptCardNo(cardNo)
		if err != nil {
			result.Errors = append(result.Errors, "加密失败: "+cardNo)
			result.Skipped++
			continue
		}

		var existing model.Card
		err = database.DB.Where("card_no = ?", encrypted).First(&existing).Error
		if err == nil {
			result.Skipped++
			continue
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			result.Errors = append(result.Errors, "重复卡密: "+cardNo)
			result.Skipped++
			continue
		}

		card := &model.Card{
			ProductID: product.ID,
			CardNoHash: encrypted,
			Status:     model.CardStatusUnsold,
		}

		if err := database.DB.Create(card).Error; err != nil {
			result.Errors = append(result.Errors, "创建失败: "+cardNo)
			continue
		}
		result.Imported++
	}

	database.DB.Model(&model.Product{}).Where("id = ?", product.ID).
		UpdateColumn("stock", gorm.Expr("stock + ?", result.Imported))

	return result, nil
}

func (s *CardService) ParseCardText(text string) []string {
	var cards []string
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			cards = append(cards, line)
		}
	}
	return cards
}

func (s *CardService) GetByID(id uint) (*model.Card, error) {
	if database.DB == nil {
		return nil, errors.New("数据库未连接")
	}

	var card model.Card
	if err := database.DB.First(&card, id).Error; err != nil {
		return nil, errors.New("卡密不存在")
	}
	decrypted, err := s.decryptCardNo(card.CardNoHash)
	if err == nil {
		card.CardNo = decrypted
	} else {
		card.CardNo = card.CardNoHash
	}
	return &card, nil
}

func (s *CardService) List(page, pageSize, productID int, status int) (*CardListResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	if database.DB == nil {
		return &CardListResult{Cards: []model.Card{}, Total: 0, Page: page, PageSize: pageSize}, nil
	}

	var cards []model.Card
	var total int64

	query := database.DB.Model(&model.Card{})
	if productID > 0 {
		query = query.Where("product_id = ?", productID)
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&cards).Error; err != nil {
		return nil, err
	}

	for i := range cards {
		decrypted, err := s.decryptCardNo(cards[i].CardNoHash)
		if err == nil {
			cards[i].CardNo = decrypted
		} else {
			cards[i].CardNo = cards[i].CardNoHash
		}
	}

	return &CardListResult{
		Cards:    cards,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *CardService) Delete(id uint) error {
	if database.DB == nil {
		return errors.New("数据库未连接")
	}
	return database.DB.Delete(&model.Card{}, id).Error
}

func (s *CardService) CountByProduct(productID uint) (int64, error) {
	if database.DB == nil {
		return 0, nil
	}

	var count int64
	err := database.DB.Model(&model.Card{}).
		Where("product_id = ? AND status = ?", productID, model.CardStatusUnsold).
		Count(&count).Error
	return count, err
}

func (s *CardService) GetAvailableCards(productID uint, limit int) ([]model.Card, error) {
	if database.DB == nil {
		return nil, errors.New("数据库未连接")
	}

	var cards []model.Card
	err := database.DB.Where("product_id = ? AND status = ?", productID, model.CardStatusUnsold).
		Limit(limit).Find(&cards).Error
	if err != nil {
		return nil, err
	}

	for i := range cards {
		decrypted, err := s.decryptCardNo(cards[i].CardNoHash)
		if err == nil {
			cards[i].CardNo = decrypted
		}
	}
	return cards, nil
}

func (s *CardService) MarkAsSold(id uint, orderNo string) error {
	if database.DB == nil {
		return errors.New("数据库未连接")
	}
	return database.DB.Model(&model.Card{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":   model.CardStatusSold,
			"order_no": orderNo,
			"sold_at":  gorm.Expr("NOW()"),
		}).Error
}

func (s *CardService) ExportCards(productID, status int) ([]model.CardExportItem, error) {
	if database.DB == nil {
		return nil, errors.New("数据库未连接")
	}
	var cards []model.Card
	query := database.DB.Model(&model.Card{})
	if productID > 0 {
		query = query.Where("product_id = ?", productID)
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	if err := query.Order("id DESC").Find(&cards).Error; err != nil {
		return nil, err
	}

	var items []model.CardExportItem
	for _, c := range cards {
		no := c.CardNoHash
		decrypted, err := s.decryptCardNo(c.CardNoHash)
		if err == nil {
			no = decrypted
		}
		statusTxt := "未使用"
		if c.Status == model.CardStatusSold {
			statusTxt = "已售出"
		}
		soldAt := ""
		if c.SoldAt != nil {
			soldAt = c.SoldAt.Format("2006-01-02 15:04:05")
		}
		items = append(items, model.CardExportItem{
			ID:        c.ID,
			CardNo:    no,
			Status:    c.Status,
			StatusTxt: statusTxt,
			OrderNo:   c.OrderNo,
			SoldAt:    soldAt,
		})
	}
	return items, nil
}

func (s *CardService) SearchByCardNo(keyword string) (*model.Card, error) {
	if database.DB == nil {
		return nil, errors.New("数据库未连接")
	}
	encrypted, err := s.encryptCardNo(keyword)
	if err != nil {
		return nil, err
	}
	var card model.Card
	err = database.DB.Where("card_no = ?", encrypted).First(&card).Error
	if err != nil {
		return nil, fmt.Errorf("卡密不存在")
	}
	decrypted, _ := s.decryptCardNo(card.CardNoHash)
	card.CardNo = decrypted
	return &card, nil
}
