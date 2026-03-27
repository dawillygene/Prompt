package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dawillygene/my-prompt-repository/internal/config"
)

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func New(cfg config.Config) *Client {
	return &Client{
		baseURL: strings.TrimRight(cfg.APIBase, "/"),
		token:   cfg.Token,
		http:    &http.Client{},
	}
}

func (c *Client) WithToken(token string) *Client {
	return &Client{
		baseURL: c.baseURL,
		token:   token,
		http:    c.http,
	}
}

// SetToken updates the token on the existing client
func (c *Client) SetToken(token string) {
	c.token = token
}

func (c *Client) Request(method, path string, payload any, authenticated bool) (map[string]any, error) {
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(raw)
	}

	req, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	if authenticated {
		if c.token == "" {
			return nil, fmt.Errorf("not logged in. Use /login or /register")
		}
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		// Hide backend server details, provide user-friendly message
		if strings.Contains(err.Error(), "connection refused") {
			return nil, fmt.Errorf("backend service is currently unavailable")
		}
		if strings.Contains(err.Error(), "dial tcp") {
			return nil, fmt.Errorf("cannot connect to backend service")
		}
		if strings.Contains(err.Error(), "timeout") {
			return nil, fmt.Errorf("backend service is not responding (timeout)")
		}
		return nil, fmt.Errorf("network error: please check your connection")
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var decoded map[string]any
	if len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, &decoded); err != nil {
			return nil, fmt.Errorf("unexpected response: %s", string(responseBody))
		}
	}

	if resp.StatusCode >= 400 {
		if message, ok := decoded["message"].(string); ok && message != "" {
			return nil, fmt.Errorf(message)
		}
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	return decoded, nil
}
