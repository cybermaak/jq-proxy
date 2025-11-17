# API Documentation

## Overview

The JQ Proxy Service provides a REST API for proxying requests to configured backend services with jq-based response transformation.

## Base URL

```
http://localhost:8080
```

## Endpoints

### Health Check

Check the service health status.

**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "healthy",
  "service": "jq-proxy-service"
}
```

**Status Codes:**
- `200 OK` - Service is healthy

---

### Metrics

Get service metrics including request counts, error rates, and response times.

**Endpoint:** `GET /metrics`

**Response:**
```json
{
  "total_requests": 150,
  "total_errors": 5,
  "average_response_time": 125000000,
  "endpoints": {
    "user-service": {
      "RequestCount": 100,
      "ErrorCount": 2,
      "TotalResponseTime": 10000000000,
      "AvgResponseTime": 100000000
    },
    "posts-service": {
      "RequestCount": 50,
      "ErrorCount": 3,
      "TotalResponseTime": 7500000000,
      "AvgResponseTime": 150000000
    }
  }
}
```

**Note:** Response times are in nanoseconds (1 second = 1,000,000,000 nanoseconds).

**Status Codes:**
- `200 OK` - Metrics retrieved successfully

---

### Configuration

Get the current service configuration including all configured endpoints.

**Endpoint:** `GET /config`

**Response:**
```json
{
  "server": {
    "port": 8080,
    "read_timeout": 30,
    "write_timeout": 30
  },
  "endpoints": {
    "user-service": {
      "name": "user-service",
      "target": "https://api.example.com/users"
    },
    "posts-service": {
      "name": "posts-service",
      "target": "https://api.example.com/posts"
    }
  }
}
```

**Status Codes:**
- `200 OK` - Configuration retrieved successfully
- `500 Internal Server Error` - Configuration not available

**Use Cases:**
- Verify endpoint configuration
- Debug routing issues
- Monitor configuration changes
- Validate deployment

---

### Proxy Request

Forward a request to a configured endpoint with optional jq transformation.

**Endpoint:** `POST /proxy/{endpoint}/{path}`

**Path Parameters:**
- `endpoint` (required) - The name of the configured endpoint
- `path` (optional) - Additional path to append to the target URL

**Query Parameters:**
All query parameters are forwarded to the target endpoint.

**Headers:**
- `Content-Type: application/json` (required)
- Custom headers are forwarded to the target endpoint
- Headers with `jpx-` prefix are filtered out (not forwarded)

**Request Body:**
```json
{
  "method": "GET|POST|PUT|PATCH|DELETE",
  "body": null | {} | [],
  "transformation_mode": "jq",
  "jq_query": "jq expression"
}
```

**Request Fields:**
- `method` (required) - HTTP method for the target request
- `body` (optional) - Request body to send to the target endpoint
- `transformation_mode` (optional) - Transformation mode, currently only "jq" is supported (default: "jq")
- `jq_query` (required) - jq query expression to transform the response

**Response:**
The transformed response data based on the jq query.

**Status Codes:**
- `200 OK` - Request successful
- `400 Bad Request` - Invalid request format or validation error
- `404 Not Found` - Endpoint not found
- `422 Unprocessable Entity` - Transformation error
- `502 Bad Gateway` - Upstream service error
- `500 Internal Server Error` - Unexpected error

---

## Examples

### Example 1: Simple Field Extraction

Extract specific fields from a user object.

**Request:**
```bash
curl -X POST http://localhost:8080/proxy/user-service/users/1 \
  -H "Content-Type: application/json" \
  -d '{
    "method": "GET",
    "jq_query": "{id, name, email}"
  }'
```

**Response:**
```json
{
  "id": 1,
  "name": "Leanne Graham",
  "email": "Sincere@example.biz"
}
```

---

### Example 2: Array Transformation

Transform an array of posts to extract titles.

**Request:**
```bash
curl -X POST http://localhost:8080/proxy/posts-service/posts \
  -H "Content-Type: application/json" \
  -d '{
    "method": "GET",
    "jq_query": "[.[] | {id, title}]"
  }'
```

**Response:**
```json
[
  {
    "id": 1,
    "title": "First Post"
  },
  {
    "id": 2,
    "title": "Second Post"
  }
]
```

---

### Example 3: Complex Aggregation

Calculate statistics from an array of posts.

**Request:**
```bash
curl -X POST http://localhost:8080/proxy/posts-service/posts \
  -H "Content-Type: application/json" \
  -d '{
    "method": "GET",
    "jq_query": "{total: length, titles: [.[].title], avg_title_length: ([.[].title | length] | add / length)}"
  }'
```

**Response:**
```json
{
  "total": 100,
  "titles": ["First Post", "Second Post", ...],
  "avg_title_length": 25.5
}
```

---

### Example 4: Filtering and Mapping

Filter posts by user ID and extract specific fields.

**Request:**
```bash
curl -X POST http://localhost:8080/proxy/posts-service/posts \
  -H "Content-Type: application/json" \
  -d '{
    "method": "GET",
    "jq_query": "[.[] | select(.userId == 1) | {id, title}]"
  }'
```

**Response:**
```json
[
  {
    "id": 1,
    "title": "Post by User 1"
  },
  {
    "id": 2,
    "title": "Another Post by User 1"
  }
]
```

---

### Example 5: POST Request with Body

Create a new resource with a POST request.

**Request:**
```bash
curl -X POST http://localhost:8080/proxy/posts-service/posts \
  -H "Content-Type: application/json" \
  -d '{
    "method": "POST",
    "body": {
      "title": "New Post",
      "body": "Post content",
      "userId": 1
    },
    "jq_query": "{id, title, created: true}"
  }'
```

**Response:**
```json
{
  "id": 101,
  "title": "New Post",
  "created": true
}
```

---

### Example 6: Query Parameters

Forward query parameters to the target endpoint.

**Request:**
```bash
curl -X POST "http://localhost:8080/proxy/posts-service/posts?userId=1&_limit=5" \
  -H "Content-Type: application/json" \
  -d '{
    "method": "GET",
    "jq_query": "[.[] | {id, title}]"
  }'
```

The query parameters `userId=1` and `_limit=5` are forwarded to the target endpoint.

---

### Example 7: Custom Headers

Forward custom headers to the target endpoint.

**Request:**
```bash
curl -X POST http://localhost:8080/proxy/user-service/users/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer token123" \
  -H "X-Custom-Header: value" \
  -d '{
    "method": "GET",
    "jq_query": "."
  }'
```

Headers `Authorization` and `X-Custom-Header` are forwarded to the target endpoint.

**Note:** Headers with `jpx-` prefix are filtered out and not forwarded.

---

## Error Responses

All error responses follow this format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {
      "additional": "context"
    }
  }
}
```

### Error Codes

| Code | Description | Status Code |
|------|-------------|-------------|
| `ENDPOINT_NOT_FOUND` | The requested endpoint is not configured | 404 |
| `INVALID_REQUEST` | Request validation failed | 400 |
| `TRANSFORMATION_ERROR` | jq transformation failed | 422 |
| `UPSTREAM_ERROR` | Target endpoint returned an error or is unreachable | 502 |
| `INTERNAL_ERROR` | Unexpected server error | 500 |

### Example Error Response

**Request:**
```bash
curl -X POST http://localhost:8080/proxy/nonexistent/users \
  -H "Content-Type: application/json" \
  -d '{
    "method": "GET",
    "jq_query": "."
  }'
```

**Response (404):**
```json
{
  "error": {
    "code": "ENDPOINT_NOT_FOUND",
    "message": "endpoint 'nonexistent' not found",
    "details": {
      "available_endpoints": [
        "user-service",
        "posts-service"
      ]
    }
  }
}
```

---

## jq Query Reference

The service uses [gojq](https://github.com/itchyny/gojq), a pure Go implementation of jq. Most standard jq features are supported.

### Common jq Patterns

**Identity (return as-is):**
```jq
.
```

**Select fields:**
```jq
{id, name, email}
```

**Array mapping:**
```jq
[.[] | {id, title}]
```

**Filtering:**
```jq
[.[] | select(.active == true)]
```

**Aggregation:**
```jq
{total: length, sum: [.[].value] | add}
```

**Nested field access:**
```jq
.user.profile.name
```

**Conditional:**
```jq
if .status == "active" then .data else null end
```

For complete jq documentation, see: https://stedolan.github.io/jq/manual/

---

## Rate Limiting

Currently, the service does not implement rate limiting. Consider implementing rate limiting at the infrastructure level (e.g., using a reverse proxy like nginx or an API gateway).

---

## CORS

The service includes CORS support with the following headers:
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: POST, GET, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type, Authorization`

OPTIONS requests are handled automatically.

---

## Best Practices

1. **Use specific jq queries**: More specific queries are faster and use less memory
2. **Test queries first**: Use the [jq playground](https://jqplay.org/) to test complex queries
3. **Handle errors**: Always check for error responses and handle them appropriately
4. **Monitor metrics**: Use the `/metrics` endpoint to monitor service health
5. **Use request IDs**: Check logs using request IDs for debugging (found in response headers or logs)
6. **Validate input**: Ensure your request body is valid JSON and includes required fields
