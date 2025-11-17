// Package client provides HTTP client functionality for making requests to target endpoints.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTPClient defines the interface for HTTP client operations
type HTTPClient interface {
	Do(ctx context.Context, method, targetURL string, headers http.Header, body interface{}) (*Response, error)
	ForwardRequest(
		ctx context.Context,
		method, baseURL, path string,
		queryParams url.Values,
		headers http.Header,
		body interface{},
	) (*Response, error)
}

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// Client implements HTTPClient with connection pooling and timeout management
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new HTTP client with connection pooling
func NewClient(timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// Do performs an HTTP request with the specified parameters
func (c *Client) Do(
	ctx context.Context,
	method, targetURL string,
	headers http.Header,
	body interface{},
) (*Response, error) {
	// Prepare request body
	var reqBody io.Reader
	if body != nil {
		switch v := body.(type) {
		case []byte:
			reqBody = bytes.NewReader(v)
		case string:
			reqBody = strings.NewReader(v)
		default:
			// JSON encode the body
			jsonData, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			reqBody = bytes.NewReader(jsonData)
		}
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, method, targetURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Copy headers
	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Set Content-Type for JSON body if not already set
	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Perform the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       respBody,
	}, nil
}

// ForwardRequest forwards a request to a target endpoint with path and query parameters
func (c *Client) ForwardRequest(
	ctx context.Context,
	method, baseURL, path string,
	queryParams url.Values,
	headers http.Header,
	body interface{},
) (*Response, error) {
	// Build target URL
	targetURL, err := buildTargetURL(baseURL, path, queryParams)
	if err != nil {
		return nil, fmt.Errorf("failed to build target URL: %w", err)
	}

	// Filter headers (remove jpx- prefixed headers)
	filteredHeaders := filterHeaders(headers)

	return c.Do(ctx, method, targetURL, filteredHeaders, body)
}

// buildTargetURL constructs the complete target URL
func buildTargetURL(baseURL, path string, queryParams url.Values) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	// Join path
	if path != "" {
		// Remove leading slash from path if base URL already has one
		if strings.HasSuffix(base.Path, "/") && strings.HasPrefix(path, "/") {
			path = path[1:]
		} else if !strings.HasSuffix(base.Path, "/") && !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		base.Path += path
	}

	// Add query parameters
	if len(queryParams) > 0 {
		base.RawQuery = queryParams.Encode()
	}

	return base.String(), nil
}

// filterHeaders removes headers with jpx- prefix
func filterHeaders(headers http.Header) http.Header {
	if headers == nil {
		return nil
	}

	filtered := make(http.Header)
	for key, values := range headers {
		// Skip headers with jpx- prefix (case insensitive)
		if !strings.HasPrefix(strings.ToLower(key), "jpx-") {
			filtered[key] = values
		}
	}

	return filtered
}

// IsJSONResponse checks if the response content type is JSON
func (r *Response) IsJSONResponse() bool {
	contentType := r.Headers.Get("Content-Type")
	return strings.Contains(strings.ToLower(contentType), "application/json")
}

// ParseJSONBody parses the response body as JSON
func (r *Response) ParseJSONBody() (interface{}, error) {
	if len(r.Body) == 0 {
		return nil, nil
	}

	var result interface{}
	if err := json.Unmarshal(r.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return result, nil
}
