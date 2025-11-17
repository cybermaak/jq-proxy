package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"jq-proxy-service/internal/client"
	"jq-proxy-service/internal/logging"
	"jq-proxy-service/internal/models"
	"jq-proxy-service/internal/proxy"
	"jq-proxy-service/internal/transform"
)

// APITestSuite contains comprehensive API tests that replace the shell-based test-api.sh script
// It can run against either a mock server (default) or a live development server
type APITestSuite struct {
	suite.Suite
	mockServer    *httptest.Server
	proxyHandler  http.Handler
	baseURL       string
	useLiveServer bool
	httpClient    *http.Client
}

// SetupSuite sets up the test suite with either mock server or live server
func (suite *APITestSuite) SetupSuite() {
	// Check if we should use live server
	suite.useLiveServer = os.Getenv("API_TEST_LIVE") == "true"
	suite.httpClient = &http.Client{Timeout: 30 * time.Second}

	if suite.useLiveServer {
		// Use live development server
		suite.baseURL = os.Getenv("API_TEST_URL")
		if suite.baseURL == "" {
			suite.baseURL = "http://localhost:8080"
		}
		suite.T().Logf("Running tests against live server: %s", suite.baseURL)

		// Verify server is running
		resp, err := suite.httpClient.Get(suite.baseURL + "/health")
		if err != nil {
			suite.T().Fatalf("Live server not available at %s: %v", suite.baseURL, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			suite.T().Fatalf("Live server health check failed: %d", resp.StatusCode)
		}
	} else {
		// Use mock server (default behavior)
		suite.setupMockServer()
	}
}

// setupMockServer creates a mock server for testing
func (suite *APITestSuite) setupMockServer() {
	// Create mock JSONPlaceholder-like server
	suite.mockServer = httptest.NewServer(http.HandlerFunc(suite.mockJSONPlaceholderHandler))
	suite.baseURL = suite.mockServer.URL

	// Create configuration with mock server
	configData := fmt.Sprintf(`{
		"server": {
			"port": 8080,
			"read_timeout": 30,
			"write_timeout": 30
		},
		"endpoints": {
			"jsonplaceholder": {
				"name": "jsonplaceholder",
				"target": "%s"
			},
			"httpbin": {
				"name": "httpbin", 
				"target": "%s"
			}
		}
	}`, suite.mockServer.URL, suite.mockServer.URL)

	configProvider := &mockConfigProvider{configData: configData}

	// Initialize components
	httpClient := client.NewClient(30 * time.Second)
	transformer := transform.NewUnifiedTransformer()

	logger, _ := logging.NewLogger("error")

	proxyService := proxy.NewService(configProvider, httpClient, transformer, logger)
	handler := proxy.NewHandler(proxyService, logger)

	suite.proxyHandler = handler.SetupRoutes()
}

// TearDownSuite cleans up the test suite
func (suite *APITestSuite) TearDownSuite() {
	if suite.mockServer != nil {
		suite.mockServer.Close()
	}
}

// mockJSONPlaceholderHandler simulates JSONPlaceholder API responses
func (suite *APITestSuite) mockJSONPlaceholderHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.URL.Path == "/posts/1" && r.Method == "GET":
		suite.handleSinglePost(w, r)
	case r.URL.Path == "/posts" && r.Method == "GET":
		suite.handleAllPosts(w, r)
	case r.URL.Path == "/posts" && r.Method == "POST":
		suite.handleCreatePost(w, r)
	case r.URL.Path == "/nonexistent-endpoint/test":
		http.NotFound(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (suite *APITestSuite) handleSinglePost(w http.ResponseWriter, _ *http.Request) {
	post := map[string]any{
		"userId": 1,
		"id":     1,
		"title":  "sunt aut facere repellat provident occaecati excepturi optio reprehenderit",
		"body":   "quia et suscipit\nsuscipit recusandae consequuntur expedita et cum\nreprehenderit molestiae ut ut quas totam\nnostrum rerum est autem sunt rem eveniet architecto",
	}
	json.NewEncoder(w).Encode(post)
}

func (suite *APITestSuite) handleAllPosts(w http.ResponseWriter, _ *http.Request) {
	posts := []map[string]any{
		{
			"userId": 1,
			"id":     1,
			"title":  "sunt aut facere repellat provident occaecati excepturi optio reprehenderit",
			"body":   "quia et suscipit\nsuscipit recusandae consequuntur expedita et cum\nreprehenderit molestiae ut ut quas totam\nnostrum rerum est autem sunt rem eveniet architecto",
		},
		{
			"userId": 1,
			"id":     2,
			"title":  "qui est esse",
			"body":   "est rerum tempore vitae\nsequi sint nihil reprehenderit dolor beatae ea dolores neque\nfugiat blanditiis voluptate porro vel nihil molestiae ut reiciendis\nqui aperiam non debitis possimus qui neque nisi nulla",
		},
		{
			"userId": 2,
			"id":     3,
			"title":  "ea molestias quasi exercitationem repellat qui ipsa sit aut",
			"body":   "et iusto sed quo iure\nvoluptatem occaecati omnis eligendi aut ad\nvoluptatem doloribus vel accusantium quis pariatur\nmolestiae porro eius odio et labore et velit aut",
		},
	}
	json.NewEncoder(w).Encode(posts)
}

func (suite *APITestSuite) handleCreatePost(w http.ResponseWriter, r *http.Request) {
	var requestBody map[string]any
	json.NewDecoder(r.Body).Decode(&requestBody)

	response := map[string]any{
		"id":     101,
		"title":  requestBody["title"],
		"body":   requestBody["body"],
		"userId": requestBody["userId"],
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// TestHealthEndpoint tests the health check endpoint (equivalent to test_health in shell script)
func (suite *APITestSuite) TestHealthEndpoint() {
	if suite.useLiveServer {
		suite.testHealthLive()
	} else {
		suite.testHealthMock()
	}
}

func (suite *APITestSuite) testHealthMock() {
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	suite.proxyHandler.ServeHTTP(rr, req)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
	assert.Equal(suite.T(), "application/json", rr.Header().Get("Content-Type"))

	var response map[string]any
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "healthy", response["status"])
	assert.Equal(suite.T(), "jq-proxy-service", response["service"])
}

func (suite *APITestSuite) testHealthLive() {
	resp, err := suite.httpClient.Get(suite.baseURL + "/health")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	assert.Equal(suite.T(), "application/json", resp.Header.Get("Content-Type"))

	var response map[string]any
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "healthy", response["status"])
	assert.Equal(suite.T(), "jq-proxy-service", response["service"])
}

// makeProxyRequest makes a proxy request that works with both mock and live servers
func (suite *APITestSuite) makeProxyRequest(endpoint, path string, requestBody map[string]any) (*http.Response, map[string]any, error) {
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, nil, err
	}

	url := fmt.Sprintf("%s/proxy/%s%s", suite.baseURL, endpoint, path)

	if suite.useLiveServer {
		// Make real HTTP request to live server
		req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := suite.httpClient.Do(req)
		if err != nil {
			return nil, nil, err
		}
		defer resp.Body.Close()

		// Read response body
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return resp, nil, err
		}

		// Try to parse as JSON
		var response map[string]any
		err = json.Unmarshal(bodyBytes, &response)
		if err != nil {
			// If JSON parsing fails, log the raw response for debugging
			logBody := string(bodyBytes)
			if len(logBody) > 200 {
				logBody = logBody[:200] + "..."
			}
			suite.T().Logf("Failed to parse JSON response (status %d): %s", resp.StatusCode, logBody)
			// For non-200 responses, try to return what we can
			if resp.StatusCode >= 400 {
				return resp, map[string]any{"error": string(bodyBytes)}, nil
			}
			return resp, nil, fmt.Errorf("failed to parse JSON response: %w", err)
		}

		return resp, response, nil
	} else {
		// Use mock server via handler
		req := httptest.NewRequest("POST", url, bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		suite.proxyHandler.ServeHTTP(rr, req)

		var response map[string]any
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		if err != nil {
			suite.T().Logf("Failed to parse mock response: %s", rr.Body.String())
			return nil, nil, err
		}

		// Create a mock response object for consistency
		mockResp := &http.Response{
			StatusCode: rr.Code,
			Header:     rr.Header(),
		}

		return mockResp, response, nil
	}
}

// TestSimpleGETRequest tests simple GET request (equivalent to test_simple in shell script)
func (suite *APITestSuite) TestSimpleGETRequest() {
	requestBody := map[string]any{
		"method":              "GET",
		"body":                nil,
		"transformation_mode": "jq",
		"jq_query":            "{result: .}",
	}

	resp, response, err := suite.makeProxyRequest("jsonplaceholder", "/posts/1", requestBody)
	require.NoError(suite.T(), err)

	if suite.useLiveServer {
		// Live server testing - external APIs might not be accessible or might return errors
		if resp.StatusCode == http.StatusBadGateway || resp.StatusCode == http.StatusServiceUnavailable {
			suite.T().Skipf("External API not accessible (status %d) - this is expected in some environments", resp.StatusCode)
			return
		}
		// If we get here, the external API is accessible
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
		if response != nil {
			if result, ok := response["result"].(map[string]any); ok {
				assert.Equal(suite.T(), float64(1), result["id"])
				assert.NotEmpty(suite.T(), result["title"])
				assert.NotEmpty(suite.T(), result["body"])
			}
		}
	} else {
		// Mock server testing - should always work
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
		if result, ok := response["result"].(map[string]any); ok {
			assert.Equal(suite.T(), float64(1), result["id"])
			assert.Equal(suite.T(), float64(1), result["userId"])
			assert.NotEmpty(suite.T(), result["title"])
			assert.NotEmpty(suite.T(), result["body"])
		} else {
			suite.T().Fatalf("Expected result to be a map, got %T", response["result"])
		}
	}
}

// TestJQTransformation tests jq transformation (equivalent to test_transform in shell script)
func (suite *APITestSuite) TestJQTransformation() {
	requestBody := map[string]any{
		"method":              "GET",
		"body":                nil,
		"transformation_mode": "jq",
		"jq_query":            "{post_ids: [.[].id], post_titles: [.[].title]}",
	}

	resp, response, err := suite.makeProxyRequest("jsonplaceholder", "/posts", requestBody)
	require.NoError(suite.T(), err)

	if suite.useLiveServer {
		// Live server testing - external APIs might not be accessible
		if resp.StatusCode == http.StatusBadGateway || resp.StatusCode == http.StatusServiceUnavailable {
			suite.T().Skipf("External API not accessible (status %d) - this is expected in some environments", resp.StatusCode)
			return
		}
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
		if response != nil {
			if postIds, ok := response["post_ids"].([]any); ok && len(postIds) > 0 {
				assert.Greater(suite.T(), len(postIds), 50, "Expected many posts from live JSONPlaceholder")
				assert.Equal(suite.T(), float64(1), postIds[0])
			}
		}
	} else {
		// Mock server testing - should always work
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
		if postIds, ok := response["post_ids"].([]any); ok {
			assert.Len(suite.T(), postIds, 3)
			assert.Equal(suite.T(), float64(1), postIds[0])
			assert.Equal(suite.T(), float64(2), postIds[1])
			assert.Equal(suite.T(), float64(3), postIds[2])
		} else {
			suite.T().Fatalf("Expected post_ids to be an array, got %T", response["post_ids"])
		}

		if postTitles, ok := response["post_titles"].([]any); ok {
			assert.Len(suite.T(), postTitles, 3)
			assert.NotEmpty(suite.T(), postTitles[0])
		} else {
			suite.T().Fatalf("Expected post_titles to be an array, got %T", response["post_titles"])
		}
	}
}

// TestJQAdvancedTransformation tests advanced jq transformation (equivalent to test_jq_transform in shell script)
func (suite *APITestSuite) TestJQAdvancedTransformation() {
	requestBody := map[string]any{
		"method":              "GET",
		"body":                nil,
		"transformation_mode": "jq",
		"jq_query":            "{posts: [.[] | {id: .id, title: .title}], count: length}",
	}

	resp, response, err := suite.makeProxyRequest("jsonplaceholder", "/posts", requestBody)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	// Verify jq transformation results
	if suite.useLiveServer {
		// Live server returns 100 posts from JSONPlaceholder
		assert.Equal(suite.T(), float64(100), response["count"])
	} else {
		// Mock server returns our test data (3 posts)
		assert.Equal(suite.T(), float64(3), response["count"])
	}

	if posts, ok := response["posts"].([]any); ok {
		if suite.useLiveServer {
			assert.Len(suite.T(), posts, 100)
		} else {
			assert.Len(suite.T(), posts, 3)
		}

		// Check first post structure
		if firstPost, ok := posts[0].(map[string]any); ok {
			assert.Equal(suite.T(), float64(1), firstPost["id"])
			assert.NotEmpty(suite.T(), firstPost["title"])
		} else {
			suite.T().Fatalf("Expected first post to be a map, got %T", posts[0])
		}
	} else {
		suite.T().Fatalf("Expected posts to be an array, got %T", response["posts"])
	}
}

// TestPOSTRequest tests POST request with body (equivalent to test_post in shell script)
func (suite *APITestSuite) TestPOSTRequest() {
	requestBody := map[string]any{
		"method": "POST",
		"body": map[string]any{
			"title":  "Test Post",
			"body":   "This is a test post",
			"userId": 1,
		},
		"transformation_mode": "jq",
		"jq_query":            "{created_id: .id, created_title: .title}",
	}

	resp, response, err := suite.makeProxyRequest("jsonplaceholder", "/posts", requestBody)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	// Verify POST response transformation
	if suite.useLiveServer {
		// Live JSONPlaceholder returns id 101 for new posts
		assert.Equal(suite.T(), float64(101), response["created_id"])
	} else {
		// Mock server returns our test data
		assert.Equal(suite.T(), float64(101), response["created_id"])
	}
	assert.Equal(suite.T(), "Test Post", response["created_title"])
}

// TestErrorHandling tests error scenarios (equivalent to test_error in shell script)
func (suite *APITestSuite) TestErrorHandling() {
	requestBody := map[string]any{
		"method":              "GET",
		"body":                nil,
		"transformation_mode": "jq",
		"jq_query":            "{result: .}",
	}

	resp, response, err := suite.makeProxyRequest("nonexistent-endpoint", "/test", requestBody)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)

	// For error responses, the response should contain error information
	if errorCode, ok := response["error"].(map[string]any); ok {
		assert.Equal(suite.T(), "ENDPOINT_NOT_FOUND", errorCode["code"])
		assert.Contains(suite.T(), errorCode["message"], "nonexistent-endpoint")
	} else {
		// Try parsing as ErrorResponse structure
		var errorResponse models.ErrorResponse
		bodyBytes, _ := json.Marshal(response)
		err = json.Unmarshal(bodyBytes, &errorResponse)
		if err == nil {
			assert.Equal(suite.T(), "ENDPOINT_NOT_FOUND", errorResponse.Error.Code)
			assert.Contains(suite.T(), errorResponse.Error.Message, "nonexistent-endpoint")
		} else {
			suite.T().Logf("Error response format: %+v", response)
		}
	}
}

// TestAllScenarios runs all test scenarios in sequence (equivalent to test_all in shell script)
func (suite *APITestSuite) TestAllScenarios() {
	suite.T().Log("Running comprehensive API test scenarios...")

	// Run all individual tests
	suite.TestHealthEndpoint()
	suite.TestSimpleGETRequest()
	suite.TestJQTransformation()
	suite.TestJQAdvancedTransformation()
	suite.TestPOSTRequest()
	suite.TestErrorHandling()

	suite.T().Log("All API test scenarios completed successfully!")
}

// TestDifferentEndpoints tests the ability to use different configured endpoints
func (suite *APITestSuite) TestDifferentEndpoints() {
	tests := []struct {
		name     string
		endpoint string
		path     string
		expected int
	}{
		{
			name:     "jsonplaceholder endpoint",
			endpoint: "jsonplaceholder",
			path:     "/posts/1",
			expected: http.StatusOK,
		},
		{
			name:     "httpbin endpoint",
			endpoint: "httpbin",
			path:     "/posts/1",
			expected: http.StatusOK,
		},
		{
			name:     "nonexistent endpoint",
			endpoint: "nonexistent",
			path:     "/test",
			expected: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			requestBody := map[string]any{
				"method":              "GET",
				"body":                nil,
				"transformation_mode": "jq",
				"jq_query":            "{result: .}",
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", fmt.Sprintf("/proxy/%s%s", tt.endpoint, tt.path), bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			suite.proxyHandler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expected, rr.Code)
		})
	}
}

// TestVerboseMode tests detailed response validation (equivalent to verbose mode in shell script)
func (suite *APITestSuite) TestVerboseMode() {
	suite.T().Log("Testing with verbose validation...")

	requestBody := map[string]any{
		"method":              "GET",
		"body":                nil,
		"transformation_mode": "jq",
		"jq_query":            "{full_response: ., post_id: .id, post_title: .title, user_id: .userId}",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/proxy/jsonplaceholder/posts/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	suite.proxyHandler.ServeHTTP(rr, req)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)

	var response map[string]any
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	// Verbose validation of all fields
	assert.Equal(suite.T(), float64(1), response["post_id"])
	assert.Equal(suite.T(), float64(1), response["user_id"])
	assert.NotEmpty(suite.T(), response["post_title"])

	// Validate full response structure
	if fullResponse, ok := response["full_response"].(map[string]any); ok {
		assert.Equal(suite.T(), float64(1), fullResponse["id"])
		assert.Equal(suite.T(), float64(1), fullResponse["userId"])
		assert.NotEmpty(suite.T(), fullResponse["title"])
		assert.NotEmpty(suite.T(), fullResponse["body"])
	} else {
		suite.T().Fatalf("Expected full_response to be a map, got %T", response["full_response"])
	}

	suite.T().Logf("Response validation completed. Response keys: %v", getKeys(response))
}

// TestComplexTransformationScenarios tests advanced transformation patterns
func (suite *APITestSuite) TestComplexTransformationScenarios() {
	tests := []struct {
		name        string
		mode        string
		query       any
		path        string
		description string
	}{
		{
			name:        "jq array filtering",
			mode:        "jq",
			query:       "{post_titles: [.[].title], post_ids: [.[].id]}",
			path:        "/posts",
			description: "Filter posts by user and extract titles",
		},
		{
			name: "jq complex aggregation",
			mode: "jq",
			query: `{
				user_stats: group_by(.userId) | map({
					user_id: .[0].userId,
					post_count: length,
					titles: [.[].title]
				}),
				total_posts: length,
				avg_title_length: [.[].title | length] | add / length
			}`,
			path:        "/posts",
			description: "Group posts by user with statistics",
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			var requestBody map[string]any

			requestBody = map[string]any{
				"method":              "GET",
				"body":                nil,
				"transformation_mode": "jq",
				"jq_query":            tt.query,
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/proxy/jsonplaceholder"+tt.path, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			suite.proxyHandler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code, "Test: %s", tt.description)

			var response map[string]any
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(t, err, "Test: %s", tt.description)

			// Basic validation that transformation worked
			assert.NotEmpty(t, response, "Test: %s should return non-empty response", tt.description)

			t.Logf("Test '%s' completed successfully. Response keys: %v", tt.name, getKeys(response))
		})
	}
}

// mockConfigProvider implements ConfigProvider for testing
type mockConfigProvider struct {
	configData string
}

func (m *mockConfigProvider) LoadConfig() (*models.ProxyConfig, error) {
	return models.ParseProxyConfig([]byte(m.configData))
}

func (m *mockConfigProvider) GetEndpoint(name string) (*models.Endpoint, bool) {
	config, err := m.LoadConfig()
	if err != nil {
		return nil, false
	}
	endpoint, exists := config.Endpoints[name]
	return endpoint, exists
}

func (m *mockConfigProvider) Reload() error {
	_, err := m.LoadConfig()
	return err
}

// Helper function to get map keys for logging
func getKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// TestAPITestSuite runs the complete API test suite
func TestAPITestSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
