package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProxyRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request ProxyRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: ProxyRequest{
				Method:             "GET",
				Body:               nil,
				TransformationMode: TransformationModeJQ,
				JQQuery:            "{result: .data}",
			},
			wantErr: false,
		},
		{
			name: "missing method",
			request: ProxyRequest{
				Body:               nil,
				TransformationMode: TransformationModeJQ,
				JQQuery:            "{result: .data}",
			},
			wantErr: true,
			errMsg:  "method is required",
		},
		{
			name: "invalid method",
			request: ProxyRequest{
				Method:             "INVALID",
				Body:               nil,
				TransformationMode: TransformationModeJQ,
				JQQuery:            "{result: .data}",
			},
			wantErr: true,
			errMsg:  "invalid HTTP method: INVALID",
		},
		{
			name: "missing jq_query",
			request: ProxyRequest{
				Method: "GET",
				Body:   nil,
			},
			wantErr: true,
			errMsg:  "jq_query is required",
		},
		{
			name: "case insensitive method",
			request: ProxyRequest{
				Method:             "post",
				Body:               map[string]interface{}{"key": "value"},
				TransformationMode: TransformationModeJQ,
				JQQuery:            "{result: .data}",
			},
			wantErr: false,
		},
		{
			name: "jq mode with valid query",
			request: ProxyRequest{
				Method:             "GET",
				Body:               nil,
				TransformationMode: TransformationModeJQ,
				JQQuery:            ".data | map(.name)",
			},
			wantErr: false,
		},
		{
			name: "jq mode without query",
			request: ProxyRequest{
				Method:             "GET",
				Body:               nil,
				TransformationMode: TransformationModeJQ,
			},
			wantErr: true,
			errMsg:  "jq_query is required",
		},
		{
			name: "invalid transformation mode",
			request: ProxyRequest{
				Method:             "GET",
				Body:               nil,
				TransformationMode: "invalid",
				JQQuery:            "{result: .data}",
			},
			wantErr: true,
			errMsg:  "invalid transformation mode: invalid. Must be 'jq'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEndpoint_Validate(t *testing.T) {
	tests := []struct {
		name     string
		endpoint Endpoint
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid endpoint",
			endpoint: Endpoint{
				Name:   "test-service",
				Target: "https://api.example.com",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			endpoint: Endpoint{
				Target: "https://api.example.com",
			},
			wantErr: true,
			errMsg:  "endpoint name is required",
		},
		{
			name: "missing target",
			endpoint: Endpoint{
				Name: "test-service",
			},
			wantErr: true,
			errMsg:  "endpoint target is required",
		},
		{
			name: "invalid target URL",
			endpoint: Endpoint{
				Name:   "test-service",
				Target: "invalid-url",
			},
			wantErr: true,
			errMsg:  "endpoint target must be a valid HTTP/HTTPS URL",
		},
		{
			name: "http target",
			endpoint: Endpoint{
				Name:   "test-service",
				Target: "http://api.example.com",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.endpoint.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestServerConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ServerConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: ServerConfig{
				Port:         8080,
				ReadTimeout:  30,
				WriteTimeout: 30,
			},
			wantErr: false,
		},
		{
			name: "invalid port - zero",
			config: ServerConfig{
				Port:         0,
				ReadTimeout:  30,
				WriteTimeout: 30,
			},
			wantErr: true,
			errMsg:  "port must be between 1 and 65535",
		},
		{
			name: "invalid port - too high",
			config: ServerConfig{
				Port:         70000,
				ReadTimeout:  30,
				WriteTimeout: 30,
			},
			wantErr: true,
			errMsg:  "port must be between 1 and 65535",
		},
		{
			name: "negative read timeout",
			config: ServerConfig{
				Port:         8080,
				ReadTimeout:  -1,
				WriteTimeout: 30,
			},
			wantErr: true,
			errMsg:  "read timeout must be non-negative",
		},
		{
			name: "negative write timeout",
			config: ServerConfig{
				Port:         8080,
				ReadTimeout:  30,
				WriteTimeout: -1,
			},
			wantErr: true,
			errMsg:  "write timeout must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProxyConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ProxyConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: ProxyConfig{
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  30,
					WriteTimeout: 30,
				},
				Endpoints: map[string]*Endpoint{
					"service1": {
						Name:   "service1",
						Target: "https://api1.example.com",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "no endpoints",
			config: ProxyConfig{
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  30,
					WriteTimeout: 30,
				},
				Endpoints: map[string]*Endpoint{},
			},
			wantErr: true,
			errMsg:  "at least one endpoint must be configured",
		},
		{
			name: "invalid server config",
			config: ProxyConfig{
				Server: ServerConfig{
					Port:         0,
					ReadTimeout:  30,
					WriteTimeout: 30,
				},
				Endpoints: map[string]*Endpoint{
					"service1": {
						Name:   "service1",
						Target: "https://api1.example.com",
					},
				},
			},
			wantErr: true,
			errMsg:  "invalid server configuration",
		},
		{
			name: "invalid endpoint",
			config: ProxyConfig{
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  30,
					WriteTimeout: 30,
				},
				Endpoints: map[string]*Endpoint{
					"service1": {
						Name:   "",
						Target: "https://api1.example.com",
					},
				},
			},
			wantErr: true,
			errMsg:  "invalid endpoint service1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseProxyRequest(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			data: `{
				"method": "GET",
				"body": null,
				"transformation_mode": "jq",
				"jq_query": "{result: .data}"
			}`,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			data:    `{"method": "GET", "body":}`,
			wantErr: true,
			errMsg:  "invalid JSON",
		},
		{
			name: "validation failure",
			data: `{
				"method": "",
				"body": null,
				"transformation_mode": "jq",
				"jq_query": "{result: .data}"
			}`,
			wantErr: true,
			errMsg:  "validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := ParseProxyRequest([]byte(tt.data))
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, req)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
			}
		})
	}
}

func TestParseProxyConfig(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			data: `{
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
			}`,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			data:    `{"server": {"port": 8080,}}`,
			wantErr: true,
			errMsg:  "invalid JSON",
		},
		{
			name: "validation failure",
			data: `{
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
			wantErr: true,
			errMsg:  "validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseProxyConfig([]byte(tt.data))
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
				require.NotNil(t, config.Endpoints["service1"])
				assert.Equal(t, "service1", config.Endpoints["service1"].Name)
			}
		})
	}
}
