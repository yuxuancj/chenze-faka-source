package service

import (
	"errors"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"
)

type PaymentService struct{}

func NewPaymentService() *PaymentService {
	return &PaymentService{}
}

type PaymentListResult struct {
	Items    []model.PaymentChannel `json:"items"`
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
}

func (s *PaymentService) Create(name string, payType string, icon string, config string, feeRate float64, sort int) (*model.PaymentChannel, error) {
	if name == "" {
		return nil, errors.New("支付名称不能为空")
	}
	if database.DB == nil {
		return nil, errors.New("数据库未连接")
	}
	ch := &model.PaymentChannel{
		Name:    name,
		Type:    payType,
		Icon:    icon,
		Config:  config,
		FeeRate: feeRate,
		Status:  model.ChannelEnabled,
		Sort:    sort,
	}
	if err := database.DB.Create(ch).Error; err != nil {
		return nil, err
	}
	return ch, nil
}

func (s *PaymentService) Update(id uint, name string, icon string, config string, feeRate float64, status int, sort int) (*model.PaymentChannel, error) {
	if database.DB == nil {
		return nil, errors.New("数据库未连接")
	}
	var ch model.PaymentChannel
	if err := database.DB.First(&ch, id).Error; err != nil {
		return nil, errors.New("支付接口不存在")
	}
	if name != "" {
		ch.Name = name
	}
	ch.Icon = icon
	ch.Config = config
	ch.FeeRate = feeRate
	if status == model.ChannelEnabled || status == model.ChannelDisabled {
		ch.Status = status
	}
	ch.Sort = sort
	if err := database.DB.Save(&ch).Error; err != nil {
		return nil, err
	}
	return &ch, nil
}

func (s *PaymentService) Delete(id uint) error {
	if database.DB == nil {
		return errors.New("数据库未连接")
	}
	return database.DB.Delete(&model.PaymentChannel{}, id).Error
}

func (s *PaymentService) GetActive() ([]model.PaymentChannel, error) {
	if database.DB == nil {
		return []model.PaymentChannel{}, nil
	}
	var channels []model.PaymentChannel
	err := database.DB.Where("status = ?", model.ChannelEnabled).Order("sort ASC, id DESC").Find(&channels).Error
	return channels, err
}

func (s *PaymentService) List(page, pageSize int, keyword string) (*PaymentListResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if database.DB == nil {
		return &PaymentListResult{Items: []model.PaymentChannel{}, Total: 0, Page: page, PageSize: pageSize}, nil
	}
	var items []model.PaymentChannel
	var total int64
	query := database.DB.Model(&model.PaymentChannel{})
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	offset := (page - 1) * pageSize
	if err := query.Order("sort ASC, id DESC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, err
	}
	return &PaymentListResult{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}
