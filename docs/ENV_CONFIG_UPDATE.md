# Environment Variable Configuration Update

## Changes Made

The environment variable endpoint configuration has been updated to provide more flexibility and clarity.

## New Behavior

### Variable Format

**Before:**
```bash
PROXY_ENDPOINT_{NAME}_TARGET=https://api.example.com
```
- The `{NAME}` was converted to lowercase with hyphens and used as both the map key and endpoint name
- Example: `PROXY_ENDPOINT_USER_SERVICE_TARGET` → key: `user-service`, name: `user-service`

**After:**
```bash
PROXY_ENDPOINT_{KEY}_TARGET=https://api.example.com  # Required
PROXY_ENDPOINT_{KEY}_NAME=display-name              # Optional
```
- The `{KEY}` is used as-is for the endpoint map key (case-sensitive)
- The `{KEY}` is what you use in the URL path: `/proxy/{KEY}/...`
- The `_NAME` variable sets the display name (optional)
- If `_NAME` is not provided, the name defaults to the key converted to lowercase with hyphens

## Examples

### Example 1: With Explicit Names

```bash
PROXY_ENDPOINT_USERS_TARGET=https://api.example.com
PROXY_ENDPOINT_USERS_NAME=user-service
```

**Result:**
- Map key: `USERS`
- Endpoint name: `user-service`
- URL path: `/proxy/USERS/...`

### Example 2: Without Names (Auto-Generated)

```bash
PROXY_ENDPOINT_MY_API_TARGET=https://api.example.com
```

**Result:**
- Map key: `MY_API`
- Endpoint name: `my-api` (auto-generated from key)
- URL path: `/proxy/MY_API/...`

### Example 3: Multiple Endpoints

```bash
PROXY_PORT=8080
PROXY_ENDPOINT_USERS_TARGET=https://users.example.com
PROXY_ENDPOINT_USERS_NAME=user-service
PROXY_ENDPOINT_POSTS_TARGET=https://posts.example.com
PROXY_ENDPOINT_POSTS_NAME=posts-service
PROXY_ENDPOINT_ADMIN_TARGET=https://admin.example.com
# ADMIN name will default to "admin"
```

**Result:**
- Three endpoints: `USERS`, `POSTS`, `ADMIN`
- Access via: `/proxy/USERS/...`, `/proxy/POSTS/...`, `/proxy/ADMIN/...`

## Benefits

1. **Clearer Separation**: The key (used in URLs) is separate from the display name
2. **More Control**: You can choose meaningful keys for URLs while having different display names
3. **Backward Compatible**: If you don't provide `_NAME`, it auto-generates from the key
4. **Case Sensitivity**: Keys preserve case, giving you more control over URL structure
5. **Flexibility**: Use short keys in URLs while having descriptive names in configuration

## Configuration Endpoint

Check your configuration with:

```bash
curl http://localhost:8080/config | jq .
```

**Example Response:**
```json
{
  "endpoints": {
    "USERS": {
      "name": "user-service",
      "target": "https://api.example.com"
    },
    "MY_API": {
      "name": "my-api",
      "target": "https://api2.example.com"
    }
  },
  "server": {
    "port": 8080,
    "read_timeout": 30,
    "write_timeout": 30
  }
}
```

## Docker Compose Example

```yaml
version: '3.8'
services:
  jq-proxy:
    image: jq-proxy-service
    ports:
      - "8080:8080"
    environment:
      # Server config
      - PROXY_PORT=8080
      
      # Endpoints with explicit names
      - PROXY_ENDPOINT_USERS_TARGET=https://jsonplaceholder.typicode.com
      - PROXY_ENDPOINT_USERS_NAME=user-service
      
      - PROXY_ENDPOINT_POSTS_TARGET=https://jsonplaceholder.typicode.com
      - PROXY_ENDPOINT_POSTS_NAME=posts-service
      
      # Endpoint with auto-generated name
      - PROXY_ENDPOINT_API_TARGET=https://api.example.com
      # Name will default to "api"
```

## Making Requests

```bash
# Using the USERS endpoint
curl -X POST http://localhost:8080/proxy/USERS/users/1 \
  -H "Content-Type: application/json" \
  -d '{
    "method": "GET",
    "jq_query": "{id, name, email}"
  }'

# Using the MY_API endpoint
curl -X POST http://localhost:8080/proxy/MY_API/data \
  -H "Content-Type: application/json" \
  -d '{
    "method": "GET",
    "jq_query": "."
  }'
```

## Testing

All tests have been updated and pass:

```bash
go test ./...
```

## Documentation Updated

- ✅ `README.md` - Updated environment variable section
- ✅ `docs/CONFIGURATION.md` - Updated with new variable patterns
- ✅ `FEATURE_ENHANCEMENTS.md` - Updated examples
- ✅ `docker-compose.env.yml` - Updated example
- ✅ Tests updated and passing

## Migration Notes

If you were using the old format, your configuration will still work, but the behavior is slightly different:

**Old behavior:**
```bash
PROXY_ENDPOINT_USER_SERVICE_TARGET=https://api.example.com
# → key: "user-service", name: "user-service", URL: /proxy/user-service/...
```

**New behavior:**
```bash
PROXY_ENDPOINT_USER_SERVICE_TARGET=https://api.example.com
# → key: "USER_SERVICE", name: "user-service", URL: /proxy/USER_SERVICE/...
```

**To maintain exact old behavior, add the NAME variable:**
```bash
PROXY_ENDPOINT_USER_SERVICE_TARGET=https://api.example.com
PROXY_ENDPOINT_USER_SERVICE_NAME=user-service
# → key: "USER_SERVICE", name: "user-service", URL: /proxy/USER_SERVICE/...
```

Or adjust your URLs to use the uppercase key.
