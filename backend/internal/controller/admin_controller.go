package controller

import (
	"chenze-faka/internal/pkg/response"
	"chenze-faka/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminController struct {
	productService *service.ProductService
	orderService   *service.OrderService
	cardService    *service.CardService
	licenseSvc     *service.LicenseService
}

func NewAdminController(licenseSvc *service.LicenseService) *AdminController {
	return &AdminController{
		productService: service.NewProductService(),
		orderService:   service.NewOrderService(),
		cardService:    service.NewCardService(),
		licenseSvc:     licenseSvc,
	}
}

func (h *AdminController) SystemStatus(c *gin.Context) {
	response.Success(c, gin.H{
		"version":   "1.0.0",
		"site_name": c.GetString("site_name"),
	})
}

func (h *AdminController) LicenseStatus(c *gin.Context) {
	if h.licenseSvc == nil {
		response.Success(c, gin.H{
			"status": "unknown",
		})
		return
	}

	cache := h.licenseSvc.GetCache()
	isVerified := h.licenseSvc.IsVerified()
	isGraceValid := h.licenseSvc.IsGracePeriodValid()

	status := "invalid"
	if isVerified {
		status = "valid"
	} else if isGraceValid {
		status = "grace_period"
	}

	response.Success(c, gin.H{
		"status":       status,
		"verified":     isVerified,
		"grace_valid":  isGraceValid,
		"last_verify":  cache.LastVerifyTime,
		"last_success": cache.LastSuccessTime,
		"last_result":  cache.LastResult,
		"last_message": cache.LastMessage,
		"expire_at":    cache.ExpireAt,
		"app_name":     cache.AppName,
	})
}

func (h *AdminController) VerifyLicense(c *gin.Context) {
	if h.licenseSvc == nil {
		response.Fail(c, "授权服务未初始化")
		return
	}

	ok, err := h.licenseSvc.Verify()
	if err != nil {
		response.FailWithCode(c, http.StatusInternalServerError, "验证错误: "+err.Error())
		return
	}

	response.Success(c, gin.H{"verified": ok})
}

func (h *AdminController) ProductList(c *gin.Context) {
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

func (h *AdminController) ProductCreate(c *gin.Context) {
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

func (h *AdminController) ProductUpdate(c *gin.Context) {
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

func (h *AdminController) ProductDelete(c *gin.Context) {
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

func (h *AdminController) CardList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	productID, _ := strconv.Atoi(c.Query("product_id"))
	status, _ := strconv.Atoi(c.Query("status"))

	result, err := h.cardService.List(page, pageSize, productID, status)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, result)
}

func (h *AdminController) CardImport(c *gin.Context) {
	var req struct {
		ProductID uint   `json:"product_id" binding:"required"`
		CardText  string `json:"card_text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "无效的请求参数")
		return
	}

	cardNos := h.cardService.ParseCardText(req.CardText)
	if len(cardNos) == 0 {
		response.Fail(c, "没有有效的卡密")
		return
	}

	result, err := h.cardService.ImportCards(req.ProductID, cardNos)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, result)
}

func (h *AdminController) CardDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, "无效的卡密ID")
		return
	}

	if err := h.cardService.Delete(uint(id)); err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *AdminController) OrderList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status, _ := strconv.Atoi(c.Query("status"))
	keyword := c.Query("keyword")

	result, err := h.orderService.List(page, pageSize, status, keyword)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, result)
}

func (h *AdminController) GetSettings(c *gin.Context) {
	response.Success(c, gin.H{
		"site_name":  c.GetString("site_name"),
		"site_desc":  c.GetString("site_desc"),
		"pay_method": c.GetString("pay_method"),
	})
}

func (h *AdminController) UpdateSettings(c *gin.Context) {
	var req struct {
		SiteName  string `json:"site_name"`
		SiteDesc  string `json:"site_desc"`
		PayMethod string `json:"pay_method"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "无效的请求参数")
		return
	}
	response.Success(c, gin.H{"updated": true})
}