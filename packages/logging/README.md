# Veil - Logging

Home to all things related to logging setup for Veil. This package provides centralized logging infrastructure using Loki and Grafana for log aggregation and visualization.

## Features

- **Structured Logging**: JSON-formatted logs with structured fields for better searchability
- **Loki Integration**: Direct log shipping to Grafana Loki for centralized log aggregation
- **Grafana Dashboards**: Pre-configured dashboards for monitoring API gateway performance
- **Request Tracing**: Detailed logging of HTTP requests, responses, and API validations
- **Error Tracking**: Comprehensive error logging with context and stack traces

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Veil Handler  │────│  Logging Pkg    │────│      Loki       │
│                 │    │                 │    │                 │
│ - HTTP Requests │    │ - Structured    │    │ - Log Storage   │
│ - API Validation│    │   Logging       │    │ - Indexing      │
│ - Error Handling│    │ - Loki Hook     │    │ - Retention     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                        │
                                                        │
                                               ┌─────────────────┐
                                               │     Grafana     │
                                               │                 │
                                               │ - Dashboards    │
                                               │ - Visualization │
                                               │ - Alerting      │
                                               └─────────────────┘
```

## Quick Start

### 1. Start the Observability Stack

```bash
cd packages/logging
docker-compose up -d
```

This will start:
- **Loki** on port 3100 (log aggregation)
- **Grafana** on port 3000 (visualization)
- **Promtail** (log collection agent)

### 2. Access Grafana

- URL: http://localhost:3000
- Username: admin
- Password: admin

The Veil API Gateway dashboard will be automatically available.

### 3. Integration with Veil Handler

The logging package is automatically integrated with the Veil Caddy handler. Logs are sent to both console and Loki.

## Configuration

### Environment Variables

```bash
# Loki endpoint (optional, defaults to http://localhost:3100)
LOKI_URL=http://localhost:3100

# Log level (optional, defaults to info)
LOG_LEVEL=info

# Additional labels for logs (optional)
LOG_LABELS='{"environment":"production","service":"veil-gateway"}'
```

### Caddyfile Configuration

No additional configuration needed in the Caddyfile. Logging is automatically enabled when the Veil handler is used.

## Log Structure

Logs are structured in JSON format with the following fields:

```json
{
  "ts": "2024-01-15T10:30:45Z",
  "level": "info",
  "msg": "request completed",
  "caller": "handlers/veil_handler.go:123",
  "method": "GET",
  "path": "/api/v1/users",
  "status": 200,
  "duration": "45.123ms",
  "remote_addr": "192.168.1.100:54321",
  "api_key_name": "client-app-key",
  "subscription": "premium",
  "upstream": "http://localhost:8080"
}
```

### Log Fields Reference

| Field | Description | Example |
|-------|-------------|---------|
| `ts` | Timestamp in RFC3339 format | `2024-01-15T10:30:45Z` |
| `level` | Log level (debug, info, warn, error) | `info` |
| `msg` | Log message | `request completed` |
| `caller` | Source file and line | `handlers/veil_handler.go:123` |
| `method` | HTTP method | `GET` |
| `path` | Request path | `/api/v1/users` |
| `status` | HTTP status code | `200` |
| `duration` | Request duration | `45.123ms` |
| `remote_addr` | Client address | `192.168.1.100:54321` |
| `api_key_name` | API key identifier | `client-app-key` |
| `subscription` | Subscription level | `premium` |
| `upstream` | Backend service URL | `http://localhost:8080` |
| `error` | Error message (if applicable) | `api key not found` |

## Grafana Dashboards

### Veil API Gateway Dashboard

Pre-configured dashboard includes:

- **Request Rate**: Requests per second over time
- **Response Status Codes**: Distribution of HTTP status codes
- **Response Time**: 95th percentile response times
- **Error Rate**: Failed requests per second
- **Requests by API Path**: Traffic breakdown by endpoint
- **Requests by Subscription Level**: Usage by subscription tier
- **Recent Error Logs**: Table of recent errors with details

### Custom Queries

Common LogQL queries for Veil logs:

```logql
# All application logs
{job="veil-application"}

# Error logs only
{job="veil-application"} |= "error"

# Requests to specific API path
{job="veil-application"} | json | path="/api/v1/users"

# Failed API key validations
{job="veil-application"} | json | msg="api key validation failed"

# High response times (over 1 second)
{job="veil-application"} | json | duration > 1s

# Requests by subscription level
{job="veil-application"} | json | subscription="premium"
```

## Advanced Configuration

### Custom Labels

Add custom labels to all logs:

```go
import "github.com/try-veil/veil/packages/logging"

labels := map[string]string{
    "environment": "production",
    "datacenter": "us-west-1",
    "version": "v1.2.3",
}

logger, err := logging.CreateLokiLogger("http://localhost:3100", labels)
```

### Log Retention

Configure log retention in `config/loki-config.yaml`:

```yaml
limits_config:
  retention_period: 30d

table_manager:
  retention_deletes_enabled: true
  retention_period: 30d
```

### Performance Tuning

For high-throughput environments, adjust these settings:

```yaml
# Loki configuration
limits_config:
  max_streams_per_user: 10000
  max_line_size: 256KB
  max_entries_limit_per_query: 5000
  ingestion_rate_mb: 4
  ingestion_burst_size_mb: 6

# Promtail configuration
server:
  http_listen_port: 9080
  
client_configs:
  - url: http://loki:3100/loki/api/v1/push
    batchwait: 1s
    batchsize: 1048576
```

## Troubleshooting

### Common Issues

1. **Logs not appearing in Grafana**
   - Check Loki container status: `docker logs veil-loki`
   - Verify Loki URL configuration
   - Ensure firewall allows port 3100

2. **High memory usage**
   - Reduce log retention period
   - Limit log line size in Loki config
   - Implement log sampling for high-volume endpoints

3. **Grafana dashboard not loading**
   - Check Grafana container status: `docker logs veil-grafana`
   - Verify datasource configuration
   - Restart Grafana container if needed

### Debug Mode

Enable debug logging:

```bash
export LOG_LEVEL=debug
```

### Health Checks

Check component health:

```bash
# Loki health
curl http://localhost:3100/ready

# Grafana health  
curl http://localhost:3000/api/health

# Promtail metrics
curl http://localhost:9080/metrics
```

## Development

### Building

```bash
cd packages/logging
go mod tidy
go build ./...
```

### Testing

```bash
go test ./...
```

### Integration Testing

```bash
# Start the stack
docker-compose up -d

# Wait for services to be ready
sleep 30

# Run integration tests
go test -tags=integration ./...
```

## API Reference

### LoggingMiddleware

```go
type LoggingMiddleware struct {
    logger *zap.Logger
}

func NewLoggingMiddleware(logger *zap.Logger) *LoggingMiddleware
func (m *LoggingMiddleware) LogRequest(r *http.Request, additionalFields ...zap.Field)
func (m *LoggingMiddleware) LogResponse(r *http.Request, statusCode int, duration time.Duration, additionalFields ...zap.Field)
func (m *LoggingMiddleware) LogAPIValidation(r *http.Request, valid bool, reason string, additionalFields ...zap.Field)
func (m *LoggingMiddleware) LogAPIKeyValidation(r *http.Request, keyName string, valid bool, subscription string, additionalFields ...zap.Field)
func (m *LoggingMiddleware) LogError(r *http.Request, err error, additionalFields ...zap.Field)
func (m *LoggingMiddleware) WrapHandler(next http.Handler) http.Handler
```

### LokiHook

```go
type LokiHook struct {
    lokiURL string
    labels  map[string]string
    client  *http.Client
}

func NewLokiHook(lokiURL string, labels map[string]string) *LokiHook
func CreateLokiLogger(lokiURL string, labels map[string]string) (*zap.Logger, error)
```

## Contributing

1. Follow the Go coding standards outlined in the repository rules
2. Add tests for new functionality
3. Update documentation for any new features
4. Ensure all logs follow the structured format