# Documentation Index

Complete documentation for the JQ Proxy Service.

## Getting Started

1. **[README.md](README.md)** - Start here! Quick start guide, features overview, and basic usage
2. **[PROJECT_SUMMARY.md](PROJECT_SUMMARY.md)** - Comprehensive project overview and highlights

## Core Documentation

### API & Usage
- **[docs/API.md](docs/API.md)** - Complete API reference with request/response examples
  - Health and metrics endpoints
  - Proxy request format
  - jq transformation examples
  - Error handling
  - Best practices

### Configuration
- **[docs/CONFIGURATION.md](docs/CONFIGURATION.md)** - Detailed configuration guide
  - Configuration file structure
  - Server settings
  - Endpoint configuration
  - Environment variables
  - Command-line flags
  - Docker configuration
  - Troubleshooting

### Testing
- **[docs/TESTING.md](docs/TESTING.md)** - Testing strategies and examples
  - Test structure
  - Running tests
  - Test coverage
  - Writing tests
  - Debugging tests
  - Performance testing

### Deployment
- **[docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)** - Production deployment guide
  - Deployment options
  - Docker deployment
  - Environment configuration
  - Production best practices
  - Monitoring and maintenance

### Logging & Monitoring
- **[docs/LOGGING.md](docs/LOGGING.md)** - Logging and metrics documentation
  - Structured logging
  - Request tracing
  - Metrics collection
  - Log levels
  - Monitoring best practices

## Configuration Examples

- **[configs/example.json](configs/example.json)** - Example configuration with comments
- **[configs/development.json](configs/development.json)** - Development environment config
- **[configs/production.json](configs/production.json)** - Production environment config
- **[configs/docker.json](configs/docker.json)** - Docker-specific configuration

## Additional Resources

### Build & Development
- **[Makefile](Makefile)** - Build targets and development commands
- **[Dockerfile](Dockerfile)** - Production Docker image
- **[Dockerfile.dev](Dockerfile.dev)** - Development Docker image
- **[docker-compose.dev.yml](docker-compose.dev.yml)** - Development compose file
- **[.air.toml](.air.toml)** - Hot reload configuration

### Code Quality
- **[.golangci.yml](.golangci.yml)** - Linter configuration
- **[TEST_RESULTS.md](TEST_RESULTS.md)** - Latest test results

## Quick Reference

### Common Commands

```bash
# Development
make dev                    # Run with hot reload
make test                   # Run all tests
make lint                   # Run linter
make build                  # Build binary

# Docker
make deploy-dev            # Deploy development environment
make deploy-prod           # Deploy production environment

# Testing
make test-api              # Run API tests
make coverage              # Generate coverage report
```

### API Endpoints

```
GET  /health               # Health check
GET  /metrics              # Service metrics
GET  /config               # Current configuration
POST /proxy/{endpoint}/{path}  # Proxy request
```

### Configuration Priority

1. Command-line flags (highest priority)
2. Environment variables
3. Configuration file (lowest priority)

### Environment Variables

```bash
PROXY_PORT=8080           # Server port
PROXY_READ_TIMEOUT=30     # Read timeout (seconds)
PROXY_WRITE_TIMEOUT=30    # Write timeout (seconds)
```

## Documentation by Use Case

### I want to...

**Get started quickly**
→ [README.md](README.md) → Quick Start section

**Understand the API**
→ [docs/API.md](docs/API.md)

**Configure the service**
→ [docs/CONFIGURATION.md](docs/CONFIGURATION.md)

**Deploy to production**
→ [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)

**Write jq transformations**
→ [docs/API.md](docs/API.md) → jq Query Reference section

**Monitor the service**
→ [docs/LOGGING.md](docs/LOGGING.md)

**Run tests**
→ [docs/TESTING.md](docs/TESTING.md)

**Troubleshoot issues**
→ [README.md](README.md) → Troubleshooting section
→ [docs/CONFIGURATION.md](docs/CONFIGURATION.md) → Troubleshooting section

**Understand the architecture**
→ [PROJECT_SUMMARY.md](PROJECT_SUMMARY.md)

## External Resources

- **jq Manual**: https://stedolan.github.io/jq/manual/
- **jq Playground**: https://jqplay.org/
- **gojq (Go implementation)**: https://github.com/itchyny/gojq
- **Go Documentation**: https://golang.org/doc/

## Contributing

When adding new features, please update:
1. Relevant documentation files
2. API examples if applicable
3. Configuration reference if new config options added
4. Test documentation if new test patterns introduced

## Documentation Standards

- Use Markdown format
- Include code examples with syntax highlighting
- Provide both simple and complex examples
- Keep examples up to date with code changes
- Use consistent formatting and structure
- Include troubleshooting sections where applicable

---

**Last Updated**: November 16, 2025
