package config

import (
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	System   SystemConfig   `yaml:"system"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	Pay      PayConfig      `yaml:"pay"`
	License  LicenseConfig  `yaml:"license"`
}

type SystemConfig struct {
	SiteName string `json:"site_name" yaml:"site_name"`
	Port     int    `json:"port" yaml:"port"`
	Mode     string `json:"mode" yaml:"mode"`
}

type DatabaseConfig struct {
	Driver   string `json:"driver" yaml:"driver"`
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	User     string `json:"username" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	DBName   string `json:"database" yaml:"dbname"`
}

type JWTConfig struct {
	Secret     string `json:"secret" yaml:"secret"`
	ExpireTime int    `json:"expire_time" yaml:"expire_hours"`
}

type PayConfig struct {
	URL      string `json:"url" yaml:"url"`
	Merchant string `json:"merchant" yaml:"merchant"`
	Key      string `json:"key" yaml:"key"`
}

type LicenseConfig struct {
	Enabled     bool   `json:"enabled" yaml:"enabled"`
	BaseURL     string `json:"base_url" yaml:"base_url"`
	AppKey      string `json:"app_key" yaml:"app_key"`
	AppSecret   string `json:"app_secret" yaml:"app_secret"`
	LicenseKey  string `json:"license_key" yaml:"license_key"`
	Domain      string `json:"domain" yaml:"domain"`
	ServerIP    string `json:"server_ip" yaml:"server_ip"`
	CacheFile   string `json:"cache_file" yaml:"cache_file"`
	Interval    int    `json:"interval" yaml:"interval"`
	GracePeriod int    `json:"grace_period" yaml:"grace_period"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return defaultConfig(), nil
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return defaultConfig(), nil
	}

	if cfg.System.Port == 0 {
		cfg.System.Port = 12398
	}
	if cfg.System.Mode == "" {
		cfg.System.Mode = "debug"
	}
	if cfg.Database.Port == 0 {
		cfg.Database.Port = 3306
	}
	if cfg.JWT.ExpireTime == 0 {
		cfg.JWT.ExpireTime = 72
	}
	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = "chenze-faka-secret"
	}

	return &cfg, nil
}

func defaultConfig() *Config {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	if port == 0 {
		port = 12398
	}

	return &Config{
		System: SystemConfig{
			SiteName: "晨泽发卡",
			Port:     port,
			Mode:     getEnv("MODE", "debug"),
		},
		Database: DatabaseConfig{
			Driver:   "mysql",
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 3306),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "chenze_faka"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "chenze-faka-secret"),
			ExpireTime: getEnvInt("JWT_EXPIRE", 72),
		},
		Pay: PayConfig{},
		License: LicenseConfig{
			Enabled:     true,
			BaseURL:     "https://auth.seanld.com",
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

func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
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
