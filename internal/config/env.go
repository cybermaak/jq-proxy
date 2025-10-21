package config

import (
	"fmt"
	"os"
	"strconv"

	"jq-proxy-service/internal/models"
)

// EnvProvider implements ConfigProvider with environment variable support
type EnvProvider struct {
	fileProvider *FileProvider
}

// NewEnvProvider creates a new environment-based configuration provider
// It wraps a FileProvider for endpoint configuration while using env vars for server config
func NewEnvProvider(filePath string) *EnvProvider {
	return &EnvProvider{
		fileProvider: NewFileProvider(filePath),
	}
}

// LoadConfig loads configuration from file and overrides server config with environment variables
func (ep *EnvProvider) LoadConfig() (*models.ProxyConfig, error) {
	// Load base configuration from file
	config, err := ep.fileProvider.LoadConfig()
	if err != nil {
		return nil, err
	}

	// Override server configuration with environment variables
	serverConfig, err := ep.loadServerConfigFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to load server config from environment: %w", err)
	}

	// Merge server config
	config.Server = *serverConfig

	return config, nil
}

// loadServerConfigFromEnv loads server configuration from environment variables
func (ep *EnvProvider) loadServerConfigFromEnv() (*models.ServerConfig, error) {
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

// GetEndpoint retrieves an endpoint by name (delegates to file provider)
func (ep *EnvProvider) GetEndpoint(name string) (*models.Endpoint, bool) {
	return ep.fileProvider.GetEndpoint(name)
}

// Reload reloads the configuration
func (ep *EnvProvider) Reload() error {
	_, err := ep.LoadConfig()
	return err
}

// GetConfig returns the current configuration
func (ep *EnvProvider) GetConfig() *models.ProxyConfig {
	// This will return the file provider's config, but we should reload to get env vars
	// For a more robust implementation, we might want to cache the merged config
	config, err := ep.LoadConfig()
	if err != nil {
		return ep.fileProvider.GetConfig() // Fallback to file config
	}
	return config
}