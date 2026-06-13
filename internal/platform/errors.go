package platform

import (
	"fmt"
	"net/http"
	"time"
)

type ErrorCode string

const (
	ErrAuthFailed     ErrorCode = "AUTH_FAILED"
	ErrNotFound       ErrorCode = "NOT_FOUND"
	ErrRateLimited    ErrorCode = "RATE_LIMITED"
	ErrTimeout        ErrorCode = "TIMEOUT"
	ErrOffline        ErrorCode = "OFFLINE"
	ErrBadRequest     ErrorCode = "BAD_REQUEST"
	ErrServerError    ErrorCode = "SERVER_ERROR"
	ErrQueueFull      ErrorCode = "QUEUE_FULL"
	ErrInvalidPayload ErrorCode = "INVALID_PAYLOAD"
	ErrMaxRetries     ErrorCode = "MAX_RETRIES"
)

type PlatformError struct {
	Code       ErrorCode
	Message    string
	StatusCode int
	RetryAfter time.Duration
	Wrapped    error
}

func (e *PlatformError) Error() string {
	if e.Wrapped != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Wrapped)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *PlatformError) Unwrap() error {
	return e.Wrapped
}

func classifyError(err error, statusCode int) *PlatformError {
	if err == nil {
		return nil
	}

	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return &PlatformError{
			Code: ErrAuthFailed, Message: "authentication failed — check your PAT",
			StatusCode: statusCode, Wrapped: err,
		}
	case http.StatusNotFound:
		return &PlatformError{
			Code: ErrNotFound, Message: "resource not found",
			StatusCode: statusCode, Wrapped: err,
		}
	case http.StatusTooManyRequests:
		return &PlatformError{
			Code: ErrRateLimited, Message: "rate limited by server",
			StatusCode: statusCode, RetryAfter: 5 * time.Second, Wrapped: err,
		}
	case http.StatusBadRequest:
		return &PlatformError{
			Code: ErrBadRequest, Message: "bad request",
			StatusCode: statusCode, Wrapped: err,
		}
	}

	if statusCode >= 500 {
		return &PlatformError{
			Code: ErrServerError, Message: fmt.Sprintf("server error (HTTP %d)", statusCode),
			StatusCode: statusCode, Wrapped: err,
		}
	}

	return &PlatformError{
		Code: ErrServerError, Message: fmt.Sprintf("unexpected HTTP %d", statusCode),
		StatusCode: statusCode, Wrapped: err,
	}
}

func IsAuthError(err error) bool {
	var pe *PlatformError
	return err != nil && asPlatformError(err, &pe) && pe.Code == ErrAuthFailed
}

func IsRateLimited(err error) bool {
	var pe *PlatformError
	return err != nil && asPlatformError(err, &pe) && pe.Code == ErrRateLimited
}

func IsOffline(err error) bool {
	var pe *PlatformError
	return err != nil && asPlatformError(err, &pe) && pe.Code == ErrOffline
}

func asPlatformError(err error, target **PlatformError) bool {
	if err == nil {
		return false
	}
	pe, ok := err.(*PlatformError)
	if !ok {
		return false
	}
	*target = pe
	return true
}
