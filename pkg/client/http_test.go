package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Do(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo back request details
		response := map[string]interface{}{
			"method":       r.Method,
			"path":         r.URL.Path,
			"query":        r.URL.RawQuery,
			"headers":      r.Header,
			"content_type": r.Header.Get("Content-Type"),
		}

		// Read body if present
		if r.Body != nil {
			body := make([]byte, r.ContentLength)
			r.Body.Read(body)
			if len(body) > 0 {
				response["body"] = string(body)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(30 * time.Second)

	tests := []struct {
		name        string
		method      string
		headers     http.Header
		body        interface{}
		expectError bool
	}{
		{
			name:   "GET request",
			method: "GET",
			headers: http.Header{
				"Authorization": []string{"Bearer token"},
			},
			body:        nil,
			expectError: false,
		},
		{
			name:   "POST with JSON body",
			method: "POST",
			headers: http.Header{
				"Authorization": []string{"Bearer token"},
			},
			body: map[string]interface{}{
				"key": "value",
				"num": 123,
			},
			expectError: false,
		},
		{
			name:        "PUT with string body",
			method:      "PUT",
			body:        "string body",
			expectError: false,
		},
		{
			name:        "POST with byte body",
			method:      "POST",
			body:        []byte(`{"test": "data"}`),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			resp, err := client.Do(ctx, tt.method, server.URL+"/test", tt.headers, tt.body)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				assert.True(t, resp.IsJSONResponse())

				// Parse response body
				var responseData map[string]interface{}
				err = json.Unmarshal(resp.Body, &responseData)
				require.NoError(t, err)

				assert.Equal(t, tt.method, responseData["method"])
				assert.Equal(t, "/test", responseData["path"])

				// Check headers were forwarded
				if tt.headers != nil {
					headers := responseData["headers"].(map[string]interface{})
					for key := range tt.headers {
						assert.Contains(t, headers, key)
					}
				}

				// Check content type was set for body requests
				if tt.body != nil {
					assert.Equal(t, "application/json", responseData["content_type"])
				}
			}
		})
	}
}

func TestClient_ForwardRequest(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"method":  r.Method,
			"path":    r.URL.Path,
			"query":   r.URL.RawQuery,
			"headers": r.Header,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(30 * time.Second)

	tests := []struct {
		name          string
		method        string
		path          string
		queryParams   url.Values
		headers       http.Header
		expectedPath  string
		expectedQuery string
	}{
		{
			name:         "simple path",
			method:       "GET",
			path:         "/api/users",
			expectedPath: "/api/users",
		},
		{
			name:   "path with query params",
			method: "GET",
			path:   "/api/users",
			queryParams: url.Values{
				"limit":  []string{"10"},
				"offset": []string{"20"},
			},
			expectedPath:  "/api/users",
			expectedQuery: "limit=10&offset=20",
		},
		{
			name:   "headers with jpx- prefix filtering",
			method: "GET",
			path:   "/api/test",
			headers: http.Header{
				"Authorization": []string{"Bearer token"},
				"jpx-debug":     []string{"true"},
				"JPX-Internal":  []string{"value"},
				"Content-Type":  []string{"application/json"},
			},
			expectedPath: "/api/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			resp, err := client.ForwardRequest(ctx, tt.method, server.URL, tt.path, tt.queryParams, tt.headers, nil)

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			// Parse response
			var responseData map[string]interface{}
			err = json.Unmarshal(resp.Body, &responseData)
			require.NoError(t, err)

			assert.Equal(t, tt.method, responseData["method"])
			assert.Equal(t, tt.expectedPath, responseData["path"])

			if tt.expectedQuery != "" {
				assert.Equal(t, tt.expectedQuery, responseData["query"])
			}

			// Check jpx- headers were filtered
			if tt.headers != nil {
				headers := responseData["headers"].(map[string]interface{})

				// Should have Authorization and Content-Type
				assert.Contains(t, headers, "Authorization")
				if tt.headers.Get("Content-Type") != "" {
					assert.Contains(t, headers, "Content-Type")
				}

				// Should NOT have jpx- prefixed headers
				assert.NotContains(t, headers, "jpx-debug")
				assert.NotContains(t, headers, "JPX-Internal")
			}
		})
	}
}

func TestBuildTargetURL(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		path        string
		queryParams url.Values
		expected    string
		expectError bool
	}{
		{
			name:     "simple URL with path",
			baseURL:  "https://api.example.com",
			path:     "/users",
			expected: "https://api.example.com/users",
		},
		{
			name:     "URL with trailing slash and path with leading slash",
			baseURL:  "https://api.example.com/",
			path:     "/users",
			expected: "https://api.example.com/users",
		},
		{
			name:     "URL without trailing slash and path without leading slash",
			baseURL:  "https://api.example.com",
			path:     "users",
			expected: "https://api.example.com/users",
		},
		{
			name:     "URL with base path",
			baseURL:  "https://api.example.com/v1",
			path:     "/users",
			expected: "https://api.example.com/v1/users",
		},
		{
			name:    "URL with query parameters",
			baseURL: "https://api.example.com",
			path:    "/users",
			queryParams: url.Values{
				"limit":  []string{"10"},
				"filter": []string{"active"},
			},
			expected: "https://api.example.com/users?filter=active&limit=10",
		},
		{
			name:        "invalid base URL",
			baseURL:     "://invalid-url",
			path:        "/users",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildTargetURL(tt.baseURL, tt.path, tt.queryParams)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFilterHeaders(t *testing.T) {
	tests := []struct {
		name     string
		headers  http.Header
		expected http.Header
	}{
		{
			name:     "nil headers",
			headers:  nil,
			expected: nil,
		},
		{
			name: "no jpx headers",
			headers: http.Header{
				"Authorization": []string{"Bearer token"},
				"Content-Type":  []string{"application/json"},
			},
			expected: http.Header{
				"Authorization": []string{"Bearer token"},
				"Content-Type":  []string{"application/json"},
			},
		},
		{
			name: "mixed headers with jpx prefix",
			headers: http.Header{
				"Authorization": []string{"Bearer token"},
				"jpx-debug":     []string{"true"},
				"JPX-Internal":  []string{"value"},
				"Content-Type":  []string{"application/json"},
				"jpx-trace-id":  []string{"123"},
			},
			expected: http.Header{
				"Authorization": []string{"Bearer token"},
				"Content-Type":  []string{"application/json"},
			},
		},
		{
			name: "only jpx headers",
			headers: http.Header{
				"jpx-debug":    []string{"true"},
				"JPX-Internal": []string{"value"},
			},
			expected: http.Header{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterHeaders(tt.headers)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResponse_IsJSONResponse(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		expected    bool
	}{
		{
			name:        "application/json",
			contentType: "application/json",
			expected:    true,
		},
		{
			name:        "application/json with charset",
			contentType: "application/json; charset=utf-8",
			expected:    true,
		},
		{
			name:        "text/plain",
			contentType: "text/plain",
			expected:    false,
		},
		{
			name:        "empty content type",
			contentType: "",
			expected:    false,
		},
		{
			name:        "case insensitive",
			contentType: "APPLICATION/JSON",
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &Response{
				Headers: http.Header{
					"Content-Type": []string{tt.contentType},
				},
			}
			assert.Equal(t, tt.expected, resp.IsJSONResponse())
		})
	}
}

func TestResponse_ParseJSONBody(t *testing.T) {
	tests := []struct {
		name        string
		body        []byte
		expected    interface{}
		expectError bool
	}{
		{
			name:     "valid JSON object",
			body:     []byte(`{"key": "value", "number": 123}`),
			expected: map[string]interface{}{"key": "value", "number": float64(123)},
		},
		{
			name:     "valid JSON array",
			body:     []byte(`[1, 2, 3]`),
			expected: []interface{}{float64(1), float64(2), float64(3)},
		},
		{
			name:     "empty body",
			body:     []byte{},
			expected: nil,
		},
		{
			name:        "invalid JSON",
			body:        []byte(`{"key": "value",}`),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &Response{Body: tt.body}
			result, err := resp.ParseJSONBody()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
