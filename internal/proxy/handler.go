// Package proxy implements the HTTP proxy service with request handling and routing.
package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"jq-proxy-service/internal/logging"
	"jq-proxy-service/internal/models"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler handles HTTP requests for the proxy service
type Handler struct {
	proxyService models.ProxyService
	logger       *logging.Logger
}

// NewHandler creates a new HTTP handler
func NewHandler(proxyService models.ProxyService, logger *logging.Logger) *Handler {
	return &Handler{
		proxyService: proxyService,
		logger:       logger,
	}
}

// SetupRoutes configures the HTTP routes
func (h *Handler) SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/health", h.healthCheck).Methods("GET")

	// Metrics endpoint
	router.HandleFunc("/metrics", h.metricsHandler).Methods("GET")

	// Config endpoint
	router.HandleFunc("/config", h.configHandler).Methods("GET")

	// Main proxy endpoint - captures endpoint name and remaining path
	router.HandleFunc("/proxy/{endpoint}/{path:.*}", h.handleProxyRequest).Methods("POST", "OPTIONS")
	router.HandleFunc("/proxy/{endpoint}", h.handleProxyRequest).Methods("POST", "OPTIONS")

	// Add middleware
	router.Use(logging.RequestLoggingMiddleware(h.logger))
	router.Use(h.corsMiddleware)

	return router
}

// handleProxyRequest handles the main proxy requests
func (h *Handler) handleProxyRequest(w http.ResponseWriter, r *http.Request) {
	// Handle OPTIONS requests for CORS
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Extract endpoint name from URL
	vars := mux.Vars(r)
	endpointName := vars["endpoint"]
	path := vars["path"]

	// Add leading slash to path if it doesn't have one
	if path != "" && !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	h.logger.WithContext(r.Context()).WithFields(logrus.Fields{
		"endpoint": endpointName,
		"path":     path,
		"method":   r.Method,
	}).Debug("Processing proxy request")

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.WithContext(r.Context()).WithError(err).Error("Failed to read request body")
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to read request body", nil)
		return
	}

	// Parse proxy request
	proxyReq, err := models.ParseProxyRequest(body)
	if err != nil {
		h.logger.WithContext(r.Context()).WithError(err).Error("Failed to parse proxy request")
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", fmt.Sprintf("Invalid request format: %v", err), nil)
		return
	}

	// Process the proxy request
	response, err := h.proxyService.HandleRequest(
		r.Context(),
		endpointName,
		path,
		r.URL.Query(),
		r.Header,
		proxyReq,
	)

	if err != nil {
		h.handleProxyError(w, err)
		return
	}

	// Write successful response
	h.writeJSONResponse(w, response.Status, response.Data)
}

// healthCheck provides a simple health check endpoint
func (h *Handler) healthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":  "healthy",
		"service": "jq-proxy-service",
	}
	h.writeJSONResponse(w, http.StatusOK, response)
}

// metricsHandler provides metrics endpoint
func (h *Handler) metricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics := h.logger.GetMetrics().GetMetrics()
	h.writeJSONResponse(w, http.StatusOK, metrics)
}

// configHandler provides current configuration endpoint
func (h *Handler) configHandler(w http.ResponseWriter, r *http.Request) {
	// Get the service's config provider
	config := h.proxyService.GetConfig()
	if config == nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "CONFIG_ERROR", "Configuration not available", nil)
		return
	}

	// Create a safe response that doesn't expose sensitive information
	response := map[string]interface{}{
		"server": map[string]interface{}{
			"port":          config.Server.Port,
			"read_timeout":  config.Server.ReadTimeout,
			"write_timeout": config.Server.WriteTimeout,
		},
		"endpoints": make(map[string]interface{}),
	}

	// Add endpoint information
	for name, endpoint := range config.Endpoints {
		response["endpoints"].(map[string]interface{})[name] = map[string]interface{}{
			"name":   endpoint.Name,
			"target": endpoint.Target,
		}
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// handleProxyError handles different types of proxy errors
func (h *Handler) handleProxyError(w http.ResponseWriter, err error) {
	if proxyErr, ok := err.(ProxyError); ok {
		h.writeErrorResponse(
			w,
			proxyErr.HTTPStatusCode(),
			proxyErr.ErrorCode(),
			proxyErr.Error(),
			proxyErr.ErrorDetails(),
		)
	} else {
		// Generic error
		h.logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Unexpected error in proxy request")
		h.writeErrorResponse(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"An unexpected error occurred",
			nil,
		)
	}
}

// writeJSONResponse writes a JSON response
func (h *Handler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
	}
}

// writeErrorResponse writes a standardized error response
func (h *Handler) writeErrorResponse(w http.ResponseWriter, statusCode int, code, message string, details interface{}) {
	errorResponse := models.ErrorResponse{
		Error: models.ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	}

	h.writeJSONResponse(w, statusCode, errorResponse)
}

// corsMiddleware adds CORS headers
func (h *Handler) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
