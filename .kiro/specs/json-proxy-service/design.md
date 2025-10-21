# Design Document

## Overview

The JQ Proxy Service is a Go-based HTTP proxy that accepts POST requests with JSON payloads, forwards them to configured target endpoints with specified HTTP methods, and transforms the responses using JQ expressions. The service uses a modular architecture with abstracted configuration management to support future extensibility.

## Architecture

The service follows a layered architecture pattern:

```
┌─────────────────────────────────────────┐
│              HTTP Handler               │
├─────────────────────────────────────────┤
│            Proxy Service                │
├─────────────────────────────────────────┤
│     Config Manager    │  Transformer    │
├─────────────────────────────────────────┤
│          HTTP Client Pool               │
└─────────────────────────────────────────┘
```

### Key Components:

- **HTTP Handler**: Receives and validates incoming requests
- **Proxy Service**: Core business logic for request routing and processing
- **Config Manager**: Abstracted configuration loading and management
- **Transformer**: JSONPath-based response transformation
- **HTTP Client Pool**: Manages connections to target endpoints

## Components and Interfaces

### 1. Configuration Interface

```go
type ConfigProvider interface {
    LoadConfig() (*ProxyConfig, error)
    GetEndpoint(name string) (*Endpoint, bool)
    Reload() error
}

type ProxyConfig struct {
    Endpoints map[string]*Endpoint `json:"endpoints"`
    Server    ServerConfig         `json:"server"`
}

type Endpoint struct {
    Name   string `json:"name"`
    Target string `json:"target"`
}

type ServerConfig struct {
    Port         int    `json:"port"`
    ReadTimeout  int    `json:"read_timeout"`
    WriteTimeout int    `json:"write_timeout"`
}
```

### 2. Request/Response Models

```go
type ProxyRequest struct {
    Method        string                 `json:"method"`
    Body          interface{}            `json:"body"`
    Transformation map[string]interface{} `json:"transformation"`
}

type ProxyResponse struct {
    Data   interface{} `json:"data"`
    Status int         `json:"status"`
}
```

### 3. Proxy Service Interface

```go
type ProxyService interface {
    HandleRequest(ctx context.Context, endpointName string, path string,
                 queryParams url.Values, headers http.Header,
                 proxyReq *ProxyRequest) (*ProxyResponse, error)
}
```

### 4. Transformer Interface

```go
type ResponseTransformer interface {
    Transform(data interface{}, transformation map[string]interface{}) (interface{}, error)
}
```

## Data Models

### Configuration File Structure (JSON)

```json
{
  "server": {
    "port": 8080,
    "read_timeout": 30,
    "write_timeout": 30
  },
  "endpoints": {
    "user-service": {
      "name": "user-service",
      "target": "https://api.users.example.com"
    },
    "order-service": {
      "name": "order-service",
      "target": "http://orders.internal:8080"
    }
  }
}
```

### Request Flow Examples

#### JSONPath Transformation
```
POST /proxy/user-service/api/v1/users?limit=10
Headers: Authorization: Bearer token, jpx-debug: true
Body: {
  "method": "GET",
  "body": null,
  "transformation": {
    "users": "$.data[*].{id: id, name: name}",
    "total": "$.total"
  }
}
```

#### jq Transformation
```
POST /proxy/user-service/api/v1/users?limit=10
Headers: Authorization: Bearer token, jpx-debug: true
Body: {
  "method": "GET",
  "body": null,
  "transformation_mode": "jq",
  "jq_query": "{users: [.data[] | {id: .id, name: .name}], total: .total}"
}
```

## Error Handling

### Error Types and Responses

1. **Configuration Errors** (500)

   - Missing or invalid configuration file
   - Invalid endpoint configuration

2. **Request Validation Errors** (400)

   - Invalid JSON in request body
   - Missing required fields (method, transformation)
   - Invalid HTTP method

3. **Routing Errors** (404)

   - Unknown endpoint name
   - Target endpoint unreachable

4. **Transformation Errors** (422)

   - Invalid JSONPath expressions
   - Transformation execution failures

5. **Upstream Errors** (502/503/504)
   - Target endpoint errors
   - Network timeouts
   - Connection failures

### Error Response Format

```json
{
  "error": {
    "code": "INVALID_ENDPOINT",
    "message": "Endpoint 'unknown-service' not found",
    "details": {
      "available_endpoints": ["user-service", "order-service"]
    }
  }
}
```

## Testing Strategy

### Unit Tests

- Configuration loading and validation
- Request parsing and validation
- JSONPath transformation logic
- HTTP client wrapper functionality
- Error handling scenarios

### Integration Tests

- End-to-end proxy request flow
- Multiple endpoint configurations
- Header forwarding behavior
- Response transformation accuracy
- Error propagation from target services

### Test Utilities

- Mock HTTP servers for target endpoints
- Configuration file fixtures
- Request/response builders
- JSONPath test cases

### Performance Tests

- Concurrent request handling
- Memory usage under load
- Response time benchmarks
- Connection pooling efficiency

## Implementation Details

### Libraries and Dependencies

1. **HTTP Framework**: Standard `net/http` with `gorilla/mux` for routing
2. **JSON Processing**: Standard `encoding/json`
3. **JSONPath**: `theory/jsonpath` (RFC 9535 compliant)
4. **jq**: `itchyny/gojq` for jq-style transformations
4. **Configuration**: Standard `encoding/json` for file-based config
5. **Logging**: `logrus` or `zap` for structured logging
6. **Testing**: Standard `testing` package with `testify` for assertions

### Directory Structure

```
jq-proxy-service/
├── cmd/
│   └── proxy/
│       └── main.go
├── internal/
│   ├── config/
│   │   ├── provider.go
│   │   └── file.go
│   ├── proxy/
│   │   ├── service.go
│   │   └── handler.go
│   ├── transform/
│   │   └── jsonpath.go
│   └── models/
│       └── types.go
├── pkg/
│   └── client/
│       └── http.go
├── configs/
│   └── example.json
├── docker/
│   └── Dockerfile
├── scripts/
│   └── dev-run.sh
├── go.mod
├── go.sum
└── README.md
```

### Development and Deployment

#### Local Development

- `make dev`: Start service with file watcher for auto-reload
- `make test`: Run all tests with coverage
- `make lint`: Run code quality checks

#### Docker Deployment

- Multi-stage Dockerfile for optimized image size
- Configuration via mounted files or environment variables
- Health check endpoint for container orchestration
- Non-root user for security

#### Configuration Management

- Environment variable override support
- Configuration validation on startup
- Graceful handling of configuration reload (future feature)

### Security Considerations

1. **Input Validation**: Strict JSON schema validation for requests
2. **Header Filtering**: Prevent forwarding of sensitive proxy headers
3. **Rate Limiting**: Configurable rate limits per endpoint (future feature)
4. **Timeout Management**: Prevent resource exhaustion from slow targets
5. **Error Information**: Avoid leaking internal details in error responses
