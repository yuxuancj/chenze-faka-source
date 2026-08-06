package license

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DefaultBaseURL = "https://auth.seanld.com"
	DefaultAppKey  = "app_1c1467945bb2_3105"
)

var (
	DefaultSecret     = ""
	defaultSecretFile = "license.secret"
)

func init() {
	if DefaultSecret == "" {
		if data, err := os.ReadFile(defaultSecretFile); err == nil {
			DefaultSecret = strings.TrimSpace(string(data))
		}
	}
	if DefaultSecret == "" {
		DefaultSecret = os.Getenv("LICENSE_APP_SECRET")
	}
}

type Config struct {
	Enabled     bool   `yaml:"enabled" json:"enabled"`
	BaseURL     string `yaml:"base_url" json:"base_url"`
	AppKey      string `yaml:"app_key" json:"app_key"`
	AppSecret   string `yaml:"app_secret" json:"app_secret"`
	LicenseKey  string `yaml:"license_key" json:"license_key"`
	Domain      string `yaml:"domain" json:"domain"`
	ServerIP    string `yaml:"server_ip" json:"server_ip"`
	CacheFile   string `yaml:"cache_file" json:"cache_file"`
	Interval    int    `yaml:"interval" json:"interval"`
	GracePeriod int    `yaml:"grace_period" json:"grace_period"`
}

type VerifyRequest struct {
	AppKey     string `json:"appKey"`
	Domain     string `json:"domain,omitempty"`
	ServerIP   string `json:"serverIp,omitempty"`
	LicenseKey string `json:"licenseKey,omitempty"`
	Timestamp  int64  `json:"timestamp"`
	Sign       string `json:"sign"`
}

type VerifyResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Result   string `json:"result"`
		AppName  string `json:"appName,omitempty"`
		ExpireAt string `json:"expireAt,omitempty"`
		Reason   string `json:"reason,omitempty"`
	} `json:"data"`
}

type CacheData struct {
	LastVerifyTime  int64  `json:"last_verify_time"`
	LastSuccessTime int64  `json:"last_success_time"`
	LastResult      string `json:"last_result"`
	LastMessage     string `json:"last_message"`
	ExpireAt        string `json:"expire_at,omitempty"`
	AppName         string `json:"app_name,omitempty"`
}

type Service struct {
	cfg      *Config
	mu       sync.RWMutex
	cache    *CacheData
	verified bool
	stopCh   chan struct{}
}

func DefaultConfig() *Config {
	return &Config{
		Enabled:     true,
		BaseURL:     DefaultBaseURL,
		AppKey:      DefaultAppKey,
		AppSecret:   DefaultSecret,
		LicenseKey:  "",
		Domain:      "",
		ServerIP:    "",
		CacheFile:   "license.cache",
		Interval:    3600,
		GracePeriod: 86400,
	}
}

func NewService(cfg *Config) *Service {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}
	if cfg.AppKey == "" {
		cfg.AppKey = DefaultAppKey
	}
	if cfg.AppSecret == "" {
		cfg.AppSecret = DefaultSecret
	}
	if cfg.Interval <= 0 {
		cfg.Interval = 3600
	}
	if cfg.GracePeriod <= 0 {
		cfg.GracePeriod = 86400
	}
	if cfg.CacheFile == "" {
		cfg.CacheFile = "license.cache"
	}

	s := &Service{
		cfg:    cfg,
		stopCh: make(chan struct{}),
	}

	s.loadCache()
	return s
}

func (s *Service) Verify() (bool, error) {
	if !s.cfg.Enabled {
		s.mu.Lock()
		s.verified = true
		s.mu.Unlock()
		return true, nil
	}

	domain := s.cfg.Domain
	serverIP := s.cfg.ServerIP

	if domain == "" {
		domain = getDefaultDomain()
	}
	if serverIP == "" {
		serverIP = getDefaultServerIP()
	}

	target := s.cfg.LicenseKey
	if target == "" {
		target = domain
	}
	if target == "" {
		target = serverIP
	}
	if target == "" {
		return false, nil
	}

	timestamp := time.Now().Unix()
	sign := md5Text(s.cfg.AppKey + target + strconv.FormatInt(timestamp, 10) + s.cfg.AppSecret)

	reqBody := VerifyRequest{
		AppKey:     s.cfg.AppKey,
		Domain:     domain,
		ServerIP:   serverIP,
		LicenseKey: s.cfg.LicenseKey,
		Timestamp:  timestamp,
		Sign:       sign,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return false, err
	}

	url := strings.TrimRight(s.cfg.BaseURL, "/") + "/api/license/verify"
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		s.handleVerifyError("network error: " + err.Error())
		return s.IsGracePeriodValid(), nil
	}
	defer resp.Body.Close()

	var result VerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.handleVerifyError("decode error: " + err.Error())
		return s.IsGracePeriodValid(), nil
	}

	now := time.Now().Unix()
	s.mu.Lock()
	s.cache.LastVerifyTime = now
	s.cache.LastResult = result.Data.Result
	s.cache.LastMessage = result.Msg

	if result.Code == 200 && result.Data.Result == "pass" {
		s.cache.LastSuccessTime = now
		s.cache.LastResult = "pass"
		s.cache.ExpireAt = result.Data.ExpireAt
		s.cache.AppName = result.Data.AppName
		s.verified = true
		s.saveCache()
		s.mu.Unlock()
		return true, nil
	}

	if result.Data.Result == "expired" {
		s.log("license expired: " + result.Msg)
	} else if result.Data.Result == "blacklisted" {
		s.log("license blacklisted: " + result.Msg)
	}

	s.verified = false
	s.mu.Unlock()
	s.saveCache()

	return s.IsGracePeriodValid(), nil
}

func (s *Service) handleVerifyError(msg string) {
	now := time.Now().Unix()
	s.mu.Lock()
	s.cache.LastVerifyTime = now
	s.cache.LastMessage = msg
	s.mu.Unlock()
	s.saveCache()
}

func (s *Service) IsVerified() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.verified
}

func (s *Service) StartPeriodicVerify() {
	if !s.cfg.Enabled {
		return
	}

	s.StopPeriodicVerify()
	s.stopCh = make(chan struct{})

	interval := time.Duration(s.cfg.Interval) * time.Second
	ticker := time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-ticker.C:
				ok, err := s.Verify()
				if err != nil {
					s.log("verify error: " + err.Error())
				} else if !ok {
					s.log("verification failed, grace period: " + strconv.Itoa(s.cfg.GracePeriod) + "s")
				} else {
					s.log("verification passed")
				}
			case <-s.stopCh:
				ticker.Stop()
				return
			}
		}
	}()

	s.log("periodic verification started, interval=" + strconv.Itoa(s.cfg.Interval) + "s")
}

func (s *Service) StopPeriodicVerify() {
	if s.stopCh != nil {
		select {
		case s.stopCh <- struct{}{}:
		default:
		}
	}
}

func (s *Service) GetCache() *CacheData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cache := &CacheData{}
	*cache = *s.cache
	return cache
}

func (s *Service) IsGracePeriodValid() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.cache.LastSuccessTime <= 0 {
		return false
	}

	now := time.Now().Unix()
	graceEnd := s.cache.LastSuccessTime + int64(s.cfg.GracePeriod)
	return now < graceEnd
}

func (s *Service) loadCache() {
	data, err := os.ReadFile(s.cfg.CacheFile)
	if err != nil {
		s.cache = &CacheData{}
		return
	}

	var cache CacheData
	if err := json.Unmarshal(data, &cache); err != nil {
		s.cache = &CacheData{}
		return
	}

	s.cache = &cache
}

func (s *Service) saveCache() {
	if s.cache == nil {
		return
	}

	data, err := json.Marshal(s.cache)
	if err != nil {
		return
	}

	os.WriteFile(s.cfg.CacheFile, data, 0644)
}

func (s *Service) log(msg string) {
	domain := s.cfg.Domain
	if domain == "" {
		domain = "localhost"
	}
	println("[License][" + domain + "] " + msg)
}

func Verify(domain, serverIP, licenseKey string) bool {
	s := &Service{
		cfg: DefaultConfig(),
	}

	s.cfg.Domain = domain
	s.cfg.ServerIP = serverIP
	s.cfg.LicenseKey = licenseKey

	if domain == "" {
		s.cfg.Domain = getDefaultDomain()
	}
	if serverIP == "" {
		s.cfg.ServerIP = getDefaultServerIP()
	}

	ok, _ := s.Verify()
	return ok
}

func md5Text(text string) string {
	sum := md5.Sum([]byte(text))
	return hex.EncodeToString(sum[:])
}

func hmacSHA256(key, message string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

func getDefaultDomain() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "localhost"
	}
	return hostname
}

func getDefaultServerIP() string {
	return "127.0.0.1"
}
