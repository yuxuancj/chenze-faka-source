package main

import (
	faka "chenze-faka"
	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"
	"chenze-faka/internal/router"
	"chenze-faka/internal/service"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"
)

func main() {
	cfg, err := loadConfig("config.yaml")
	if err != nil {
		log.Printf("警告: 加载配置文件失败,使用默认配置: %v", err)
	}

	installed := checkInstalled()

	licenseSvc := initLicenseService(cfg, installed)

	if installed {
		if err := database.Init(&cfg.Database); err != nil {
			log.Printf("警告: 数据库连接失败: %v", err)
			log.Printf("服务器将以降级模式运行,请检查数据库配置")
		} else {
			defer database.Close()
			if err := database.AutoMigrate(); err != nil {
				log.Printf("警告: 数据库迁移失败: %v", err)
			}
			startCleanupTask()
		}
	}

	r := router.Setup(cfg, licenseSvc, faka.WebFS)

	addr := fmt.Sprintf(":%d", cfg.System.Port)
	log.Printf("服务器启动于 %s", addr)

	go func() {
		if err := r.Run(addr); err != nil {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务器...")
	licenseSvc.StopPeriodicVerify()
	log.Println("服务器已停止")
}

func loadConfig(path string) (*model.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return defaultConfig(), nil
	}

	var cfg model.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return defaultConfig(), nil
	}

	if cfg.System.Port == 0 {
		cfg.System.Port = 12398
	}
	if cfg.System.Mode == "" {
		cfg.System.Mode = "debug"
	}
	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = "chenze-faka-secret"
	}
	if cfg.JWT.ExpireTime == 0 {
		cfg.JWT.ExpireTime = 72
	}
	if cfg.Database.Port == 0 {
		cfg.Database.Port = 3306
	}

	return &cfg, nil
}

func defaultConfig() *model.Config {
	port, _ := strconv.Atoi(getEnv("PORT", "12398"))
	return &model.Config{
		System: model.SystemConfig{
			SiteName: "晨泽发卡",
			Port:     port,
			Mode:     getEnv("MODE", "debug"),
		},
		Database: model.DatabaseConfig{
			Driver:   getEnv("DB_DRIVER", "mysql"),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 3306),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "chenze_faka"),
			SQLite:   getEnv("DB_SQLITE_PATH", ""),
		},
		JWT: model.JWTConfig{
			Secret:     getEnv("JWT_SECRET", "chenze-faka-secret"),
			ExpireTime: getEnvInt("JWT_EXPIRE", 72),
		},
		Pay: model.PayConfig{},
		License: model.LicenseConfig{
			Enabled:     true,
			BaseURL:     "https://auth.seanld.com",
			BackupURL:   "http://220.167.100.148:19127",
			AppKey:      "app_1c1467945bb2_3105",
			AppSecret:   getEnv("LICENSE_APP_SECRET", ""),
			LicenseKey:  getEnv("LICENSE_KEY", ""),
			Domain:      getEnv("LICENSE_DOMAIN", ""),
			ServerIP:    getEnv("LICENSE_SERVER_IP", ""),
			CacheFile:   "license.cache",
			Interval:    3600,
			GracePeriod: 86400,
		},
	}
}

func initLicenseService(cfg *model.Config, installed bool) *service.LicenseService {
	svc := service.NewLicenseService(&cfg.License)

	if cfg.License.Enabled && installed {
		log.Println("正在验证授权...")
		ok, err := svc.Verify()
		if err != nil {
			log.Printf("授权验证错误: %v", err)
		}
		if ok {
			log.Println("授权验证通过")
		} else {
			if svc.IsGracePeriodValid() {
				log.Println("授权验证失败,但处于宽限期内")
			} else {
				log.Println("授权验证失败,系统将以受限模式运行")
			}
		}

		svc.StartPeriodicVerify()
	} else if !cfg.License.Enabled {
		log.Println("授权验证已禁用")
	}

	return svc
}

func checkInstalled() bool {
	_, err := os.Stat("install.lock")
	return err == nil
}

func startCleanupTask() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			log.Println("执行过期订单清理...")
		}
	}()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}