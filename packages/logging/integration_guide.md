# Veil Logging Integration Guide

This guide shows how to integrate the Veil logging package with your applications and how to extend the existing Caddy handler to use structured logging with Loki.

## Quick Integration

### 1. Add Logging Dependency

Add the logging package to your Go module:

```bash
cd packages/caddy
go mod edit -require github.com/try-veil/veil/packages/logging@latest
go mod tidy
```

### 2. Update Veil Handler

Here's how to integrate the logging package with the existing Veil handler:

```go
// In packages/caddy/internal/handlers/veil_handler.go

import (
    "time"
    "os"
    
    "github.com/try-veil/veil/packages/logging"
    "go.uber.org/zap"
)

// Add to VeilHandler struct
type VeilHandler struct {
    // ... existing fields ...
    loggingMiddleware *logging.LoggingMiddleware
}

// Update Provision method
func (h *VeilHandler) Provision(ctx caddy.Context) error {
    // ... existing code ...
    
    // Get Loki URL from environment or use default
    lokiURL := os.Getenv("LOKI_URL")
    if lokiURL == "" {
        lokiURL = "http://localhost:3100"
    }
    
    // Create labels for this handler instance
    labels := map[string]string{
        "service": "veil-gateway",
        "handler": "veil_handler",
        "db_path": h.DBPath,
    }
    
    // Create Loki logger
    lokiLogger, err := logging.CreateLokiLogger(lokiURL, labels)
    if err != nil {
        h.logger.Warn("Failed to create Loki logger, falling back to console",
            zap.Error(err))
        lokiLogger = h.logger // Use existing logger as fallback
    }
    
    // Create logging middleware
    h.loggingMiddleware = logging.NewLoggingMiddleware(lokiLogger)
    h.logger = lokiLogger // Replace the existing logger
    
    // ... rest of existing code ...
}

// Update ServeHTTP method for enhanced logging
func (h *VeilHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
    start := time.Now()
    
    // Log request start
    h.loggingMiddleware.LogRequest(r, 
        zap.String("handler", "veil_handler"),
        zap.String("subscription_key", h.SubscriptionKey))
    
    // Validate API
    api, err := h.validateAPIKey(r.URL.Path, r.Header.Get(h.SubscriptionKey))
    if err != nil {
        h.loggingMiddleware.LogAPIValidation(r, false, err.Error(),
            zap.String("api_key_header", h.SubscriptionKey))
        h.loggingMiddleware.LogResponse(r, http.StatusForbidden, time.Since(start))
        return caddyhttp.Error(http.StatusForbidden, err)
    }
    
    if api != nil {
        h.loggingMiddleware.LogAPIValidation(r, true, "api key validated",
            zap.String("api_path", api.Path),
            zap.String("subscription", api.RequiredSubscription),
            zap.String("upstream", api.Upstream))
    }
    
    // Call next handler
    err = next.ServeHTTP(w, r)
    
    // Log completion
    status := http.StatusOK
    if err != nil {
        status = http.StatusInternalServerError
        h.loggingMiddleware.LogError(r, err,
            zap.String("upstream", api.Upstream))
    }
    
    duration := time.Since(start)
    additionalFields := []zap.Field{}
    if api != nil {
        additionalFields = append(additionalFields,
            zap.String("upstream", api.Upstream),
            zap.String("subscription", api.RequiredSubscription))
    }
    
    h.loggingMiddleware.LogResponse(r, status, duration, additionalFields...)
    
    return err
}
```

### 3. Enhanced API Operations Logging

Update API operations with detailed logging:

```go
// In the API creation/update methods
func (h *VeilHandler) handleCreateAPI(w http.ResponseWriter, r *http.Request) error {
    h.loggingMiddleware.LogRequest(r, 
        zap.String("operation", "create_api"))
    
    // ... existing API creation logic ...
    
    if err != nil {
        h.loggingMiddleware.LogError(r, err,
            zap.String("operation", "create_api"),
            zap.String("api_path", req.Path))
        return err
    }
    
    h.logger.Info("API created successfully",
        zap.String("path", req.Path),
        zap.String("upstream", req.Upstream),
        zap.String("subscription", req.RequiredSubscription),
        zap.Int("api_keys_count", len(req.APIKeys)))
    
    return nil
}
```

## Advanced Usage

### Custom Log Fields

Add custom fields to all logs from a specific component:

```go
// Create logger with custom labels
labels := map[string]string{
    "component": "api_validator",
    "version": "v1.2.3",
    "deployment": "production",
}

logger, err := logging.CreateLokiLogger(lokiURL, labels)
```

### Request Context Logging

Add request-specific context to logs:

```go
func (h *VeilHandler) validateSubscription(r *http.Request, required string, provided string) error {
    fields := []zap.Field{
        zap.String("required_subscription", required),
        zap.String("provided_subscription", provided),
        zap.String("user_agent", r.UserAgent()),
        zap.String("remote_addr", r.RemoteAddr),
    }
    
    if required == provided {
        h.logger.Info("subscription validation successful", fields...)
        return nil
    }
    
    h.logger.Warn("subscription validation failed", fields...)
    return fmt.Errorf("insufficient subscription level")
}
```

### Performance Monitoring

Log performance metrics:

```go
func (h *VeilHandler) handleUpstreamRequest(upstream string, r *http.Request) error {
    start := time.Now()
    
    // Make upstream request
    resp, err := h.client.Do(upstreamReq)
    
    // Log performance
    h.logger.Info("upstream request completed",
        zap.String("upstream", upstream),
        zap.Duration("duration", time.Since(start)),
        zap.Int("status", resp.StatusCode),
        zap.String("method", r.Method),
        zap.String("path", r.URL.Path))
    
    return err
}
```

## Environment Configuration

Set these environment variables to configure logging:

```bash
# Required
export LOKI_URL="http://localhost:3100"

# Optional
export LOG_LEVEL="info"
export LOG_LABELS='{"environment":"production","datacenter":"us-west"}'
export LOKI_TIMEOUT="10"
```

## Docker Integration

Add logging configuration to your docker-compose.yml:

```yaml
version: '3.8'
services:
  veil:
    build: .
    environment:
      - LOKI_URL=http://loki:3100
      - LOG_LEVEL=info
      - LOG_LABELS={"service":"veil-gateway","env":"docker"}
    depends_on:
      - loki
    networks:
      - veil-network
      - veil-observability
```

## Log Queries for Monitoring

Common LogQL queries for Veil monitoring:

```logql
# All requests to specific API
{job="veil-application"} | json | path="/api/v1/users"

# Failed API key validations
{job="veil-application"} | json | msg="api key validation failed"

# High response times (over 1 second)
{job="veil-application"} | json | duration > 1s

# Error rate by API path
rate({job="veil-application"} | json | level="error" [5m]) by (path)

# Subscription usage breakdown
sum by (subscription) (rate({job="veil-application"} | json | subscription != "" [5m]))
```

## Testing Integration

Test the logging integration:

```go
func TestLoggingIntegration(t *testing.T) {
    // Start test Loki instance
    lokiURL := "http://localhost:3100"
    
    // Create test logger
    labels := map[string]string{
        "test": "true",
        "component": "veil_handler_test",
    }
    
    logger, err := logging.CreateLokiLogger(lokiURL, labels)
    require.NoError(t, err)
    
    // Create logging middleware
    middleware := logging.NewLoggingMiddleware(logger)
    
    // Create test request
    req := httptest.NewRequest("GET", "/api/test", nil)
    
    // Test logging
    middleware.LogRequest(req, zap.String("test_case", "basic_request"))
    
    // Verify logs appear in Loki
    // (implementation depends on your testing setup)
}
```

## Troubleshooting

### Logs Not Appearing in Loki

1. Check Loki connectivity:
```bash
curl http://localhost:3100/ready
```

2. Check logs for errors:
```bash
docker logs veil-loki
```

3. Verify handler configuration:
```go
h.logger.Debug("Loki configuration",
    zap.String("loki_url", lokiURL),
    zap.Any("labels", labels))
```

### High Memory Usage

1. Implement log sampling for high-volume endpoints:
```go
// Sample logs for high-volume endpoints
if shouldSampleLog(r.URL.Path) {
    h.loggingMiddleware.LogRequest(r)
}
```

2. Use async logging for better performance:
```go
// Use buffered channels for async logging
go func() {
    h.loggingMiddleware.LogResponse(r, status, duration)
}()
```

## Best Practices

1. **Use Structured Fields**: Always use structured fields instead of string formatting
2. **Include Context**: Add relevant context (user ID, session ID, etc.) to logs
3. **Log at Appropriate Levels**: Use DEBUG for development, INFO for normal operations, WARN for issues, ERROR for failures
4. **Avoid Sensitive Data**: Never log passwords, API keys, or personal information
5. **Use Correlation IDs**: Include request IDs to trace requests across services
6. **Monitor Log Volume**: Implement sampling for high-volume endpoints to control costs