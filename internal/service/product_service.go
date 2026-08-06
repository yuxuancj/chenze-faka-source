package service

import (
	"errors"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"

	"gorm.io/gorm"
)

type ProductService struct{}

func NewProductService() *ProductService {
	return &ProductService{}
}

type CreateProductRequest struct {
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	Image       string  `json:"image"`
	Stock       int     `json:"stock"`
	Status      int     `json:"status"`
	Sort        int     `json:"sort"`
}

type UpdateProductRequest struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	Image       string  `json:"image"`
	Status      int     `json:"status"`
	Sort        int     `json:"sort"`
}

type ProductListResult struct {
	Products []model.ProductVO `json:"products"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

func (s *ProductService) Create(req *CreateProductRequest) (*model.Product, error) {
	if req.Name == "" {
		return nil, errors.New("产品名称不能为空")
	}
	if req.Price <= 0 {
		return nil, errors.New("产品价格必须大于0")
	}

	product := &model.Product{
		Name:        req.Name,
		Category:    req.Category,
		Price:       req.Price,
		Description: req.Description,
		Image:       req.Image,
		Stock:       req.Stock,
		Status:      req.Status,
		Sort:        req.Sort,
	}

	if product.Status == 0 {
		product.Status = model.ProductStatusOnShelf
	}

	if database.DB == nil {
		return nil, errors.New("数据库未连接")
	}

	if err := database.DB.Create(product).Error; err != nil {
		return nil, err
	}

	return product, nil
}

func (s *ProductService) Update(req *UpdateProductRequest) (*model.Product, error) {
	if database.DB == nil {
		return nil, errors.New("数据库未连接")
	}

	var product model.Product
	if err := database.DB.First(&product, req.ID).Error; err != nil {
		return nil, errors.New("产品不存在")
	}

	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Category != "" {
		product.Category = req.Category
	}
	if req.Price > 0 {
		product.Price = req.Price
	}
	product.Description = req.Description
	if req.Image != "" {
		product.Image = req.Image
	}
	if req.Status != 0 {
		product.Status = req.Status
	}
	product.Sort = req.Sort

	if err := database.DB.Save(&product).Error; err != nil {
		return nil, err
	}

	return &product, nil
}

func (s *ProductService) Delete(id uint) error {
	if database.DB == nil {
		return errors.New("数据库未连接")
	}
	return database.DB.Delete(&model.Product{}, id).Error
}

func (s *ProductService) GetByID(id uint) (*model.Product, error) {
	if database.DB == nil {
		return nil, errors.New("数据库未连接")
	}

	var product model.Product
	if err := database.DB.First(&product, id).Error; err != nil {
		return nil, errors.New("产品不存在")
	}
	return &product, nil
}

func (s *ProductService) List(page, pageSize int, keyword string) (*ProductListResult, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	if database.DB == nil {
		return s.mockList(page, pageSize, keyword), nil
	}

	var products []model.Product
	var total int64

	query := database.DB.Model(&model.Product{})
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("sort ASC, id DESC").Offset(offset).Limit(pageSize).Find(&products).Error; err != nil {
		return nil, err
	}

	vos := make([]model.ProductVO, 0, len(products))
	for _, p := range products {
		vos = append(vos, p.ToVO())
	}

	return &ProductListResult{
		Products: vos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *ProductService) ListOnShelf() ([]model.ProductVO, error) {
	if database.DB == nil {
		result, _ := s.mockOnShelf()
		return result, nil
	}

	var products []model.Product
	err := database.DB.Where("status = ?", model.ProductStatusOnShelf).
		Order("sort ASC, id DESC").Find(&products).Error
	if err != nil {
		return nil, err
	}

	vos := make([]model.ProductVO, 0, len(products))
	for _, p := range products {
		vos = append(vos, p.ToVO())
	}
	return vos, nil
}

func (s *ProductService) ListOnShelfGrouped() ([]model.ProductGroup, error) {
	if database.DB == nil {
		result, _ := s.mockOnShelfGrouped()
		return result, nil
	}

	var products []model.Product
	err := database.DB.Where("status = ?", model.ProductStatusOnShelf).
		Order("category ASC, sort ASC, id DESC").Find(&products).Error
	if err != nil {
		return nil, err
	}

	groups := make(map[string][]model.ProductVO)
	for _, p := range products {
		cat := p.Category
		if cat == "" {
			cat = "其他"
		}
		groups[cat] = append(groups[cat], p.ToVO())
	}

	result := make([]model.ProductGroup, 0)
	for cat, items := range groups {
		result = append(result, model.ProductGroup{
			Category: cat,
			Products: items,
		})
	}
	return result, nil
}

func (s *ProductService) UpdateStock(id uint, delta int) error {
	if database.DB == nil {
		return errors.New("数据库未连接")
	}
	return database.DB.Model(&model.Product{}).Where("id = ?", id).
		UpdateColumn("stock", gorm.Expr("stock + ?", delta)).Error
}

func (s *ProductService) mockList(page, pageSize int, keyword string) *ProductListResult {
	mock := mockProducts()
	var filtered []model.ProductVO
	for _, p := range mock {
		if keyword == "" || containsStr(p.Name, keyword) {
			filtered = append(filtered, p)
		}
	}
	total := int64(len(filtered))
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > int(total) {
		end = int(total)
	}
	if start > int(total) {
		start = int(total)
	}
	return &ProductListResult{
		Products: filtered[start:end],
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}

func (s *ProductService) mockOnShelf() ([]model.ProductVO, error) {
	var result []model.ProductVO
	for _, p := range mockProducts() {
		if p.Status == model.ProductStatusOnShelf {
			result = append(result, p)
		}
	}
	return result, nil
}

func (s *ProductService) mockOnShelfGrouped() ([]model.ProductGroup, error) {
	var products []model.ProductVO
	for _, p := range mockProducts() {
		if p.Status == model.ProductStatusOnShelf {
			products = append(products, p)
		}
	}

	groups := make(map[string][]model.ProductVO)
	for _, p := range products {
		cat := p.Category
		if cat == "" {
			cat = "其他"
		}
		groups[cat] = append(groups[cat], p)
	}

	var result []model.ProductGroup
	for cat, items := range groups {
		result = append(result, model.ProductGroup{Category: cat, Products: items})
	}
	return result, nil
}

func mockProducts() []model.ProductVO {
	return []model.ProductVO{
		{ID: 1, Name: "QQ币-10个", Category: "虚拟币", Price: 10.00, Description: "QQ币10个,自动发货", Image: "", Stock: 9999, Status: model.ProductStatusOnShelf, Sort: 1},
		{ID: 2, Name: "QQ币-50个", Category: "虚拟币", Price: 50.00, Description: "QQ币50个,自动发货", Image: "", Stock: 9999, Status: model.ProductStatusOnShelf, Sort: 2},
		{ID: 3, Name: "QQ币-100个", Category: "虚拟币", Price: 100.00, Description: "QQ币100个,自动发货", Image: "", Stock: 9999, Status: model.ProductStatusOnShelf, Sort: 3},
		{ID: 4, Name: "爱奇艺VIP月卡", Category: "影视会员", Price: 15.00, Description: "爱奇艺黄金VIP月卡", Image: "", Stock: 500, Status: model.ProductStatusOnShelf, Sort: 4},
		{ID: 5, Name: "腾讯视频VIP月卡", Category: "影视会员", Price: 20.00, Description: "腾讯视频VIP月卡", Image: "", Stock: 500, Status: model.ProductStatusOnShelf, Sort: 5},
		{ID: 6, Name: "哔哩哔哩大会员月卡", Category: "影视会员", Price: 25.00, Description: "B站大会员月卡", Image: "", Stock: 300, Status: model.ProductStatusOnShelf, Sort: 6},
		{ID: 7, Name: "网易云音乐会员月卡", Category: "音乐会员", Price: 18.00, Description: "网易云音乐黑胶会员月卡", Image: "", Stock: 200, Status: model.ProductStatusOnShelf, Sort: 7},
		{ID: 8, Name: "百度网盘超级会员月卡", Category: "实用工具", Price: 30.00, Description: "百度网盘超级会员月卡", Image: "", Stock: 100, Status: model.ProductStatusOnShelf, Sort: 8},
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}