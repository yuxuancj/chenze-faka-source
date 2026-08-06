package router

import (
	"chenze-faka/internal/config"
	"chenze-faka/internal/handler"
	"chenze-faka/internal/license"
	"chenze-faka/internal/middleware"
	"chenze-faka/internal/utils"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func Setup(cfg *config.Config, licenseSvc *license.Service) *gin.Engine {
	r := gin.Default()

	installHandler := handler.NewInstallHandler(cfg, licenseSvc)
	authHandler := handler.NewAuthHandler(cfg)
	productHandler := handler.NewProductHandler()
	cardHandler := handler.NewCardHandler()
	orderHandler := handler.NewOrderHandler(cfg)
	systemHandler := handler.NewSystemHandler(cfg, licenseSvc)

	authMiddleware := middleware.NewAuthMiddleware(cfg.JWT)
	licenseMiddleware := middleware.NewLicenseMiddleware(licenseSvc)
	installMiddleware := middleware.NewInstallMiddleware()

	webDir := getWebDir()

	// Install check middleware - highest priority, runs on ALL requests
	r.Use(installMiddleware.Handle())

	// License middleware - blocks API when license invalid
	r.Use(licenseMiddleware.Handle())

	r.Static("/static", filepath.Join(webDir, "static"))

	// Install page routes (always accessible, even when not installed)
	r.GET("/install-page", func(c *gin.Context) {
		c.File(filepath.Join(webDir, "install.html"))
	})

	// /install redirects to install-page for HTML display
	r.GET("/install", func(c *gin.Context) {
		c.Redirect(302, "/install-page")
	})

	// Auth install API routes
	r.GET("/install/license-status", installHandler.GetLicenseStatus)
	r.POST("/install/verify-license", installHandler.VerifyLicense)
	r.POST("/install/test-db", installHandler.TestDatabase)
	r.POST("/install", installHandler.Install)

	// Homepage
	r.GET("/", func(c *gin.Context) {
		c.File(filepath.Join(webDir, "index.html"))
	})

	// Order page
	r.GET("/order", func(c *gin.Context) {
		c.File(filepath.Join(webDir, "order.html"))
	})

	// Compatibility with old-style PHP URLs
	r.GET("/order.php", func(c *gin.Context) {
		c.Redirect(302, "/order" + "?" + c.Request.URL.RawQuery)
	})

	// Query page
	r.GET("/query", func(c *gin.Context) {
		c.File(filepath.Join(webDir, "query.html"))
	})

	r.GET("/query.php", func(c *gin.Context) {
		c.Redirect(302, "/query")
	})

	// Admin routes
	r.GET("/admin", func(c *gin.Context) {
		c.Redirect(302, "/admin/index.html")
	})

	r.GET("/admin/index.html", func(c *gin.Context) {
		c.File(filepath.Join(webDir, "admin", "index.html"))
	})

	r.GET("/admin/login.html", func(c *gin.Context) {
		c.File(filepath.Join(webDir, "admin", "login.html"))
	})

	// API routes
	api := r.Group("/api")
	{
		api.GET("/site/config", authHandler.GetSiteConfig)
		api.GET("/system/status", systemHandler.GetSystemStatus)
		api.GET("/system/license", systemHandler.GetLicenseStatus)
		api.POST("/system/license/verify", systemHandler.VerifyLicense)
		api.GET("/system/version/check", systemHandler.CheckVersion)

		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/profile", authMiddleware.AuthRequired(), authHandler.GetProfile)
		}

		products := api.Group("/products")
		{
			products.GET("", productHandler.List)
			products.GET("/on-shelf", productHandler.ListOnShelf)
			products.GET("/on-shelf-grouped", productHandler.ListOnShelfGrouped)
			products.GET("/:id", productHandler.GetByID)
		}

		orders := api.Group("/orders")
		{
			orders.POST("", orderHandler.CreateOrder)
			orders.GET("/query", orderHandler.QueryOrder)
			orders.GET("/:order_no", orderHandler.GetByOrderNo)
			orders.POST("/notify", orderHandler.Notify)
			orders.GET("/return", orderHandler.Return)
		}

		admin := api.Group("/admin", authMiddleware.AuthRequired())
		{
			productAdmin := admin.Group("/products")
			{
				productAdmin.POST("", productHandler.Create)
				productAdmin.PUT("", productHandler.Update)
				productAdmin.DELETE("/:id", productHandler.Delete)
				productAdmin.GET("", productHandler.List)
				productAdmin.GET("/:id", productHandler.GetByID)
			}

			cardAdmin := admin.Group("/cards")
			{
				cardAdmin.POST("/import", cardHandler.Import)
				cardAdmin.GET("", cardHandler.List)
				cardAdmin.GET("/:id", cardHandler.GetByID)
				cardAdmin.DELETE("/:id", cardHandler.Delete)
				cardAdmin.GET("/product/:product_id/count", cardHandler.CountByProduct)
			}

			orderAdmin := admin.Group("/orders")
			{
				orderAdmin.GET("", orderHandler.List)
			}
		}
	}

	// Fallback for undefined routes
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// API routes get JSON response
		if len(path) >= 4 && path[:4] == "/api/" {
			c.JSON(404, utils.FailResponse("not found"))
			return
		}

		c.JSON(404, gin.H{"code": 404, "message": "not found"})
	})

	return r
}

func checkInstalled() bool {
	lockFile := "install.lock"
	_, err := os.Stat(lockFile)
	return err == nil
}

func getWebDir() string {
	candidates := []string{
		"web",
		"assets/web",
		"chenze_faka_web",
	}
	for _, d := range candidates {
		if _, err := os.Stat(d); err == nil {
			return d
		}
	}
	return "web"
}
