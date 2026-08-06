package handler

import (
	"chenze-faka/internal/config"
	"chenze-faka/internal/service"
	"chenze-faka/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService *service.OrderService
	cfg          *config.Config
}

func NewOrderHandler(cfg *config.Config) *OrderHandler {
	return &OrderHandler{
		orderService: service.NewOrderService(cfg.Pay),
		cfg:          cfg,
	}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req service.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "invalid request parameters")
		return
	}

	result, err := h.orderService.CreateOrder(&req)
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.Success(c, result)
}

func (h *OrderHandler) Notify(c *gin.Context) {
	params := make(map[string]string)
	for key, values := range c.Request.Form {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	if params["sign"] != "" {
		if !h.orderService.VerifyNotify(params) {
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

func (h *OrderHandler) Return(c *gin.Context) {
	c.HTML(200, "return.html", gin.H{
		"order_no": c.Query("out_trade_no"),
	})
}

func (h *OrderHandler) QueryOrder(c *gin.Context) {
	orderNo := c.Query("order_no")
	contact := c.Query("contact")

	result, err := h.orderService.QueryOrder(orderNo, contact)
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.Success(c, result)
}

func (h *OrderHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status, _ := strconv.Atoi(c.Query("status"))
	keyword := c.Query("keyword")

	orders, total, err := h.orderService.List(page, pageSize, status, keyword)
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"orders":    orders,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *OrderHandler) GetByOrderNo(c *gin.Context) {
	orderNo := c.Param("order_no")
	contact := c.Query("contact")

	result, err := h.orderService.QueryOrder(orderNo, contact)
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.Success(c, result)
}
