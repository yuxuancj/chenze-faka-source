package license

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type VersionCheckRequest struct {
	AppKey        string `json:"appKey"`
	CurrentVersion string `json:"currentVersion"`
	Domain        string `json:"domain,omitempty"`
	ServerIP      string `json:"serverIp,omitempty"`
	LicenseKey    string `json:"licenseKey,omitempty"`
	Timestamp     int64  `json:"timestamp"`
	Sign          string `json:"sign"`
}

type VersionInfo struct {
	HasUpdate      bool   `json:"hasUpdate"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	Version        string `json:"version"`
	Title          string `json:"title"`
	Changelog      string `json:"changelog"`
	DownloadURL    string `json:"downloadUrl"`
	FileSizeMb     float64 `json:"fileSizeMb"`
	FileMD5        string `json:"fileMd5"`
	ForceUpdate    bool   `json:"forceUpdate"`
	MinVersion     string `json:"minVersion"`
	PublishedAt    string `json:"publishedAt"`
	Updates        []struct {
		Version   string `json:"version"`
		Title     string `json:"title"`
		Changelog string `json:"changelog"`
		UpdateSQL string `json:"updateSql"`
	} `json:"updates"`
}

type VersionCheckResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data *VersionInfo `json:"data,omitempty"`
}

func (s *Service) CheckVersion(currentVersion string) (*VersionInfo, error) {
	if !s.cfg.Enabled {
		return nil, nil
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

	timestamp := time.Now().Unix()
	signMessage := s.cfg.AppKey + "\n" + currentVersion + "\n" + target + "\n" + strconv.FormatInt(timestamp, 10)
	sign := hmacSHA256(s.cfg.AppSecret, signMessage)

	reqBody := VersionCheckRequest{
		AppKey:         s.cfg.AppKey,
		CurrentVersion: currentVersion,
		Domain:         domain,
		ServerIP:       serverIP,
		LicenseKey:     s.cfg.LicenseKey,
		Timestamp:      timestamp,
		Sign:           sign,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := strings.TrimRight(s.cfg.BaseURL, "/") + "/api/app/version/check"
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result VersionCheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Code == 200 && result.Data != nil {
		return result.Data, nil
	}

	return nil, nil
}

func (s *Service) GetCurrentVersion() string {
	return version
}

const version = "1.0.0"
