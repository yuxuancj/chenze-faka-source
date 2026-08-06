package handler

import (
	"chenze-faka/internal/service"
	"chenze-faka/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CardHandler struct {
	cardService *service.CardService
}

func NewCardHandler() *CardHandler {
	return &CardHandler{
		cardService: service.NewCardService(),
	}
}

func (h *CardHandler) Import(c *gin.Context) {
	var req struct {
		ProductID uint   `json:"product_id" binding:"required"`
		CardText  string `json:"card_text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "invalid request parameters")
		return
	}

	cardNos := h.cardService.ParseCardText(req.CardText)
	if len(cardNos) == 0 {
		utils.Fail(c, "no valid card numbers provided")
		return
	}

	result, err := h.cardService.ImportCards(req.ProductID, cardNos)
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.Success(c, result)
}

func (h *CardHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Fail(c, "invalid card id")
		return
	}

	card, err := h.cardService.GetByID(uint(id))
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}

	utils.Success(c, card)
}

func (h *CardHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	productID, _ := strconv.Atoi(c.Query("product_id"))
	status, _ := strconv.Atoi(c.Query("status"))

	result, err := h.cardService.List(page, pageSize, uint(productID), status)
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.Success(c, result)
}

func (h *CardHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.Fail(c, "invalid card id")
		return
	}

	if err := h.cardService.Delete(uint(id)); err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

func (h *CardHandler) CountByProduct(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("product_id"))
	if err != nil {
		utils.Fail(c, "invalid product id")
		return
	}

	count, err := h.cardService.CountByProduct(uint(productID))
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"product_id": productID,
		"count":      count,
	})
}
