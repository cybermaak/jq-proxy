#!/bin/bash

# Development run script for JQ Proxy Service
# This script provides easy development server startup with various options

set -e

# Default values
CONFIG_FILE="configs/example.json"
PORT=""
LOG_LEVEL="debug"
AUTO_RELOAD=false

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -c, --config FILE     Configuration file path (default: configs/example.json)"
    echo "  -p, --port PORT       Port to listen on (overrides config)"
    echo "  -l, --log-level LEVEL Log level: debug, info, warn, error (default: debug)"
    echo "  -r, --reload          Enable auto-reload with Air (requires Air to be installed)"
    echo "  -h, --help            Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Run with default settings"
    echo "  $0 -p 8081                           # Run on port 8081"
    echo "  $0 -c configs/production.json        # Use production config"
    echo "  $0 -r                                # Enable auto-reload"
    echo "  $0 -l info -p 9000                   # Info logging on port 9000"
}

# Function to check if Air is installed
check_air() {
    if command -v air > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Function to install Air
install_air() {
    print_info "Installing Air for auto-reload..."
    go install github.com/cosmtrek/air@latest
    if [ $? -eq 0 ]; then
        print_success "Air installed successfully"
        return 0
    else
        print_error "Failed to install Air"
        return 1
    fi
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -c|--config)
            CONFIG_FILE="$2"
            shift 2
            ;;
        -p|--port)
            PORT="$2"
            shift 2
            ;;
        -l|--log-level)
            LOG_LEVEL="$2"
            shift 2
            ;;
        -r|--reload)
            AUTO_RELOAD=true
            shift
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Validate configuration file exists
if [ ! -f "$CONFIG_FILE" ]; then
    print_error "Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Validate log level
case $LOG_LEVEL in
    debug|info|warn|error)
        ;;
    *)
        print_error "Invalid log level: $LOG_LEVEL. Must be one of: debug, info, warn, error"
        exit 1
        ;;
esac

# Build command arguments
CMD_ARGS="-config $CONFIG_FILE -log-level $LOG_LEVEL"
if [ -n "$PORT" ]; then
    CMD_ARGS="$CMD_ARGS -port $PORT"
fi

print_info "Starting JQ Proxy Service..."
print_info "Configuration: $CONFIG_FILE"
print_info "Log Level: $LOG_LEVEL"
if [ -n "$PORT" ]; then
    print_info "Port: $PORT"
fi

# Check if we should use auto-reload
if [ "$AUTO_RELOAD" = true ]; then
    if check_air; then
        print_info "Using Air for auto-reload"
        # Update Air config with current arguments
        export AIR_ARGS="$CMD_ARGS"
        air -c .air.toml
    else
        print_warning "Air not found. Would you like to install it? (y/n)"
        read -r response
        if [[ "$response" =~ ^[Yy]$ ]]; then
            if install_air; then
                print_info "Using Air for auto-reload"
                export AIR_ARGS="$CMD_ARGS"
                air -c .air.toml
            else
                print_warning "Falling back to normal mode"
                go run cmd/proxy/main.go $CMD_ARGS
            fi
        else
            print_warning "Falling back to normal mode"
            go run cmd/proxy/main.go $CMD_ARGS
        fi
    fi
else
    print_info "Running in normal mode"
    go run cmd/proxy/main.go $CMD_ARGS
fi