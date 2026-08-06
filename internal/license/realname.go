package license

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type RealNameStatusRequest struct {
	AppKey     string `json:"appKey"`
	LicenseKey string `json:"licenseKey,omitempty"`
	Domain     string `json:"domain,omitempty"`
	ServerIP   string `json:"serverIp,omitempty"`
	Timestamp  int64  `json:"timestamp"`
	Sign       string `json:"sign"`
}

type RealNameStatusResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		NeedRealName  bool   `json:"needRealName"`
		Account       string `json:"account,omitempty"`
		RealNameType  string `json:"realNameType,omitempty"`
		RealNameDesc  string `json:"realNameDesc,omitempty"`
	} `json:"data"`
}

type RealNameResult struct {
	NeedRealName bool   `json:"need_real_name"`
	Account      string `json:"account,omitempty"`
	Desc         string `json:"desc,omitempty"`
}

func (s *Service) CheckRealNameStatus(licenseKey, domain, serverIP string) (*RealNameResult, error) {
	if !s.cfg.Enabled {
		return &RealNameResult{NeedRealName: false}, nil
	}

	if domain == "" {
		domain = getDefaultDomain()
	}
	if serverIP == "" {
		serverIP = getDefaultServerIP()
	}

	target := licenseKey
	if target == "" {
		target = domain
	}
	if target == "" {
		target = serverIP
	}

	timestamp := time.Now().Unix()
	sign := md5Text(s.cfg.AppKey + target + strconv.FormatInt(timestamp, 10) + s.cfg.AppSecret)

	reqBody := RealNameStatusRequest{
		AppKey:     s.cfg.AppKey,
		LicenseKey: licenseKey,
		Domain:     domain,
		ServerIP:   serverIP,
		Timestamp:  timestamp,
		Sign:       sign,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := strings.TrimRight(s.cfg.BaseURL, "/") + "/api/license/realname-status"
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return &RealNameResult{NeedRealName: false}, nil
	}
	defer resp.Body.Close()

	var result RealNameStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &RealNameResult{NeedRealName: false}, nil
	}

	if result.Code == 200 && result.Data.NeedRealName {
		return &RealNameResult{
			NeedRealName: true,
			Account:      result.Data.Account,
			Desc:         result.Data.RealNameDesc,
		}, nil
	}

	return &RealNameResult{NeedRealName: false}, nil
}

func (s *Service) QuickVerify(licenseKey, domain, serverIP string) (bool, *RealNameResult, error) {
	if domain == "" {
		domain = getDefaultDomain()
	}
	if serverIP == "" {
		serverIP = getDefaultServerIP()
	}

	target := licenseKey
	if target == "" {
		target = domain
	}
	if target == "" {
		target = serverIP
	}

	if target == "" {
		return false, &RealNameResult{NeedRealName: false}, nil
	}

	timestamp := time.Now().Unix()
	sign := md5Text(s.cfg.AppKey + target + strconv.FormatInt(timestamp, 10) + s.cfg.AppSecret)

	reqBody := VerifyRequest{
		AppKey:     s.cfg.AppKey,
		Domain:     domain,
		ServerIP:   serverIP,
		LicenseKey: licenseKey,
		Timestamp:  timestamp,
		Sign:       sign,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return false, &RealNameResult{NeedRealName: false}, err
	}

	url := strings.TrimRight(s.cfg.BaseURL, "/") + "/api/license/verify"
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return false, &RealNameResult{NeedRealName: false}, nil
	}
	defer resp.Body.Close()

	var result VerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, &RealNameResult{NeedRealName: false}, nil
	}

	if result.Code == 200 && result.Data.Result == "pass" {
		return true, &RealNameResult{NeedRealName: false}, nil
	}

	realNameResult, _ := s.CheckRealNameStatus(licenseKey, domain, serverIP)
	return false, realNameResult, nil
}
