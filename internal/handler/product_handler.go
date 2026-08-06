package handler

import (
	"chenze-faka/internal/service"
	"chenze-faka/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productService *service.ProductService
}

func NewProductHandler() *ProductHandler {
	return &ProductHandler{
		productService: service.NewProductService(),
	}
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req service.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "invalid request parameters")
		return
	}

	product, err := h.productService.Create(&req)
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.Success(c, product)
}

func (h *ProductHandler) Update(c *gin.Context) {
	var req service.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "invalid request parameters")
		return
	}

	product, err := h.productService.Update(&req)
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.Success(c, product)
}

func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Fail(c, "invalid product id")
		return
	}

	if err := h.productService.Delete(uint(id)); err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Fail(c, "invalid product id")
		return
	}

	product, err := h.productService.GetByID(uint(id))
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}

	utils.Success(c, product)
}

func (h *ProductHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")

	result, err := h.productService.List(page, pageSize, keyword)
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.Success(c, result)
}

func (h *ProductHandler) ListOnShelf(c *gin.Context) {
	products, err := h.productService.ListOnShelf()
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.Success(c, products)
}

func (h *ProductHandler) ListOnShelfGrouped(c *gin.Context) {
	products, err := h.productService.ListOnShelfGrouped()
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.Success(c, products)
}
