// Package config provides configuration loading and management functionality.
package config

import (
	"fmt"
	"os"
	"sync"

	"jq-proxy-service/internal/models"
)

// FileProvider implements ConfigProvider for file-based configuration
type FileProvider struct {
	filePath string
	config   *models.ProxyConfig
	mutex    sync.RWMutex
}

// NewFileProvider creates a new file-based configuration provider
func NewFileProvider(filePath string) *FileProvider {
	return &FileProvider{
		filePath: filePath,
	}
}

// LoadConfig loads configuration from the file
func (fp *FileProvider) LoadConfig() (*models.ProxyConfig, error) {
	fp.mutex.Lock()
	defer fp.mutex.Unlock()

	// Check if file exists
	if _, err := os.Stat(fp.filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", fp.filePath)
	}

	// Read file content
	data, err := os.ReadFile(fp.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	// Parse configuration
	config, err := models.ParseProxyConfig(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse configuration: %w", err)
	}

	// Store the loaded configuration
	fp.config = config

	return config, nil
}

// GetEndpoint retrieves an endpoint by name
func (fp *FileProvider) GetEndpoint(name string) (*models.Endpoint, bool) {
	fp.mutex.RLock()
	defer fp.mutex.RUnlock()

	if fp.config == nil {
		return nil, false
	}

	endpoint, exists := fp.config.Endpoints[name]
	return endpoint, exists
}

// Reload reloads the configuration from the file
func (fp *FileProvider) Reload() error {
	_, err := fp.LoadConfig()
	return err
}

// GetConfig returns the current configuration (thread-safe)
func (fp *FileProvider) GetConfig() *models.ProxyConfig {
	fp.mutex.RLock()
	defer fp.mutex.RUnlock()
	return fp.config
}
