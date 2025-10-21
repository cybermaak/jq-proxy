package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"jq-proxy-service/internal/models"
	"jq-proxy-service/internal/proxy"
	"jq-proxy-service/internal/transform"
	"jq-proxy-service/pkg/client"
)

// IntegrationTestSuite contains integration tests for the complete proxy service
type IntegrationTestSuite struct {
	suite.Suite
	mockServer   *httptest.Server
	proxyHandler http.Handler
	configData   string
}

// SetupSuite sets up the test suite
func (suite *IntegrationTestSuite) SetupSuite() {
	// Create mock target server
	suite.mockServer = httptest.NewServer(http.HandlerFunc(suite.mockServerHandler))

	// Create configuration with mock server URL
	suite.configData = fmt.Sprintf(`{
		"server": {
			"port": 8080,
			"read_timeout": 30,
			"write_timeout": 30
		},
		"endpoints": {
			"test-api": {
				"name": "test-api",
				"target": "%s"
			},
			"json-api": {
				"name": "json-api", 
				"target": "%s"
			}
		}
	}`, suite.mockServer.URL, suite.mockServer.URL)

	// Create temporary config file
	configProvider := &mockConfigProvider{configData: suite.configData}

	// Initialize components
	httpClient := client.NewClient(30 * time.Second)
	transformer := transform.NewUnifiedTransformer()

	// Create logger for tests
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce log noise in tests

	proxyService := proxy.NewService(configProvider, httpClient, transformer, logger)
	handler := proxy.NewHandler(proxyService, logger)

	suite.proxyHandler = handler.SetupRoutes()
}

// TearDownSuite cleans up the test suite
func (suite *IntegrationTestSuite) TearDownSuite() {
	if suite.mockServer != nil {
		suite.mockServer.Close()
	}
}

// mockServerHandler handles requests to the mock target server
func (suite *IntegrationTestSuite) mockServerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.URL.Path {
	case "/users":
		suite.handleUsersEndpoint(w, r)
	case "/posts":
		suite.handlePostsEndpoint(w, r)
	case "/error":
		http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
	case "/timeout":
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "delayed response"})
	default:
		http.NotFound(w, r)
	}
}

func (suite *IntegrationTestSuite) handleUsersEndpoint(w http.ResponseWriter, _ *http.Request) {
	users := []map[string]interface{}{
		{
			"id":    1,
			"name":  "John Doe",
			"email": "john@example.com",
			"profile": map[string]interface{}{
				"age":  30,
				"city": "New York",
			},
		},
		{
			"id":    2,
			"name":  "Jane Smith",
			"email": "jane@example.com",
			"profile": map[string]interface{}{
				"age":  25,
				"city": "Los Angeles",
			},
		},
	}

	response := map[string]interface{}{
		"data":  users,
		"total": len(users),
		"page":  1,
	}

	json.NewEncoder(w).Encode(response)
}

func (suite *IntegrationTestSuite) handlePostsEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Handle POST request - create new post
		var requestBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&requestBody)

		response := map[string]interface{}{
			"id":     101,
			"title":  requestBody["title"],
			"body":   requestBody["body"],
			"userId": requestBody["userId"],
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Handle GET request - return posts
	posts := []map[string]interface{}{
		{
			"id":     1,
			"title":  "First Post",
			"body":   "This is the first post",
			"userId": 1,
		},
		{
			"id":     2,
			"title":  "Second Post",
			"body":   "This is the second post",
			"userId": 2,
		},
	}

	json.NewEncoder(w).Encode(posts)
}

// TestHealthEndpoint tests the health check endpoint
func (suite *IntegrationTestSuite) TestHealthEndpoint() {
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	suite.proxyHandler.ServeHTTP(rr, req)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
	assert.Equal(suite.T(), "application/json", rr.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "healthy", response["status"])
	assert.Equal(suite.T(), "jq-proxy-service", response["service"])
}

// TestJSONPathTransformation tests JSONPath-based transformations
func (suite *IntegrationTestSuite) TestJSONPathTransformation() {
	requestBody := map[string]interface{}{
		"method": "GET",
		"body":   nil,
		"transformation": map[string]interface{}{
			"user_count":      "$.total",
			"user_names":      "$.data[*].name",
			"first_user_id":   "$.data[0].id",
			"first_user_name": "$.data[0].name",
		},
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/proxy/test-api/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	rr := httptest.NewRecorder()
	suite.proxyHandler.ServeHTTP(rr, req)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), float64(2), response["user_count"])
	assert.Equal(suite.T(), []interface{}{"John Doe", "Jane Smith"}, response["user_names"])
	assert.Equal(suite.T(), float64(1), response["first_user_id"])
	assert.Equal(suite.T(), "John Doe", response["first_user_name"])
}

// TestJQTransformation tests jq-based transformations
func (suite *IntegrationTestSuite) TestJQTransformation() {
	requestBody := map[string]interface{}{
		"method":              "GET",
		"body":                nil,
		"transformation_mode": "jq",
		"jq_query":            "{user_count: .total, user_names: [.data[].name], avg_age: (.data | map(.profile.age) | add / length)}",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/proxy/test-api/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	suite.proxyHandler.ServeHTTP(rr, req)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), float64(2), response["user_count"])
	assert.Equal(suite.T(), []interface{}{"John Doe", "Jane Smith"}, response["user_names"])
	assert.Equal(suite.T(), 27.5, response["avg_age"])
}

// TestPOSTRequest tests POST requests with body forwarding
func (suite *IntegrationTestSuite) TestPOSTRequest() {
	requestBody := map[string]interface{}{
		"method": "POST",
		"body": map[string]interface{}{
			"title":  "New Post",
			"body":   "This is a new post",
			"userId": 1,
		},
		"transformation": map[string]interface{}{
			"created_id":    "$.id",
			"created_title": "$.title",
		},
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/proxy/test-api/posts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	suite.proxyHandler.ServeHTTP(rr, req)

	assert.Equal(suite.T(), http.StatusCreated, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), float64(101), response["created_id"])
	assert.Equal(suite.T(), "New Post", response["created_title"])
}

// TestHeaderForwarding tests that headers are properly forwarded
func (suite *IntegrationTestSuite) TestHeaderForwarding() {
	// Create a mock server that echoes headers
	echoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"headers": map[string]string{
				"authorization": r.Header.Get("Authorization"),
				"custom-header": r.Header.Get("Custom-Header"),
				"jpx-debug":     r.Header.Get("jpx-debug"), // Should be empty
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer echoServer.Close()

	// Update config to use echo server
	configData := fmt.Sprintf(`{
		"server": {"port": 8080, "read_timeout": 30, "write_timeout": 30},
		"endpoints": {
			"echo-api": {"name": "echo-api", "target": "%s"}
		}
	}`, echoServer.URL)

	configProvider := &mockConfigProvider{configData: configData}
	httpClient := client.NewClient(30 * time.Second)
	transformer := transform.NewUnifiedTransformer()

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	proxyService := proxy.NewService(configProvider, httpClient, transformer, logger)
	handler := proxy.NewHandler(proxyService, logger)
	proxyHandler := handler.SetupRoutes()

	requestBody := map[string]interface{}{
		"method": "GET",
		"body":   nil,
		"transformation": map[string]interface{}{
			"result": "$.headers",
		},
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/proxy/echo-api/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Custom-Header", "test-value")
	req.Header.Set("jpx-debug", "true") // Should be filtered out

	rr := httptest.NewRecorder()
	proxyHandler.ServeHTTP(rr, req)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	if result, ok := response["result"].(map[string]interface{}); ok {
		assert.Equal(suite.T(), "Bearer test-token", result["authorization"])
		assert.Equal(suite.T(), "test-value", result["custom-header"])
		assert.Empty(suite.T(), result["jpx-debug"]) // Should be filtered out
	} else {
		suite.T().Logf("Response: %+v", response)
		suite.T().Fail()
	}
}

// TestErrorHandling tests various error scenarios
func (suite *IntegrationTestSuite) TestErrorHandling() {
	tests := []struct {
		name           string
		endpoint       string
		path           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:     "nonexistent endpoint",
			endpoint: "nonexistent",
			path:     "/test",
			requestBody: map[string]interface{}{
				"method":         "GET",
				"transformation": map[string]interface{}{"result": "$.data"},
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "ENDPOINT_NOT_FOUND",
		},
		{
			name:     "invalid transformation",
			endpoint: "test-api",
			path:     "/users",
			requestBody: map[string]interface{}{
				"method":         "GET",
				"transformation": map[string]interface{}{"result": "$.invalid["},
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "TRANSFORMATION_ERROR",
		},
		{
			name:     "invalid jq query",
			endpoint: "test-api",
			path:     "/users",
			requestBody: map[string]interface{}{
				"method":              "GET",
				"transformation_mode": "jq",
				"jq_query":            ".invalid | map(",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "TRANSFORMATION_ERROR",
		},
		{
			name:     "upstream error",
			endpoint: "test-api",
			path:     "/error",
			requestBody: map[string]interface{}{
				"method":         "GET",
				"transformation": map[string]interface{}{"result": "$.error"},
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", fmt.Sprintf("/proxy/%s%s", tt.endpoint, tt.path), bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			suite.proxyHandler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedError != "" {
				var errorResponse models.ErrorResponse
				err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedError, errorResponse.Error.Code)
			}
		})
	}
}

// TestCORSHeaders tests CORS functionality
func (suite *IntegrationTestSuite) TestCORSHeaders() {
	req := httptest.NewRequest("OPTIONS", "/proxy/test-api/users", nil)
	req.Header.Set("Origin", "https://example.com")

	rr := httptest.NewRecorder()
	suite.proxyHandler.ServeHTTP(rr, req)

	assert.Equal(suite.T(), http.StatusOK, rr.Code)
	assert.Equal(suite.T(), "*", rr.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(suite.T(), rr.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(suite.T(), rr.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
}

// TestComplexTransformationScenarios tests complex real-world transformation scenarios
func (suite *IntegrationTestSuite) TestComplexTransformationScenarios() {
	tests := []struct {
		name     string
		mode     string
		query    interface{}
		expected map[string]interface{}
	}{
		{
			name: "JSONPath complex nested transformation",
			mode: "jsonpath",
			query: map[string]interface{}{
				"total":  "$.total",
				"page":   "$.page",
				"names":  "$.data[*].name",
				"cities": "$.data[*].profile.city",
			},
			expected: map[string]interface{}{
				"total":  float64(2),
				"page":   float64(1),
				"names":  []interface{}{"John Doe", "Jane Smith"},
				"cities": []interface{}{"New York", "Los Angeles"},
			},
		},
		{
			name:  "jq complex aggregation",
			mode:  "jq",
			query: "{summary: {total: .total, page: .page}, users: [.data[] | {id: .id, name: .name, age: .profile.age}], avg_age: (.data | map(.profile.age) | add / length), cities: [.data[].profile.city] | unique}",
			expected: map[string]interface{}{
				"summary": map[string]interface{}{"total": float64(2), "page": float64(1)},
				"users": []interface{}{
					map[string]interface{}{"id": float64(1), "name": "John Doe", "age": float64(30)},
					map[string]interface{}{"id": float64(2), "name": "Jane Smith", "age": float64(25)},
				},
				"avg_age": 27.5,
				"cities":  []interface{}{"Los Angeles", "New York"},
			},
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			var requestBody map[string]interface{}

			if tt.mode == "jq" {
				requestBody = map[string]interface{}{
					"method":              "GET",
					"body":                nil,
					"transformation_mode": "jq",
					"jq_query":            tt.query,
				}
			} else {
				requestBody = map[string]interface{}{
					"method":         "GET",
					"body":           nil,
					"transformation": tt.query,
				}
			}

			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", "/proxy/test-api/users", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			suite.proxyHandler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)

			var response map[string]interface{}
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(t, err)

			for key, expectedValue := range tt.expected {
				assert.Equal(t, expectedValue, response[key], "Mismatch for key: %s", key)
			}
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

// TestIntegrationSuite runs the integration test suite
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
