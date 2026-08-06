package service

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

	"chenze-faka/internal/model"
)

type LicenseService struct {
	cfg       *model.LicenseConfig
	mu        sync.RWMutex
	cache     *LicenseCache
	verified  bool
	stopCh    chan struct{}
	baseURLs  []string
}

type LicenseCache struct {
	LastVerifyTime  int64  `json:"last_verify_time"`
	LastSuccessTime int64  `json:"last_success_time"`
	LastResult      string `json:"last_result"`
	LastMessage     string `json:"last_message"`
	ExpireAt        string `json:"expire_at,omitempty"`
	AppName         string `json:"app_name,omitempty"`
	UsedServer      string `json:"used_server,omitempty"`
}

type verifyRequest struct {
	AppKey     string `json:"appKey"`
	Domain     string `json:"domain,omitempty"`
	ServerIP   string `json:"serverIp,omitempty"`
	LicenseKey string `json:"licenseKey,omitempty"`
	Timestamp  int64  `json:"timestamp"`
	Sign       string `json:"sign"`
}

type verifyResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Result   string `json:"result"`
		AppName  string `json:"appName,omitempty"`
		ExpireAt string `json:"expireAt,omitempty"`
		Reason   string `json:"reason,omitempty"`
	} `json:"data"`
}

type QuickVerifyResult struct {
	Verified       bool   `json:"verified"`
	AppName        string `json:"app_name,omitempty"`
	ExpireAt       string `json:"expire_at,omitempty"`
	NeedRealName   bool   `json:"need_real_name,omitempty"`
	Account        string `json:"account,omitempty"`
	Desc           string `json:"desc,omitempty"`
	LicenseKey     string `json:"license_key,omitempty"`
}

func NewLicenseService(cfg *model.LicenseConfig) *LicenseService {
	if cfg == nil {
		cfg = &model.LicenseConfig{}
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://auth.seanld.com"
	}
	if cfg.AppKey == "" {
		cfg.AppKey = "app_1c1467945bb2_3105"
	}
	if cfg.AppSecret == "" {
		if v := os.Getenv("LICENSE_APP_SECRET"); v != "" {
			cfg.AppSecret = v
		} else if data, err := os.ReadFile("license.secret"); err == nil {
			cfg.AppSecret = strings.TrimSpace(string(data))
		}
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

	baseURLs := []string{}
	if cfg.BaseURL != "" {
		baseURLs = append(baseURLs, cfg.BaseURL)
	}
	if cfg.BackupURL != "" && cfg.BackupURL != cfg.BaseURL {
		baseURLs = append(baseURLs, cfg.BackupURL)
	}
	if len(baseURLs) == 0 {
		baseURLs = append(baseURLs, "https://auth.seanld.com")
		baseURLs = append(baseURLs, "http://220.167.100.148:19127")
	}

	s := &LicenseService{
		cfg:      cfg,
		stopCh:   make(chan struct{}),
		baseURLs: baseURLs,
	}

	s.loadCache()
	return s
}

func (s *LicenseService) Verify() (bool, error) {
	if !s.cfg.Enabled {
		s.mu.Lock()
		s.verified = true
		s.mu.Unlock()
		return true, nil
	}

	domain := s.cfg.Domain
	serverIP := s.cfg.ServerIP

	if domain == "" {
		domain = getHostname()
	}
	if serverIP == "" {
		serverIP = "127.0.0.1"
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
	sign := md5Hex(s.cfg.AppKey + target + strconv.FormatInt(timestamp, 10) + s.cfg.AppSecret)

	reqBody := verifyRequest{
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

	var lastError error
	for _, baseURL := range s.baseURLs {
		url := strings.TrimRight(baseURL, "/") + "/api/license/verify"
		resp, err := http.Post(url, "application/json", bytes.NewReader(body))
		if err != nil {
			lastError = err
			continue
		}

		var result verifyResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			lastError = err
			continue
		}
		resp.Body.Close()

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
			s.cache.UsedServer = baseURL
			s.verified = true
			s.saveCache()
			s.mu.Unlock()
			return true, nil
		}

		lastError = nil
		s.mu.Unlock()
	}

	s.verified = false
	s.handleError("所有授权站验证失败")

	if lastError != nil {
		return s.IsGracePeriodValid(), nil
	}
	return s.IsGracePeriodValid(), nil
}

func (s *LicenseService) QuickVerify(licenseKey, domain, serverIP string) (*QuickVerifyResult, error) {
	cfg := &model.LicenseConfig{
		Enabled:     s.cfg.Enabled,
		BaseURL:     s.cfg.BaseURL,
		BackupURL:   s.cfg.BackupURL,
		AppKey:      s.cfg.AppKey,
		AppSecret:   s.cfg.AppSecret,
		LicenseKey:  licenseKey,
		Domain:      domain,
		ServerIP:    serverIP,
		CacheFile:   s.cfg.CacheFile,
		Interval:    s.cfg.Interval,
		GracePeriod: s.cfg.GracePeriod,
	}

	svc := NewLicenseService(cfg)
	ok, err := svc.Verify()
	if err != nil {
		return nil, err
	}

	result := &QuickVerifyResult{
		Verified:   ok,
		LicenseKey: licenseKey,
	}

	if ok {
		cache := svc.GetCache()
		result.AppName = cache.AppName
		result.ExpireAt = cache.ExpireAt
	}

	return result, nil
}

func (s *LicenseService) IsVerified() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.verified
}

func (s *LicenseService) IsGracePeriodValid() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.cache.LastSuccessTime <= 0 {
		return false
	}

	now := time.Now().Unix()
	graceEnd := s.cache.LastSuccessTime + int64(s.cfg.GracePeriod)
	return now < graceEnd
}

func (s *LicenseService) GetCache() *LicenseCache {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cache := &LicenseCache{}
	*cache = *s.cache
	return cache
}

func (s *LicenseService) StartPeriodicVerify() {
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
				s.Verify()
			case <-s.stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *LicenseService) StopPeriodicVerify() {
	if s.stopCh != nil {
		select {
		case s.stopCh <- struct{}{}:
		default:
		}
	}
}

func (s *LicenseService) handleError(msg string) {
	now := time.Now().Unix()
	s.mu.Lock()
	s.cache.LastVerifyTime = now
	s.cache.LastMessage = msg
	s.mu.Unlock()
	s.saveCache()
}

func (s *LicenseService) loadCache() {
	data, err := os.ReadFile(s.cfg.CacheFile)
	if err != nil {
		s.cache = &LicenseCache{}
		return
	}

	var cache LicenseCache
	if err := json.Unmarshal(data, &cache); err != nil {
		s.cache = &LicenseCache{}
		return
	}

	s.cache = &cache
}

func (s *LicenseService) saveCache() {
	if s.cache == nil {
		return
	}

	data, err := json.Marshal(s.cache)
	if err != nil {
		return
	}

	os.WriteFile(s.cfg.CacheFile, data, 0644)
}

func md5Hex(text string) string {
	sum := md5.Sum([]byte(text))
	return hex.EncodeToString(sum[:])
}

func hmacSHA256(key, message string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "localhost"
	}
	return hostname
}
