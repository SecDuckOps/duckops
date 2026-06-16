package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type APIClient interface {
	Register(ctx context.Context, req any, dst any) error
	Connect(ctx context.Context, req any, dst any) error
	Heartbeat(ctx context.Context, req any, dst any) error
	GetAgent(ctx context.Context, agentID string, dst any) error
	SetToken(token string)
	SetBaseURL(url string)
}

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client

	maxRetries int
	baseBackoff time.Duration
	maxBackoff  time.Duration
}

type Option func(*Client)

func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

func WithMaxRetries(n int) Option {
	return func(c *Client) {
		c.maxRetries = n
	}
}

func New(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 30 * time.Second},
		maxRetries: 5,
		baseBackoff: 500 * time.Millisecond,
		maxBackoff:  30 * time.Second,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) SetToken(token string) {
	c.token = token
}

func (c *Client) SetBaseURL(url string) {
	c.baseURL = strings.TrimRight(url, "/")
}

func (c *Client) Register(ctx context.Context, req any, dst any) error {
	return c.do(ctx, http.MethodPost, "/api/v1/agents/register", req, dst)
}

func (c *Client) Connect(ctx context.Context, req any, dst any) error {
	return c.do(ctx, http.MethodPost, "/api/v1/agents/connect", req, dst)
}

func (c *Client) Heartbeat(ctx context.Context, req any, dst any) error {
	return c.do(ctx, http.MethodPost, "/api/v1/agents/heartbeat", req, dst)
}

func (c *Client) GetAgent(ctx context.Context, agentID string, dst any) error {
	return c.do(ctx, http.MethodGet, "/api/v1/agents/"+agentID, nil, dst)
}

func (c *Client) do(ctx context.Context, method, path string, body, dst any) error {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	return c.retryDo(ctx, method, url, bodyReader, dst)
}

func (c *Client) retryDo(ctx context.Context, method, url string, body io.Reader, dst any) error {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := c.computeBackoff(attempt)
			slog.Debug("retrying request",
				"method", method,
				"url", url,
				"attempt", attempt,
				"max_retries", c.maxRetries,
				"backoff", backoff,
			)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		respBody, err := c.doOnce(ctx, method, url, body)
		if err != nil {
			lastErr = err
			if isRetryable(err) {
				continue
			}
			return err
		}

		if dst != nil {
			if err := json.Unmarshal(respBody, dst); err != nil {
				return fmt.Errorf("parse response: %w", err)
			}
		}

		return nil
	}

	return fmt.Errorf("max retries exceeded (%d): %w", c.maxRetries, lastErr)
}

func (c *Client) doOnce(ctx context.Context, method, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "duckops-agent/1.0")

	slog.Debug("sending request", "method", method, "url", maskURL(url))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Warn("request failed", "error", err)
		return nil, classifyNetworkError(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return respBody, nil
	}

	return nil, classifyHTTPError(resp.StatusCode, respBody)
}

func (c *Client) computeBackoff(attempt int) time.Duration {
	exp := math.Pow(2, float64(attempt))
	backoff := float64(c.baseBackoff) * exp
	if backoff > float64(c.maxBackoff) {
		backoff = float64(c.maxBackoff)
	}
	jitter := (rand.Float64()*2 - 1) * 0.2 * backoff
	return time.Duration(backoff + jitter)
}

func maskURL(url string) string {
	if idx := strings.Index(url, "token="); idx != -1 {
		return url[:idx+6] + "***"
	}
	return url
}

func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	var ce *ClientError
	if errorsAs(err, &ce) {
		switch ce.Code {
		case ErrRateLimited, ErrServerError, ErrOffline, ErrTimeout:
			return true
		case ErrAuthFailed, ErrNotFound, ErrBadRequest:
			return false
		}
	}

	return false
}

func errorsAs(err error, target any) bool {
	for {
		if e, ok := err.(interface{ As(any) bool }); ok {
			if e.As(target) {
				return true
			}
		}
		if u, ok := err.(interface{ Unwrap() error }); ok {
			err = u.Unwrap()
		} else {
			return false
		}
	}
}
