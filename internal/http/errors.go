package http

import (
	"fmt"
)

// EncryptBodyError представляет ошибку шифрования тела запроса.
type EncryptBodyError struct {
	PayloadSize int
	Err         error
}

func (e *EncryptBodyError) Error() string {
	return fmt.Sprintf("encrypt body (payload %d bytes): %v", e.PayloadSize, e.Err)
}

func (e *EncryptBodyError) Unwrap() error {
	return e.Err
}
