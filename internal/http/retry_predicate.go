package http

import (
	"context"
	"errors"
	"io"
	"net"
	"net/url"
	"syscall"
)

// NetworkRetryPredicate определяет, можно ли повторить сетевую попытку.
func NetworkRetryPredicate(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	if errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, io.EOF) {
		return true
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) && urlErr.Err != nil {
		return NetworkRetryPredicate(urlErr.Err)
	}

	return false
}
