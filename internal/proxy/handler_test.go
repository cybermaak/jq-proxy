package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"jq-proxy-service/internal/logging"
	"jq-proxy-service/internal/models"
)

// MockProxyService for testing
type MockProxyService struct {
	mock.Mock
}

func (m *MockProxyService) HandleRequest(ctx context.Context, endpointName string, path string, queryParams url.Values, headers http.Header, proxyReq *models.ProxyRequest) (*models.ProxyResponse, error) {
	args := m.Called(ctx, endpointName, path, queryParams, headers, proxyReq)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProxyResponse), args.Error(1)
}

func (m *MockProxyService) GetConfig() *models.ProxyConfig {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.ProxyConfig)
}

// Helper function to create a test logger
func createTestLogger() *logging.Logger {
	logger, _ := logging.NewLogger("error")
	return logger
}

func TestHandler_HandleProxyRequest_Success(t *testing.T) {
	// Setup
	mockService := &MockProxyService{}
	logger := createTestLogger()

	handler := NewHandler(mockService, logger)
	router := handler.SetupRoutes()

	// Test data
	requestBody := map[string]interface{}{
		"method":              "GET",
		"body":                nil,
		"transformation_mode": "jq",
		"jq_query":            "{users: [.data[].name]}",
	}

	responseData := map[string]interface{}{
		"users": []interface{}{"John", "Jane"},
	}

	proxyResponse := &models.ProxyResponse{
		Data:   responseData,
		Status: 200,
	}

	// Setup expectations
	mockService.On("HandleRequest",
		mock.Anything, // context
		"user-service",
		"/api/users",
		url.Values{"limit": []string{"10"}},
		mock.AnythingOfType("http.Header"),
		mock.MatchedBy(func(req *models.ProxyRequest) bool {
			return req.Method == "GET" && req.Body == nil
		}),
	).Return(proxyResponse, nil)

	// Create request
	reqBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/proxy/user-service/api/users?limit=10", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token")

	// Execute
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, responseData, response)

	mockService.AssertExpectations(t)
}

func TestHandler_HandleProxyRequest_EndpointNotFound(t *testing.T) {
	// Setup
	mockService := &MockProxyService{}
	logger := createTestLogger()

	handler := NewHandler(mockService, logger)
	router := handler.SetupRoutes()

	// Test data
	requestBody := map[string]interface{}{
		"method":              "GET",
		"body":                nil,
		"transformation_mode": "jq",
		"jq_query":            "{result: .data}",
	}

	endpointError := &EndpointNotFoundError{
		EndpointName:       "nonexistent-service",
		AvailableEndpoints: []string{"user-service", "order-service"},
	}

	// Setup expectations
	mockService.On("HandleRequest",
		mock.Anything,
		"nonexistent-service",
		"/api/data",
		url.Values{},
		mock.AnythingOfType("http.Header"),
		mock.Anything,
	).Return((*models.ProxyResponse)(nil), endpointError)

	// Create request
	reqBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/proxy/nonexistent-service/api/data", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "ENDPOINT_NOT_FOUND", errorResponse.Error.Code)
	assert.Contains(t, errorResponse.Error.Message, "nonexistent-service")

	details, ok := errorResponse.Error.Details.(map[string]interface{})
	require.True(t, ok)
	availableEndpoints, ok := details["available_endpoints"].([]interface{})
	require.True(t, ok)
	assert.Len(t, availableEndpoints, 2)

	mockService.AssertExpectations(t)
}

func TestHandler_HandleProxyRequest_InvalidJSON(t *testing.T) {
	// Setup
	mockService := &MockProxyService{}
	logger := createTestLogger()

	handler := NewHandler(mockService, logger)
	router := handler.SetupRoutes()

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/proxy/user-service/api/users", bytes.NewReader([]byte(`{"method": "GET", "body":}`)))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "INVALID_REQUEST", errorResponse.Error.Code)
	assert.Contains(t, errorResponse.Error.Message, "Invalid request format")

	// Should not call the service
	mockService.AssertNotCalled(t, "HandleRequest")
}

func TestHandler_HandleProxyRequest_MissingRequiredFields(t *testing.T) {
	// Setup
	mockService := &MockProxyService{}
	logger := createTestLogger()

	handler := NewHandler(mockService, logger)
	router := handler.SetupRoutes()

	// Test data missing required fields
	requestBody := map[string]interface{}{
		"body": nil,
		// Missing method and transformation
	}

	// Create request
	reqBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/proxy/user-service/api/users", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "INVALID_REQUEST", errorResponse.Error.Code)

	mockService.AssertNotCalled(t, "HandleRequest")
}

func TestHandler_HandleProxyRequest_TransformationError(t *testing.T) {
	// Setup
	mockService := &MockProxyService{}
	logger := createTestLogger()

	handler := NewHandler(mockService, logger)
	router := handler.SetupRoutes()

	// Test data
	requestBody := map[string]interface{}{
		"method":              "GET",
		"body":                nil,
		"transformation_mode": "jq",
		"jq_query":            ".invalid[",
	}

	transformationError := &TransformationError{
		Message: "Invalid jq expression",
		Details: map[string]interface{}{
			"expression": ".invalid[",
		},
	}

	// Setup expectations
	mockService.On("HandleRequest",
		mock.Anything,
		"user-service",
		"",
		url.Values{},
		mock.AnythingOfType("http.Header"),
		mock.Anything,
	).Return((*models.ProxyResponse)(nil), transformationError)

	// Create request
	reqBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/proxy/user-service", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "TRANSFORMATION_ERROR", errorResponse.Error.Code)
	assert.Contains(t, errorResponse.Error.Message, "Invalid jq expression")

	mockService.AssertExpectations(t)
}

func TestHandler_HandleProxyRequest_UpstreamError(t *testing.T) {
	// Setup
	mockService := &MockProxyService{}
	logger := createTestLogger()

	handler := NewHandler(mockService, logger)
	router := handler.SetupRoutes()

	// Test data
	requestBody := map[string]interface{}{
		"method":              "GET",
		"body":                nil,
		"transformation_mode": "jq",
		"jq_query":            "{result: .data}",
	}

	upstreamError := &UpstreamError{
		Message:    "Connection timeout",
		StatusCode: http.StatusGatewayTimeout,
		Details: map[string]interface{}{
			"endpoint": "user-service",
		},
	}

	// Setup expectations
	mockService.On("HandleRequest",
		mock.Anything,
		"user-service",
		"/api/users",
		url.Values{},
		mock.AnythingOfType("http.Header"),
		mock.Anything,
	).Return((*models.ProxyResponse)(nil), upstreamError)

	// Create request
	reqBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/proxy/user-service/api/users", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusGatewayTimeout, rr.Code)

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "UPSTREAM_ERROR", errorResponse.Error.Code)
	assert.Contains(t, errorResponse.Error.Message, "Connection timeout")

	mockService.AssertExpectations(t)
}

func TestHandler_HealthCheck(t *testing.T) {
	// Setup
	mockService := &MockProxyService{}
	logger := createTestLogger()

	handler := NewHandler(mockService, logger)
	router := handler.SetupRoutes()

	// Create request
	req := httptest.NewRequest("GET", "/health", nil)

	// Execute
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "jq-proxy-service", response["service"])
}

func TestHandler_CORS(t *testing.T) {
	// Setup
	mockService := &MockProxyService{}
	logger := createTestLogger()

	handler := NewHandler(mockService, logger)
	router := handler.SetupRoutes()

	// Test OPTIONS request
	req := httptest.NewRequest("OPTIONS", "/proxy/user-service/api/users", nil)
	req.Header.Set("Origin", "https://example.com")

	// Execute
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, rr.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, rr.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
}

func TestHandler_MetricsEndpoint(t *testing.T) {
	mockService := &MockProxyService{}
	logger := createTestLogger()

	handler := NewHandler(mockService, logger)
	router := handler.SetupRoutes()

	req := httptest.NewRequest("GET", "/metrics", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "total_requests")
	assert.Contains(t, response, "total_errors")
}

func TestHandler_ConfigEndpoint_Success(t *testing.T) {
	mockService := &MockProxyService{}
	logger := createTestLogger()

	config := &models.ProxyConfig{
		Server: models.ServerConfig{
			Port:         8080,
			ReadTimeout:  30,
			WriteTimeout: 30,
		},
		Endpoints: map[string]*models.Endpoint{
			"svc": {Name: "svc", Target: "https://api.example.com"},
		},
	}
	mockService.On("GetConfig").Return(config)

	handler := NewHandler(mockService, logger)
	router := handler.SetupRoutes()

	req := httptest.NewRequest("GET", "/config", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "server")
	assert.Contains(t, response, "endpoints")

	endpoints, ok := response["endpoints"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, endpoints, "svc")

	mockService.AssertExpectations(t)
}

func TestHandler_ConfigEndpoint_NilConfig(t *testing.T) {
	mockService := &MockProxyService{}
	logger := createTestLogger()

	mockService.On("GetConfig").Return((*models.ProxyConfig)(nil))

	handler := NewHandler(mockService, logger)
	router := handler.SetupRoutes()

	req := httptest.NewRequest("GET", "/config", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	var errResp models.ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "CONFIG_ERROR", errResp.Error.Code)

	mockService.AssertExpectations(t)
}

func TestHandler_HandleProxyError_GenericError(t *testing.T) {
	mockService := &MockProxyService{}
	logger := createTestLogger()

	handler := NewHandler(mockService, logger)
	router := handler.SetupRoutes()

	// Return a plain error (not implementing ProxyError) to exercise the else branch.
	requestBody := map[string]interface{}{
		"method":              "GET",
		"body":                nil,
		"transformation_mode": "jq",
		"jq_query":            ".",
	}
	mockService.On("HandleRequest",
		mock.Anything,
		"svc",
		"",
		url.Values{},
		mock.AnythingOfType("http.Header"),
		mock.Anything,
	).Return((*models.ProxyResponse)(nil), assert.AnError)

	reqBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/proxy/svc", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	var errResp models.ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "INTERNAL_ERROR", errResp.Error.Code)

	mockService.AssertExpectations(t)
}

func TestHandler_HandleProxyRequest_BodyReadError(t *testing.T) {
	mockService := &MockProxyService{}
	logger := createTestLogger()

	handler := NewHandler(mockService, logger)
	router := handler.SetupRoutes()

	// errReader forces io.ReadAll to fail so we hit the read-error branch.
	req := httptest.NewRequest("POST", "/proxy/svc/path", errReader{})
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp models.ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "INVALID_REQUEST", errResp.Error.Code)

	mockService.AssertNotCalled(t, "HandleRequest")
}

// errReader is an io.Reader that always returns an error.
type errReader struct{}

func (errReader) Read([]byte) (int, error) {
	return 0, assert.AnError
}

func TestHandler_PathExtraction(t *testing.T) {
	// Setup
	mockService := &MockProxyService{}
	logger := createTestLogger()

	handler := NewHandler(mockService, logger)
	router := handler.SetupRoutes()

	tests := []struct {
		name         string
		url          string
		expectedPath string
	}{
		{
			name:         "no path",
			url:          "/proxy/user-service",
			expectedPath: "",
		},
		{
			name:         "simple path",
			url:          "/proxy/user-service/api/users",
			expectedPath: "/api/users",
		},
		{
			name:         "nested path",
			url:          "/proxy/user-service/api/v1/users/123",
			expectedPath: "/api/v1/users/123",
		},
		{
			name:         "path with query params",
			url:          "/proxy/user-service/api/users?limit=10",
			expectedPath: "/api/users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody := map[string]interface{}{
				"method":              "GET",
				"body":                nil,
				"transformation_mode": "jq",
				"jq_query":            "{result: .data}",
			}

			mockService.On("HandleRequest",
				mock.Anything,
				"user-service",
				tt.expectedPath,
				mock.Anything,
				mock.AnythingOfType("http.Header"),
				mock.Anything,
			).Return(&models.ProxyResponse{Data: map[string]interface{}{"result": "test"}, Status: 200}, nil).Once()

			reqBody, _ := json.Marshal(requestBody)
			req := httptest.NewRequest("POST", tt.url, bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
		})
	}

	mockService.AssertExpectations(t)
}

func TestHandler_LivezAlias(t *testing.T) {
	mockService := &MockProxyService{}
	logger := createTestLogger()

	handler := NewHandler(mockService, logger)
	router := handler.SetupRoutes()

	req := httptest.NewRequest("GET", "/livez", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
}

func TestHandler_Readyz_ConfigUnavailable(t *testing.T) {
	mockService := &MockProxyService{}
	logger := createTestLogger()
	mockService.On("GetConfig").Return((*models.ProxyConfig)(nil))

	handler := NewHandler(mockService, logger)
	router := handler.SetupRoutes()

	req := httptest.NewRequest("GET", "/readyz", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	var errResp models.ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "NOT_READY", errResp.Error.Code)
	mockService.AssertExpectations(t)
}

func TestHandler_Readyz_ConfigReady_DefaultMode(t *testing.T) {
	mockService := &MockProxyService{}
	logger := createTestLogger()
	config := &models.ProxyConfig{Endpoints: map[string]*models.Endpoint{}}
	mockService.On("GetConfig").Return(config)

	handler := NewHandler(mockService, logger)
	router := handler.SetupRoutes()

	req := httptest.NewRequest("GET", "/readyz", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ready", response["status"])
	checks, ok := response["checks"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "ok", checks["config"])
	assert.NotContains(t, checks, "dependencies")
	mockService.AssertExpectations(t)
}

func TestHandler_Readyz_DependencyCheckFailure(t *testing.T) {
	mockService := &MockProxyService{}
	logger := createTestLogger()

	downstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer downstream.Close()

	config := &models.ProxyConfig{Endpoints: map[string]*models.Endpoint{
		"downstream": {Name: "downstream", Target: downstream.URL},
	}}
	mockService.On("GetConfig").Return(config)

	handler := NewHandler(mockService, logger)
	router := handler.SetupRoutes()

	req := httptest.NewRequest("GET", "/readyz?check_deps=true", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "not_ready", response["status"])

	checks := response["checks"].(map[string]interface{})
	deps := checks["dependencies"].(map[string]interface{})
	dep := deps["downstream"].(map[string]interface{})
	assert.Equal(t, "error", dep["status"])
	mockService.AssertExpectations(t)
}

func TestHandler_Readyz_DependencyCheckSuccess(t *testing.T) {
	mockService := &MockProxyService{}
	logger := createTestLogger()

	downstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Fatalf("expected HEAD method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer downstream.Close()

	config := &models.ProxyConfig{Endpoints: map[string]*models.Endpoint{
		"downstream": {Name: "downstream", Target: downstream.URL},
	}}
	mockService.On("GetConfig").Return(config)

	handler := NewHandler(mockService, logger)
	router := handler.SetupRoutes()

	req := httptest.NewRequest("GET", "/readyz?check_deps=1", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ready", response["status"])

	checks := response["checks"].(map[string]interface{})
	deps := checks["dependencies"].(map[string]interface{})
	dep := deps["downstream"].(map[string]interface{})
	assert.Equal(t, "ok", dep["status"])
	mockService.AssertExpectations(t)
}

func TestShouldCheckDependencies(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{name: "default false", url: "/readyz", want: false},
		{name: "true", url: "/readyz?check_deps=true", want: true},
		{name: "one", url: "/readyz?check_deps=1", want: true},
		{name: "case insensitive", url: "/readyz?check_deps=TRUE", want: true},
		{name: "other value", url: "/readyz?check_deps=" + url.QueryEscape("yes"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			got := shouldCheckDependencies(req.URL.Query())
			if got != tt.want {
				t.Fatalf("shouldCheckDependencies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandler_Readyz_DependencyInvalidTarget(t *testing.T) {
	mockService := &MockProxyService{}
	logger := createTestLogger()

	config := &models.ProxyConfig{Endpoints: map[string]*models.Endpoint{
		"bad": {Name: "bad", Target: "://bad-url"},
	}}
	mockService.On("GetConfig").Return(config)

	handler := NewHandler(mockService, logger)
	router := handler.SetupRoutes()

	req := httptest.NewRequest("GET", "/readyz?check_deps=true", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "not_ready", response["status"])
	mockService.AssertExpectations(t)
}

func TestHandler_Readyz_DependencyCheck_LimitsToKeyUpstreams(t *testing.T) {
	mockService := &MockProxyService{}
	logger := createTestLogger()

	called := 0
	downstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
		w.WriteHeader(http.StatusNoContent)
	}))
	defer downstream.Close()

	config := &models.ProxyConfig{Endpoints: map[string]*models.Endpoint{
		"a": {Name: "a", Target: downstream.URL + "/a"},
		"b": {Name: "b", Target: downstream.URL + "/b"},
		"c": {Name: "c", Target: downstream.URL + "/c"},
		"d": {Name: "d", Target: downstream.URL + "/d"},
	}}
	mockService.On("GetConfig").Return(config)

	handler := NewHandler(mockService, logger)
	router := handler.SetupRoutes()

	req := httptest.NewRequest("GET", "/readyz?check_deps=true", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, 3, called)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	checks := response["checks"].(map[string]interface{})
	deps := checks["dependencies"].(map[string]interface{})
	assert.Len(t, deps, 3)
	mockService.AssertExpectations(t)
}
