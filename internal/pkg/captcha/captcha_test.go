package captcha

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	id, imgBase64 := Generate()
	assert.NotEmpty(t, id)
	assert.NotEmpty(t, imgBase64)
	assert.Greater(t, len(id), 0)
	assert.Greater(t, len(imgBase64), 0)
}

func TestVerifyCorrect(t *testing.T) {
	id, _ := Generate()

	val, ok := store.Load(id)
	assert.True(t, ok)
	entry := val.(captchaEntry)

	result := Verify(id, entry.answer)
	assert.True(t, result)
}

func TestVerifyWrong(t *testing.T) {
	id, _ := Generate()

	result := Verify(id, "wrong-answer")
	assert.False(t, result)
}

func TestVerifyExpired(t *testing.T) {
	id, _ := Generate()

	val, ok := store.Load(id)
	assert.True(t, ok)
	entry := val.(captchaEntry)

	store.Store(id, captchaEntry{
		answer: entry.answer,
		expire: time.Now().Add(-1 * time.Minute),
	})

	result := Verify(id, entry.answer)
	assert.False(t, result)

	store.Delete(id)
}

func TestVerifyNonexistent(t *testing.T) {
	result := Verify("nonexistent-id-12345", "answer")
	assert.False(t, result)
}

func TestDelete(t *testing.T) {
	id, _ := Generate()

	val, ok := store.Load(id)
	assert.True(t, ok)
	assert.NotNil(t, val)

	Delete(id)

	_, ok = store.Load(id)
	assert.False(t, ok)
}

func TestVerifyAfterDelete(t *testing.T) {
	id, _ := Generate()

	Delete(id)

	_, ok := store.Load(id)
	assert.False(t, ok)

	result := Verify(id, "any-answer")
	assert.False(t, result)
}

func TestMultipleGenerations(t *testing.T) {
	for i := 0; i < 10; i++ {
		id, imgBase64 := Generate()
		assert.NotEmpty(t, id)
		assert.NotEmpty(t, imgBase64)

		val, ok := store.Load(id)
		assert.True(t, ok)
		entry := val.(captchaEntry)

		result := Verify(id, entry.answer)
		assert.True(t, result)
	}
}

func TestVerifyConsumesCaptcha(t *testing.T) {
	id, _ := Generate()

	val, ok := store.Load(id)
	assert.True(t, ok)
	entry := val.(captchaEntry)

	result := Verify(id, entry.answer)
	assert.True(t, result)

	_, ok = store.Load(id)
	assert.False(t, ok)

	result2 := Verify(id, entry.answer)
	assert.False(t, result2)
}
