// Package config provides configuration loading and management functionality.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"jq-proxy-service/internal/models"
)

// FullEnvProvider implements ConfigProvider loading everything from environment variables
type FullEnvProvider struct {
	mu     sync.RWMutex
	config *models.ProxyConfig
}

// NewFullEnvProvider creates a provider that loads all config from environment variables
func NewFullEnvProvider() *FullEnvProvider {
	return &FullEnvProvider{}
}

// LoadConfig loads configuration entirely from environment variables
func (fep *FullEnvProvider) LoadConfig() (*models.ProxyConfig, error) {
	fep.mu.Lock()
	defer fep.mu.Unlock()

	// Load server configuration
	serverConfig, err := loadServerConfigFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to load server config from environment: %w", err)
	}

	// Load endpoints from environment
	endpoints, err := loadEndpointsFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to load endpoints from environment: %w", err)
	}

	config := &models.ProxyConfig{
		Server:    *serverConfig,
		Endpoints: endpoints,
	}

	// Validate the configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	fep.config = config
	return config, nil
}

// loadServerConfigFromEnv loads server configuration from environment variables
func loadServerConfigFromEnv() (*models.ServerConfig, error) {
	config := &models.ServerConfig{
		Port:         8080, // Default port
		ReadTimeout:  30,   // Default read timeout in seconds
		WriteTimeout: 30,   // Default write timeout in seconds
	}

	// Load port from environment
	if portStr := os.Getenv("PROXY_PORT"); portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid PROXY_PORT value: %s", portStr)
		}
		config.Port = port
	}

	// Load read timeout from environment
	if timeoutStr := os.Getenv("PROXY_READ_TIMEOUT"); timeoutStr != "" {
		timeout, err := strconv.Atoi(timeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid PROXY_READ_TIMEOUT value: %s", timeoutStr)
		}
		config.ReadTimeout = timeout
	}

	// Load write timeout from environment
	if timeoutStr := os.Getenv("PROXY_WRITE_TIMEOUT"); timeoutStr != "" {
		timeout, err := strconv.Atoi(timeoutStr)
		if err != nil {
			return nil, fmt.Errorf("invalid PROXY_WRITE_TIMEOUT value: %s", timeoutStr)
		}
		config.WriteTimeout = timeout
	}

	// Validate the configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// loadEndpointsFromEnv loads endpoint configurations from environment variables
// Supports two formats:
//  1. PROXY_ENDPOINTS_JSON - JSON string with all endpoints
//  2. PROXY_ENDPOINT_{KEY}_TARGET and PROXY_ENDPOINT_{KEY}_NAME - Individual endpoint configuration
//     where {KEY} is used as the map key for the endpoint
func loadEndpointsFromEnv() (map[string]*models.Endpoint, error) {
	endpoints := make(map[string]*models.Endpoint)

	// Try loading from JSON first
	if endpointsJSON := os.Getenv("PROXY_ENDPOINTS_JSON"); endpointsJSON != "" {
		var jsonEndpoints map[string]*models.Endpoint
		if err := json.Unmarshal([]byte(endpointsJSON), &jsonEndpoints); err != nil {
			return nil, fmt.Errorf("invalid PROXY_ENDPOINTS_JSON: %w", err)
		}
		endpoints = jsonEndpoints
	}

	// Load individual endpoints in a single loop (these override JSON if both are present)
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, "PROXY_ENDPOINT_") || !strings.Contains(env, "_TARGET=") {
			continue
		}

		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		// Extract key from PROXY_ENDPOINT_{KEY}_TARGET
		varName := parts[0]
		key := strings.TrimPrefix(varName, "PROXY_ENDPOINT_")
		key = strings.TrimSuffix(key, "_TARGET")
		if key == "" {
			continue
		}

		target := parts[1]
		if target == "" {
			continue
		}

		// Get the name from PROXY_ENDPOINT_{KEY}_NAME
		nameVar := fmt.Sprintf("PROXY_ENDPOINT_%s_NAME", key)
		name := os.Getenv(nameVar)

		// Determine the map key and endpoint name
		var mapKey string
		if name != "" {
			// If name is provided, use it as both the map key and endpoint name
			mapKey = name
		} else {
			// If name is not provided, use the variable key as map key
			// and convert it to lowercase with hyphens for the endpoint name
			mapKey = key
			name = strings.ToLower(key)
			name = strings.ReplaceAll(name, "_", "-")
		}

		endpoint := &models.Endpoint{
			Name:   name,
			Target: target,
		}

		// Validate the endpoint
		if err := endpoint.Validate(); err != nil {
			return nil, fmt.Errorf("invalid endpoint %s: %w", mapKey, err)
		}

		endpoints[mapKey] = endpoint
	}

	if len(endpoints) == 0 {
		return nil, fmt.Errorf("no endpoints configured (use PROXY_ENDPOINTS_JSON or PROXY_ENDPOINT_{KEY}_TARGET)")
	}

	return endpoints, nil
}

// GetEndpoint retrieves an endpoint by name
func (fep *FullEnvProvider) GetEndpoint(name string) (*models.Endpoint, bool) {
	fep.mu.RLock()
	defer fep.mu.RUnlock()

	if fep.config == nil {
		return nil, false
	}

	endpoint, exists := fep.config.Endpoints[name]
	return endpoint, exists
}

// Reload reloads the configuration from environment variables
func (fep *FullEnvProvider) Reload() error {
	_, err := fep.LoadConfig()
	return err
}

// GetConfig returns the current configuration
func (fep *FullEnvProvider) GetConfig() *models.ProxyConfig {
	fep.mu.RLock()
	defer fep.mu.RUnlock()
	return fep.config
}
