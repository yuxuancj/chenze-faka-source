package crypto

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAesEncryptVeryShortKey(t *testing.T) {
	key := "a"
	plaintext := "test with very short key"

	encrypted, err := AesEncrypt(plaintext, key)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := AesDecrypt(encrypted, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestAesEncryptUnicodeSpecialChars(t *testing.T) {
	key := "this-is-a-32-byte-key!!!!!!"
	plaintexts := []string{
		"Hello, 世界!",
		"Привет, мир!",
		"こんにちは世界",
		"🎉🚀💻测试特殊字符",
		"Line1\nLine2\tTabbed",
		"Quote\"Escaped'Back\\Slash",
		"中文标点：，。！？；：\"\"''（）【】",
	}

	for _, pt := range plaintexts {
		encrypted, err := AesEncrypt(pt, key)
		require.NoError(t, err)
		assert.NotEmpty(t, encrypted)

		decrypted, err := AesDecrypt(encrypted, key)
		require.NoError(t, err)
		assert.Equal(t, pt, decrypted)
	}
}

func TestAesEncryptMultipleCycles(t *testing.T) {
	key := "cycle-test-key-1234567890123456"
	original := "original plaintext for cycle test"

	firstEncrypted, err := AesEncrypt(original, key)
	require.NoError(t, err)

	firstDecrypted, err := AesDecrypt(firstEncrypted, key)
	require.NoError(t, err)
	assert.Equal(t, original, firstDecrypted)

	secondEncrypted, err := AesEncrypt(firstDecrypted, key)
	require.NoError(t, err)

	secondDecrypted, err := AesDecrypt(secondEncrypted, key)
	require.NoError(t, err)
	assert.Equal(t, original, secondDecrypted)

	thirdEncrypted, err := AesEncrypt(secondDecrypted, key)
	require.NoError(t, err)

	thirdDecrypted, err := AesDecrypt(thirdEncrypted, key)
	require.NoError(t, err)
	assert.Equal(t, original, thirdDecrypted)
}

func TestAesEncryptDecryptDeterministicCycle(t *testing.T) {
	key := "deterministic-key-123456789012345"
	plaintext := "deterministic cycle test data"

	for i := 0; i < 10; i++ {
		encrypted, err := AesEncrypt(plaintext, key)
		require.NoError(t, err)

		decrypted, err := AesDecrypt(encrypted, key)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	}
}

func TestAesConcurrent(t *testing.T) {
	key := "concurrent-test-key-12345678901"
	plaintexts := []string{
		"concurrent-message-1",
		"concurrent-message-2",
		"concurrent-message-3",
		"concurrent-message-4",
		"concurrent-message-5",
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			pt := plaintexts[idx%len(plaintexts)]

			encrypted, err := AesEncrypt(pt, key)
			if err != nil {
				errCh <- err
				return
			}

			decrypted, err := AesDecrypt(encrypted, key)
			if err != nil {
				errCh <- err
				return
			}

			if decrypted != pt {
				errCh <- err
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		assert.NoError(t, err)
	}
}

func TestAesConcurrentStress(t *testing.T) {
	key := "stress-concurrent-key-12345678"

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			pt := "stress-test-message-" + string(rune('0'+n%10))

			encrypted, err := AesEncrypt(pt, key)
			assert.NoError(t, err)

			decrypted, err := AesDecrypt(encrypted, key)
			assert.NoError(t, err)
			assert.Equal(t, pt, decrypted)
		}(i)
	}
	wg.Wait()
}

func TestAesEncryptDecryptEmptyUnicode(t *testing.T) {
	key := "unicode-empty-test-key!!!!!!"

	encrypted, err := AesEncrypt("", key)
	require.NoError(t, err)

	decrypted, err := AesDecrypt(encrypted, key)
	require.NoError(t, err)
	assert.Equal(t, "", decrypted)
}

func TestAesDecryptWithModifiedCiphertext(t *testing.T) {
	key := "modified-ciphertext-key!!!!!!"
	plaintext := "original plaintext for modification test"

	encrypted, err := AesEncrypt(plaintext, key)
	require.NoError(t, err)

	decrypted, err := AesDecrypt(encrypted, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)

	cipherBytes := []byte(encrypted)
	if len(cipherBytes) > 0 {
		cipherBytes[len(cipherBytes)-1] ^= 0x01
		modified := string(cipherBytes)

		_, err = AesDecrypt(modified, key)
		assert.Error(t, err)
	}
}
