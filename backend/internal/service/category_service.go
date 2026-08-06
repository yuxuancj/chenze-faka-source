package service

import (
	"errors"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"
)

type CategoryService struct{}

func NewCategoryService() *CategoryService {
	return &CategoryService{}
}

type CategoryListResult struct {
	Items    []model.Category `json:"items"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

func (s *CategoryService) Create(name string, icon string, sort int) (*model.Category, error) {
	if name == "" {
		return nil, errors.New("分类名称不能为空")
	}
	if database.DB == nil {
		return nil, errors.New("数据库未连接")
	}
	cat := &model.Category{Name: name, Icon: icon, Sort: sort, Status: model.CategoryEnabled}
	if err := database.DB.Create(cat).Error; err != nil {
		return nil, err
	}
	return cat, nil
}

func (s *CategoryService) Update(id uint, name string, icon string, sort int, status int) (*model.Category, error) {
	if database.DB == nil {
		return nil, errors.New("数据库未连接")
	}
	var cat model.Category
	if err := database.DB.First(&cat, id).Error; err != nil {
		return nil, errors.New("分类不存在")
	}
	if name != "" {
		cat.Name = name
	}
	cat.Icon = icon
	cat.Sort = sort
	if status != 0 || status == model.CategoryDisabled {
		cat.Status = status
	}
	if err := database.DB.Save(&cat).Error; err != nil {
		return nil, err
	}
	return &cat, nil
}

func (s *CategoryService) Delete(id uint) error {
	if database.DB == nil {
		return errors.New("数据库未连接")
	}
	return database.DB.Delete(&model.Category{}, id).Error
}

func (s *CategoryService) GetAll() ([]model.Category, error) {
	if database.DB == nil {
		return []model.Category{}, nil
	}
	var cats []model.Category
	err := database.DB.Where("status = ?", model.CategoryEnabled).Order("sort ASC, id DESC").Find(&cats).Error
	return cats, err
}

func (s *CategoryService) List(page, pageSize int, keyword string) (*CategoryListResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if database.DB == nil {
		return &CategoryListResult{Items: []model.Category{}, Total: 0, Page: page, PageSize: pageSize}, nil
	}
	var cats []model.Category
	var total int64
	query := database.DB.Model(&model.Category{})
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	offset := (page - 1) * pageSize
	if err := query.Order("sort ASC, id DESC").Offset(offset).Limit(pageSize).Find(&cats).Error; err != nil {
		return nil, err
	}
	return &CategoryListResult{Items: cats, Total: total, Page: page, PageSize: pageSize}, nil
}
