package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"jq-proxy-service/internal/models"
	"jq-proxy-service/internal/transform"
	"jq-proxy-service/pkg/client"

	"github.com/sirupsen/logrus"
)

// Service implements the ProxyService interface
type Service struct {
	configProvider models.ConfigProvider
	httpClient     client.HTTPClient
	transformer    models.ResponseTransformer
	logger         *logrus.Logger
}

// NewService creates a new proxy service instance
func NewService(configProvider models.ConfigProvider, httpClient client.HTTPClient, transformer models.ResponseTransformer, logger *logrus.Logger) models.ProxyService {
	return &Service{
		configProvider: configProvider,
		httpClient:     httpClient,
		transformer:    transformer,
		logger:         logger,
	}
}

// HandleRequest processes a proxy request and returns the transformed response
func (s *Service) HandleRequest(ctx context.Context, endpointName string, path string, queryParams url.Values, headers http.Header, proxyReq *models.ProxyRequest) (*models.ProxyResponse, error) {
	// Log the incoming request
	s.logger.WithFields(logrus.Fields{
		"endpoint": endpointName,
		"path":     path,
		"method":   proxyReq.Method,
	}).Info("Processing proxy request")

	// Resolve endpoint
	endpoint, exists := s.configProvider.GetEndpoint(endpointName)
	if !exists {
		s.logger.WithField("endpoint", endpointName).Warn("Endpoint not found")
		return nil, &EndpointNotFoundError{
			EndpointName:       endpointName,
			AvailableEndpoints: s.getAvailableEndpoints(),
		}
	}

	// Validate transformation before making the request
	if err := s.validateTransformation(proxyReq); err != nil {
		s.logger.WithError(err).Error("Invalid transformation")
		return nil, &TransformationError{
			Message: fmt.Sprintf("Invalid transformation: %v", err),
			Details: map[string]interface{}{
				"transformation_mode": proxyReq.TransformationMode,
				"jq_query":            proxyReq.JQQuery,
			},
		}
	}

	// Forward request to target endpoint
	response, err := s.forwardRequest(ctx, endpoint, path, queryParams, headers, proxyReq)
	if err != nil {
		s.logger.WithError(err).Error("Failed to forward request")
		return nil, err
	}

	// Parse response body if it's JSON
	var responseData interface{}
	if response.IsJSONResponse() {
		responseData, err = response.ParseJSONBody()
		if err != nil {
			s.logger.WithError(err).Error("Failed to parse JSON response")
			return nil, &UpstreamError{
				Message:    "Failed to parse response from target endpoint",
				StatusCode: response.StatusCode,
				Details: map[string]interface{}{
					"endpoint": endpointName,
					"error":    err.Error(),
				},
			}
		}
	} else {
		// For non-JSON responses, use the raw body as string
		responseData = string(response.Body)
	}

	// Apply transformation using the unified transformer
	var transformedData interface{}
	if unifiedTransformer, ok := s.transformer.(*transform.UnifiedTransformer); ok {
		transformedData, err = unifiedTransformer.TransformRequest(responseData, proxyReq)
	} else {
		return nil, &TransformationError{
			Message: "Invalid transformer type - only jq transformations are supported",
			Details: map[string]interface{}{
				"transformation_mode": proxyReq.TransformationMode,
				"jq_query":            proxyReq.JQQuery,
			},
		}
	}

	if err != nil {
		s.logger.WithError(err).Error("Failed to transform response")
		return nil, &TransformationError{
			Message: fmt.Sprintf("Failed to transform response: %v", err),
			Details: map[string]interface{}{
				"transformation_mode": proxyReq.TransformationMode,
				"jq_query":            proxyReq.JQQuery,
				"error":               err.Error(),
			},
		}
	}

	s.logger.WithFields(logrus.Fields{
		"endpoint":    endpointName,
		"status_code": response.StatusCode,
	}).Info("Successfully processed proxy request")

	return &models.ProxyResponse{
		Data:   transformedData,
		Status: response.StatusCode,
	}, nil
}

// forwardRequest forwards the request to the target endpoint
func (s *Service) forwardRequest(ctx context.Context, endpoint *models.Endpoint, path string, queryParams url.Values, headers http.Header, proxyReq *models.ProxyRequest) (*client.Response, error) {
	// Create a timeout context for the request
	requestCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Forward the request
	response, err := s.httpClient.ForwardRequest(
		requestCtx,
		proxyReq.Method,
		endpoint.Target,
		path,
		queryParams,
		headers,
		proxyReq.Body,
	)

	if err != nil {
		return nil, &UpstreamError{
			Message:    "Failed to connect to target endpoint",
			StatusCode: http.StatusBadGateway,
			Details: map[string]interface{}{
				"endpoint": endpoint.Name,
				"target":   endpoint.Target,
				"error":    err.Error(),
			},
		}
	}

	// Check for HTTP error status codes
	if response.StatusCode >= 400 {
		s.logger.WithFields(logrus.Fields{
			"endpoint":    endpoint.Name,
			"status_code": response.StatusCode,
		}).Warn("Target endpoint returned error status")

		// For 4xx and 5xx errors, we still want to transform the response if possible
		// but we'll preserve the original status code
	}

	return response, nil
}

// validateTransformation validates the transformation rules
func (s *Service) validateTransformation(req *models.ProxyRequest) error {
	// Use the unified transformer's validation if available
	if unifiedTransformer, ok := s.transformer.(*transform.UnifiedTransformer); ok {
		return unifiedTransformer.ValidateTransformation(req)
	}

	return fmt.Errorf("invalid transformer type - only jq transformations are supported")
}

// getAvailableEndpoints returns a list of available endpoint names
func (s *Service) getAvailableEndpoints() []string {
	// Try to load config to get available endpoints
	config, err := s.configProvider.LoadConfig()
	if err != nil || config == nil {
		return []string{}
	}

	endpoints := make([]string, 0, len(config.Endpoints))
	for name := range config.Endpoints {
		endpoints = append(endpoints, name)
	}
	return endpoints
}

// Error types for different failure scenarios

// EndpointNotFoundError represents an error when the requested endpoint is not found
type EndpointNotFoundError struct {
	EndpointName       string
	AvailableEndpoints []string
}

func (e *EndpointNotFoundError) Error() string {
	return fmt.Sprintf("endpoint '%s' not found", e.EndpointName)
}

func (e *EndpointNotFoundError) HTTPStatusCode() int {
	return http.StatusNotFound
}

func (e *EndpointNotFoundError) ErrorCode() string {
	return "ENDPOINT_NOT_FOUND"
}

func (e *EndpointNotFoundError) ErrorDetails() interface{} {
	return map[string]interface{}{
		"available_endpoints": e.AvailableEndpoints,
	}
}

// TransformationError represents an error during response transformation
type TransformationError struct {
	Message string
	Details map[string]interface{}
}

func (e *TransformationError) Error() string {
	return e.Message
}

func (e *TransformationError) HTTPStatusCode() int {
	return http.StatusUnprocessableEntity
}

func (e *TransformationError) ErrorCode() string {
	return "TRANSFORMATION_ERROR"
}

func (e *TransformationError) ErrorDetails() interface{} {
	return e.Details
}

// UpstreamError represents an error from the target endpoint
type UpstreamError struct {
	Message    string
	StatusCode int
	Details    map[string]interface{}
}

func (e *UpstreamError) Error() string {
	return e.Message
}

func (e *UpstreamError) HTTPStatusCode() int {
	if e.StatusCode >= 400 {
		return e.StatusCode
	}
	return http.StatusBadGateway
}

func (e *UpstreamError) ErrorCode() string {
	return "UPSTREAM_ERROR"
}

func (e *UpstreamError) ErrorDetails() interface{} {
	return e.Details
}

// ProxyError interface for structured error handling
type ProxyError interface {
	error
	HTTPStatusCode() int
	ErrorCode() string
	ErrorDetails() interface{}
}
