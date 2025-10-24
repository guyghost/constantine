package dydx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/guyghost/constantine/internal/ratelimit"
)

const (
	// dYdX rate limits (conservative estimates)
	// Public endpoints: ~175 requests per 10 seconds = 17.5 req/s
	// Private endpoints: ~175 requests per 10 seconds = 17.5 req/s
	dydxRateLimit = 15.0 // requests per second (conservative)
)

// HTTPClient handles REST API requests to dYdX
type HTTPClient struct {
	baseURL     string
	apiKey      string
	apiSecret   string
	httpClient  *http.Client
	rateLimiter ratelimit.Limiter
}

// NewHTTPClient creates a new HTTP client for dYdX
func NewHTTPClient(baseURL, apiKey, apiSecret string) *HTTPClient {
	// Create rate limiter with burst capability
	limiter := ratelimit.NewTokenBucket(dydxRateLimit, int(dydxRateLimit*2))

	return &HTTPClient{
		baseURL:     baseURL,
		apiKey:      apiKey,
		apiSecret:   apiSecret,
		rateLimiter: limiter,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an HTTP request
func (c *HTTPClient) doRequest(ctx context.Context, method, path string, body any, result any) error {
	// Apply rate limiting before making the request
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit wait failed: %w", err)
	}

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Add API key if provided
	if c.apiKey != "" {
		req.Header.Set("X-API-KEY", c.apiKey)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	// Parse response
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// get performs a GET request
func (c *HTTPClient) get(ctx context.Context, path string, result any) error {
	return c.doRequest(ctx, http.MethodGet, path, nil, result)
}

// post performs a POST request
func (c *HTTPClient) post(ctx context.Context, path string, body any, result any) error {
	return c.doRequest(ctx, http.MethodPost, path, body, result)
}

// delete performs a DELETE request
func (c *HTTPClient) delete(ctx context.Context, path string, result any) error {
	return c.doRequest(ctx, http.MethodDelete, path, nil, result)
}
