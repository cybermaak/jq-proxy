# JQ Proxy Service

A configurable JSON REST API proxy service written in Go that accepts requests, transforms them according to configuration, forwards them to target endpoints, and applies response transformations using jq.

## Features

- Configurable endpoint mappings
- jq-based response transformation
- Header forwarding with filtering
- Docker deployment support
- Comprehensive error handling
- Structured logging

## Quick Start

### Local Development

```bash
# Run the service with auto-reload
make dev

# Run with custom configuration
make dev CONFIG=configs/production.json PORT=9000 LOG_LEVEL=info

# Run tests
make test

# Build binary
make build
```

### Docker

```bash
# Build image
docker build -t jq-proxy-service .

# Run container
docker run -p 8080:8080 -v $(pwd)/configs:/app/configs jq-proxy-service
```

## Configuration

The service supports configuration through JSON files and environment variables.

### JSON Configuration

See `configs/example.json` for the complete configuration format. The JSON file defines:
- **Endpoints**: Target service mappings (required)
- **Server**: Default server settings (can be overridden by environment variables)

### Environment Variables

Server configuration can be overridden using environment variables, making it Docker-friendly:

| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| `PROXY_PORT` | Port to listen on | 8080 |
| `PROXY_READ_TIMEOUT` | Read timeout in seconds | 30 |
| `PROXY_WRITE_TIMEOUT` | Write timeout in seconds | 30 |

**Note**: Endpoints are always loaded from the JSON configuration file. Only server settings can be overridden with environment variables.

### Docker Compose

The project includes two Docker Compose configurations:

- `docker-compose.dev.yml` - Development environment with hot-reload
- `docker-compose.prod.yml` - Production environment

```bash
# Development
docker compose -f docker-compose.dev.yml up -d

# Production
docker compose -f docker-compose.prod.yml up -d
```

### Production Deployment

For production deployment with the `production.json` config:

```bash
# Using Make
make deploy-prod

# Or using Docker Compose
docker compose -f docker-compose.prod.yml up -d
```

See [DEPLOYMENT.md](DEPLOYMENT.md) for detailed deployment instructions.

## API Usage

### jq Transformation
```bash
POST /proxy/{endpoint-name}/path/to/resource?query=params
Content-Type: application/json

{
  "method": "GET",
  "body": null,

  "jq_query": "{result: [.data[] | {id: .id, name: .name}]}"
}
```

## Testing

The service includes comprehensive Go-based tests:

```bash
# Run all tests
make test

# Run API tests specifically
make test-api

# Run API tests against live development server
make test-api-live
```

See [TESTING.md](TESTING.md) for detailed testing information.

## Documentation

- [Testing Guide](TESTING.md)
- [Configuration Reference](docs/configuration.md)
- [API Documentation](docs/api.md)
- [Development Guide](docs/development.md)