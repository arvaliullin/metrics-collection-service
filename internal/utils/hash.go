package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
)

// ErrKeyNotSet сообщает о том, что ключ для хэширования не задан.
var ErrKeyNotSet = errors.New("не задан ключ")

// Hash выполняет расчет хэш
func Hash(src []byte, key string) ([]byte, error) {
	var dst []byte

	if key == "" {
		return dst, ErrKeyNotSet
	}

	h := hmac.New(sha256.New, []byte(key))

	if _, err := h.Write(src); err != nil {
		return dst, err
	}

	dst = h.Sum(nil)

	return dst, nil
}
