# JQ Proxy Service

A configurable JSON REST API proxy service written in Go that accepts requests, transforms them according to configuration, forwards them to target endpoints, and applies response transformations using jq.

## Features

- **Configurable endpoint mappings** - Define multiple backend services
- **jq-based response transformation** - Powerful JSON transformation using jq syntax
- **Header forwarding with filtering** - Control which headers are forwarded
- **Request tracing** - Unique request IDs for debugging
- **Metrics collection** - Track request counts, errors, and response times
- **Docker deployment support** - Production-ready containerization
- **Comprehensive error handling** - Detailed error messages with context
- **Structured logging** - JSON-formatted logs for easy parsing

## Why jq?

This service uses [jq](https://stedolan.github.io/jq/) for response transformation, providing:
- **Powerful transformations** - Filter, map, reduce, and reshape JSON data
- **Familiar syntax** - Industry-standard jq query language
- **Complex operations** - Support for conditionals, functions, and aggregations
- **Type safety** - Strong typing with clear error messages

Example jq transformations:
- Extract fields: `{id, name, email}`
- Filter arrays: `[.[] | select(.active == true)]`
- Aggregate data: `{total: length, sum: [.[].value] | add}`
- Nested access: `.user.profile.settings.theme`

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

### Basic Request Format

```bash
POST /proxy/{endpoint-name}/{path}
Content-Type: application/json

{
  "method": "GET|POST|PUT|PATCH|DELETE",
  "body": null | {} | [],
  "transformation_mode": "jq",
  "jq_query": "jq expression"
}
```

### Examples

#### Simple Field Extraction
```bash
curl -X POST http://localhost:8080/proxy/user-service/users/1 \
  -H "Content-Type: application/json" \
  -d '{
    "method": "GET",
    "jq_query": "{id, name, email}"
  }'
```

Response:
```json
{
  "id": 1,
  "name": "Leanne Graham",
  "email": "Sincere@april.biz"
}
```

#### Array Transformation
```bash
curl -X POST http://localhost:8080/proxy/posts-service/posts \
  -H "Content-Type: application/json" \
  -d '{
    "method": "GET",
    "jq_query": "[.[] | {id, title}]"
  }'
```

#### Filtering and Aggregation
```bash
curl -X POST http://localhost:8080/proxy/posts-service/posts \
  -H "Content-Type: application/json" \
  -d '{
    "method": "GET",
    "jq_query": "{total: length, user_posts: [.[] | select(.userId == 1) | {id, title}]}"
  }'
```

#### POST Request with Body
```bash
curl -X POST http://localhost:8080/proxy/posts-service/posts \
  -H "Content-Type: application/json" \
  -d '{
    "method": "POST",
    "body": {
      "title": "New Post",
      "body": "Post content",
      "userId": 1
    },
    "jq_query": "{id, title, created: true}"
  }'
```

For more examples and detailed API documentation, see [API Documentation](docs/API.md).

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

- [API Documentation](docs/API.md) - Complete API reference with examples
- [Configuration Reference](docs/CONFIGURATION.md) - Detailed configuration guide
- [Testing Guide](docs/TESTING.md) - Testing strategies and examples
- [Deployment Guide](docs/DEPLOYMENT.md) - Production deployment instructions
- [Logging & Monitoring](docs/LOGGING.md) - Logging and metrics documentation

## Troubleshooting

### Common Issues

**Service won't start:**
- Check if the port is already in use: `lsof -i :8080` (Linux/Mac) or `netstat -ano | findstr :8080` (Windows)
- Verify configuration file exists and is valid JSON
- Check file permissions

**Endpoint not found:**
- Verify the endpoint name in your configuration matches the URL
- Check that the configuration file is being loaded correctly
- Review logs for configuration loading errors

**Transformation errors:**
- Test your jq query using [jqplay.org](https://jqplay.org/)
- Check that the response from the target endpoint is valid JSON
- Review error details in the response for specific jq syntax errors

**Connection refused:**
- Verify the target endpoint URL is correct and accessible
- Check network connectivity to the target service
- Ensure the target service is running

For more detailed troubleshooting, enable debug logging:
```bash
./proxy -config configs/config.json -log-level debug
```

## Contributing

Contributions are welcome! Please ensure:
1. All tests pass: `make test`
2. Code is properly formatted: `make fmt`
3. Linting passes: `make lint`
4. Documentation is updated for new features

