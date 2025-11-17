package config

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFullEnvProvider_LoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
	}{
		{
			name: "valid configuration with individual endpoints",
			envVars: map[string]string{
				"PROXY_PORT":                         "9000",
				"PROXY_READ_TIMEOUT":                 "60",
				"PROXY_WRITE_TIMEOUT":                "60",
				"PROXY_ENDPOINT_USER_SERVICE_TARGET": "https://api.example.com/users",
				"PROXY_ENDPOINT_POST_SERVICE_TARGET": "https://api.example.com/posts",
			},
			wantErr: false,
		},
		{
			name: "valid configuration with JSON endpoints",
			envVars: map[string]string{
				"PROXY_PORT":           "8080",
				"PROXY_ENDPOINTS_JSON": `{"api":{"name":"api","target":"https://api.example.com"}}`,
			},
			wantErr: false,
		},
		{
			name: "no endpoints configured",
			envVars: map[string]string{
				"PROXY_PORT": "8080",
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			envVars: map[string]string{
				"PROXY_PORT":                "invalid",
				"PROXY_ENDPOINT_API_TARGET": "https://api.example.com",
			},
			wantErr: true,
		},
		{
			name: "invalid endpoint target",
			envVars: map[string]string{
				"PROXY_PORT":                "8080",
				"PROXY_ENDPOINT_API_TARGET": "not-a-url",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			clearEnv()

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer clearEnv()

			provider := NewFullEnvProvider()
			config, err := provider.LoadConfig()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, config)
			assert.NotEmpty(t, config.Endpoints)
		})
	}
}

func TestFullEnvProvider_GetEndpoint(t *testing.T) {
	clearEnv()
	os.Setenv("PROXY_PORT", "8080")
	os.Setenv("PROXY_ENDPOINT_USER_SERVICE_TARGET", "https://api.example.com/users")
	os.Setenv("PROXY_ENDPOINT_USER_SERVICE_NAME", "user-api")
	defer clearEnv()

	provider := NewFullEnvProvider()
	_, err := provider.LoadConfig()
	require.NoError(t, err)

	// The key is USER_SERVICE, not the name
	endpoint, exists := provider.GetEndpoint("USER_SERVICE")
	assert.True(t, exists)
	assert.NotNil(t, endpoint)
	assert.Equal(t, "user-api", endpoint.Name)
	assert.Equal(t, "https://api.example.com/users", endpoint.Target)

	_, exists = provider.GetEndpoint("nonexistent")
	assert.False(t, exists)
}

func TestLoadEndpointsFromEnv_IndividualEndpoints(t *testing.T) {
	clearEnv()
	os.Setenv("PROXY_ENDPOINT_USER_API_TARGET", "https://users.example.com")
	os.Setenv("PROXY_ENDPOINT_USER_API_NAME", "user-service")
	os.Setenv("PROXY_ENDPOINT_POST_API_TARGET", "https://posts.example.com")
	os.Setenv("PROXY_ENDPOINT_POST_API_NAME", "post-service")
	defer clearEnv()

	endpoints, err := loadEndpointsFromEnv()
	require.NoError(t, err)
	assert.Len(t, endpoints, 2)
	assert.Contains(t, endpoints, "USER_API")
	assert.Contains(t, endpoints, "POST_API")
	assert.Equal(t, "user-service", endpoints["USER_API"].Name)
	assert.Equal(t, "post-service", endpoints["POST_API"].Name)
}

func TestLoadEndpointsFromEnv_JSONFormat(t *testing.T) {
	clearEnv()
	jsonEndpoints := `{
		"api1": {"name": "api1", "target": "https://api1.example.com"},
		"api2": {"name": "api2", "target": "https://api2.example.com"}
	}`
	os.Setenv("PROXY_ENDPOINTS_JSON", jsonEndpoints)
	defer clearEnv()

	endpoints, err := loadEndpointsFromEnv()
	require.NoError(t, err)
	assert.Len(t, endpoints, 2)
	assert.Contains(t, endpoints, "api1")
	assert.Contains(t, endpoints, "api2")
}

func TestLoadEndpointsFromEnv_MixedFormats(t *testing.T) {
	clearEnv()
	// Individual endpoints override JSON
	jsonEndpoints := `{"API1": {"name": "api1", "target": "https://old.example.com"}}`
	os.Setenv("PROXY_ENDPOINTS_JSON", jsonEndpoints)
	os.Setenv("PROXY_ENDPOINT_API1_TARGET", "https://new.example.com")
	os.Setenv("PROXY_ENDPOINT_API1_NAME", "api-one")
	defer clearEnv()

	endpoints, err := loadEndpointsFromEnv()
	require.NoError(t, err)
	assert.Len(t, endpoints, 1)
	assert.Equal(t, "https://new.example.com", endpoints["API1"].Target)
	assert.Equal(t, "api-one", endpoints["API1"].Name)
}

func TestLoadEndpointsFromEnv_NameDefaultsToKey(t *testing.T) {
	clearEnv()
	// If NAME is not provided, it should default to the key (converted to lowercase with hyphens)
	os.Setenv("PROXY_ENDPOINT_MY_API_TARGET", "https://api.example.com")
	defer clearEnv()

	endpoints, err := loadEndpointsFromEnv()
	require.NoError(t, err)
	assert.Len(t, endpoints, 1)
	assert.Contains(t, endpoints, "MY_API")
	assert.Equal(t, "my-api", endpoints["MY_API"].Name)
	assert.Equal(t, "https://api.example.com", endpoints["MY_API"].Target)
}

func clearEnv() {
	os.Unsetenv("PROXY_PORT")
	os.Unsetenv("PROXY_READ_TIMEOUT")
	os.Unsetenv("PROXY_WRITE_TIMEOUT")
	os.Unsetenv("PROXY_ENDPOINTS_JSON")

	// Clear all PROXY_ENDPOINT_* variables
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) > 0 && strings.HasPrefix(parts[0], "PROXY_ENDPOINT_") {
			os.Unsetenv(parts[0])
		}
	}
}
