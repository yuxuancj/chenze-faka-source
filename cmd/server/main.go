package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"chenze-faka/internal/config"
	"chenze-faka/internal/license"
	"chenze-faka/internal/router"
	"chenze-faka/pkg/database"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Printf("warning: failed to load config.yaml, using defaults: %v", err)
	}

	installed := checkInstalled()

	licenseSvc := initLicenseService(cfg)

	if installed {
		if err := database.Init(&cfg.Database); err != nil {
			log.Printf("warning: failed to connect database: %v", err)
			log.Printf("server will start in degraded mode, please check database configuration")
		} else {
			defer database.Close()
			startOrderCleanupTask()
		}
	}

	r := router.Setup(cfg, licenseSvc)

	addr := fmt.Sprintf(":%d", cfg.System.Port)
	log.Printf("server starting on %s", addr)

	go func() {
		if err := r.Run(addr); err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	licenseSvc.StopPeriodicVerify()
	log.Println("server stopped")
}

func initLicenseService(cfg *config.Config) *license.Service {
	svc := license.NewService(&license.Config{
		Enabled:     cfg.License.Enabled,
		BaseURL:     cfg.License.BaseURL,
		AppKey:      cfg.License.AppKey,
		AppSecret:   cfg.License.AppSecret,
		LicenseKey:  cfg.License.LicenseKey,
		Domain:      cfg.License.Domain,
		ServerIP:    cfg.License.ServerIP,
		CacheFile:   cfg.License.CacheFile,
		Interval:    cfg.License.Interval,
		GracePeriod: cfg.License.GracePeriod,
	})

	if cfg.License.Enabled {
		log.Println("verifying license...")
		ok, err := svc.Verify()
		if err != nil {
			log.Printf("license verification error: %v", err)
		}
		if ok {
			log.Println("license verification passed")
		} else {
			if svc.IsGracePeriodValid() {
				log.Println("license verification failed but within grace period")
			} else {
				log.Println("license verification failed, system will run in restricted mode")
			}
		}

		svc.StartPeriodicVerify()
	} else {
		log.Println("license verification disabled")
	}

	return svc
}

func checkInstalled() bool {
	_, err := os.Stat("install.lock")
	return err == nil
}

func startOrderCleanupTask() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			log.Println("running expired order cleanup task")
		}
	}()
}
