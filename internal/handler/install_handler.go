package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"os"

	"chenze-faka/internal/config"
	"chenze-faka/internal/license"
	"chenze-faka/internal/model"
	"chenze-faka/internal/repository"
	"chenze-faka/internal/utils"
	"chenze-faka/pkg/database"

	"github.com/gin-gonic/gin"
)

type InstallHandler struct {
	cfg        *config.Config
	licenseSvc *license.Service
}

func NewInstallHandler(cfg *config.Config, licenseSvc *license.Service) *InstallHandler {
	return &InstallHandler{
		cfg:        cfg,
		licenseSvc: licenseSvc,
	}
}

func (h *InstallHandler) CheckInstall(c *gin.Context) {
	lockFile := "install.lock"
	if _, err := os.Stat(lockFile); os.IsNotExist(err) {
		utils.Success(c, gin.H{
			"installed": false,
		})
		return
	}
	utils.Success(c, gin.H{
		"installed": true,
	})
}

func (h *InstallHandler) GetLicenseStatus(c *gin.Context) {
	lockFile := "install.lock"
	installed := false
	if _, err := os.Stat(lockFile); err == nil {
		installed = true
	}

	licenseKey := h.cfg.License.LicenseKey
	verified := h.licenseSvc.IsVerified()

	utils.Success(c, gin.H{
		"configured":  licenseKey != "",
		"license_key": licenseKey,
		"verified":    verified,
		"installed":   installed,
	})
}

func (h *InstallHandler) VerifyLicense(c *gin.Context) {
	var req struct {
		LicenseKey string `json:"license_key"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "invalid request parameters")
		return
	}

	if req.LicenseKey == "" {
		utils.Fail(c, "license key is required")
		return
	}

	cfg := license.Config{
		Enabled:     h.cfg.License.Enabled,
		BaseURL:     h.cfg.License.BaseURL,
		AppKey:      h.cfg.License.AppKey,
		AppSecret:   h.cfg.License.AppSecret,
		LicenseKey:  req.LicenseKey,
		Domain:      h.cfg.License.Domain,
		ServerIP:    h.cfg.License.ServerIP,
		CacheFile:   h.cfg.License.CacheFile,
		Interval:    h.cfg.License.Interval,
		GracePeriod: h.cfg.License.GracePeriod,
	}

	svc := license.NewService(&cfg)
	verified, realNameInfo, err := svc.QuickVerify(req.LicenseKey, h.cfg.License.Domain, h.cfg.License.ServerIP)
	if err != nil {
		utils.Fail(c, "verification error: "+err.Error())
		return
	}

	if verified {
		cache := svc.GetCache()
		utils.Success(c, gin.H{
			"verified":   true,
			"app_name":   cache.AppName,
			"expire_at":  cache.ExpireAt,
			"domain":     h.cfg.License.Domain,
			"server_ip":  h.cfg.License.ServerIP,
			"license_key": req.LicenseKey,
		})
		return
	}

	if realNameInfo != nil && realNameInfo.NeedRealName {
		utils.Success(c, gin.H{
			"verified":      false,
			"need_real_name": true,
			"account":       realNameInfo.Account,
			"desc":          realNameInfo.Desc,
			"license_key":   req.LicenseKey,
		})
		return
	}

	utils.Success(c, gin.H{
		"verified":    false,
		"license_key": req.LicenseKey,
	})
}

func (h *InstallHandler) TestDatabase(c *gin.Context) {
	var dbConfig config.DatabaseConfig
	if err := c.ShouldBindJSON(&dbConfig); err != nil {
		utils.Fail(c, "invalid database configuration")
		return
	}

	testCfg := &config.Config{
		Database: dbConfig,
	}

	if err := database.Init(&testCfg.Database); err != nil {
		utils.Fail(c, "database connection failed: "+err.Error())
		return
	}

	utils.Success(c, gin.H{
		"connected": true,
	})
}

func (h *InstallHandler) Install(c *gin.Context) {
	var req struct {
		SiteName   string              `json:"site_name"`
		LicenseKey string              `json:"license_key"`
		Database   config.DatabaseConfig `json:"database"`
		JWT        config.JWTConfig    `json:"jwt"`
		Pay        config.PayConfig    `json:"pay"`
		Username   string              `json:"username"`
		Password   string              `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "invalid request parameters")
		return
	}

	if req.LicenseKey == "" {
		utils.Fail(c, "license key is required")
		return
	}

	verified, realNameInfo, err := h.licenseSvc.QuickVerify(req.LicenseKey, h.cfg.License.Domain, h.cfg.License.ServerIP)
	if err != nil {
		utils.Fail(c, "license verification error: "+err.Error())
		return
	}

	if !verified {
		if realNameInfo != nil && realNameInfo.NeedRealName {
			utils.Fail(c, "license requires real name verification, account: "+realNameInfo.Account)
		} else {
			utils.Fail(c, "license verification failed, please check your license key")
		}
		return
	}

	if req.Username == "" || req.Password == "" {
		utils.Fail(c, "admin username and password are required")
		return
	}

	if err := database.Init(&req.Database); err != nil {
		utils.Fail(c, "database connection failed: "+err.Error())
		return
	}

	if err := autoMigrate(); err != nil {
		utils.Fail(c, "database migration failed: "+err.Error())
		return
	}

	userRepo := repository.NewUserRepository()
	existingUser, err := userRepo.GetByUsername(req.Username)
	if err == nil && existingUser != nil {
		utils.Fail(c, "username already exists")
		return
	}

	salt := generateSalt()
	hashedPassword := hashPassword(req.Password, salt)

	user := &model.User{
		Username: req.Username,
		Password: hashedPassword,
		Salt:     salt,
	}

	if err := userRepo.Create(user); err != nil {
		utils.Fail(c, "failed to create admin user")
		return
	}

	h.cfg.System.SiteName = req.SiteName
	h.cfg.Database = req.Database
	h.cfg.JWT = req.JWT
	h.cfg.Pay = req.Pay
	h.cfg.License.LicenseKey = req.LicenseKey

	if err := h.cfg.Save("config.yaml"); err != nil {
		utils.Fail(c, "failed to save config.yaml")
		return
	}

	lockFile := "install.lock"
	if err := os.WriteFile(lockFile, []byte("installed"), 0644); err != nil {
		utils.Fail(c, "failed to create install.lock")
		return
	}

	utils.Success(c, gin.H{
		"message": "installation completed successfully",
	})
}

func autoMigrate() error {
	if database.DB == nil {
		return nil
	}
	return database.DB.AutoMigrate(
		&model.User{},
		&model.Product{},
		&model.Card{},
		&model.Order{},
	)
}

func generateSalt() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func hashPassword(password, salt string) string {
	hash := sha256.New()
	hash.Write([]byte(password + salt))
	return hex.EncodeToString(hash.Sum(nil))
}
