package utils

import (
	"net/url"
	"regexp"
	"testing"

	"chenze-faka/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestGenerateOrderNo(t *testing.T) {
	orderNo := GenerateOrderNo()
	assert.NotEmpty(t, orderNo)
	assert.True(t, len(orderNo) > 15)

	assert.Equal(t, "CZ", orderNo[:2])

	matched, _ := regexp.MatchString(`^CZ\d{20}$`, orderNo)
	assert.True(t, matched)
}

func TestGenerateSalt(t *testing.T) {
	salt := GenerateSalt()
	assert.Equal(t, 32, len(salt))

	matched, _ := regexp.MatchString(`^[0-9a-f]{32}$`, salt)
	assert.True(t, matched)

	salt2 := GenerateSalt()
	assert.NotEqual(t, salt, salt2)
}

func TestHashPassword(t *testing.T) {
	hash1 := HashPassword("password123", "salt1")
	hash2 := HashPassword("password123", "salt2")
	hash3 := HashPassword("password123", "salt1")

	assert.NotEqual(t, hash1, hash2)
	assert.Equal(t, hash1, hash3)
	assert.Equal(t, 64, len(hash1))
}

func TestMD5Sum(t *testing.T) {
	result := MD5Sum("hello")
	assert.Equal(t, "5d41402abc4b2a76b9719d911017c592", result)
	assert.Equal(t, 32, len(result))

	resultEmpty := MD5Sum("")
	assert.Equal(t, "d41d8cd98f00b204e9800998ecf8427e", resultEmpty)
}

func TestBuildPayURL(t *testing.T) {
	config := model.PayConfig{
		URL:      "https://example.com",
		Merchant: "1001",
		Key:      "secretkey",
	}

	payURL := BuildPayURL("ORDER123", 99.99, "alipay", "Test Product", config)
	assert.NotEmpty(t, payURL)
	assert.Contains(t, payURL, "https://example.com/submit.php?")
	assert.Contains(t, payURL, "out_trade_no=ORDER123")
	assert.Contains(t, payURL, "money=99.99")
	assert.Contains(t, payURL, "pid=1001")
	assert.Contains(t, payURL, "name=Test Product")
	assert.Contains(t, payURL, "sign=")

	emptyConfig := model.PayConfig{}
	url2 := BuildPayURL("ORDER123", 99.99, "alipay", "Test Product", emptyConfig)
	assert.Empty(t, url2)
}

func TestVerifyPaySign(t *testing.T) {
	config := model.PayConfig{
		URL:      "https://example.com",
		Merchant: "1001",
		Key:      "secretkey",
	}

	payURL := BuildPayURL("ORDER123", 99.99, "alipay", "Test Product", config)
	assert.NotEmpty(t, payURL)

	parsedURL, err := url.Parse(payURL)
	assert.NoError(t, err)

	query := parsedURL.Query()
	params := make(map[string]string)
	for k, v := range query {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	assert.True(t, VerifyPaySign(params, "secretkey"))
	assert.False(t, VerifyPaySign(params, "wrongkey"))
	assert.False(t, VerifyPaySign(params, ""))

	delete(params, "sign")
	assert.False(t, VerifyPaySign(params, "secretkey"))
}