package router

import (
	"chenze-faka/internal/controller"
	"chenze-faka/internal/middleware"
	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/response"
	"chenze-faka/internal/service"
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Setup(cfg *model.Config, licenseSvc *service.LicenseService, webFS embed.FS) *gin.Engine {
	r := gin.Default()

	productCtrl := controller.NewProductController()
	orderCtrl := controller.NewOrderController(cfg.Pay)
	cardCtrl := controller.NewCardController()
	authCtrl := controller.NewAuthController(cfg.JWT.Secret, cfg.JWT.ExpireTime, cfg.System.SiteName, &cfg.License, &cfg.Database)
	adminCtrl := controller.NewAdminController(licenseSvc)

	authMw := middleware.NewAuthMiddleware(cfg.JWT.Secret, cfg.JWT.ExpireTime)
	licenseMw := middleware.NewLicenseMiddleware(licenseSvc)
	installMw := middleware.NewInstallMiddleware()

	r.Use(installMw.Handle())
	r.Use(licenseMw.Handle())

	webContent, err := fs.Sub(webFS, "assets")
	if err == nil {
		_, assetsErr := fs.Sub(webContent, "assets")
		if assetsErr == nil {
			r.GET("/assets/*path", func(c *gin.Context) {
				filePath := strings.TrimPrefix(c.Param("path"), "/")
				data, err := webFS.ReadFile("assets/assets/" + filePath)
				if err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
					return
				}
				contentType := detectContentType(filePath)
				c.Data(http.StatusOK, contentType, data)
			})
		}
		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path

			if strings.HasPrefix(path, "/api/") {
				c.JSON(http.StatusNotFound, response.Response{
					Code:    1,
					Message: "接口不存在",
				})
				return
			}

			serveEmbeddedFile(c, webFS, "assets/index.html")
		})
	} else {
		r.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path

			if strings.HasPrefix(path, "/api/") {
				c.JSON(http.StatusNotFound, response.Response{
					Code:    1,
					Message: "接口不存在",
				})
				return
			}

			c.File("assets/index.html")
		})
	}

	api := r.Group("/api")
	{
		api.GET("/site/config", authCtrl.GetSiteConfig)

		api.GET("/captcha", authCtrl.GetCaptcha)

		auth := api.Group("/auth")
		{
			auth.POST("/login", authCtrl.Login)
			auth.POST("/register", authCtrl.Register)
			auth.GET("/profile", authMw.AuthRequired(), authCtrl.GetProfile)
		}

		products := api.Group("/products")
		{
			products.GET("", productCtrl.List)
			products.GET("/on-shelf", productCtrl.OnShelf)
			products.GET("/on-shelf-grouped", productCtrl.OnShelfGrouped)
			products.GET("/:id", productCtrl.GetByID)
		}

		orders := api.Group("/orders")
		{
			orders.POST("", orderCtrl.Create)
			orders.GET("/query", orderCtrl.Query)
			orders.GET("/:order_no", orderCtrl.GetByOrderNo)
			orders.POST("/notify", orderCtrl.Notify)
			orders.GET("/return", orderCtrl.Return)
		}

		cards := api.Group("/cards")
		{
			cards.GET("/product/:id/count", cardCtrl.CountByProduct)
		}

		install := api.Group("/install")
		{
			install.GET("/env", authCtrl.CheckEnv)
			install.GET("/license-status", authCtrl.GetLicenseStatus)
			install.POST("/verify-license", authCtrl.VerifyLicense)
			install.POST("/test-database", authCtrl.TestDatabase)
			install.POST("", authCtrl.Install)
		}

		admin := api.Group("/admin", authMw.AuthRequired())
		{
			admin.GET("/system/status", adminCtrl.SystemStatus)
			admin.GET("/system/license", adminCtrl.LicenseStatus)
			admin.POST("/system/license/verify", adminCtrl.VerifyLicense)
			admin.GET("/dashboard", adminCtrl.Dashboard)
			admin.GET("/dashboard/order-status", adminCtrl.OrderStatusCounts)

			categoryAdmin := admin.Group("/categories")
			{
				categoryAdmin.GET("", adminCtrl.CategoryList)
				categoryAdmin.GET("/all", adminCtrl.CategoryAll)
				categoryAdmin.POST("", adminCtrl.CategoryCreate)
				categoryAdmin.PUT("", adminCtrl.CategoryUpdate)
				categoryAdmin.DELETE("/:id", adminCtrl.CategoryDelete)
			}

			productAdmin := admin.Group("/products")
			{
				productAdmin.GET("", adminCtrl.ProductList)
				productAdmin.POST("", adminCtrl.ProductCreate)
				productAdmin.PUT("", adminCtrl.ProductUpdate)
				productAdmin.DELETE("/:id", adminCtrl.ProductDelete)
			}

			cardAdmin := admin.Group("/cards")
			{
				cardAdmin.POST("/import", adminCtrl.CardImport)
				cardAdmin.GET("", adminCtrl.CardList)
				cardAdmin.DELETE("/:id", adminCtrl.CardDelete)
				cardAdmin.GET("/export", adminCtrl.CardExport)
			}

			orderAdmin := admin.Group("/orders")
			{
				orderAdmin.GET("", adminCtrl.OrderList)
				orderAdmin.GET("/logs", adminCtrl.OrderLogs)
			}

			paymentAdmin := admin.Group("/payments")
			{
				paymentAdmin.GET("", adminCtrl.PaymentList)
				paymentAdmin.GET("/all", adminCtrl.PaymentAll)
				paymentAdmin.POST("", adminCtrl.PaymentCreate)
				paymentAdmin.PUT("", adminCtrl.PaymentUpdate)
				paymentAdmin.DELETE("/:id", adminCtrl.PaymentDelete)
			}

			emailAdmin := admin.Group("/emails")
			{
				emailAdmin.GET("", adminCtrl.EmailList)
				emailAdmin.POST("", adminCtrl.EmailCreate)
				emailAdmin.PUT("", adminCtrl.EmailUpdate)
				emailAdmin.DELETE("/:id", adminCtrl.EmailDelete)
				emailAdmin.POST("/test/:id", adminCtrl.EmailTest)
				emailAdmin.GET("/logs", adminCtrl.EmailLogs)
			}

			logAdmin := admin.Group("/logs")
			{
				logAdmin.GET("/operations", adminCtrl.OperationLogs)
				logAdmin.GET("/logins", adminCtrl.LoginLogs)
			}

			nodeAdmin := admin.Group("/nodes")
			{
				nodeAdmin.GET("", adminCtrl.NodeList)
				nodeAdmin.POST("", adminCtrl.NodeCreate)
				nodeAdmin.PUT("", adminCtrl.NodeUpdate)
				nodeAdmin.DELETE("/:id", adminCtrl.NodeDelete)
				nodeAdmin.POST("/ping/:id", adminCtrl.NodePing)
			}

			siteAdmin := admin.Group("/settings")
			{
				siteAdmin.GET("", adminCtrl.GetSettings)
				siteAdmin.PUT("", adminCtrl.UpdateSettings)
			}

			upgradeAdmin := admin.Group("/upgrade")
			{
				upgradeAdmin.GET("/version", adminCtrl.GetVersion)
				upgradeAdmin.GET("/check", adminCtrl.CheckUpdate)
				upgradeAdmin.POST("/upload", adminCtrl.UploadPackage)
				upgradeAdmin.POST("/apply", adminCtrl.ApplyUpgrade)
				upgradeAdmin.GET("/logs", adminCtrl.UpgradeLogs)
			}

			uploadAdmin := admin.Group("/upload")
			{
				uploadAdmin.POST("", adminCtrl.UploadFile)
				uploadAdmin.GET("", adminCtrl.ListFiles)
				uploadAdmin.GET("/:id", adminCtrl.GetFile)
				uploadAdmin.DELETE("/:id", adminCtrl.DeleteFile)
			}
		}
	}

	return r
}

func serveEmbeddedFile(c *gin.Context, fsys embed.FS, name string) {
	data, err := fsys.ReadFile(name)
	if err != nil {
		c.JSON(http.StatusNotFound, response.Response{
			Code:    1,
			Message: "页面不存在",
		})
		return
	}

	contentType := "text/html; charset=utf-8"
	if strings.HasSuffix(name, ".js") {
		contentType = "application/javascript; charset=utf-8"
	} else if strings.HasSuffix(name, ".css") {
		contentType = "text/css; charset=utf-8"
	} else if strings.HasSuffix(name, ".png") {
		contentType = "image/png"
	} else if strings.HasSuffix(name, ".jpg") || strings.HasSuffix(name, ".jpeg") {
		contentType = "image/jpeg"
	} else if strings.HasSuffix(name, ".svg") {
		contentType = "image/svg+xml"
	} else if strings.HasSuffix(name, ".ico") {
		contentType = "image/x-icon"
	}

	c.Data(http.StatusOK, contentType, data)
}

func detectContentType(name string) string {
	if strings.HasSuffix(name, ".js") {
		return "application/javascript; charset=utf-8"
	} else if strings.HasSuffix(name, ".css") {
		return "text/css; charset=utf-8"
	} else if strings.HasSuffix(name, ".png") {
		return "image/png"
	} else if strings.HasSuffix(name, ".jpg") || strings.HasSuffix(name, ".jpeg") {
		return "image/jpeg"
	} else if strings.HasSuffix(name, ".svg") {
		return "image/svg+xml"
	} else if strings.HasSuffix(name, ".ico") {
		return "image/x-icon"
	} else if strings.HasSuffix(name, ".map") {
		return "application/json"
	} else if strings.HasSuffix(name, ".json") {
		return "application/json"
	}
	return "application/octet-stream"
}
