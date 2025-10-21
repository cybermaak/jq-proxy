#!/bin/bash

# Example script showing how to run the service with environment variables
# This demonstrates Docker-friendly configuration

export PROXY_PORT=9090
export PROXY_READ_TIMEOUT=60
export PROXY_WRITE_TIMEOUT=45
export LOG_LEVEL=debug

echo "Starting JQ Proxy Service with environment configuration:"
echo "  Port: $PROXY_PORT"
echo "  Read Timeout: $PROXY_READ_TIMEOUT seconds"
echo "  Write Timeout: $PROXY_WRITE_TIMEOUT seconds"
echo "  Log Level: $LOG_LEVEL"
echo ""

# Run the service
./bin/proxy -config configs/example.json -log-level $LOG_LEVEL