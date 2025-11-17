# Configuration Reference

## Overview

The JQ Proxy Service supports configuration through JSON files and environment variables. This document provides a complete reference for all configuration options.

## Configuration File

The service requires a JSON configuration file that defines endpoints and server settings.

### File Location

By default, the service looks for `configs/config.json`. You can specify a custom path using the `-config` flag:

```bash
./proxy -config /path/to/config.json
```

### Configuration Structure

```json
{
  "server": {
    "port": 8080,
    "read_timeout": 30,
    "write_timeout": 30
  },
  "endpoints": {
    "endpoint-name": {
      "name": "endpoint-name",
      "target": "https://api.example.com"
    }
  }
}
```

---

## Server Configuration

Server settings control how the proxy service operates.

### `server.port`

**Type:** Integer  
**Required:** Yes  
**Default:** 8080  
**Range:** 1-65535  
**Environment Variable:** `PROXY_PORT`

The port number the service listens on.

**Example:**
```json
{
  "server": {
    "port": 9000
  }
}
```

**Environment Override:**
```bash
PROXY_PORT=9000 ./proxy -config configs/config.json
```

---

### `server.read_timeout`

**Type:** Integer  
**Required:** Yes  
**Default:** 30  
**Unit:** Seconds  
**Environment Variable:** `PROXY_READ_TIMEOUT`

Maximum duration for reading the entire request, including the body.

**Example:**
```json
{
  "server": {
    "read_timeout": 60
  }
}
```

**Environment Override:**
```bash
PROXY_READ_TIMEOUT=60 ./proxy -config configs/config.json
```

---

### `server.write_timeout`

**Type:** Integer  
**Required:** Yes  
**Default:** 30  
**Unit:** Seconds  
**Environment Variable:** `PROXY_WRITE_TIMEOUT`

Maximum duration before timing out writes of the response.

**Example:**
```json
{
  "server": {
    "write_timeout": 60
  }
}
```

**Environment Override:**
```bash
PROXY_WRITE_TIMEOUT=60 ./proxy -config configs/config.json
```

---

## Endpoint Configuration

Endpoints define the target services that the proxy can forward requests to.

### `endpoints`

**Type:** Object (map of endpoint configurations)  
**Required:** Yes  
**Minimum:** 1 endpoint required

A map of endpoint names to their configurations.

**Example:**
```json
{
  "endpoints": {
    "user-service": {
      "name": "user-service",
      "target": "https://jsonplaceholder.typicode.com"
    },
    "posts-service": {
      "name": "posts-service",
      "target": "https://jsonplaceholder.typicode.com"
    }
  }
}
```

---

### `endpoints[name].name`

**Type:** String  
**Required:** Yes  
**Pattern:** Must match the key in the endpoints map

The unique identifier for the endpoint. This is used in the proxy URL path.

**Example:**
```json
{
  "endpoints": {
    "my-api": {
      "name": "my-api",
      "target": "https://api.example.com"
    }
  }
}
```

**Usage:**
```bash
POST /proxy/my-api/users
```

---

### `endpoints[name].target`

**Type:** String (URL)  
**Required:** Yes  
**Format:** Must be a valid HTTP or HTTPS URL

The base URL of the target service. All requests to this endpoint will be forwarded to this URL.

**Example:**
```json
{
  "endpoints": {
    "api": {
      "name": "api",
      "target": "https://api.example.com/v1"
    }
  }
}
```

**URL Handling:**
- Trailing slashes are handled automatically
- Path from proxy request is appended to target URL
- Query parameters are preserved

**Example Request Flow:**
```
Proxy Request:  POST /proxy/api/users/123?active=true
Target Request: POST https://api.example.com/v1/users/123?active=true
```

---

## Environment Variables

Environment variables can override server configuration settings. This is particularly useful for Docker deployments.

### Available Environment Variables

#### Server Configuration

| Variable | Description | Type | Default |
|----------|-------------|------|---------|
| `PROXY_PORT` | Server port | Integer | 8080 |
| `PROXY_READ_TIMEOUT` | Read timeout in seconds | Integer | 30 |
| `PROXY_WRITE_TIMEOUT` | Write timeout in seconds | Integer | 30 |

#### Endpoint Configuration

Endpoints can be configured using environment variables in two ways:

**Method 1: Individual Endpoint Variables**

| Variable Pattern | Description | Example |
|-----------------|-------------|---------|
| `PROXY_ENDPOINT_{KEY}_TARGET` | Target URL for endpoint (required) | `PROXY_ENDPOINT_USERS_TARGET=https://api.example.com` |
| `PROXY_ENDPOINT_{KEY}_NAME` | Display name for endpoint (optional) | `PROXY_ENDPOINT_USERS_NAME=user-service` |

**How it works:**
- The `{KEY}` part is used as the endpoint identifier in the URL path
- If `_NAME` is not provided, the name defaults to the key converted to lowercase with hyphens
- The `{KEY}` is case-sensitive and used exactly as specified

**Examples:**
```bash
# With explicit name
PROXY_ENDPOINT_USERS_TARGET=https://api.example.com
PROXY_ENDPOINT_USERS_NAME=user-service
# → Key: "USERS", Name: "user-service", URL: /proxy/USERS/...

# Without name (defaults to lowercase with hyphens)
PROXY_ENDPOINT_MY_API_TARGET=https://api.example.com
# → Key: "MY_API", Name: "my-api", URL: /proxy/MY_API/...
```

**Method 2: JSON Configuration**

| Variable | Description | Format |
|----------|-------------|--------|
| `PROXY_ENDPOINTS_JSON` | All endpoints as JSON | JSON object |

**Example:**
```bash
PROXY_ENDPOINTS_JSON='{"user-service":{"name":"user-service","target":"https://api.example.com"},"posts-service":{"name":"posts-service","target":"https://api2.example.com"}}'
```

**Note:** Individual endpoint variables override JSON configuration if both are present.

### Precedence

1. **With config file**: Environment variables override server settings from the file
2. **Without config file**: All configuration comes from environment variables

**Example with config file:**
```bash
# Configuration file has port: 8080
# Environment variable overrides it
PROXY_PORT=9000 ./proxy -config configs/config.json
# Service will listen on port 9000
```

**Example without config file:**
```bash
# No config file needed - everything from environment
PROXY_PORT=8080 \
PROXY_ENDPOINT_API_TARGET=https://api.example.com \
./proxy
# Service starts with env-only configuration
```

---

## Command-Line Flags

The service supports several command-line flags for runtime configuration.

### `-config`

**Type:** String  
**Default:** `configs/config.json`

Path to the configuration file.

**Example:**
```bash
./proxy -config /etc/jq-proxy/production.json
```

---

### `-port`

**Type:** String  
**Default:** (from config file or environment)

Override the port number. Takes precedence over both config file and environment variables.

**Example:**
```bash
./proxy -config configs/config.json -port 9000
```

---

### `-log-level`

**Type:** String  
**Default:** `info`  
**Options:** `debug`, `info`, `warn`, `error`

Set the logging level.

**Example:**
```bash
./proxy -config configs/config.json -log-level debug
```

**Log Levels:**
- `debug` - Detailed information for debugging
- `info` - General informational messages
- `warn` - Warning messages
- `error` - Error messages only

---

## Configuration Examples

### Minimal Configuration

```json
{
  "server": {
    "port": 8080,
    "read_timeout": 30,
    "write_timeout": 30
  },
  "endpoints": {
    "api": {
      "name": "api",
      "target": "https://api.example.com"
    }
  }
}
```

---

### Multiple Endpoints

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
      "target": "https://users.example.com/api/v1"
    },
    "posts-service": {
      "name": "posts-service",
      "target": "https://posts.example.com/api/v1"
    },
    "comments-service": {
      "name": "comments-service",
      "target": "https://comments.example.com/api/v1"
    }
  }
}
```

---

### Production Configuration

```json
{
  "server": {
    "port": 8080,
    "read_timeout": 60,
    "write_timeout": 60
  },
  "endpoints": {
    "user-api": {
      "name": "user-api",
      "target": "https://api.production.example.com/users"
    },
    "order-api": {
      "name": "order-api",
      "target": "https://api.production.example.com/orders"
    },
    "inventory-api": {
      "name": "inventory-api",
      "target": "https://api.production.example.com/inventory"
    }
  }
}
```

**Run with:**
```bash
PROXY_PORT=8080 \
PROXY_READ_TIMEOUT=60 \
PROXY_WRITE_TIMEOUT=60 \
./proxy -config configs/production.json -log-level info
```

---

### Development Configuration

```json
{
  "server": {
    "port": 3000,
    "read_timeout": 30,
    "write_timeout": 30
  },
  "endpoints": {
    "local-api": {
      "name": "local-api",
      "target": "http://localhost:4000"
    },
    "mock-api": {
      "name": "mock-api",
      "target": "http://localhost:5000"
    }
  }
}
```

**Run with:**
```bash
./proxy -config configs/development.json -log-level debug
```

---

### Environment-Only Configuration

You can run the service without a configuration file by providing all settings via environment variables.

```bash
# Using individual endpoint variables
PROXY_PORT=8080 \
PROXY_READ_TIMEOUT=60 \
PROXY_WRITE_TIMEOUT=60 \
PROXY_ENDPOINT_USER_SERVICE_TARGET=https://users.example.com \
PROXY_ENDPOINT_POSTS_SERVICE_TARGET=https://posts.example.com \
./proxy -log-level info
```

```bash
# Using JSON endpoints
PROXY_PORT=8080 \
PROXY_ENDPOINTS_JSON='{"api":{"name":"api","target":"https://api.example.com"}}' \
./proxy -log-level info
```

**Benefits:**
- No configuration file needed
- Easy to configure in containerized environments
- Simple to manage in orchestration platforms (Kubernetes, Docker Swarm)
- Configuration as code in docker-compose files

---

## Docker Configuration

### Option 1: Environment Variables Only (Recommended)

No configuration file needed - everything configured via environment variables:

```yaml
# docker-compose.yml
version: '3.8'
services:
  jq-proxy:
    image: jq-proxy-service
    ports:
      - "8080:8080"
    environment:
      # Server configuration
      - PROXY_PORT=8080
      - PROXY_READ_TIMEOUT=60
      - PROXY_WRITE_TIMEOUT=60
      
      # Endpoint configuration
      - PROXY_ENDPOINT_USER_SERVICE_TARGET=https://users.example.com
      - PROXY_ENDPOINT_POSTS_SERVICE_TARGET=https://posts.example.com
      - PROXY_ENDPOINT_API_TARGET=https://api.example.com
    # No volumes needed!
```

### Option 2: JSON Endpoints in Environment

```yaml
version: '3.8'
services:
  jq-proxy:
    image: jq-proxy-service
    ports:
      - "8080:8080"
    environment:
      - PROXY_PORT=8080
      - PROXY_ENDPOINTS_JSON={"user-service":{"name":"user-service","target":"https://api.example.com"},"posts-service":{"name":"posts-service","target":"https://api2.example.com"}}
```

### Option 3: Config File with Environment Overrides

```yaml
version: '3.8'
services:
  jq-proxy:
    image: jq-proxy-service
    ports:
      - "8080:8080"
    environment:
      - PROXY_PORT=8080
      - PROXY_READ_TIMEOUT=60
      - PROXY_WRITE_TIMEOUT=60
    volumes:
      - ./configs:/app/configs
    command: ["-config", "/app/configs/production.json", "-log-level", "info"]
```

### Option 4: Config File Only

```yaml
version: '3.8'
services:
  jq-proxy:
    image: jq-proxy-service
    ports:
      - "8080:8080"
    volumes:
      - ./configs/production.json:/app/configs/config.json:ro
    command: ["-config", "/app/configs/config.json"]
```

---

## Configuration Validation

The service validates configuration on startup and will fail with descriptive errors if the configuration is invalid.

### Common Validation Errors

**Missing endpoints:**
```
Failed to load configuration: configuration must have at least one endpoint
```

**Invalid port:**
```
Failed to load configuration: port must be between 1 and 65535
```

**Invalid target URL:**
```
Failed to load configuration: endpoint 'api': target must be a valid HTTP or HTTPS URL
```

**Missing required fields:**
```
Failed to load configuration: endpoint 'api': name is required
```

---

## Best Practices

1. **Use environment variables for deployment-specific settings**: Port, timeouts, etc.
2. **Keep endpoints in the config file**: Don't try to configure endpoints via environment variables
3. **Use descriptive endpoint names**: Makes URLs more readable and logs easier to understand
4. **Set appropriate timeouts**: Consider your target services' response times
5. **Use HTTPS for target URLs**: Ensure secure communication with backend services
6. **Version your config files**: Keep different configs for dev, staging, and production
7. **Validate configs before deployment**: Test configuration changes in a non-production environment first
8. **Document custom endpoints**: Maintain documentation for your specific endpoint configurations
9. **Use consistent naming**: Follow a naming convention for endpoints (e.g., `service-name-api`)
10. **Monitor configuration changes**: Log and track when configuration is updated

---

## Troubleshooting

### Service won't start

**Check configuration file path:**
```bash
./proxy -config configs/config.json
# Ensure the file exists and is readable
```

**Validate JSON syntax:**
```bash
cat configs/config.json | jq .
# Should output formatted JSON without errors
```

**Check port availability:**
```bash
# On Linux/Mac
lsof -i :8080
# On Windows
netstat -ano | findstr :8080
```

### Environment variables not working

**Verify variable names:**
```bash
# Correct
PROXY_PORT=9000 ./proxy -config configs/config.json

# Incorrect (wrong prefix)
PORT=9000 ./proxy -config configs/config.json
```

**Check precedence:**
- Command-line flags override everything
- Environment variables override config file
- Config file is the base

### Endpoints not found

**Verify endpoint name matches:**
```json
{
  "endpoints": {
    "my-api": {  // This is the name to use in URLs
      "name": "my-api",  // Must match the key
      "target": "https://api.example.com"
    }
  }
}
```

**Use correct URL format:**
```bash
# Correct
POST /proxy/my-api/users

# Incorrect (wrong endpoint name)
POST /proxy/myapi/users
```
