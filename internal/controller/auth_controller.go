package controller

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"runtime"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/captcha"
	"chenze-faka/internal/pkg/database"
	"chenze-faka/internal/pkg/response"
	"chenze-faka/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gopkg.in/yaml.v3"
)

var mysqlVersionRegexp = regexp.MustCompile(`^(\d+)\.(\d+)`)

type AuthController struct {
	authService  *service.AuthService
	jwtSecret    string
	expireHours  int
	siteName     string
	licenseCfg   *model.LicenseConfig
	dbCfg        *model.DatabaseConfig
}

func NewAuthController(jwtSecret string, expireHours int, siteName string, licenseCfg *model.LicenseConfig, dbCfg *model.DatabaseConfig) *AuthController {
	return &AuthController{
		authService: service.NewAuthService(),
		jwtSecret:   jwtSecret,
		expireHours: expireHours,
		siteName:    siteName,
		licenseCfg:  licenseCfg,
		dbCfg:       dbCfg,
	}
}

func (h *AuthController) Login(c *gin.Context) {
	var req struct {
		Username  string `json:"username" binding:"required"`
		Password  string `json:"password" binding:"required"`
		CaptchaID string `json:"captcha_id"`
		Captcha   string `json:"captcha"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "无效的请求参数")
		return
	}

	if req.CaptchaID != "" && req.Captcha != "" {
		if !captcha.Verify(req.CaptchaID, req.Captcha) {
			response.Fail(c, "验证码错误")
			return
		}
	}

	token, user, err := h.authService.Login(req.Username, req.Password, h.jwtSecret, h.expireHours)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

func (h *AuthController) GetCaptcha(c *gin.Context) {
	id, imgBase64 := captcha.Generate()
	response.Success(c, gin.H{
		"id":  id,
		"img": "data:image/png;base64," + imgBase64,
	})
}

func (h *AuthController) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "无效的请求参数")
		return
	}

	user, err := h.authService.Register(req.Username, req.Password)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
	})
}

func (h *AuthController) GetProfile(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		response.Unauthorized(c, "未登录")
		return
	}

	u := user.(*model.User)
	response.Success(c, gin.H{
		"id":          u.ID,
		"username":    u.Username,
		"role":        u.Role,
		"last_login":  u.LastLoginAt,
		"created_at":  u.CreatedAt,
	})
}

func (h *AuthController) GetSiteConfig(c *gin.Context) {
	response.Success(c, gin.H{
		"site_name": h.siteName,
	})
}

type InstallRequest struct {
	SiteName   string              `json:"site_name"`
	LicenseKey string              `json:"license_key"`
	Database   model.DatabaseConfig `json:"database"`
	JWT        model.JWTConfig    `json:"jwt"`
	Pay        model.PayConfig    `json:"pay"`
	Username   string              `json:"username"`
	Password   string              `json:"password"`
	Force      bool                `json:"force"`
	Skip       bool                `json:"skip"`
}

func (h *AuthController) Install(c *gin.Context) {
	var req InstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "无效的请求参数")
		return
	}

	if req.Skip {
		if err := os.WriteFile("install.lock", []byte("skipped"), 0644); err != nil {
			response.Fail(c, "创建安装锁失败")
			return
		}
		response.Success(c, gin.H{"message": "跳过安装成功"})
		return
	}

	if req.LicenseKey == "" {
		response.Fail(c, "授权密钥不能为空")
		return
	}

	if req.Username == "" || req.Password == "" {
		response.Fail(c, "管理员用户名和密码不能为空")
		return
	}

	licenseSvc := service.NewLicenseService(&model.LicenseConfig{
		Enabled:    true,
		LicenseKey: req.LicenseKey,
	})

	result, err := licenseSvc.QuickVerify(req.LicenseKey, "", "")
	if err != nil || !result.Verified {
		errMsg := "授权码无效或已过期"
		if err != nil {
			errMsg = "授权验证失败: " + err.Error()
		}
		response.Fail(c, errMsg)
		return
	}

	if err := database.Init(&req.Database); err != nil {
		response.Fail(c, "数据库连接失败: "+err.Error())
		return
	}

	if req.Force {
		if err := database.WipeTables(); err != nil {
			response.Fail(c, "清空数据表失败: "+err.Error())
			return
		}
	}

	if err := database.AutoMigrate(); err != nil {
		response.Fail(c, "数据库迁移失败: "+err.Error())
		return
	}

	var count int64
	database.DB.Model(&model.User{}).Count(&count)
	if count > 0 && !req.Force {
		response.Fail(c, "系统已安装,请直接登录")
		return
	}

	salt := generateSalt()
	hash := hashPassword(req.Password, salt)

	user := &model.User{
		Username:     req.Username,
		PasswordHash: hash,
		Salt:         salt,
		Role:         model.RoleAdmin,
	}

	if err := database.DB.Create(user).Error; err != nil {
		response.Fail(c, "创建管理员用户失败")
		return
	}

	cfg := model.Config{
		System: model.SystemConfig{
			SiteName: req.SiteName,
			Port:     12398,
			Mode:     "debug",
		},
		Database: req.Database,
		JWT:      req.JWT,
		Pay:      req.Pay,
		License: model.LicenseConfig{
			Enabled:    true,
			LicenseKey: req.LicenseKey,
		},
	}

	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = "chenze-faka-secret"
	}
	if cfg.JWT.ExpireTime == 0 {
		cfg.JWT.ExpireTime = 72
	}

	if err := saveConfig("config.yaml", &cfg); err != nil {
		response.Fail(c, "保存配置失败")
		return
	}

	if err := os.WriteFile("install.lock", []byte("installed"), 0644); err != nil {
		response.Fail(c, "创建安装锁失败")
		return
	}

	response.Success(c, gin.H{
		"message": "安装成功",
	})
}

func (h *AuthController) GetLicenseStatus(c *gin.Context) {
	installed := false
	if _, err := os.Stat("install.lock"); err == nil {
		installed = true
	}

	response.Success(c, gin.H{
		"installed": installed,
		"site_name": h.siteName,
	})
}

type TestDatabaseRequest struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type EnvCheckResponse struct {
	OS           string `json:"os"`
	GOVersion    string `json:"go_version"`
	MySQLStatus  string `json:"mysql_status"`
	MySQLVersion string `json:"mysql_version"`
	Memory       string `json:"memory"`
}

func (h *AuthController) CheckEnv(c *gin.Context) {
	resp := EnvCheckResponse{
		OS:           runtime.GOOS,
		GOVersion:    runtime.Version(),
		MySQLStatus:  "not-installed",
		MySQLVersion: "未安装",
		Memory:       "2GB+",
	}

	cfg := h.dbCfg
	if cfg == nil {
		response.Success(c, resp)
		return
	}

	if cfg.Driver == "sqlite" {
		resp.MySQLStatus = "normal"
		resp.MySQLVersion = "SQLite 3.x"
		response.Success(c, resp)
		return
	}

	host := cfg.Host
	if host == "" {
		host = "localhost"
	}
	port := cfg.Port
	if port == 0 {
		port = 3306
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/mysql?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, host, port)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		response.Success(c, resp)
		return
	}
	sqlDB, err := db.DB()
	if err != nil {
		response.Success(c, resp)
		return
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		response.Success(c, resp)
		return
	}

	var version string
	row := sqlDB.QueryRow("SELECT VERSION()")
	if err := row.Scan(&version); err != nil {
		version = "unknown"
	}

	resp.MySQLVersion = version
	resp.MySQLStatus = checkMysqlVersionStatusEn(version)
	response.Success(c, resp)
}

func checkMysqlVersionStatusEn(version string) string {
	match := mysqlVersionRegexp.FindStringSubmatch(version)
	if match == nil {
		return "error"
	}
	var major, minor int
	fmt.Sscanf(match[1], "%d", &major)
	fmt.Sscanf(match[2], "%d", &minor)
	if major >= 8 {
		return "normal"
	}
	if major == 5 && minor >= 7 {
		return "normal"
	}
	return "error"
}

func checkMysqlVersionStatus(version string) string {
	match := mysqlVersionRegexp.FindStringSubmatch(version)
	if match == nil {
		return "异常"
	}
	var major, minor int
	fmt.Sscanf(match[1], "%d", &major)
	fmt.Sscanf(match[2], "%d", &minor)
	if major >= 8 {
		return "正常"
	}
	if major == 5 && minor >= 7 {
		return "正常"
	}
	return "异常"
}

func (h *AuthController) TestDatabase(c *gin.Context) {
	var req TestDatabaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "无效的请求参数")
		return
	}

	if req.Host == "" {
		req.Host = "localhost"
	}
	if req.Port == 0 {
		req.Port = 3306
	}
	if req.Database == "" {
		response.Fail(c, "数据库名不能为空")
		return
	}
	if req.Username == "" {
		response.Fail(c, "用户名不能为空")
		return
	}

	cfg := &model.DatabaseConfig{
		Driver:   "mysql",
		Host:     req.Host,
		Port:     req.Port,
		DBName:   req.Database,
		User:     req.Username,
		Password: req.Password,
	}

	result, err := database.TestConnection(cfg)
	if err != nil {
		response.Fail(c, "数据库连接失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"message":     "数据库连接成功",
		"version":     result.Version,
		"has_data":    result.HasData,
		"table_count": result.TableCount,
	})
}

func (h *AuthController) VerifyLicense(c *gin.Context) {
	var req struct {
		LicenseKey string `json:"license_key"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "无效的请求参数")
		return
	}

	if req.LicenseKey == "" {
		response.Fail(c, "授权密钥不能为空")
		return
	}

	cfg := h.licenseCfg
	if cfg == nil {
		cfg = &model.LicenseConfig{}
	}

	verifyCfg := &model.LicenseConfig{
		Enabled:     true,
		BaseURL:     cfg.BaseURL,
		BackupURL:   cfg.BackupURL,
		AppKey:      cfg.AppKey,
		AppSecret:   cfg.AppSecret,
		LicenseKey:  req.LicenseKey,
		Domain:      cfg.Domain,
		ServerIP:    cfg.ServerIP,
		CacheFile:   cfg.CacheFile,
		Interval:    cfg.Interval,
		GracePeriod: cfg.GracePeriod,
	}

	licenseSvc := service.NewLicenseService(verifyCfg)

	result, err := licenseSvc.QuickVerify(req.LicenseKey, "", "")
	if err != nil {
		response.Fail(c, "授权验证失败: "+err.Error())
		return
	}

	response.Success(c, result)
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

func saveConfig(path string, cfg *model.Config) error {
	data, err := yamlMarshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func yamlMarshal(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}