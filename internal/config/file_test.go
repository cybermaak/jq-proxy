package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileProvider_LoadConfig(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := ioutil.TempDir("", "config_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name        string
		configData  string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid configuration",
			configData: `{
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
			}`,
			expectError: false,
		},
		{
			name: "invalid JSON",
			configData: `{
				"server": {
					"port": 8080,
				}
			}`,
			expectError: true,
			errorMsg:    "failed to parse configuration",
		},
		{
			name: "invalid configuration - missing endpoints",
			configData: `{
				"server": {
					"port": 8080,
					"read_timeout": 30,
					"write_timeout": 30
				},
				"endpoints": {}
			}`,
			expectError: true,
			errorMsg:    "failed to parse configuration",
		},
		{
			name: "invalid configuration - invalid port",
			configData: `{
				"server": {
					"port": 0,
					"read_timeout": 30,
					"write_timeout": 30
				},
				"endpoints": {
					"service1": {
						"name": "service1",
						"target": "https://api1.example.com"
					}
				}
			}`,
			expectError: true,
			errorMsg:    "failed to parse configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			configFile := filepath.Join(tempDir, "config.json")
			err := ioutil.WriteFile(configFile, []byte(tt.configData), 0644)
			require.NoError(t, err)

			// Create provider and load config
			provider := NewFileProvider(configFile)
			config, err := provider.LoadConfig()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
				assert.Equal(t, 8080, config.Server.Port)
				assert.Contains(t, config.Endpoints, "service1")
			}
		})
	}
}

func TestFileProvider_LoadConfig_FileNotFound(t *testing.T) {
	provider := NewFileProvider("nonexistent.json")
	config, err := provider.LoadConfig()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuration file not found")
	assert.Nil(t, config)
}

func TestFileProvider_GetEndpoint(t *testing.T) {
	// Create a temporary config file
	tempDir, err := ioutil.TempDir("", "config_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

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

	provider := NewFileProvider(configFile)
	_, err = provider.LoadConfig()
	require.NoError(t, err)

	tests := []struct {
		name         string
		endpointName string
		expectFound  bool
		expectedURL  string
	}{
		{
			name:         "existing endpoint",
			endpointName: "service1",
			expectFound:  true,
			expectedURL:  "https://api1.example.com",
		},
		{
			name:         "another existing endpoint",
			endpointName: "service2",
			expectFound:  true,
			expectedURL:  "http://api2.example.com",
		},
		{
			name:         "non-existing endpoint",
			endpointName: "service3",
			expectFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpoint, found := provider.GetEndpoint(tt.endpointName)

			assert.Equal(t, tt.expectFound, found)
			if tt.expectFound {
				assert.NotNil(t, endpoint)
				assert.Equal(t, tt.expectedURL, endpoint.Target)
				assert.Equal(t, tt.endpointName, endpoint.Name)
			} else {
				assert.Nil(t, endpoint)
			}
		})
	}
}

func TestFileProvider_GetEndpoint_NoConfigLoaded(t *testing.T) {
	provider := NewFileProvider("nonexistent.json")
	endpoint, found := provider.GetEndpoint("service1")

	assert.False(t, found)
	assert.Nil(t, endpoint)
}

func TestFileProvider_Reload(t *testing.T) {
	// Create a temporary config file
	tempDir, err := ioutil.TempDir("", "config_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	initialConfig := `{
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

	updatedConfig := `{
		"server": {
			"port": 9090,
			"read_timeout": 60,
			"write_timeout": 60
		},
		"endpoints": {
			"service1": {
				"name": "service1",
				"target": "https://api1.example.com"
			},
			"service2": {
				"name": "service2",
				"target": "https://api2.example.com"
			}
		}
	}`

	configFile := filepath.Join(tempDir, "config.json")
	err = ioutil.WriteFile(configFile, []byte(initialConfig), 0644)
	require.NoError(t, err)

	provider := NewFileProvider(configFile)

	// Load initial config
	config, err := provider.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 8080, config.Server.Port)
	assert.Len(t, config.Endpoints, 1)

	// Update the file
	err = ioutil.WriteFile(configFile, []byte(updatedConfig), 0644)
	require.NoError(t, err)

	// Reload config
	err = provider.Reload()
	require.NoError(t, err)

	// Verify updated config
	config = provider.GetConfig()
	assert.Equal(t, 9090, config.Server.Port)
	assert.Len(t, config.Endpoints, 2)
	assert.Contains(t, config.Endpoints, "service2")
}

func TestFileProvider_ThreadSafety(t *testing.T) {
	// Create a temporary config file
	tempDir, err := ioutil.TempDir("", "config_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

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

	provider := NewFileProvider(configFile)
	_, err = provider.LoadConfig()
	require.NoError(t, err)

	// Test concurrent access
	done := make(chan bool, 10)

	// Start multiple goroutines reading endpoints
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_, _ = provider.GetEndpoint("service1")
			}
			done <- true
		}()
	}

	// Start multiple goroutines reloading config
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				_ = provider.Reload()
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify final state
	endpoint, found := provider.GetEndpoint("service1")
	assert.True(t, found)
	assert.NotNil(t, endpoint)
}
