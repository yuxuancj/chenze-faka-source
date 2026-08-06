package repository

import (
	"chenze-faka/internal/model"
	"chenze-faka/pkg/database"

	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository() *ProductRepository {
	return &ProductRepository{db: database.DB}
}

func (r *ProductRepository) Create(product *model.Product) error {
	return r.db.Create(product).Error
}

func (r *ProductRepository) Update(product *model.Product) error {
	return r.db.Save(product).Error
}

func (r *ProductRepository) Delete(id uint) error {
	return r.db.Delete(&model.Product{}, id).Error
}

func (r *ProductRepository) GetByID(id uint) (*model.Product, error) {
	var product model.Product
	err := r.db.First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) List(page, pageSize int, keyword string) ([]model.Product, int64, error) {
	if r.db == nil {
		return []model.Product{}, 0, nil
	}
	var products []model.Product
	var total int64

	query := r.db.Model(&model.Product{})
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("sort ASC, id DESC").Offset(offset).Limit(pageSize).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *ProductRepository) ListOnShelf() ([]model.Product, error) {
	if r.db == nil {
		return []model.Product{}, nil
	}
	var products []model.Product
	err := r.db.Where("status = ?", model.ProductStatusOnShelf).Order("sort ASC, id DESC").Find(&products).Error
	return products, err
}

func (r *ProductRepository) ListOnShelfGrouped() ([]map[string]interface{}, error) {
	if r.db == nil {
		return []map[string]interface{}{}, nil
	}
	var products []model.Product
	err := r.db.Where("status = ?", model.ProductStatusOnShelf).Order("category ASC, sort ASC, id DESC").Find(&products).Error
	if err != nil {
		return nil, err
	}

	groups := make(map[string][]model.Product)
	for _, p := range products {
		cat := p.Category
		if cat == "" {
			cat = "其他"
		}
		groups[cat] = append(groups[cat], p)
	}

	result := make([]map[string]interface{}, 0)
	for cat, items := range groups {
		result = append(result, map[string]interface{}{
			"category": cat,
			"products": items,
		})
	}

	return result, nil
}

func (r *ProductRepository) UpdateStock(id uint, delta int) error {
	return r.db.Model(&model.Product{}).Where("id = ?", id).
		UpdateColumn("stock", gorm.Expr("stock + ?", delta)).Error
}

func (r *ProductRepository) GetOnShelf() ([]model.Product, error) {
	return r.ListOnShelf()
}
