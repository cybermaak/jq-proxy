#!/bin/bash

# Docker deployment script for JQ Proxy Service

set -e

# Default values
ENVIRONMENT="production"
BUILD_ARGS=""
COMPOSE_FILE="docker-compose.yml"
IMAGE_TAG="latest"
REGISTRY=""

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
    echo "Usage: $0 [OPTIONS] [COMMAND]"
    echo ""
    echo "Options:"
    echo "  -e, --env ENV         Environment (development, production) (default: production)"
    echo "  -t, --tag TAG         Docker image tag (default: latest)"
    echo "  -r, --registry URL    Docker registry URL"
    echo "  -h, --help            Show this help message"
    echo ""
    echo "Commands:"
    echo "  build                 Build Docker image"
    echo "  up                    Start services"
    echo "  down                  Stop services"
    echo "  restart               Restart services"
    echo "  logs                  Show logs"
    echo "  push                  Push image to registry"
    echo "  deploy                Build and deploy"
    echo ""
    echo "Examples:"
    echo "  $0 build                              # Build production image"
    echo "  $0 -e development up                 # Start development environment"
    echo "  $0 -t v1.0.0 build                   # Build with specific tag"
    echo "  $0 -r registry.example.com push      # Push to registry"
}

# Function to build Docker image
build_image() {
    print_info "Building Docker image for $ENVIRONMENT environment..."
    
    if [ "$ENVIRONMENT" = "development" ]; then
        docker build -f Dockerfile.dev -t jq-proxy-service:$IMAGE_TAG .
    else
        docker build -t jq-proxy-service:$IMAGE_TAG .
    fi
    
    print_success "Docker image built: jq-proxy-service:$IMAGE_TAG"
}

# Function to start services
start_services() {
    print_info "Starting services for $ENVIRONMENT environment..."
    
    if [ "$ENVIRONMENT" = "development" ]; then
        docker compose -f docker-compose.dev.yml up -d
    else
        docker compose -f $COMPOSE_FILE up -d
    fi
    
    print_success "Services started"
    print_info "Health check: curl http://localhost:8080/health"
}

# Function to stop services
stop_services() {
    print_info "Stopping services..."
    
    if [ "$ENVIRONMENT" = "development" ]; then
        docker compose -f docker-compose.dev.yml down
    else
        docker compose -f $COMPOSE_FILE down
    fi
    
    print_success "Services stopped"
}

# Function to restart services
restart_services() {
    print_info "Restarting services..."
    stop_services
    start_services
}

# Function to show logs
show_logs() {
    print_info "Showing logs..."
    
    if [ "$ENVIRONMENT" = "development" ]; then
        docker compose -f docker-compose.dev.yml logs -f
    else
        docker compose -f $COMPOSE_FILE logs -f
    fi
}

# Function to push image
push_image() {
    if [ -z "$REGISTRY" ]; then
        print_error "Registry URL is required for push command"
        exit 1
    fi
    
    print_info "Pushing image to registry..."
    
    # Tag for registry
    docker tag jq-proxy-service:$IMAGE_TAG $REGISTRY/jq-proxy-service:$IMAGE_TAG
    
    # Push to registry
    docker push $REGISTRY/jq-proxy-service:$IMAGE_TAG
    
    print_success "Image pushed to $REGISTRY/jq-proxy-service:$IMAGE_TAG"
}

# Function to deploy (build and start)
deploy() {
    print_info "Deploying JQ Proxy Service..."
    build_image
    start_services
    print_success "Deployment complete!"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--env)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -t|--tag)
            IMAGE_TAG="$2"
            shift 2
            ;;
        -r|--registry)
            REGISTRY="$2"
            shift 2
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        build|up|down|restart|logs|push|deploy)
            COMMAND="$1"
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Validate environment
case $ENVIRONMENT in
    development|production)
        ;;
    *)
        print_error "Invalid environment: $ENVIRONMENT. Must be 'development' or 'production'"
        exit 1
        ;;
esac

# Set compose file based on environment
case $ENVIRONMENT in
    development)
        COMPOSE_FILE="docker-compose.dev.yml"
        ;;
    production)
        COMPOSE_FILE="docker-compose.prod.yml"
        ;;
    *)
        COMPOSE_FILE="docker-compose.yml"  # Default/example
        ;;
esac

# Default command if none specified
if [ -z "$COMMAND" ]; then
    COMMAND="deploy"
fi

print_info "Environment: $ENVIRONMENT"
print_info "Image Tag: $IMAGE_TAG"
if [ -n "$REGISTRY" ]; then
    print_info "Registry: $REGISTRY"
fi
echo ""

# Execute the command
case $COMMAND in
    build)
        build_image
        ;;
    up)
        start_services
        ;;
    down)
        stop_services
        ;;
    restart)
        restart_services
        ;;
    logs)
        show_logs
        ;;
    push)
        push_image
        ;;
    deploy)
        deploy
        ;;
    *)
        print_error "Unknown command: $COMMAND"
        show_usage
        exit 1
        ;;
esac