package controller

import (
	"chenze-faka/internal/pkg/response"
	"chenze-faka/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductController struct {
	productService *service.ProductService
}

func NewProductController() *ProductController {
	return &ProductController{
		productService: service.NewProductService(),
	}
}

func (h *ProductController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")

	result, err := h.productService.List(page, pageSize, keyword)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, result)
}

func (h *ProductController) OnShelf(c *gin.Context) {
	products, err := h.productService.ListOnShelf()
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, products)
}

func (h *ProductController) OnShelfGrouped(c *gin.Context) {
	products, err := h.productService.ListOnShelfGrouped()
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, products)
}

func (h *ProductController) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, "无效的产品ID")
		return
	}

	product, err := h.productService.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, product)
}

func (h *ProductController) Create(c *gin.Context) {
	var req service.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "无效的请求参数")
		return
	}

	product, err := h.productService.Create(&req)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, product)
}

func (h *ProductController) Update(c *gin.Context) {
	var req service.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "无效的请求参数")
		return
	}

	product, err := h.productService.Update(&req)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, product)
}

func (h *ProductController) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, "无效的产品ID")
		return
	}

	if err := h.productService.Delete(uint(id)); err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, nil)
}