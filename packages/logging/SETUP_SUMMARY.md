# Veil Logging Infrastructure - Complete Setup Summary

This document provides a comprehensive overview of the Loki and Grafana logging integration that has been implemented for the Veil project.

## 🎯 What Was Implemented

### 1. Observability Stack
- **Loki**: Log aggregation and storage
- **Grafana**: Visualization and dashboards
- **Promtail**: Log collection agent
- **Docker Compose**: Container orchestration

### 2. Go Logging Package
- **Structured Logging**: JSON-formatted logs with zap
- **Loki Hook**: Direct log shipping to Loki
- **Middleware**: HTTP request/response logging
- **API Integration**: Ready-to-use logging for Caddy handlers

### 3. Monitoring Dashboards
- **API Gateway Dashboard**: Pre-configured Grafana dashboard
- **Real-time Metrics**: Request rates, response times, error rates
- **Business Metrics**: Usage by API path and subscription level

### 4. Developer Tools
- **Automated Setup**: Scripts and Makefiles for easy deployment
- **Configuration Management**: Environment-based configuration
- **Health Checks**: Service monitoring and validation

## 📁 Directory Structure

```
packages/logging/
├── README.md                          # Main documentation
├── docker-compose.yml                 # Observability stack
├── Makefile                          # Management commands
├── go.mod                            # Go module definition
├── integration_guide.md              # Developer integration guide
├── SETUP_SUMMARY.md                  # This file
├── .env.example                      # Environment configuration example
├── config/
│   ├── loki-config.yaml              # Loki configuration
│   ├── promtail-config.yaml          # Promtail configuration
│   └── grafana/
│       ├── provisioning/
│       │   ├── datasources/
│       │   │   └── loki.yml          # Loki datasource
│       │   └── dashboards/
│       │       └── dashboard.yml     # Dashboard provisioning
│       └── dashboards/
│           └── veil-api-gateway.json # Pre-built dashboard
├── scripts/
│   └── start-observability.sh       # Startup script
├── loki_hook.go                      # Loki integration for zap
└── middleware.go                     # HTTP logging middleware
```

## 🚀 Quick Start

### 1. Start the Observability Stack

```bash
cd packages/logging
make quick-start
```

This command will:
- Install and validate dependencies
- Build the Go logging module
- Start Loki, Grafana, and Promtail
- Wait for services to be ready
- Display access information

### 2. Access Grafana

- **URL**: http://localhost:3000
- **Username**: admin
- **Password**: admin

The Veil API Gateway dashboard will be automatically available.

### 3. Service Endpoints

- **Grafana Dashboard**: http://localhost:3000
- **Loki API**: http://localhost:3100
- **Promtail Metrics**: http://localhost:9080

## 🔧 Configuration

### Environment Variables

Copy `.env.example` to `.env` and customize:

```bash
cp .env.example .env
```

Key configuration options:
- `LOKI_URL`: Loki endpoint (default: http://localhost:3100)
- `LOG_LEVEL`: Logging level (debug, info, warn, error)
- `LOG_LABELS`: Additional labels in JSON format
- `LOG_RETENTION_DAYS`: Log retention period

### Docker Compose Override

Create `docker-compose.override.yml` for custom configurations:

```yaml
version: '3.8'
services:
  grafana:
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=your_secure_password
  loki:
    volumes:
      - ./custom-loki-config.yaml:/etc/loki/local-config.yaml
```

## 📊 Features

### Structured Logging
- JSON-formatted logs with consistent fields
- Automatic extraction of HTTP request details
- Support for custom fields and labels
- Integration with zap logger

### Real-time Dashboards
- Request rate and volume metrics
- Response time percentiles (95th, 99th)
- Error rate tracking
- API usage by path and subscription level
- Live error log viewer

### Log Queries
Pre-configured LogQL queries for common monitoring needs:

```logql
# High response times
{job="veil-application"} | json | duration > 1s

# Failed API validations
{job="veil-application"} | json | msg="api key validation failed"

# Usage by subscription level
sum by (subscription) (rate({job="veil-application"} | json [5m]))
```

### Performance Features
- Async log shipping to prevent blocking
- Configurable batch sizes and timeouts
- Log sampling for high-volume endpoints
- Automatic retry and fallback mechanisms

## 🔌 Integration with Veil Handler

The logging package integrates seamlessly with the existing Veil Caddy handler:

### Key Integration Points

1. **Request Logging**: Every HTTP request is logged with structured fields
2. **API Validation**: Success/failure of API key validations
3. **Performance Tracking**: Response times and upstream latencies
4. **Error Handling**: Detailed error context and stack traces
5. **Business Metrics**: Usage patterns by subscription level

### Log Schema

Each log entry includes:
```json
{
  "ts": "2024-01-15T10:30:45Z",
  "level": "info",
  "msg": "request completed",
  "method": "GET",
  "path": "/api/v1/users",
  "status": 200,
  "duration": "45.123ms",
  "subscription": "premium",
  "upstream": "http://localhost:8080",
  "remote_addr": "192.168.1.100:54321"
}
```

## 🛠️ Management Commands

### Makefile Targets

```bash
make help              # Show all available commands
make start             # Start the observability stack
make stop              # Stop the observability stack
make status            # Check service status
make logs              # View all service logs
make health-check      # Check service health
make config-validate   # Validate configurations
make clean             # Clean up resources
```

### Script Options

```bash
./scripts/start-observability.sh          # Start stack
./scripts/start-observability.sh --stop   # Stop stack
./scripts/start-observability.sh --status # Check status
```

## 🔍 Monitoring and Alerting

### Built-in Dashboards

1. **Veil API Gateway**: Main operational dashboard
   - Request rates and patterns
   - Response time distributions
   - Error rates and status codes
   - API usage breakdowns

### Custom Metrics

Create custom panels using LogQL:
- API usage trends
- Performance degradation alerts
- Security event monitoring
- Business KPI tracking

### Health Checks

Automated health monitoring for:
- Loki ingestion pipeline
- Grafana dashboard availability
- Log shipping connectivity
- Storage and retention policies

## 🚨 Troubleshooting

### Common Issues

1. **Services Not Starting**
   ```bash
   make ports-check  # Check port conflicts
   make logs         # View error logs
   ```

2. **Logs Not Appearing**
   ```bash
   curl http://localhost:3100/ready  # Check Loki health
   make health-check                 # Full health check
   ```

3. **Performance Issues**
   - Adjust batch sizes in configuration
   - Implement log sampling
   - Check retention policies

### Debug Mode

Enable debug logging:
```bash
export LOG_LEVEL=debug
make restart
```

### Log Validation

Test log ingestion:
```bash
# Send test log to Loki
curl -X POST http://localhost:3100/loki/api/v1/push \
  -H "Content-Type: application/json" \
  -d '{"streams":[{"stream":{"job":"test"},"values":[["'$(date +%s%N)'","test message"]]}]}'
```

## 📈 Performance Considerations

### Resource Requirements
- **Loki**: 512MB RAM minimum, 1GB recommended
- **Grafana**: 256MB RAM minimum, 512MB recommended
- **Promtail**: 128MB RAM minimum, 256MB recommended

### Storage Planning
- Log retention: 30 days default (configurable)
- Estimated storage: ~1GB per million log entries
- Compression: ~10:1 ratio with Loki

### Scaling Recommendations
- For high-volume environments (>10K RPS):
  - Use multiple Loki replicas
  - Implement log sampling
  - Consider log aggregation strategies

## 🔐 Security Considerations

### Access Control
- Grafana admin password protection
- Network isolation with Docker networks
- Optional TLS encryption for Loki API

### Data Privacy
- No sensitive data logging (passwords, keys)
- IP address anonymization options
- GDPR-compliant retention policies

### Audit Trail
- All configuration changes logged
- Access monitoring via Grafana
- Log integrity verification

## 🔄 Backup and Recovery

### Configuration Backup
```bash
make backup-config  # Create timestamped backup
```

### Data Recovery
- Loki data stored in Docker volumes
- Grafana dashboards version controlled
- Configuration restored from backups

## 📚 Documentation Links

- [README.md](./README.md): Main documentation
- [integration_guide.md](./integration_guide.md): Developer integration guide
- [Grafana Documentation](https://grafana.com/docs/grafana/latest/)
- [Loki Documentation](https://grafana.com/docs/loki/latest/)
- [LogQL Query Language](https://grafana.com/docs/loki/latest/logql/)

## 🎉 Success Metrics

After implementation, you should expect to see:

1. **Operational Visibility**
   - Real-time request monitoring
   - Performance trend analysis
   - Error detection and alerting

2. **Developer Experience**
   - Structured log searching
   - Request tracing capabilities
   - Performance bottleneck identification

3. **Business Insights**
   - API usage patterns
   - Subscription level analysis
   - Customer behavior tracking

## 🚀 Next Steps

1. **Integrate with Existing Handler**: Follow the integration guide
2. **Customize Dashboards**: Add business-specific metrics
3. **Set Up Alerting**: Configure Grafana alerts for critical events
4. **Performance Tuning**: Optimize based on actual load patterns
5. **Extend Monitoring**: Add application-specific log sources

---

The Veil logging infrastructure is now ready for production use. Start with `make quick-start` and begin monitoring your API gateway with enterprise-grade observability!