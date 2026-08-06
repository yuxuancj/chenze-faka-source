package controller

import (
	"chenze-faka/internal/pkg/response"
	"chenze-faka/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CardController struct {
	cardService *service.CardService
}

func NewCardController() *CardController {
	return &CardController{
		cardService: service.NewCardService(),
	}
}

func (h *CardController) Import(c *gin.Context) {
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

func (h *CardController) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, "无效的卡密ID")
		return
	}

	card, err := h.cardService.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, card)
}

func (h *CardController) List(c *gin.Context) {
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

func (h *CardController) Delete(c *gin.Context) {
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

func (h *CardController) CountByProduct(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, "无效的产品ID")
		return
	}

	count, err := h.cardService.CountByProduct(uint(id))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"product_id": id,
		"count":      count,
	})
}