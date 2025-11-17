.PHONY: build test dev clean lint deps run-config docker-build docker-run install-tools check coverage benchmark

# Build the binary
build:
	@echo "Building binary..."
	go build -o bin/proxy cmd/proxy/main.go
	@echo "Binary built: bin/proxy"

# Build all binaries
build-all: build

# Run tests with coverage
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run API tests specifically (replaces test-api.sh)
test-api:
	@echo "Running API tests..."
	go test -v -race ./test/api/...
	@echo "API tests completed"

# Run API tests against live development server
test-api-live:
	@echo "Running API tests against live development server..."
	@echo "Make sure the development server is running: make dev"
	API_TEST_LIVE=true go test -v -race ./test/api/...
	@echo "Live API tests completed"

# Run tests without verbose output
test-quiet:
	go test -race -coverprofile=coverage.out ./...

# Run in development mode with auto-reload
dev:
	@echo "Starting development server..."
	@if command -v air > /dev/null 2>&1; then \
		air -c .air.toml; \
	else \
		echo "Air not installed. Running without auto-reload..."; \
		go run cmd/proxy/main.go -config configs/example.json -log-level debug; \
	fi

# Run in development mode without auto-reload
dev-simple:
	@echo "Starting development server (simple mode)..."
	go run cmd/proxy/main.go -config configs/example.json -log-level debug

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -rf tmp/
	@echo "Clean complete"

# Run linting and formatting
lint:
	@echo "Running linters..."
	go fmt ./...
	go vet ./...
	@if command -v golangci-lint > /dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: make install-tools"; \
	fi

# Install development dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download
	@echo "Dependencies installed"

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development tools installed"

# Run with custom config
run-config:
	@if [ -z "$(CONFIG)" ]; then \
		echo "Usage: make run-config CONFIG=path/to/config.json"; \
		exit 1; \
	fi
	go run cmd/proxy/main.go -config $(CONFIG)

# Run with custom port
run-port:
	@if [ -z "$(PORT)" ]; then \
		echo "Usage: make run-port PORT=8081"; \
		exit 1; \
	fi
	go run cmd/proxy/main.go -config configs/example.json -port $(PORT)

# Check code quality
check: lint test-quiet
	@echo "Code quality check complete"

# Generate coverage report
coverage:
	@echo "Generating coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out
	@echo "Coverage report: coverage.html"

# Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t jq-proxy-service .
	@echo "Docker image built: jq-proxy-service"

# Docker run
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 -v $(PWD)/configs:/app/configs jq-proxy-service

# Production deployment
deploy-prod:
	@echo "Deploying to production..."
	./scripts/docker-deploy.sh -e production deploy

# Development deployment
deploy-dev:
	@echo "Deploying to development..."
	./scripts/docker-deploy.sh -e development deploy

# Start production services
prod-up:
	@echo "Starting production services..."
	docker compose -f docker-compose.prod.yml up -d

# Stop production services
prod-down:
	@echo "Stopping production services..."
	docker compose -f docker-compose.prod.yml down

# Production logs
prod-logs:
	@echo "Showing production logs..."
	docker compose -f docker-compose.prod.yml logs -f

# Docker run with custom port
docker-run-port:
	@if [ -z "$(PORT)" ]; then \
		echo "Usage: make docker-run-port PORT=8081"; \
		exit 1; \
	fi
	docker run -p $(PORT):8080 -v $(PWD)/configs:/app/configs jq-proxy-service

# Create example request for testing
example-request:
	@echo "Example jq proxy request:"
	@echo "curl -X POST http://localhost:8080/proxy/user-service/posts \\"
	@echo "  -H 'Content-Type: application/json' \\"
	@echo "  -d '{"
	@echo "    \"method\": \"GET\","
	@echo "    \"body\": null,"
	@echo "    \"transformation\": {"
	@echo "      \"posts\": \"$[*].{id: id, title: title}\","
	@echo "      \"count\": \"$.length\""
	@echo "    }"
	@echo "  }'"
	@echo ""
	@echo "Example jq proxy request:"
	@echo "curl -X POST http://localhost:8080/proxy/user-service/posts \\"
	@echo "  -H 'Content-Type: application/json' \\"
	@echo "  -d '{"
	@echo "    \"method\": \"GET\","
	@echo "    \"body\": null,"
	@echo "    \"transformation_mode\": \"jq\","
	@echo "    \"jq_query\": \"{posts: [.[] | {id: .id, title: .title}], count: length}\""
	@echo "  }'"

# Show help
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  build-all      - Build all binaries"
	@echo "  test           - Run tests with coverage"
	@echo "  test-api       - Run API tests (replaces test-api.sh)"
	@echo "  test-api-live  - Run API tests against live development server"
	@echo "  test-quiet     - Run tests without verbose output"
	@echo "  dev            - Run in development mode with auto-reload"
	@echo "  dev-simple     - Run in development mode without auto-reload"
	@echo "  clean          - Clean build artifacts"
	@echo "  lint           - Run linting and formatting"
	@echo "  deps           - Install dependencies"
	@echo "  install-tools  - Install development tools"
	@echo "  check          - Run linting and tests"
	@echo "  coverage       - Generate coverage report"
	@echo "  benchmark      - Run benchmarks"
	@echo "  run-config     - Run with custom config (CONFIG=path)"
	@echo "  run-port       - Run with custom port (PORT=8081)"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  docker-run-port- Run Docker container with custom port (PORT=8081)"
	@echo "  deploy-prod    - Deploy to production environment"
	@echo "  deploy-dev     - Deploy to development environment"
	@echo "  prod-up        - Start production services"
	@echo "  prod-down      - Stop production services"
	@echo "  prod-logs      - Show production logs"
	@echo "  example-request- Show example request"
	@echo "  help           - Show this help"