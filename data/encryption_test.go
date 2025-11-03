package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptDecrypt(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef") // 32 bytes for AES-256
	plaintext := "hello world"

	encrypted, err := Encrypt(plaintext, key)
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := Decrypt(encrypted, key)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptDecryptFields(t *testing.T) {
	type PII struct {
		Name    string `pii:"true"`
		Address string `pii:"true"`
		Other   string
	}

	key := []byte("0123456789abcdef0123456789abcdef") // 32 bytes for AES-256

	data := &PII{
		Name:    "John Doe",
		Address: "123 Main St",
		Other:   "some other data",
	}

	err := EncryptFields(data, key)
	assert.NoError(t, err)

	assert.NotEqual(t, "John Doe", data.Name)
	assert.NotEqual(t, "123 Main St", data.Address)
	assert.Equal(t, "some other data", data.Other)

	err = DecryptFields(data, key)
	assert.NoError(t, err)

	assert.Equal(t, "John Doe", data.Name)
	assert.Equal(t, "123 Main St", data.Address)
	assert.Equal(t, "some other data", data.Other)
}
