package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvProvider_LoadConfig(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := ioutil.TempDir("", "config_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test config file
	configData := `{
		"server": {
			"port": 8080,
			"read_timeout": 30,
			"write_timeout": 30
		},
		"endpoints": {
			"service1": {
				"name": "service1",
				"target": "https://api1.example.com"
			}
		}
	}`

	configFile := filepath.Join(tempDir, "config.json")
	err = ioutil.WriteFile(configFile, []byte(configData), 0644)
	require.NoError(t, err)

	tests := []struct {
		name           string
		envVars        map[string]string
		expectedPort   int
		expectedRead   int
		expectedWrite  int
	}{
		{
			name:           "no environment variables - use defaults",
			envVars:        map[string]string{},
			expectedPort:   8080, // Default from env provider
			expectedRead:   30,   // Default from env provider
			expectedWrite:  30,   // Default from env provider
		},
		{
			name: "override port only",
			envVars: map[string]string{
				"PROXY_PORT": "9090",
			},
			expectedPort:   9090,
			expectedRead:   30,
			expectedWrite:  30,
		},
		{
			name: "override all server settings",
			envVars: map[string]string{
				"PROXY_PORT":          "3000",
				"PROXY_READ_TIMEOUT":  "60",
				"PROXY_WRITE_TIMEOUT": "45",
			},
			expectedPort:   3000,
			expectedRead:   60,
			expectedWrite:  45,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables
			os.Unsetenv("PROXY_PORT")
			os.Unsetenv("PROXY_READ_TIMEOUT")
			os.Unsetenv("PROXY_WRITE_TIMEOUT")

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			// Create provider and load config
			provider := NewEnvProvider(configFile)
			config, err := provider.LoadConfig()

			require.NoError(t, err)
			assert.NotNil(t, config)

			// Check server configuration
			assert.Equal(t, tt.expectedPort, config.Server.Port)
			assert.Equal(t, tt.expectedRead, config.Server.ReadTimeout)
			assert.Equal(t, tt.expectedWrite, config.Server.WriteTimeout)

			// Check that endpoints are still loaded from file
			assert.Contains(t, config.Endpoints, "service1")
			assert.Equal(t, "https://api1.example.com", config.Endpoints["service1"].Target)
		})
	}
}

func TestEnvProvider_LoadConfig_InvalidEnvVars(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := ioutil.TempDir("", "config_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test config file
	configData := `{
		"server": {
			"port": 8080,
			"read_timeout": 30,
			"write_timeout": 30
		},
		"endpoints": {
			"service1": {
				"name": "service1",
				"target": "https://api1.example.com"
			}
		}
	}`

	configFile := filepath.Join(tempDir, "config.json")
	err = ioutil.WriteFile(configFile, []byte(configData), 0644)
	require.NoError(t, err)

	tests := []struct {
		name     string
		envVars  map[string]string
		errorMsg string
	}{
		{
			name: "invalid port",
			envVars: map[string]string{
				"PROXY_PORT": "invalid",
			},
			errorMsg: "invalid PROXY_PORT value",
		},
		{
			name: "invalid read timeout",
			envVars: map[string]string{
				"PROXY_READ_TIMEOUT": "invalid",
			},
			errorMsg: "invalid PROXY_READ_TIMEOUT value",
		},
		{
			name: "invalid write timeout",
			envVars: map[string]string{
				"PROXY_WRITE_TIMEOUT": "invalid",
			},
			errorMsg: "invalid PROXY_WRITE_TIMEOUT value",
		},
		{
			name: "port out of range",
			envVars: map[string]string{
				"PROXY_PORT": "70000",
			},
			errorMsg: "port must be between 1 and 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables
			os.Unsetenv("PROXY_PORT")
			os.Unsetenv("PROXY_READ_TIMEOUT")
			os.Unsetenv("PROXY_WRITE_TIMEOUT")

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			// Create provider and load config
			provider := NewEnvProvider(configFile)
			config, err := provider.LoadConfig()

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMsg)
			assert.Nil(t, config)
		})
	}
}

func TestEnvProvider_GetEndpoint(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := ioutil.TempDir("", "config_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test config file
	configData := `{
		"server": {
			"port": 8080,
			"read_timeout": 30,
			"write_timeout": 30
		},
		"endpoints": {
			"service1": {
				"name": "service1",
				"target": "https://api1.example.com"
			},
			"service2": {
				"name": "service2",
				"target": "http://api2.example.com"
			}
		}
	}`

	configFile := filepath.Join(tempDir, "config.json")
	err = ioutil.WriteFile(configFile, []byte(configData), 0644)
	require.NoError(t, err)

	provider := NewEnvProvider(configFile)
	_, err = provider.LoadConfig()
	require.NoError(t, err)

	// Test getting existing endpoint
	endpoint, found := provider.GetEndpoint("service1")
	assert.True(t, found)
	assert.NotNil(t, endpoint)
	assert.Equal(t, "https://api1.example.com", endpoint.Target)

	// Test getting non-existing endpoint
	endpoint, found = provider.GetEndpoint("nonexistent")
	assert.False(t, found)
	assert.Nil(t, endpoint)
}

func TestEnvProvider_FileNotFound(t *testing.T) {
	provider := NewEnvProvider("nonexistent.json")
	config, err := provider.LoadConfig()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuration file not found")
	assert.Nil(t, config)
}