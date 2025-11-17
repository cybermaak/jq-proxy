# Logging and Monitoring

The JQ Proxy Service includes comprehensive logging and monitoring capabilities to help with debugging, performance analysis, and operational visibility.

## Features

### Structured Logging

All logs are output in JSON format with the following fields:
- `timestamp`: ISO 8601 timestamp with nanosecond precision
- `level`: Log level (debug, info, warn, error)
- `message`: Human-readable log message
- `request_id`: Unique identifier for request tracing
- Additional context fields specific to each log entry

### Request Tracing

Every HTTP request is assigned a unique `request_id` that is included in all related log entries. This makes it easy to trace a request through the entire system.

Example log entries for a single request:
```json
{"level":"info","message":"Request started","method":"POST","path":"/proxy/user-service/users/1","request_id":"f585a6fc-0448-4fd2-979b-a1308ebaa035","timestamp":"2025-11-16T16:34:59.653612124-08:00"}
{"endpoint":"user-service","level":"info","message":"Processing proxy request","method":"GET","path":"/users/1","request_id":"f585a6fc-0448-4fd2-979b-a1308ebaa035","timestamp":"2025-11-16T16:34:59.653649799-08:00"}
{"duration_ms":149,"endpoint":"user-service","level":"info","message":"Successfully processed proxy request","request_id":"f585a6fc-0448-4fd2-979b-a1308ebaa035","status_code":200,"timestamp":"2025-11-16T16:34:59.802766596-08:00"}
```

### Metrics Collection

The service collects real-time metrics for:
- Total request count
- Total error count
- Average response time
- Per-endpoint metrics:
  - Request count
  - Error count
  - Average response time

Access metrics via the `/metrics` endpoint:

```bash
curl http://localhost:8080/metrics
```

Example response:
```json
{
  "total_requests": 10,
  "total_errors": 2,
  "average_response_time": 125000000,
  "endpoints": {
    "user-service": {
      "RequestCount": 8,
      "ErrorCount": 1,
      "TotalResponseTime": 1000000000,
      "AvgResponseTime": 125000000
    },
    "posts-service": {
      "RequestCount": 2,
      "ErrorCount": 1,
      "TotalResponseTime": 250000000,
      "AvgResponseTime": 125000000
    }
  }
}
```

Note: Response times are in nanoseconds.

## Log Levels

Configure the log level using the `-log-level` flag:

```bash
./proxy -log-level debug
```

Available levels:
- `debug`: Detailed information for debugging
- `info`: General informational messages (default)
- `warn`: Warning messages
- `error`: Error messages

## Log Output

Logs are written to stdout in JSON format, making them easy to parse and integrate with log aggregation systems like:
- ELK Stack (Elasticsearch, Logstash, Kibana)
- Splunk
- Datadog
- CloudWatch Logs

## Health Check

The service provides a health check endpoint at `/health`:

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "healthy",
  "service": "jq-proxy-service"
}
```

## Monitoring Best Practices

1. **Use request IDs for debugging**: When investigating issues, search logs by `request_id` to see the complete request flow

2. **Monitor metrics regularly**: Set up alerts based on:
   - High error rates
   - Slow response times
   - Unusual traffic patterns

3. **Adjust log levels appropriately**:
   - Use `info` or `warn` in production
   - Use `debug` for troubleshooting specific issues
   - Avoid `debug` in high-traffic production environments

4. **Integrate with monitoring tools**: Export metrics to Prometheus, Grafana, or similar tools for visualization and alerting

## Example: Filtering Logs

Since logs are in JSON format, you can easily filter them using `jq`:

```bash
# Show only error logs
./proxy | jq 'select(.level == "error")'

# Show logs for a specific request ID
./proxy | jq 'select(.request_id == "f585a6fc-0448-4fd2-979b-a1308ebaa035")'

# Show logs for a specific endpoint
./proxy | jq 'select(.endpoint == "user-service")'

# Show slow requests (> 100ms)
./proxy | jq 'select(.duration_ms > 100)'
```
