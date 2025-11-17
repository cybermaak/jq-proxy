package proxy

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"jq-proxy-service/internal/client"
	"jq-proxy-service/internal/models"
	"jq-proxy-service/internal/transform"
)

// Mock implementations for testing

type MockConfigProvider struct {
	mock.Mock
}

func (m *MockConfigProvider) LoadConfig() (*models.ProxyConfig, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ProxyConfig), args.Error(1)
}

func (m *MockConfigProvider) GetEndpoint(name string) (*models.Endpoint, bool) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*models.Endpoint), args.Bool(1)
}

func (m *MockConfigProvider) Reload() error {
	args := m.Called()
	return args.Error(0)
}

type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(ctx context.Context, method, targetURL string, headers http.Header, body interface{}) (*client.Response, error) {
	args := m.Called(ctx, method, targetURL, headers, body)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.Response), args.Error(1)
}

func (m *MockHTTPClient) ForwardRequest(ctx context.Context, method, baseURL, path string, queryParams url.Values, headers http.Header, body interface{}) (*client.Response, error) {
	args := m.Called(ctx, method, baseURL, path, queryParams, headers, body)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.Response), args.Error(1)
}

func TestService_HandleRequest_Success(t *testing.T) {
	// Setup mocks
	mockConfig := &MockConfigProvider{}
	mockClient := &MockHTTPClient{}
	transformer := transform.NewUnifiedTransformer()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce log noise in tests

	service := NewService(mockConfig, mockClient, transformer, logger)

	// Test data
	endpoint := &models.Endpoint{
		Name:   "test-service",
		Target: "https://api.example.com",
	}

	proxyReq := &models.ProxyRequest{
		Method:             "GET",
		Body:               nil,
		TransformationMode: models.TransformationModeJQ,
		JQQuery:            "{result: .data}",
	}

	transformedData := map[string]interface{}{
		"result": []interface{}{
			map[string]interface{}{"id": float64(1), "name": "John"},
			map[string]interface{}{"id": float64(2), "name": "Jane"},
		},
	}

	httpResponse := &client.Response{
		StatusCode: 200,
		Headers:    http.Header{"Content-Type": []string{"application/json"}},
		Body:       []byte(`{"data":[{"id":1,"name":"John"},{"id":2,"name":"Jane"}]}`),
	}

	// Setup expectations
	mockConfig.On("GetEndpoint", "test-service").Return(endpoint, true)
	mockClient.On("ForwardRequest", mock.Anything, "GET", "https://api.example.com", "/users", url.Values(nil), http.Header(nil), nil).Return(httpResponse, nil)

	// Execute
	ctx := context.Background()
	result, err := service.HandleRequest(ctx, "test-service", "/users", nil, nil, proxyReq)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 200, result.Status)
	assert.Equal(t, transformedData, result.Data)

	// Verify all expectations were met
	mockConfig.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

func TestService_HandleRequest_EndpointNotFound(t *testing.T) {
	// Setup mocks
	mockConfig := &MockConfigProvider{}
	mockClient := &MockHTTPClient{}
	transformer := transform.NewUnifiedTransformer()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewService(mockConfig, mockClient, transformer, logger)

	proxyReq := &models.ProxyRequest{
		Method:             "GET",
		Body:               nil,
		TransformationMode: models.TransformationModeJQ,
		JQQuery:            "{result: .data}",
	}

	config := &models.ProxyConfig{
		Endpoints: map[string]*models.Endpoint{
			"existing-service": {Name: "existing-service", Target: "https://api.example.com"},
		},
	}

	// Setup expectations
	mockConfig.On("GetEndpoint", "nonexistent-service").Return((*models.Endpoint)(nil), false)
	mockConfig.On("LoadConfig").Return(config, nil)

	// Execute
	ctx := context.Background()
	result, err := service.HandleRequest(ctx, "nonexistent-service", "/users", nil, nil, proxyReq)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	endpointErr, ok := err.(*EndpointNotFoundError)
	require.True(t, ok)
	assert.Equal(t, "nonexistent-service", endpointErr.EndpointName)
	assert.Contains(t, endpointErr.AvailableEndpoints, "existing-service")
	assert.Equal(t, http.StatusNotFound, endpointErr.HTTPStatusCode())
	assert.Equal(t, "ENDPOINT_NOT_FOUND", endpointErr.ErrorCode())

	mockConfig.AssertExpectations(t)
}

func TestService_HandleRequest_InvalidTransformation(t *testing.T) {
	// Setup mocks
	mockConfig := &MockConfigProvider{}
	mockClient := &MockHTTPClient{}
	transformer := transform.NewUnifiedTransformer()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewService(mockConfig, mockClient, transformer, logger)

	endpoint := &models.Endpoint{
		Name:   "test-service",
		Target: "https://api.example.com",
	}

	proxyReq := &models.ProxyRequest{
		Method:             "GET",
		Body:               nil,
		TransformationMode: models.TransformationModeJQ,
		JQQuery:            ".invalid[",
	}

	// Setup expectations
	mockConfig.On("GetEndpoint", "test-service").Return(endpoint, true)

	// Execute
	ctx := context.Background()
	result, err := service.HandleRequest(ctx, "test-service", "/users", nil, nil, proxyReq)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	transformErr, ok := err.(*TransformationError)
	require.True(t, ok)
	assert.Contains(t, transformErr.Message, "Invalid transformation")
	assert.Equal(t, http.StatusUnprocessableEntity, transformErr.HTTPStatusCode())
	assert.Equal(t, "TRANSFORMATION_ERROR", transformErr.ErrorCode())

	mockConfig.AssertExpectations(t)
}

func TestService_HandleRequest_UpstreamError(t *testing.T) {
	// Setup mocks
	mockConfig := &MockConfigProvider{}
	mockClient := &MockHTTPClient{}
	transformer := transform.NewUnifiedTransformer()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewService(mockConfig, mockClient, transformer, logger)

	endpoint := &models.Endpoint{
		Name:   "test-service",
		Target: "https://api.example.com",
	}

	proxyReq := &models.ProxyRequest{
		Method:             "GET",
		Body:               nil,
		TransformationMode: models.TransformationModeJQ,
		JQQuery:            "{result: .data}",
	}

	// Setup expectations
	mockConfig.On("GetEndpoint", "test-service").Return(endpoint, true)
	mockClient.On("ForwardRequest", mock.Anything, "GET", "https://api.example.com", "/users", url.Values(nil), http.Header(nil), nil).Return((*client.Response)(nil), errors.New("connection refused"))

	// Execute
	ctx := context.Background()
	result, err := service.HandleRequest(ctx, "test-service", "/users", nil, nil, proxyReq)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	upstreamErr, ok := err.(*UpstreamError)
	require.True(t, ok)
	assert.Contains(t, upstreamErr.Message, "Failed to connect to target endpoint")
	assert.Equal(t, http.StatusBadGateway, upstreamErr.HTTPStatusCode())
	assert.Equal(t, "UPSTREAM_ERROR", upstreamErr.ErrorCode())

	mockConfig.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

func TestService_HandleRequest_TransformationFailure(t *testing.T) {
	// Setup mocks
	mockConfig := &MockConfigProvider{}
	mockClient := &MockHTTPClient{}
	transformer := transform.NewUnifiedTransformer()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewService(mockConfig, mockClient, transformer, logger)

	endpoint := &models.Endpoint{
		Name:   "test-service",
		Target: "https://api.example.com",
	}

	proxyReq := &models.ProxyRequest{
		Method:             "GET",
		Body:               nil,
		TransformationMode: models.TransformationModeJQ,
		JQQuery:            ".invalid_syntax[",
	}

	// Setup expectations - validation happens before HTTP request
	mockConfig.On("GetEndpoint", "test-service").Return(endpoint, true)

	// Execute
	ctx := context.Background()
	result, err := service.HandleRequest(ctx, "test-service", "/users", nil, nil, proxyReq)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	transformErr, ok := err.(*TransformationError)
	require.True(t, ok)
	assert.Contains(t, transformErr.Message, "Invalid transformation")
	assert.Equal(t, http.StatusUnprocessableEntity, transformErr.HTTPStatusCode())

	mockConfig.AssertExpectations(t)
}

func TestService_HandleRequest_NonJSONResponse(t *testing.T) {
	// Setup mocks
	mockConfig := &MockConfigProvider{}
	mockClient := &MockHTTPClient{}
	transformer := transform.NewUnifiedTransformer()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewService(mockConfig, mockClient, transformer, logger)

	endpoint := &models.Endpoint{
		Name:   "test-service",
		Target: "https://api.example.com",
	}

	proxyReq := &models.ProxyRequest{
		Method:             "GET",
		Body:               nil,
		TransformationMode: models.TransformationModeJQ,
		JQQuery:            "{result: .}",
	}

	transformedData := map[string]interface{}{
		"result": "plain text response",
	}

	httpResponse := &client.Response{
		StatusCode: 200,
		Headers:    http.Header{"Content-Type": []string{"text/plain"}},
		Body:       []byte("plain text response"),
	}

	// Setup expectations
	mockConfig.On("GetEndpoint", "test-service").Return(endpoint, true)
	mockClient.On("ForwardRequest", mock.Anything, "GET", "https://api.example.com", "/data", url.Values(nil), http.Header(nil), nil).Return(httpResponse, nil)

	// Execute
	ctx := context.Background()
	result, err := service.HandleRequest(ctx, "test-service", "/data", nil, nil, proxyReq)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 200, result.Status)
	assert.Equal(t, transformedData, result.Data)

	mockConfig.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

func TestService_HandleRequest_HTTPErrorStatus(t *testing.T) {
	// Setup mocks
	mockConfig := &MockConfigProvider{}
	mockClient := &MockHTTPClient{}
	transformer := transform.NewUnifiedTransformer()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewService(mockConfig, mockClient, transformer, logger)

	endpoint := &models.Endpoint{
		Name:   "test-service",
		Target: "https://api.example.com",
	}

	proxyReq := &models.ProxyRequest{
		Method:             "GET",
		Body:               nil,
		TransformationMode: models.TransformationModeJQ,
		JQQuery:            "{error: .message}",
	}

	transformedData := map[string]interface{}{
		"error": "Not found",
	}

	httpResponse := &client.Response{
		StatusCode: 404,
		Headers:    http.Header{"Content-Type": []string{"application/json"}},
		Body:       []byte(`{"message":"Not found"}`),
	}

	// Setup expectations
	mockConfig.On("GetEndpoint", "test-service").Return(endpoint, true)
	mockClient.On("ForwardRequest", mock.Anything, "GET", "https://api.example.com", "/users/999", url.Values(nil), http.Header(nil), nil).Return(httpResponse, nil)

	// Execute
	ctx := context.Background()
	result, err := service.HandleRequest(ctx, "test-service", "/users/999", nil, nil, proxyReq)

	// Assert - should still succeed but preserve the error status code
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 404, result.Status)
	assert.Equal(t, transformedData, result.Data)

	mockConfig.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}

func TestService_HandleRequest_WithQueryParamsAndHeaders(t *testing.T) {
	// Setup mocks
	mockConfig := &MockConfigProvider{}
	mockClient := &MockHTTPClient{}
	transformer := transform.NewUnifiedTransformer()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := NewService(mockConfig, mockClient, transformer, logger)

	endpoint := &models.Endpoint{
		Name:   "test-service",
		Target: "https://api.example.com",
	}

	proxyReq := &models.ProxyRequest{
		Method:             "POST",
		Body:               map[string]interface{}{"name": "John"},
		TransformationMode: models.TransformationModeJQ,
		JQQuery:            "{id: .id}",
	}

	queryParams := url.Values{
		"limit": []string{"10"},
		"page":  []string{"1"},
	}

	headers := http.Header{
		"Authorization": []string{"Bearer token"},
		"Content-Type":  []string{"application/json"},
	}

	transformedData := map[string]interface{}{
		"id": float64(123),
	}

	httpResponse := &client.Response{
		StatusCode: 201,
		Headers:    http.Header{"Content-Type": []string{"application/json"}},
		Body:       []byte(`{"id":123,"name":"John"}`),
	}

	// Setup expectations
	mockConfig.On("GetEndpoint", "test-service").Return(endpoint, true)
	mockClient.On("ForwardRequest", mock.Anything, "POST", "https://api.example.com", "/users", queryParams, headers, map[string]interface{}{"name": "John"}).Return(httpResponse, nil)

	// Execute
	ctx := context.Background()
	result, err := service.HandleRequest(ctx, "test-service", "/users", queryParams, headers, proxyReq)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 201, result.Status)
	assert.Equal(t, transformedData, result.Data)

	mockConfig.AssertExpectations(t)
	mockClient.AssertExpectations(t)
}
