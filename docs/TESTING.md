# Testing Guide

This document describes the testing approach for the JQ Proxy Service.

## Test Structure

The project uses Go's built-in testing framework with the following test organization:

- `test/api/` - Comprehensive API tests (replaces the old shell-based `test-api.sh`)
- `test/integration/` - Integration tests for complete service functionality
- `test/benchmark/` - Performance benchmarks
- `internal/*/` - Unit tests alongside source code

## API Testing

### Go-based API Tests

The API tests in `test/api/api_test.go` provide comprehensive testing of all proxy service endpoints and functionality. These tests replace the previous shell-based `scripts/test-api.sh` script.

**Run API tests:**
```bash
make test-api
```

**Test coverage includes:**
- Health endpoint validation
- Simple GET requests
- JSONPath transformations
- jq transformations  
- POST requests with body forwarding
- Error handling scenarios
- Different endpoint configurations
- Complex transformation patterns

### Live Server Testing

The API tests can run against either a mock server (default) or a live development server:

**Run against mock server (default):**
```bash
make test-api
```

**Run against live development server:**
```bash
# Start the development server in one terminal
make dev

# Run tests against live server in another terminal
make test-api-live

# Or set environment variables manually
API_TEST_LIVE=true API_TEST_URL=http://localhost:8080 go test -v ./test/api/...
```

**Benefits of live testing:**
- Tests against real external APIs (JSONPlaceholder, HTTPBin)
- Validates actual network behavior and response formats
- Ensures compatibility with real-world API responses
- Tests the complete request/response cycle

**Environment variables for live testing:**
- `API_TEST_LIVE=true` - Enable live server testing mode
- `API_TEST_URL` - Base URL of the live server (default: http://localhost:8080)

**Note about live testing:**
Live tests may skip certain test cases if external APIs are not accessible (returning 502/503 errors). This is expected behavior in environments where external internet access is restricted. The health endpoint and core proxy functionality will always be tested.

## Migration from Shell Script

The previous `scripts/test-api.sh` shell script has been replaced with:

1. **Go-based API tests** (`test/api/api_test.go`) - For automated testing in CI/CD
2. **Live server testing mode** - For testing against real development servers

### Benefits of the Go-based approach:

- **Better integration** with Go testing ecosystem
- **Type safety** and compile-time validation
- **Consistent tooling** with the rest of the project
- **Easier maintenance** and extension
- **Better error handling** and reporting
- **Cross-platform compatibility** without shell dependencies

### Equivalent functionality:

| Shell Script Command | Go Test Equivalent | Live Test Command |
|---------------------|-------------------|-------------------|
| `./test-api.sh health` | `TestHealthEndpoint` | `make test-api-live` |
| `./test-api.sh simple` | `TestSimpleGETRequest` | `make test-api-live` |
| `./test-api.sh transform` | `TestJSONPathTransformation` | `make test-api-live` |
| `./test-api.sh jq-transform` | `TestJQTransformation` | `make test-api-live` |
| `./test-api.sh post` | `TestPOSTRequest` | `make test-api-live` |
| `./test-api.sh error` | `TestErrorHandling` | `make test-api-live` |
| `./test-api.sh all` | `TestAllScenarios` | `make test-api-live` |

## Running All Tests

```bash
# Run all tests with coverage
make test

# Run only API tests
make test-api

# Run API tests against live development server
make test-api-live

# Run tests quietly (no verbose output)
make test-quiet

# Generate coverage report
make coverage
```

## Test Configuration

Tests use mock servers to simulate external APIs, ensuring:
- **Deterministic results** independent of external services
- **Fast execution** without network dependencies
- **Reliable CI/CD** pipeline execution

The mock servers simulate JSONPlaceholder-like APIs with predictable responses for consistent test validation.