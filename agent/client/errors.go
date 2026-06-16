package client

import (
	"fmt"
	"net"
	"strings"
)

type ErrorCode int

const (
	ErrUnknown    ErrorCode = 0
	ErrBadRequest ErrorCode = 400
	ErrAuthFailed ErrorCode = 401
	ErrNotFound   ErrorCode = 404
	ErrRateLimited ErrorCode = 429
	ErrServerError ErrorCode = 500
	ErrOffline    ErrorCode = -1
	ErrTimeout    ErrorCode = -2
)

type ClientError struct {
	Code    ErrorCode
	Message string
	Wrapped error
}

func (e *ClientError) Error() string {
	if e.Wrapped != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Wrapped)
	}
	return e.Message
}

func (e *ClientError) Unwrap() error {
	return e.Wrapped
}

func classifyHTTPError(statusCode int, body []byte) *ClientError {
	msg := string(body)
	if msg == "" {
		msg = httpStatusText(statusCode)
	}

	switch {
	case statusCode == 400:
		return &ClientError{Code: ErrBadRequest, Message: msg}
	case statusCode == 401 || statusCode == 403:
		return &ClientError{Code: ErrAuthFailed, Message: msg}
	case statusCode == 404:
		return &ClientError{Code: ErrNotFound, Message: msg}
	case statusCode == 429:
		return &ClientError{Code: ErrRateLimited, Message: msg}
	case statusCode >= 500:
		return &ClientError{Code: ErrServerError, Message: msg}
	default:
		return &ClientError{Code: ErrUnknown, Message: fmt.Sprintf("HTTP %d: %s", statusCode, msg)}
	}
}

func classifyNetworkError(err error) *ClientError {
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return &ClientError{
			Code:    ErrTimeout,
			Message: "request timed out",
			Wrapped: err,
		}
	}

	if strings.Contains(err.Error(), "connection refused") ||
		strings.Contains(err.Error(), "no such host") ||
		strings.Contains(err.Error(), "i/o timeout") {
		return &ClientError{
			Code:    ErrOffline,
			Message: "server unreachable",
			Wrapped: err,
		}
	}

	return &ClientError{
		Code:    ErrUnknown,
		Message: "network error",
		Wrapped: err,
	}
}

func httpStatusText(code int) string {
	switch code {
	case 400:
		return "bad request"
	case 401:
		return "unauthorized"
	case 403:
		return "forbidden"
	case 404:
		return "not found"
	case 429:
		return "too many requests"
	case 500:
		return "internal server error"
	case 502:
		return "bad gateway"
	case 503:
		return "service unavailable"
	default:
		return fmt.Sprintf("HTTP %d", code)
	}
}
