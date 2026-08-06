package database

import (
	"fmt"
	"log"

	"chenze-faka/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(cfg *model.DatabaseConfig) error {
	var err error

	switch cfg.Driver {
	case "sqlite":
		DB, err = initSQLite(cfg)
	default:
		DB, err = initMySQL(cfg)
	}

	if err != nil {
		DB = nil
		return err
	}

	sqlDB, err := DB.DB()
	if err != nil {
		DB = nil
		return err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	log.Printf("[DB] connected via %s", cfg.Driver)
	return nil
}

func initMySQL(cfg *model.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func initSQLite(cfg *model.DatabaseConfig) (*gorm.DB, error) {
	path := cfg.SQLite
	if path == "" {
		path = "chenze_faka.db"
	}

	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func AutoMigrate() error {
	if DB == nil {
		return nil
	}
	return DB.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.Product{},
		&model.Card{},
		&model.Order{},
		&model.PaymentChannel{},
		&model.EmailConfig{},
		&model.Node{},
		&model.OperationLog{},
		&model.OrderLog{},
		&model.EmailLog{},
		&model.UpgradeLog{},
		&model.FileUpload{},
	)
}

func Close() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

func IsAvailable() bool {
	return DB != nil
}

type TestResult struct {
	Version    string `json:"version"`
	HasData    bool   `json:"has_data"`
	TableCount int    `json:"table_count"`
}

func TestConnection(cfg *model.DatabaseConfig) (*TestResult, error) {
	var testDB *gorm.DB
	var err error

	switch cfg.Driver {
	case "sqlite":
		testDB, err = initSQLite(cfg)
	default:
		testDB, err = initMySQL(cfg)
	}

	if err != nil {
		return nil, err
	}

	sqlDB, err := testDB.DB()
	if err != nil {
		return nil, err
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	var version string
	if cfg.Driver != "sqlite" {
		row := sqlDB.QueryRow("SELECT VERSION()")
		if err := row.Scan(&version); err != nil {
			version = "unknown"
		}
	} else {
		version = "3.x"
	}

	hasData := false
	tableCount := 0

	if cfg.Driver != "sqlite" && cfg.DBName != "" {
		row := sqlDB.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE()")
		if err := row.Scan(&tableCount); err == nil && tableCount > 0 {
			hasData = true
		}
	} else if cfg.Driver == "sqlite" {
		var count int
		if err := sqlDB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table'").Scan(&count); err == nil {
			tableCount = count
			if count > 0 {
				hasData = true
			}
		}
	}

	return &TestResult{
		Version:    version,
		HasData:    hasData,
		TableCount: tableCount,
	}, nil
}