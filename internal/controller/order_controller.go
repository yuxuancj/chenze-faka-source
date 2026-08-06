package controller

import (
	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/response"
	"chenze-faka/internal/pkg/utils"
	"chenze-faka/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OrderController struct {
	orderService *service.OrderService
	payConfig    model.PayConfig
}

func NewOrderController(payConfig model.PayConfig) *OrderController {
	return &OrderController{
		orderService: service.NewOrderService(),
		payConfig:    payConfig,
	}
}

func (h *OrderController) Create(c *gin.Context) {
	var req service.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "无效的请求参数")
		return
	}

	result, err := h.orderService.CreateOrder(&req, h.payConfig)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, result)
}

func (h *OrderController) Query(c *gin.Context) {
	orderNo := c.Query("order_no")
	contact := c.Query("contact")

	result, err := h.orderService.QueryOrder(orderNo, contact)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, result)
}

func (h *OrderController) GetByOrderNo(c *gin.Context) {
	orderNo := c.Param("order_no")
	contact := c.Query("contact")

	result, err := h.orderService.QueryOrder(orderNo, contact)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, result)
}

func (h *OrderController) Notify(c *gin.Context) {
	params := make(map[string]string)
	for key, values := range c.Request.Form {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	if params["sign"] != "" {
		if !h.orderService.VerifyNotify(params, h.payConfig.Key) {
			c.String(200, "fail")
			return
		}
	}

	outTradeNo := params["out_trade_no"]
	payNo := params["trade_no"]
	amount := 0.0
	if m := params["money"]; m != "" {
		if v, err := strconv.ParseFloat(m, 64); err == nil {
			amount = v
		}
	}

	success, err := h.orderService.HandlePaymentCallback(payNo, outTradeNo, amount)
	if err != nil || !success {
		c.String(200, "fail")
		return
	}

	c.String(200, "success")
}

func (h *OrderController) Return(c *gin.Context) {
	orderNo := c.Query("out_trade_no")
	if orderNo != "" {
		result, err := h.orderService.QueryOrder(orderNo, "")
		if err == nil {
			response.Success(c, result)
			return
		}
	}
	response.Success(c, gin.H{
		"order_no": c.Query("out_trade_no"),
		"status":   "unknown",
	})
}

func (h *OrderController) List(c *gin.Context) {
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

func (h *OrderController) BuildPayURL(orderNo string, amount float64, payMethod string, productName string) string {
	return utils.BuildPayURL(orderNo, amount, payMethod, productName, h.payConfig)
}