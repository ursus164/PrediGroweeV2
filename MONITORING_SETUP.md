# Monitoring & Security Logging Setup

## Overview

This monitoring stack provides comprehensive observability and security logging for PrediGrowee:

- **Prometheus** - Metrics collection and storage
- **Grafana** - Visualization and dashboards
- **Loki** - Log aggregation
- **Promtail** - Log collection from containers
- **Node Exporter** - Host system metrics
- **cAdvisor** - Container metrics

## Quick Start

### 1. Start Monitoring Stack

```bash
# Start main application first
docker-compose up -d

# Start monitoring stack
docker-compose -f docker-compose.monitoring.yml up -d
```

### 2. Access Dashboards

- **Grafana**: http://localhost:3002
  - Username: `admin`
  - Password: `admin123` (change this!)

- **Prometheus**: http://prometheus:9090 (internal only)

### 3. Configure Application Logging

Add logging label to main docker-compose.yml services:

```yaml
services:
  auth:
    labels:
      logging: "promtail"
```

## Security Monitoring Features

### 1. Failed Authentication Tracking
- Monitors failed login attempts
- Alerts on suspicious patterns
- Tracks unauthorized access (401)

### 2. HTTP Error Monitoring
- 4xx client errors
- 5xx server errors
- Response time tracking

### 3. Container Security
- Resource usage monitoring
- Container health checks
- Network traffic analysis

### 4. Log Analysis
- Real-time log streaming
- Security event filtering
- Pattern matching for threats

## Prometheus Metrics

### Application Metrics (to implement in Go services)

Add Prometheus client library to your Go services:

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "endpoint", "status_code"},
    )

    authFailedAttempts = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "auth_failed_attempts_total",
            Help: "Total failed authentication attempts",
        },
    )

    dbConnectionErrors = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "db_connection_errors_total",
            Help: "Total database connection errors",
        },
    )
)

func init() {
    prometheus.MustRegister(httpRequestsTotal)
    prometheus.MustRegister(authFailedAttempts)
    prometheus.MustRegister(dbConnectionErrors)
}

// Expose /metrics endpoint
http.Handle("/metrics", promhttp.Handler())
```

### Key Metrics to Track

1. **Authentication**
   - `auth_failed_attempts_total` - Failed login attempts
   - `auth_successful_logins_total` - Successful logins
   - `auth_token_validations_total` - Token validation attempts

2. **HTTP Traffic**
   - `http_requests_total` - Total requests by status code
   - `http_request_duration_seconds` - Response times
   - `http_requests_in_flight` - Concurrent requests

3. **Database**
   - `db_connections_active` - Active connections
   - `db_query_duration_seconds` - Query performance
   - `db_connection_errors_total` - Connection failures

4. **Container Health**
   - `container_cpu_usage_seconds_total` - CPU usage
   - `container_memory_usage_bytes` - Memory usage
   - `container_network_receive_bytes_total` - Network RX
   - `container_network_transmit_bytes_total` - Network TX

## Loki Log Queries

### Security Event Queries

```logql
# Failed authentication attempts
{service="auth"} |= "failed" |= "login"

# Unauthorized access
{service=~"auth|admin"} | json | status_code="401"

# Error logs from all services
{service=~".*"} | json | level="error"

# Suspicious patterns
{service="auth"} |~ "(?i)(sql injection|xss|unauthorized|forbidden)"

# High error rate
sum(rate({service=~".*"} | json | level="error" [5m])) by (service)

# Security events with alert flag
{alert="true"}
```

### Performance Queries

```logql
# Slow requests (>1s)
{service=~".*"} | json | duration > 1s

# 5xx errors
{service="nginx"} | json | status_code =~ "5.."

# Database connection issues
{service=~"auth|quiz|stats|images|admin"} |= "database" |= "error"
```

## Grafana Dashboards

### Pre-configured Dashboards

1. **Security Dashboard** - `/var/lib/grafana/dashboards/security-dashboard.json`
   - Failed auth attempts
   - HTTP error rates
   - Container resource usage
   - Recent security events

### Custom Dashboard Creation

1. Go to Grafana (http://localhost:3002)
2. Click "+" → "Dashboard"
3. Add panels with queries from above
4. Save dashboard

### Recommended Dashboards to Import

Import these community dashboards by ID:

- **893** - Docker and system monitoring
- **1860** - Node Exporter Full
- **13639** - Loki Dashboard
- **12633** - Docker Container & Host Metrics

## Alerting (Optional)

### Add Alertmanager

1. Uncomment alerting section in `prometheus.yml`
2. Create `alertmanager.yml`:

```yaml
global:
  smtp_smarthost: 'localhost:587'
  smtp_from: 'alerts@predigrowee.com'

route:
  receiver: 'team-email'

receivers:
  - name: 'team-email'
    email_configs:
      - to: 'team@predigrowee.com'
```

3. Add to docker-compose.monitoring.yml:

```yaml
alertmanager:
  image: prom/alertmanager:latest
  ports:
    - "9093:9093"
  volumes:
    - ./monitoring/alertmanager/alertmanager.yml:/etc/alertmanager/alertmanager.yml
```

### Alert Rules

Create `prometheus/alerts/security.yml`:

```yaml
groups:
  - name: security
    interval: 30s
    rules:
      - alert: HighFailedAuthRate
        expr: rate(auth_failed_attempts_total[5m]) > 10
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High failed authentication rate"

      - alert: HighErrorRate
        expr: sum(rate(http_requests_total{status_code=~"5.."}[5m])) > 5
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High 5xx error rate"
```

## Security Best Practices

### 1. Change Default Credentials

```bash
# Update in docker-compose.monitoring.yml
- GF_SECURITY_ADMIN_PASSWORD=<strong-password>
```

### 2. Enable HTTPS

Add Nginx reverse proxy for Grafana with TLS.

### 3. Restrict Access

Use firewall to limit access:

```bash
# Allow only from localhost
sudo ufw allow from 127.0.0.1 to any port 3002
```

### 4. Regular Log Review

Schedule regular reviews of:
- Failed authentication attempts
- Unusual access patterns
- Resource usage spikes
- Error rate increases

### 5. Data Retention

Adjust retention in `loki-config.yml`:

```yaml
limits_config:
  retention_period: 744h  # 31 days (adjust as needed)
```

## Troubleshooting

### Prometheus Not Scraping

```bash
# Check Prometheus targets
curl http://localhost:9090/api/v1/targets

# Check service connectivity
docker-compose -f docker-compose.monitoring.yml exec prometheus wget -O- http://auth:8080/metrics
```

### Loki Not Receiving Logs

```bash
# Check Promtail status
docker-compose -f docker-compose.monitoring.yml logs promtail

# Verify Docker socket access
docker-compose -f docker-compose.monitoring.yml exec promtail ls -la /var/run/docker.sock
```

### Grafana Dashboard Issues

```bash
# Check datasource connection
docker-compose -f docker-compose.monitoring.yml logs grafana

# Verify Prometheus/Loki URLs
docker-compose -f docker-compose.monitoring.yml exec grafana wget -O- http://prometheus:9090/-/healthy
```

## Performance Considerations

### Resource Usage

Monitoring stack resource requirements:
- Prometheus: ~256-512MB RAM
- Grafana: ~256-512MB RAM
- Loki: ~256-512MB RAM
- Promtail: ~64-128MB RAM
- Total: ~1-2GB RAM

### Storage

- Prometheus: ~1GB per day (15s scrape interval)
- Loki: Depends on log volume (~500MB per day for typical usage)

### Optimization

1. Increase scrape intervals for less critical metrics
2. Add retention policies
3. Use recording rules for expensive queries
4. Enable compression

## Integration with CI/CD

Add monitoring checks to CI/CD:

```yaml
# .github/workflows/monitoring-check.yml
- name: Check Prometheus Targets
  run: |
    curl -sf http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | select(.health!="up")'
```

## Next Steps

1. ✅ Start monitoring stack
2. ⏳ Add Prometheus metrics to Go services
3. ⏳ Configure alerting rules
4. ⏳ Set up log rotation
5. ⏳ Create custom dashboards
6. ⏳ Integrate with incident response

## Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Loki Documentation](https://grafana.com/docs/loki/)
- [PromQL Guide](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [LogQL Guide](https://grafana.com/docs/loki/latest/logql/)
