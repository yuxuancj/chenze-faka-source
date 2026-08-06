package service

import (
	"errors"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"
)

type LogService struct{}

func NewLogService() *LogService {
	return &LogService{}
}

func (s *LogService) WriteOperation(userID uint, username, action, targetType, targetID, detail, ip, userAgent string, status int) {
	if database.DB == nil {
		return
	}
	log := &model.OperationLog{
		UserID:     userID,
		Username:   username,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Detail:     detail,
		IP:         ip,
		UserAgent:  userAgent,
		Status:     status,
	}
	database.DB.Create(log)
}

func (s *LogService) WriteOrder(orderNo, action, detail, operator string) {
	if database.DB == nil {
		return
	}
	log := &model.OrderLog{
		OrderNo:  orderNo,
		Action:   action,
		Detail:   detail,
		Operator: operator,
	}
	database.DB.Create(log)
}

type OperationLogListResult struct {
	Items    []model.OperationLog `json:"items"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

func (s *LogService) ListOperation(page, pageSize int, username, action string) (*OperationLogListResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if database.DB == nil {
		return &OperationLogListResult{Items: []model.OperationLog{}, Total: 0, Page: page, PageSize: pageSize}, nil
	}
	var items []model.OperationLog
	var total int64
	query := database.DB.Model(&model.OperationLog{})
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, err
	}
	return &OperationLogListResult{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

type OrderLogListResult struct {
	Items    []model.OrderLog `json:"items"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

func (s *LogService) ListOrder(orderNo string, page, pageSize int) (*OrderLogListResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if database.DB == nil {
		return &OrderLogListResult{Items: []model.OrderLog{}, Total: 0, Page: page, PageSize: pageSize}, nil
	}
	var items []model.OrderLog
	var total int64
	query := database.DB.Model(&model.OrderLog{})
	if orderNo != "" {
		query = query.Where("order_no = ?", orderNo)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, err
	}
	return &OrderLogListResult{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

type EmailLogListResult struct {
	Items    []model.EmailLog `json:"items"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

func (s *LogService) ListEmail(page, pageSize int, email string) (*EmailLogListResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if database.DB == nil {
		return &EmailLogListResult{Items: []model.EmailLog{}, Total: 0, Page: page, PageSize: pageSize}, nil
	}
	var items []model.EmailLog
	var total int64
	query := database.DB.Model(&model.EmailLog{})
	if email != "" {
		query = query.Where("to_email LIKE ?", "%"+email+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, err
	}
	return &EmailLogListResult{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

type LoginLogListResult struct {
	Items    []model.OperationLog `json:"items"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

func (s *LogService) ListLogin(page, pageSize int, username string) (*LoginLogListResult, error) {
	if database.DB == nil {
		return nil, errors.New("数据库未连接")
	}
	var items []model.OperationLog
	var total int64
	query := database.DB.Model(&model.OperationLog{}).Where("action = ?", "login")
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, err
	}
	return &LoginLogListResult{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}
