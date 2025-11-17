# Testing Guide

This document describes the testing approach for the JQ Proxy Service.

## Test Structure

The project uses Go's built-in testing framework with the following test organization:

- `test/api/` - Comprehensive API tests
- `test/integration/` - Integration tests for complete service functionality
- `test/benchmark/` - Performance benchmarks
- `internal/*/` - Unit tests alongside source code

## Running Tests

### All Tests

```bash
# Run all tests with coverage
make test

# Run tests with race detection
go test -race ./...

# Run tests with verbose output
go test -v ./...

# Generate coverage report
make coverage
```

### API Tests

```bash
# Run API tests
make test-api

# Run API tests with verbose output
go test -v ./test/api/...
```

### Integration Tests

```bash
# Run integration tests
go test -v ./test/integration/...
```

### Unit Tests

```bash
# Run unit tests for a specific package
go test -v ./internal/proxy/...
go test -v ./internal/transform/...
go test -v ./internal/config/...
```

## Test Coverage

The project maintains high test coverage across all packages:

- **internal/client**: 91.4% coverage
- **internal/config**: 88.1% coverage
- **internal/logging**: 67.2% coverage
- **internal/models**: 98.2% coverage
- **internal/proxy**: 85.0% coverage
- **internal/transform**: 91.7% coverage

**Overall coverage: 78.2%**

## API Test Coverage

The API tests provide comprehensive coverage of:

- **Health endpoint** - Service health checks
- **Metrics endpoint** - Request metrics and statistics
- **Simple GET requests** - Basic proxy functionality
- **jq transformations** - Field extraction, filtering, aggregation
- **POST requests** - Request body forwarding
- **Error handling** - Invalid requests, missing endpoints, transformation errors
- **Different endpoints** - Multiple backend service configurations
- **Complex transformations** - Nested data, array operations, conditionals
- **Header forwarding** - Custom headers and filtering
- **Query parameters** - Parameter forwarding to target services

## Integration Test Coverage

Integration tests verify:

- **Complete request flow** - End-to-end request processing
- **jq transformations** - Complex transformation scenarios
- **Error handling** - Upstream errors, invalid queries
- **Header forwarding** - jpx- prefix filtering
- **CORS support** - Cross-origin request handling
- **POST requests** - Body forwarding and transformation

## Writing Tests

### Unit Test Example

```go
func TestJQTransformer_TransformWithQuery(t *testing.T) {
    transformer := transform.NewJQTransformer()
    
    data := map[string]interface{}{
        "id": 1,
        "name": "John",
        "email": "john@example.com",
    }
    
    result, err := transformer.TransformWithQuery(data, "{id, name}")
    
    require.NoError(t, err)
    assert.Equal(t, map[string]interface{}{
        "id": float64(1),
        "name": "John",
    }, result)
}
```

### Integration Test Example

```go
func (suite *IntegrationTestSuite) TestJQTransformation() {
    requestBody := map[string]interface{}{
        "method": "GET",
        "jq_query": "{id, name, email}",
    }
    
    response := suite.makeProxyRequest("user-service", "/users/1", requestBody)
    
    suite.Equal(200, response.StatusCode)
    suite.Contains(response.Body, "id")
    suite.Contains(response.Body, "name")
    suite.Contains(response.Body, "email")
}
```

## Test Best Practices

1. **Use table-driven tests** for multiple test cases
2. **Mock external dependencies** to ensure test reliability
3. **Test error cases** as thoroughly as success cases
4. **Use descriptive test names** that explain what is being tested
5. **Keep tests independent** - no shared state between tests
6. **Use test fixtures** for complex test data
7. **Test edge cases** - empty data, null values, large datasets
8. **Verify error messages** not just error presence

## Continuous Integration

Tests are designed to run in CI/CD pipelines:

```bash
# CI test command
go test -race -coverprofile=coverage.out ./...

# Generate coverage report
go tool cover -html=coverage.out -o coverage.html
```

## Performance Testing

### Benchmarks

Run performance benchmarks:

```bash
# Run all benchmarks
go test -bench=. ./test/benchmark/...

# Run specific benchmark
go test -bench=BenchmarkProxyRequest ./test/benchmark/...

# Run with memory profiling
go test -bench=. -benchmem ./test/benchmark/...
```

### Load Testing

For load testing, use tools like:
- **Apache Bench (ab)**: `ab -n 1000 -c 10 http://localhost:8080/health`
- **wrk**: `wrk -t4 -c100 -d30s http://localhost:8080/health`
- **hey**: `hey -n 1000 -c 10 http://localhost:8080/health`

## Debugging Tests

### Verbose Output

```bash
# Run with verbose output
go test -v ./...

# Run specific test
go test -v -run TestJQTransformation ./test/integration/...
```

### Debug Logging

Enable debug logging in tests:

```go
logger, _ := logging.NewLogger("debug")
```

### Test Isolation

Run a single test:

```bash
go test -v -run TestName ./path/to/package/...
```

## Test Data

### Mock Servers

Tests use mock HTTP servers to simulate external APIs:

```go
mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "id": 1,
        "name": "Test User",
    })
}))
defer mockServer.Close()
```

### Test Fixtures

Complex test data is defined in test fixtures:

```go
var testUser = map[string]interface{}{
    "id": 1,
    "name": "Leanne Graham",
    "email": "Sincere@april.biz",
    "address": map[string]interface{}{
        "city": "Gwenborough",
    },
}
```

## Common Test Scenarios

### Testing jq Transformations

```go
tests := []struct {
    name     string
    query    string
    input    interface{}
    expected interface{}
}{
    {
        name:  "field extraction",
        query: "{id, name}",
        input: map[string]interface{}{"id": 1, "name": "John", "email": "john@example.com"},
        expected: map[string]interface{}{"id": 1, "name": "John"},
    },
    {
        name:  "array filtering",
        query: "[.[] | select(.active == true)]",
        input: []interface{}{
            map[string]interface{}{"id": 1, "active": true},
            map[string]interface{}{"id": 2, "active": false},
        },
        expected: []interface{}{
            map[string]interface{}{"id": 1, "active": true},
        },
    },
}
```

### Testing Error Handling

```go
func TestErrorHandling(t *testing.T) {
    // Test invalid jq query
    _, err := transformer.TransformWithQuery(data, "invalid{")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "invalid jq query")
    
    // Test endpoint not found
    response := makeRequest("/proxy/nonexistent/users")
    assert.Equal(t, 404, response.StatusCode)
    assert.Contains(t, response.Body, "ENDPOINT_NOT_FOUND")
}
```

## Troubleshooting Tests

### Tests Failing Locally

1. **Check Go version**: Ensure you're using Go 1.21 or later
2. **Clean build cache**: `go clean -testcache`
3. **Update dependencies**: `go mod tidy`
4. **Check for port conflicts**: Ensure test ports are available

### Race Condition Warnings

If you see race condition warnings:

```bash
# Run with race detector
go test -race ./...
```

Fix any reported race conditions before committing.

### Coverage Issues

To identify untested code:

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

Open the HTML report to see which lines are not covered.

## Test Maintenance

1. **Update tests when changing functionality**
2. **Add tests for new features**
3. **Remove obsolete tests**
4. **Keep test data up to date**
5. **Review test coverage regularly**
6. **Refactor tests to reduce duplication**

## Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [jq Manual](https://stedolan.github.io/jq/manual/)
