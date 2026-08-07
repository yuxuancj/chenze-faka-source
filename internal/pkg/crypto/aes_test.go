package crypto

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAesEncryptDecryptRoundtrip(t *testing.T) {
	key := "this-is-a-32-byte-key!!!!!!"
	plaintext := "Hello, World! This is a test message."

	encrypted, err := AesEncrypt(plaintext, key)
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := AesDecrypt(encrypted, key)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestAesEncryptKeyShorterThan32(t *testing.T) {
	key := "short-key"
	plaintext := "test message with short key"

	encrypted, err := AesEncrypt(plaintext, key)
	assert.NoError(t, err)

	decrypted, err := AesDecrypt(encrypted, key)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestAesEncryptKeyLongerThan32(t *testing.T) {
	key := "this-is-a-very-long-key-that-exceeds-32-bytes!!!"
	plaintext := "test message with long key"

	encrypted, err := AesEncrypt(plaintext, key)
	assert.NoError(t, err)

	decrypted, err := AesDecrypt(encrypted, key)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestAesEncryptEmptyPlaintext(t *testing.T) {
	key := "this-is-a-32-byte-key!!!!!!"

	encrypted, err := AesEncrypt("", key)
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := AesDecrypt(encrypted, key)
	assert.NoError(t, err)
	assert.Equal(t, "", decrypted)
}

func TestAesDecryptWrongKey(t *testing.T) {
	key1 := "this-is-a-32-byte-key!!!!!!"
	key2 := "this-is-a-different-32byte!"
	plaintext := "secret message"

	encrypted, err := AesEncrypt(plaintext, key1)
	assert.NoError(t, err)

	_, err = AesDecrypt(encrypted, key2)
	assert.Error(t, err)
}

func TestAesDecryptInvalidBase64(t *testing.T) {
	key := "this-is-a-32-byte-key!!!!!!"

	_, err := AesDecrypt("!!!not-valid-base64!!!", key)
	assert.Error(t, err)
}

func TestAesDecryptShortCiphertext(t *testing.T) {
	key := "this-is-a-32-byte-key!!!!!!"

	_, err := AesDecrypt("YQ==", key)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "密文长度不足")
}

func TestAesEncryptVeryLongPlaintext(t *testing.T) {
	key := "this-is-a-32-byte-key!!!!!!"
	plaintext := strings.Repeat("A", 10000)

	encrypted, err := AesEncrypt(plaintext, key)
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := AesDecrypt(encrypted, key)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}