# jq-proxy — Claude Code Guide

## Overview

**jq-proxy** is an HTTP proxy service that forwards requests to upstream endpoints and transforms responses using [jq](https://jqlang.github.io/jq/) queries. Clients send a single `POST /proxy/{endpoint}/{path}` request carrying the HTTP method, optional body, and a jq transformation expression; the proxy fetches from the real upstream, runs the jq filter, and returns the shaped result.

## Quick Start

```bash
# Install dependencies
go mod download

# Build
make build          # produces bin/proxy

# Run (requires a config file or env vars)
make dev            # hot-reload with air (file-based config)
make dev-simple     # plain go run

# Tests
make test           # all unit tests + coverage report
make test-api       # API integration tests (needs running server)
make benchmark      # performance benchmarks
make coverage       # open coverage.html in browser
```

## Project Structure

```
cmd/proxy/          — entry point (main.go)
internal/
  client/           — HTTP client for forwarding requests
  config/           — configuration loading
    file.go         — file-based config (JSON)
    env.go          — env-var overrides on top of file config
    env_endpoints.go — full env-var config (no file needed)
  logging/          — structured logging (logrus), metrics, HTTP middleware
  models/           — shared types and validation (ProxyRequest, ProxyConfig, …)
  proxy/            — HTTP handler (handler.go) and business logic (service.go)
  transform/        — jq execution engine (jq.go) and unified facade (unified.go)
test/
  api/              — API-level integration tests (testify suite)
  integration/      — end-to-end integration tests
  benchmark/        — performance benchmarks
configs/            — sample configuration files
docs/               — architecture documentation
scripts/            — deployment helpers
```

## Configuration

Configuration can come from a JSON file or environment variables:

| Env var | Default | Description |
|---|---|---|
| `PROXY_PORT` | `8080` | Listening port |
| `PROXY_READ_TIMEOUT` | `30` | Read timeout (seconds) |
| `PROXY_WRITE_TIMEOUT` | `30` | Write timeout (seconds) |
| `PROXY_ENDPOINTS_JSON` | — | Full endpoints map as JSON |
| `PROXY_ENDPOINT_{KEY}_TARGET` | — | Target URL for a single endpoint |
| `PROXY_ENDPOINT_{KEY}_NAME` | — | Display name for a single endpoint |

Example config file: `configs/config.example.json`.

## API

```
POST /proxy/{endpoint}
POST /proxy/{endpoint}/{path...}

GET  /health
GET  /metrics
GET  /config
```

Request body (JSON):
```json
{
  "method": "GET",
  "body": null,
  "transformation_mode": "jq",
  "jq_query": "{users: [.data[].name]}"
}
```

## Running Tests

```bash
# Unit tests with race detector and coverage
go test -race ./internal/... -coverprofile=coverage.out

# View per-function coverage
go tool cover -func=coverage.out

# HTML report
go tool cover -html=coverage.out -o coverage.html
```

## Architecture Notes

- **ConfigProvider** (`models.ConfigProvider`) is an interface satisfied by `FileProvider`, `EnvProvider`, and `FullEnvProvider`. Swap implementations at startup in `main.go`.
- **ProxyService** (`models.ProxyService`) is also an interface, making it easy to mock in handler tests.
- Errors are typed (`EndpointNotFoundError`, `TransformationError`, `UpstreamError`) and implement the `ProxyError` interface so `handleProxyError` can map them to HTTP status codes without a type-switch ladder.
- The jq transformer compiles and caches queries for performance; see `internal/transform/jq.go`.
- `RequestLoggingMiddleware` in `internal/logging/middleware.go` wraps every HTTP handler: it generates a UUID request ID, injects it into the context, and logs request start/completion with duration.

## Lint & Format

```bash
make lint           # golangci-lint (config: .golangci.yml)
gofmt -w .
```
