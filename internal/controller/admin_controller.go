package controller

import (
	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/response"
	"chenze-faka/internal/service"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminController struct {
	productService  *service.ProductService
	orderService    *service.OrderService
	cardService     *service.CardService
	categoryService *service.CategoryService
	paymentService  *service.PaymentService
	emailService    *service.EmailService
	logService      *service.LogService
	dashboardSvc    *service.DashboardService
	nodeService     *service.NodeService
	licenseSvc      *service.LicenseService
	upgradeService  *service.UpgradeService
	uploadService   *service.UploadService
}

func NewAdminController(licenseSvc *service.LicenseService) *AdminController {
	return &AdminController{
		productService:  service.NewProductService(),
		orderService:    service.NewOrderService(),
		cardService:     service.NewCardService(),
		categoryService: service.NewCategoryService(),
		paymentService:  service.NewPaymentService(),
		emailService:    service.NewEmailService(),
		logService:      service.NewLogService(),
		dashboardSvc:    service.NewDashboardService(),
		nodeService:     service.NewNodeService(),
		licenseSvc:      licenseSvc,
		upgradeService:  service.NewUpgradeService(),
		uploadService:   service.NewUploadService(),
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
		response.Success(c, gin.H{"status": "unknown"})
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

func (h *AdminController) Dashboard(c *gin.Context) {
	stats, err := h.dashboardSvc.GetStats()
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, stats)
}

func (h *AdminController) OrderStatusCounts(c *gin.Context) {
	counts, err := h.dashboardSvc.GetOrderStatusCounts()
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, counts)
}

// ===== Category =====

func (h *AdminController) CategoryList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")
	result, err := h.categoryService.List(page, pageSize, keyword)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AdminController) CategoryAll(c *gin.Context) {
	cats, err := h.categoryService.GetAll()
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, gin.H{"items": cats})
}

func (h *AdminController) CategoryCreate(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
		Icon string `json:"icon"`
		Sort int    `json:"sort"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "无效的请求参数")
		return
	}
	cat, err := h.categoryService.Create(req.Name, req.Icon, req.Sort)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, cat)
}

func (h *AdminController) CategoryUpdate(c *gin.Context) {
	var req struct {
		ID     uint   `json:"id"`
		Name   string `json:"name"`
		Icon   string `json:"icon"`
		Sort   int    `json:"sort"`
		Status int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "无效的请求参数")
		return
	}
	cat, err := h.categoryService.Update(req.ID, req.Name, req.Icon, req.Sort, req.Status)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, cat)
}

func (h *AdminController) CategoryDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, "无效的ID")
		return
	}
	if err := h.categoryService.Delete(uint(id)); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, nil)
}

// ===== Product =====

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

// ===== Card =====

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

func (h *AdminController) CardExport(c *gin.Context) {
	productID, _ := strconv.Atoi(c.Query("product_id"))
	status, _ := strconv.Atoi(c.Query("status"))
	items, err := h.cardService.ExportCards(productID, status)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, gin.H{"items": items})
}

// ===== Order =====

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

func (h *AdminController) OrderLogs(c *gin.Context) {
	orderNo := c.Query("order_no")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	result, err := h.logService.ListOrder(orderNo, page, pageSize)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, result)
}

// ===== Payment =====

func (h *AdminController) PaymentList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")
	result, err := h.paymentService.List(page, pageSize, keyword)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AdminController) PaymentAll(c *gin.Context) {
	channels, err := h.paymentService.GetActive()
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, gin.H{"items": channels})
}

func (h *AdminController) PaymentCreate(c *gin.Context) {
	var req struct {
		Name    string  `json:"name"`
		Type    string  `json:"type"`
		Icon    string  `json:"icon"`
		Config  string  `json:"config"`
		FeeRate float64 `json:"fee_rate"`
		Sort    int     `json:"sort"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "无效的请求参数")
		return
	}
	ch, err := h.paymentService.Create(req.Name, req.Type, req.Icon, req.Config, req.FeeRate, req.Sort)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, ch)
}

func (h *AdminController) PaymentUpdate(c *gin.Context) {
	var req struct {
		ID      uint    `json:"id"`
		Name    string  `json:"name"`
		Icon    string  `json:"icon"`
		Config  string  `json:"config"`
		FeeRate float64 `json:"fee_rate"`
		Status  int     `json:"status"`
		Sort    int     `json:"sort"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "无效的请求参数")
		return
	}
	ch, err := h.paymentService.Update(req.ID, req.Name, req.Icon, req.Config, req.FeeRate, req.Status, req.Sort)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, ch)
}

func (h *AdminController) PaymentDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, "无效的ID")
		return
	}
	if err := h.paymentService.Delete(uint(id)); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, nil)
}

// ===== Email =====

func (h *AdminController) EmailList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	result, err := h.emailService.List(page, pageSize)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AdminController) EmailCreate(c *gin.Context) {
	var cfg model.EmailConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.Fail(c, "无效的请求参数")
		return
	}
	saved, err := h.emailService.Create(&cfg)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, saved)
}

func (h *AdminController) EmailUpdate(c *gin.Context) {
	var req struct {
		ID uint `json:"id"`
		model.EmailConfig
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "无效的请求参数")
		return
	}
	cfg, err := h.emailService.Update(req.ID, &req.EmailConfig)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, cfg)
}

func (h *AdminController) EmailDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, "无效的ID")
		return
	}
	if err := h.emailService.Delete(uint(id)); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *AdminController) EmailTest(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, "无效的ID")
		return
	}
	if err := h.emailService.TestConnection(uint(id)); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *AdminController) EmailLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	email := c.Query("email")
	result, err := h.logService.ListEmail(page, pageSize, email)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, result)
}

// ===== Logs =====

func (h *AdminController) OperationLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	username := c.Query("username")
	action := c.Query("action")
	result, err := h.logService.ListOperation(page, pageSize, username, action)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AdminController) LoginLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	username := c.Query("username")
	result, err := h.logService.ListLogin(page, pageSize, username)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, result)
}

// ===== Node =====

func (h *AdminController) NodeList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	keyword := c.Query("keyword")
	result, err := h.nodeService.List(page, pageSize, keyword)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AdminController) NodeCreate(c *gin.Context) {
	var req struct {
		Name   string `json:"name"`
		URL    string `json:"url"`
		Weight int    `json:"weight"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "无效的请求参数")
		return
	}
	node, err := h.nodeService.Create(req.Name, req.URL, req.Weight)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, node)
}

func (h *AdminController) NodeUpdate(c *gin.Context) {
	var req struct {
		ID     uint `json:"id"`
		Name   string `json:"name"`
		URL    string `json:"url"`
		Weight int    `json:"weight"`
		Status int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "无效的请求参数")
		return
	}
	node, err := h.nodeService.Update(req.ID, req.Name, req.URL, req.Weight, req.Status)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, node)
}

func (h *AdminController) NodeDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, "无效的ID")
		return
	}
	if err := h.nodeService.Delete(uint(id)); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *AdminController) NodePing(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, "无效的ID")
		return
	}
	node, err := h.nodeService.Ping(uint(id))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, node)
}

// ===== Settings =====

func (h *AdminController) GetSettings(c *gin.Context) {
	response.Success(c, gin.H{
		"site_name":        c.GetString("site_name"),
		"site_logo":        c.GetString("site_logo"),
		"site_description": c.GetString("site_desc"),
		"alipay_enabled":   c.GetBool("alipay_enabled"),
		"alipay_app_id":    c.GetString("alipay_app_id"),
		"alipay_private_key": c.GetString("alipay_private_key"),
		"wechat_enabled":   c.GetBool("wechat_enabled"),
		"wechat_app_id":    c.GetString("wechat_app_id"),
		"wechat_mch_id":    c.GetString("wechat_mch_id"),
		"order_expire":     c.GetInt("order_expire"),
		"card_prefix":      c.GetString("card_prefix"),
		"maintenance":      c.GetBool("maintenance"),
	})
}

func (h *AdminController) UpdateSettings(c *gin.Context) {
	var req struct {
		SiteName        string `json:"site_name"`
		SiteLogo        string `json:"site_logo"`
		SiteDescription string `json:"site_description"`
		AlipayEnabled   bool   `json:"alipay_enabled"`
		AlipayAppID     string `json:"alipay_app_id"`
		AlipayPrivateKey string `json:"alipay_private_key"`
		WechatEnabled   bool   `json:"wechat_enabled"`
		WechatAppID     string `json:"wechat_app_id"`
		WechatMchID     string `json:"wechat_mch_id"`
		OrderExpire     int    `json:"order_expire"`
		CardPrefix      string `json:"card_prefix"`
		Maintenance     bool   `json:"maintenance"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "无效的请求参数")
		return
	}

	response.Success(c, gin.H{
		"updated": true,
		"data":    req,
	})
}

// ===== Upgrade =====

func (h *AdminController) GetVersion(c *gin.Context) {
	response.Success(c, h.upgradeService.GetVersion())
}

func (h *AdminController) CheckUpdate(c *gin.Context) {
	response.Success(c, h.upgradeService.CheckUpdate())
}

func (h *AdminController) UploadPackage(c *gin.Context) {
	file, header, err := c.Request.FormFile("package")
	if err != nil {
		response.Fail(c, "无效的请求参数: "+err.Error())
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		response.Fail(c, "读取文件失败")
		return
	}

	if err := h.upgradeService.UploadPackage(header.Filename, data); err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "升级包上传成功"})
}

func (h *AdminController) ApplyUpgrade(c *gin.Context) {
	if err := h.upgradeService.ApplyUpgrade(); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, gin.H{"message": "升级成功"})
}

func (h *AdminController) UpgradeLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	result, err := h.upgradeService.ListUpgradeLogs(page, pageSize)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, result)
}

// ===== Upload =====

func (h *AdminController) UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Fail(c, "无效的请求参数: "+err.Error())
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	allowedImages := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".svg": true}
	allowedDocs := map[string]bool{".pdf": true, ".txt": true}
	isValid := allowedImages[ext] || allowedDocs[ext]
	if !isValid {
		response.Fail(c, "不支持的文件类型")
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		response.Fail(c, "读取文件失败")
		return
	}

	if len(data) > 10*1024*1024 {
		response.Fail(c, "文件大小超过限制(10MB)")
		return
	}

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	user, _ := c.Get("user")
	uploaderID := uint(0)
	if u, ok := user.(*model.User); ok {
		uploaderID = u.ID
	}

	f, err := h.uploadService.SaveFile(header.Filename, data, mimeType, uploaderID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"id":           f.ID,
		"original_name": f.OriginalName,
		"url":          "/uploads/" + f.StoredName,
		"size":         f.Size,
		"mime_type":    f.MimeType,
		"type":         f.Type,
	})
}

func (h *AdminController) ListFiles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	fileType := c.Query("type")
	result, err := h.uploadService.ListFiles(page, pageSize, fileType)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *AdminController) DeleteFile(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, "无效的ID")
		return
	}
	if err := h.uploadService.DeleteFile(uint(id)); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *AdminController) GetFile(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, "无效的ID")
		return
	}
	f, err := h.uploadService.GetFile(uint(id))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	filePath := filepath.Join(".", f.Path)
	data, err := os.ReadFile(filePath)
	if err != nil {
		response.Fail(c, "文件不存在")
		return
	}
	c.Data(http.StatusOK, f.MimeType, data)
}
