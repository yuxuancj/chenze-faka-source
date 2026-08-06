package utils

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"chenze-faka/internal/model"
)

func GenerateOrderNo() string {
	now := time.Now()
	return fmt.Sprintf("CZ%s%06d", now.Format("20060102150405"), now.Nanosecond()/1000)
}

func GenerateSalt() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func HashPassword(password, salt string) string {
	hash := sha256.New()
	hash.Write([]byte(password + salt))
	return hex.EncodeToString(hash.Sum(nil))
}

func MD5Sum(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

func BuildPayURL(orderNo string, amount float64, payMethod string, productName string, payConfig model.PayConfig) string {
	if payConfig.URL == "" || payConfig.Merchant == "" || payConfig.Key == "" {
		return ""
	}

	params := map[string]string{
		"pid":          payConfig.Merchant,
		"type":         payMethod,
		"out_trade_no": orderNo,
		"notify_url":   "/api/orders/notify",
		"return_url":   "/api/orders/return",
		"name":         productName,
		"money":        fmt.Sprintf("%.2f", amount),
	}

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}

	signStr := strings.Join(parts, "&") + "&key=" + payConfig.Key
	sign := MD5Sum(signStr)
	params["sign"] = sign
	params["sign_type"] = "MD5"

	var queryParts []string
	for k, v := range params {
		queryParts = append(queryParts, k+"="+v)
	}

	return payConfig.URL + "/submit.php?" + strings.Join(queryParts, "&")
}

func VerifyPaySign(params map[string]string, payKey string) bool {
	if payKey == "" {
		return false
	}

	sign := params["sign"]
	signType := params["sign_type"]

	cleaned := make(map[string]string)
	for k, v := range params {
		if k == "sign" || k == "sign_type" {
			continue
		}
		cleaned[k] = v
	}

	keys := make([]string, 0, len(cleaned))
	for k := range cleaned {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		if cleaned[k] != "" {
			parts = append(parts, k+"="+cleaned[k])
		}
	}

	signStr := strings.Join(parts, "&") + "&key=" + payKey
	calcSign := MD5Sum(signStr)

	if signType == "MD5" {
		return sign == calcSign
	}
	return false
}