package service

import (
	"errors"

	"chenze-faka/internal/model"
	"chenze-faka/internal/repository"
)

type ProductService struct {
	productRepo *repository.ProductRepository
}

func NewProductService() *ProductService {
	return &ProductService{
		productRepo: repository.NewProductRepository(),
	}
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

type ProductListResponse struct {
	Products []model.Product `json:"products"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

func (s *ProductService) Create(req *CreateProductRequest) (*model.Product, error) {
	if req.Name == "" {
		return nil, errors.New("product name cannot be empty")
	}
	if req.Price <= 0 {
		return nil, errors.New("price must be greater than 0")
	}

	product := &model.Product{
		Name:        req.Name,
		Category:    req.Category,
		Price:       req.Price,
		Description: req.Description,
		Image:       req.Image,
		Stock:       req.Stock,
		Status:      model.ProductStatusOnShelf,
		Sort:        req.Sort,
	}

	if err := s.productRepo.Create(product); err != nil {
		return nil, err
	}

	return product, nil
}

func (s *ProductService) Update(req *UpdateProductRequest) (*model.Product, error) {
	product, err := s.productRepo.GetByID(req.ID)
	if err != nil {
		return nil, errors.New("product not found")
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

	if err := s.productRepo.Update(product); err != nil {
		return nil, err
	}

	return product, nil
}

func (s *ProductService) Delete(id uint) error {
	return s.productRepo.Delete(id)
}

func (s *ProductService) GetByID(id uint) (*model.Product, error) {
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("product not found")
	}
	return product, nil
}

func (s *ProductService) List(page, pageSize int, keyword string) (*ProductListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	products, total, err := s.productRepo.List(page, pageSize, keyword)
	if err != nil {
		return nil, err
	}

	return &ProductListResponse{
		Products: products,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *ProductService) ListOnShelf() ([]model.Product, error) {
	return s.productRepo.ListOnShelf()
}

func (s *ProductService) ListOnShelfGrouped() ([]map[string]interface{}, error) {
	return s.productRepo.ListOnShelfGrouped()
}

func (s *ProductService) UpdateStock(id uint, delta int) error {
	return s.productRepo.UpdateStock(id, delta)
}
