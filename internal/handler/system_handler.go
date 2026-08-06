package handler

import (
	"chenze-faka/internal/config"
	"chenze-faka/internal/license"
	"chenze-faka/internal/utils"

	"github.com/gin-gonic/gin"
)

type SystemHandler struct {
	cfg        *config.Config
	licenseSvc *license.Service
}

func NewSystemHandler(cfg *config.Config, licenseSvc *license.Service) *SystemHandler {
	return &SystemHandler{
		cfg:        cfg,
		licenseSvc: licenseSvc,
	}
}

func (h *SystemHandler) GetSystemStatus(c *gin.Context) {
	data := gin.H{
		"site_name": h.cfg.System.SiteName,
		"port":      h.cfg.System.Port,
		"mode":      h.cfg.System.Mode,
		"version":   h.licenseSvc.GetCurrentVersion(),
	}

	utils.Success(c, data)
}

func (h *SystemHandler) GetLicenseStatus(c *gin.Context) {
	cache := h.licenseSvc.GetCache()
	isVerified := h.licenseSvc.IsVerified()
	isGraceValid := h.licenseSvc.IsGracePeriodValid()

	status := "invalid"
	if isVerified {
		status = "valid"
	} else if isGraceValid {
		status = "grace_period"
	}

	data := gin.H{
		"status":       status,
		"verified":     isVerified,
		"grace_valid":  isGraceValid,
		"last_verify":  cache.LastVerifyTime,
		"last_success": cache.LastSuccessTime,
		"last_result":  cache.LastResult,
		"last_message": cache.LastMessage,
		"expire_at":    cache.ExpireAt,
		"app_name":     cache.AppName,
		"domain":       h.cfg.License.Domain,
		"server_ip":    h.cfg.License.ServerIP,
	}

	utils.Success(c, data)
}

func (h *SystemHandler) VerifyLicense(c *gin.Context) {
	ok, err := h.licenseSvc.Verify()
	if err != nil {
		utils.FailWithCode(c, 500, "verification error: "+err.Error())
		return
	}

	utils.Success(c, gin.H{"verified": ok})
}

func (h *SystemHandler) CheckVersion(c *gin.Context) {
	currentVersion := c.Query("version")
	if currentVersion == "" {
		currentVersion = h.licenseSvc.GetCurrentVersion()
	}

	info, err := h.licenseSvc.CheckVersion(currentVersion)
	if err != nil {
		utils.FailWithCode(c, 500, "version check error: "+err.Error())
		return
	}

	utils.Success(c, info)
}
