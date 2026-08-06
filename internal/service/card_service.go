package service

import (
	"bufio"
	"errors"
	"strings"

	"chenze-faka/internal/model"
	"chenze-faka/internal/repository"

	"gorm.io/gorm"
)

type CardService struct {
	cardRepo    *repository.CardRepository
	productRepo *repository.ProductRepository
}

func NewCardService() *CardService {
	return &CardService{
		cardRepo:    repository.NewCardRepository(),
		productRepo: repository.NewProductRepository(),
	}
}

type ImportCardsResponse struct {
	TotalCount int   `json:"total_count"`
	Imported   int   `json:"imported"`
	Skipped    int   `json:"skipped"`
	Errors     []string `json:"errors,omitempty"`
}

type CardListResponse struct {
	Cards    []model.Card `json:"cards"`
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
}

func (s *CardService) ImportCards(productID uint, cardNos []string) (*ImportCardsResponse, error) {
	product, err := s.productRepo.GetByID(productID)
	if err != nil {
		return nil, errors.New("product not found")
	}

	result := &ImportCardsResponse{
		TotalCount: len(cardNos),
		Errors:     make([]string, 0),
	}

	for _, cardNo := range cardNos {
		cardNo = strings.TrimSpace(cardNo)
		if cardNo == "" {
			result.Skipped++
			continue
		}

		existingCard, err := s.cardRepo.GetByCardNo(cardNo)
		if err == nil && existingCard != nil {
			result.Skipped++
			continue
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			result.Errors = append(result.Errors, "duplicate card: "+cardNo)
			result.Skipped++
			continue
		}

		card := &model.Card{
			ProductID: product.ID,
			CardNo:    cardNo,
			Status:    model.CardStatusUnused,
		}

		if err := s.cardRepo.Create(card); err != nil {
			result.Errors = append(result.Errors, "create failed: "+cardNo)
			continue
		}
		result.Imported++
	}

	s.productRepo.UpdateStock(product.ID, int(result.Imported))
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
	card, err := s.cardRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("card not found")
	}
	return card, nil
}

func (s *CardService) List(page, pageSize int, productID uint, status int) (*CardListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	cards, total, err := s.cardRepo.List(page, pageSize, productID, status)
	if err != nil {
		return nil, err
	}

	return &CardListResponse{
		Cards:    cards,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *CardService) Delete(id uint) error {
	return s.cardRepo.Delete(id)
}

func (s *CardService) GetAvailableCards(productID uint, limit int) ([]*model.Card, error) {
	return s.cardRepo.GetAvailableCards(productID, limit)
}

func (s *CardService) MarkAsSold(id uint, orderNo string) error {
	return s.cardRepo.MarkAsSold(id, orderNo)
}

func (s *CardService) CountByProduct(productID uint) (int64, error) {
	return s.cardRepo.CountByProduct(productID)
}
