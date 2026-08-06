package handler

import (
	"chenze-faka/internal/config"
	"chenze-faka/internal/model"
	"chenze-faka/internal/service"
	"chenze-faka/internal/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
	cfg         *config.Config
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: service.NewAuthService(cfg.JWT),
		cfg:         cfg,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "invalid request parameters")
		return
	}

	user, err := h.authService.Register(req.Username, req.Password)
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "invalid request parameters")
		return
	}

	token, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"token": token,
	})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		utils.Unauthorized(c, "not logged in")
		return
	}

	u := user.(*model.User)
	utils.Success(c, gin.H{
		"id":         u.ID,
		"username":   u.Username,
		"last_login": u.LastLogin,
		"created_at": u.CreatedAt,
	})
}

func (h *AuthHandler) GetSiteConfig(c *gin.Context) {
	utils.Success(c, gin.H{
		"site_name": h.cfg.System.SiteName,
	})
}
