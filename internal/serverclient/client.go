package serverclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
	Retries    int
}

type APIResponse struct {
	Status  string          `json:"status"`
	Data    json.RawMessage `json:"data"`
	Results int             `json:"results,omitempty"`
	Message string          `json:"message,omitempty"`
	Token   string          `json:"token,omitempty"`
}

func New(baseURL, token string) *Client {
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Retries: 3,
	}
}

func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	var buf io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		buf = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, buf)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (c *Client) do(req *http.Request) ([]byte, error) {
	var lastErr error
	for i := 0; i <= c.Retries; i++ {
		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			time.Sleep(time.Duration(i+1) * 500 * time.Millisecond)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read response: %w", err)
		}

		if resp.StatusCode >= 500 && i < c.Retries {
			lastErr = fmt.Errorf("server error %d: %s", resp.StatusCode, string(body))
			time.Sleep(time.Duration(i+1) * 500 * time.Millisecond)
			continue
		}

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
		}

		return body, nil
	}
	return nil, lastErr
}

func (c *Client) doRequest(method, path string, body, out interface{}) error {
	req, err := c.newRequest(method, path, body)
	if err != nil {
		return err
	}

	respBody, err := c.do(req)
	if err != nil {
		return err
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}

	if out != nil && apiResp.Data != nil {
		if err := json.Unmarshal(apiResp.Data, out); err != nil {
			return fmt.Errorf("unmarshal data: %w", err)
		}
	}

	return nil
}

func (c *Client) Get(path string, out interface{}) error {
	return c.doRequest("GET", path, nil, out)
}

func (c *Client) Post(path string, body, out interface{}) error {
	return c.doRequest("POST", path, body, out)
}

func (c *Client) Patch(path string, body, out interface{}) error {
	return c.doRequest("PATCH", path, body, out)
}

func (c *Client) Put(path string, body, out interface{}) error {
	return c.doRequest("PUT", path, body, out)
}

func (c *Client) Delete(path string, out interface{}) error {
	return c.doRequest("DELETE", path, nil, out)
}

func (c *Client) GetWithContext(ctx context.Context, path string, out interface{}) error {
	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	respBody, err := c.do(req)
	if err != nil {
		return err
	}
	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return fmt.Errorf("parse response: %w", err)
	}
	if out != nil && apiResp.Data != nil {
		return json.Unmarshal(apiResp.Data, out)
	}
	return nil
}
