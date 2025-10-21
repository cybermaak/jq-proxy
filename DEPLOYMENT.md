# Deployment Guide

This guide explains how to deploy the JQ Proxy Service in different environments.

## Configuration Files

The service uses different configuration files for different environments:

- `configs/example.json` - Example/development configuration
- `configs/development.json` - Development environment
- `configs/production.json` - Production environment
- `configs/docker.json` - Docker-specific configuration

## Environment Variables

Server configuration can be overridden using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PROXY_PORT` | Server port | 8080 |
| `PROXY_READ_TIMEOUT` | Read timeout (seconds) | 30 |
| `PROXY_WRITE_TIMEOUT` | Write timeout (seconds) | 30 |
| `LOG_LEVEL` | Log level (debug, info, warn, error) | info |

## Deployment Options

### 1. Local Development

```bash
# Using Make
make dev

# Or directly with Go
go run cmd/proxy/main.go -config configs/development.json -log-level debug
```

### 2. Docker Development

```bash
# Using Make
make deploy-dev

# Or using docker-compose directly
docker compose -f docker-compose.dev.yml up -d
```

### 3. Production Deployment

#### Option A: Using Make (Recommended)

```bash
# Deploy to production
make deploy-prod

# Start production services
make prod-up

# View logs
make prod-logs

# Stop services
make prod-down
```

#### Option B: Using Docker Compose directly

```bash
# Start production services
docker compose -f docker-compose.prod.yml up -d

# View logs
docker compose -f docker-compose.prod.yml logs -f

# Stop services
docker compose -f docker-compose.prod.yml down
```

#### Option C: Using deployment script

```bash
# Deploy production
./scripts/docker-deploy.sh -e production deploy

# Start production services
./scripts/docker-deploy.sh -e production up

# View logs
./scripts/docker-deploy.sh -e production logs

# Stop services
./scripts/docker-deploy.sh -e production down
```

### 4. Custom Environment Variables

You can override configuration using environment files:

```bash
# Copy and customize environment file
cp .env.production .env.local

# Edit .env.local with your settings
# Then use it with docker-compose
docker compose -f docker-compose.prod.yml --env-file .env.local up -d
```

### 5. Kubernetes Deployment

Use the provided Kubernetes manifests:

```bash
# Apply Kubernetes resources
kubectl apply -f k8s/

# Check deployment status
kubectl get pods -l app=jq-proxy-service

# View logs
kubectl logs -l app=jq-proxy-service -f
```

## Configuration Priority

The service loads configuration in this order (later overrides earlier):

1. JSON configuration file (endpoints + default server settings)
2. Environment variables (server settings only)
3. Command line flags (port override)

## Production Checklist

- [ ] Update `configs/production.json` with correct endpoint URLs
- [ ] Set appropriate environment variables in `.env.production`
- [ ] Configure resource limits in `docker-compose.prod.yml`
- [ ] Set up monitoring and logging
- [ ] Configure health checks
- [ ] Set up SSL/TLS termination (reverse proxy)
- [ ] Configure firewall rules
- [ ] Set up backup procedures

## Monitoring

The service provides a health check endpoint:

```bash
# Check service health
curl http://localhost:8080/health

# Docker health check
docker ps  # Shows health status
```

## Troubleshooting

### Common Issues

1. **Port already in use**
   ```bash
   # Change port in environment variables
   export PROXY_PORT=8081
   ```

2. **Configuration file not found**
   ```bash
   # Ensure config file exists and is mounted correctly
   ls -la configs/
   ```

3. **Permission denied**
   ```bash
   # Check file permissions
   chmod +x scripts/docker-deploy.sh
   ```

4. **Container won't start**
   ```bash
   # Check logs
   docker compose -f docker-compose.prod.yml logs jq-proxy
   ```

### Debug Mode

Enable debug logging:

```bash
# Set environment variable
export LOG_LEVEL=debug

# Or in docker-compose
environment:
  - LOG_LEVEL=debug
```

## Security Considerations

- Use non-root user in containers (already configured)
- Set resource limits to prevent DoS
- Use HTTPS in production (configure reverse proxy)
- Regularly update base images
- Monitor for security vulnerabilities
- Use secrets management for sensitive configuration