package data

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"
)

func Encrypt(stringToEncrypt string, key []byte) (string, error) {
	if stringToEncrypt == "" {
		return "", nil
	}

	c, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	return hex.EncodeToString(gcm.Seal(nonce, nonce, []byte(stringToEncrypt), nil)), nil
}

func Decrypt(encryptedString string, key []byte) (string, error) {
	if encryptedString == "" {
		return "", nil
	}

	encrypted, err := hex.DecodeString(encryptedString)
	if err != nil {
		return "", err
	}

	c, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(encrypted) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// LookupHash returns a deterministic HMAC-SHA256 hex string for the
// provided value using the given key. Normalize (trim+lower) before HMAC
// to ensure stable, case-insensitive lookups for usernames.
// This is one-way (non-decryptable) and intended for equality lookups.
func LookupHash(value string, key []byte) (string, error) {
	if value == "" {
		return "", nil
	}

	normalized := strings.ToLower(strings.TrimSpace(value))

	mac := hmac.New(sha256.New, key)
	if _, err := mac.Write([]byte(normalized)); err != nil {
		return "", err
	}

	return hex.EncodeToString(mac.Sum(nil)), nil
}

// LookupFields walks a struct or slice of structs and replaces string fields
// tagged with `pii:"lookup"` with their deterministic lookup hash using
// the provided key. Use this when you want to store a stable, indexable
// token instead of the original plaintext (note: original value cannot be
// recovered from the HMAC).
func LookupFields(s interface{}, key []byte) error {
	val := reflect.ValueOf(s)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() == reflect.Slice {
		for i := 0; i < val.Len(); i++ {
			if err := LookupFields(val.Index(i).Addr().Interface(), key); err != nil {
				return err
			}
		}
		return nil
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		typeField := val.Type().Field(i)

		if tag := typeField.Tag.Get("pii"); tag == "lookup" {
			if field.Kind() == reflect.String && field.CanSet() {
				processed, err := LookupHash(field.String(), key)
				if err != nil {
					return err
				}
				field.SetString(processed)
			}
		}
	}

	return nil
}

func processStruct(s interface{}, process func(string, []byte) (string, error), key []byte) error {
	val := reflect.ValueOf(s)

	// If it's a pointer, get the element it points to.
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// If it's a slice, iterate over it.
	if val.Kind() == reflect.Slice {
		for i := 0; i < val.Len(); i++ {
			if err := processStruct(val.Index(i).Addr().Interface(), process, key); err != nil {
				return err
			}
		}
		return nil
	}

	if val.Kind() != reflect.Struct {
		return nil // Not a struct, nothing to do.
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		typeField := val.Type().Field(i)

		if tag := typeField.Tag.Get("pii"); tag == "true" {
			if field.Kind() == reflect.String {
				if field.CanSet() {
					processed, err := process(field.String(), key)
					if err != nil {
						return err
					}
					field.SetString(processed)
				}
			} else if field.Type() == reflect.TypeOf(time.Time{}) {
				if field.CanSet() {
					date := field.Interface().(time.Time)
					if date.IsZero() {
						continue
					}
					processed, err := process(date.Format(time.RFC3339Nano), key)
					if err != nil {
						return err
					}
					parsedTime, err := time.Parse(time.RFC3339Nano, processed)
					if err != nil {
						return err
					}
					field.Set(reflect.ValueOf(parsedTime))
				}
			}
		}
	}

	return nil
}

func EncryptFields(s interface{}, key []byte) error {
	return processStruct(s, Encrypt, key)
}

func DecryptFields(s interface{}, key []byte) error {
	return processStruct(s, Decrypt, key)
}
