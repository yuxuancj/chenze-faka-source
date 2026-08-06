package model

type Config struct {
	System   SystemConfig   `yaml:"system"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	Pay      PayConfig      `yaml:"pay"`
	License  LicenseConfig  `yaml:"license"`
}

type SystemConfig struct {
	SiteName string `yaml:"site_name"`
	Port     int    `yaml:"port"`
	Mode     string `yaml:"mode"`
}

type DatabaseConfig struct {
	Driver   string `yaml:"driver"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SQLite   string `yaml:"sqlite_path"`
}

type JWTConfig struct {
	Secret     string `yaml:"secret"`
	ExpireTime int    `yaml:"expire_hours"`
}

type PayConfig struct {
	URL      string `yaml:"url"`
	Merchant string `yaml:"merchant"`
	Key      string `yaml:"key"`
}

type LicenseConfig struct {
	Enabled    bool   `yaml:"enabled"`
	BaseURL    string `yaml:"base_url"`
	AppKey     string `yaml:"app_key"`
	AppSecret  string `yaml:"app_secret"`
	LicenseKey string `yaml:"license_key"`
	Domain     string `yaml:"domain"`
	ServerIP   string `yaml:"server_ip"`
	CacheFile  string `yaml:"cache_file"`
	Interval   int    `yaml:"interval"`
	GracePeriod int   `yaml:"grace_period"`
}