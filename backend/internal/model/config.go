package model

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
	SQLite   string `json:"sqlite_path" yaml:"sqlite_path"`
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
	BackupURL   string `json:"backup_url" yaml:"backup_url"`
	AppKey      string `json:"app_key" yaml:"app_key"`
	AppSecret   string `json:"app_secret" yaml:"app_secret"`
	LicenseKey  string `json:"license_key" yaml:"license_key"`
	Domain      string `json:"domain" yaml:"domain"`
	ServerIP    string `json:"server_ip" yaml:"server_ip"`
	CacheFile   string `json:"cache_file" yaml:"cache_file"`
	Interval    int    `json:"interval" yaml:"interval"`
	GracePeriod int    `json:"grace_period" yaml:"grace_period"`
}