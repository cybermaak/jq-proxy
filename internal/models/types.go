// Package models defines the core data structures and types used throughout the proxy service.
package models

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// ConfigProvider defines the interface for configuration management
type ConfigProvider interface {
	LoadConfig() (*ProxyConfig, error)
	GetEndpoint(name string) (*Endpoint, bool)
	Reload() error
}

// ProxyService defines the interface for the core proxy functionality
type ProxyService interface {
	HandleRequest(ctx context.Context, endpointName string, path string,
		queryParams url.Values, headers http.Header,
		proxyReq *ProxyRequest) (*ProxyResponse, error)
	GetConfig() *ProxyConfig
}

// ProxyConfig represents the complete service configuration
type ProxyConfig struct {
	Endpoints map[string]*Endpoint `json:"endpoints"`
	Server    ServerConfig         `json:"server"`
}

// Endpoint represents a target endpoint configuration
type Endpoint struct {
	Name   string `json:"name"`
	Target string `json:"target"`
}

// ServerConfig represents server-specific configuration
type ServerConfig struct {
	Port         int `json:"port"`
	ReadTimeout  int `json:"read_timeout"`
	WriteTimeout int `json:"write_timeout"`
}

// TransformationMode represents the type of transformation to apply
type TransformationMode string

const (
	TransformationModeJQ TransformationMode = "jq"
)

// ProxyRequest represents the incoming request payload
type ProxyRequest struct {
	Method             string             `json:"method"`
	Body               interface{}        `json:"body"`
	TransformationMode TransformationMode `json:"transformation_mode,omitempty"`
	JQQuery            string             `json:"jq_query,omitempty"`
}

// ProxyResponse represents the response returned to the client
type ProxyResponse struct {
	Data   interface{} `json:"data"`
	Status int         `json:"status"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Validate validates the ProxyRequest
func (pr *ProxyRequest) Validate() error {
	if pr.Method == "" {
		return fmt.Errorf("method is required")
	}

	// Validate HTTP method
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	methodValid := false
	upperMethod := strings.ToUpper(pr.Method)
	for _, method := range validMethods {
		if upperMethod == method {
			methodValid = true
			break
		}
	}
	if !methodValid {
		return fmt.Errorf("invalid HTTP method: %s", pr.Method)
	}

	// Set default transformation mode if not specified
	if pr.TransformationMode == "" {
		pr.TransformationMode = TransformationModeJQ
	}

	// Validate transformation mode
	if pr.TransformationMode != TransformationModeJQ {
		return fmt.Errorf("invalid transformation mode: %s. Must be 'jq'", pr.TransformationMode)
	}

	// Validate jq query is provided
	if pr.JQQuery == "" {
		return fmt.Errorf("jq_query is required")
	}

	return nil
}

// Validate validates the ProxyConfig
func (pc *ProxyConfig) Validate() error {
	if len(pc.Endpoints) == 0 {
		return fmt.Errorf("at least one endpoint must be configured")
	}

	// Check for duplicate endpoint names
	names := make(map[string]bool)
	for name, endpoint := range pc.Endpoints {
		if names[name] {
			return fmt.Errorf("duplicate endpoint name: %s", name)
		}
		names[name] = true

		if err := endpoint.Validate(); err != nil {
			return fmt.Errorf("invalid endpoint %s: %w", name, err)
		}
	}

	if err := pc.Server.Validate(); err != nil {
		return fmt.Errorf("invalid server configuration: %w", err)
	}

	return nil
}

// Validate validates the Endpoint
func (e *Endpoint) Validate() error {
	if e.Name == "" {
		return fmt.Errorf("endpoint name is required")
	}

	if e.Target == "" {
		return fmt.Errorf("endpoint target is required")
	}

	// Basic URL validation
	if !strings.HasPrefix(e.Target, "http://") && !strings.HasPrefix(e.Target, "https://") {
		return fmt.Errorf("endpoint target must be a valid HTTP/HTTPS URL")
	}

	return nil
}

// Validate validates the ServerConfig
func (sc *ServerConfig) Validate() error {
	if sc.Port <= 0 || sc.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	if sc.ReadTimeout < 0 {
		return fmt.Errorf("read timeout must be non-negative")
	}

	if sc.WriteTimeout < 0 {
		return fmt.Errorf("write timeout must be non-negative")
	}

	return nil
}

// ParseProxyRequest parses JSON data into a ProxyRequest
func ParseProxyRequest(data []byte) (*ProxyRequest, error) {
	var req ProxyRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &req, nil
}

// ParseProxyConfig parses JSON data into a ProxyConfig
func ParseProxyConfig(data []byte) (*ProxyConfig, error) {
	var config ProxyConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &config, nil
}
