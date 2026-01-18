# Monitoring & Logging Guide

## Overview

Comprehensive monitoring, error tracking, and logging infrastructure.

## 1. Error Tracking (Sentry)

### Setup

#### Create Sentry Account

1. Go to https://sentry.io and create account
2. Create new project for each service:
   - `qr-dragonfly-frontend`
   - `qr-dragonfly-user-service`
   - `qr-dragonfly-qr-service`
   - `qr-dragonfly-click-service`
3. Copy DSN for each project

#### Configure Backend

Set environment variables:

```bash
# .env
SENTRY_DSN=https://xxxxx@o123456.ingest.sentry.io/xxxxx
ENVIRONMENT=production
RELEASE_VERSION=1.0.0
```

#### Configure Frontend

Create `.env.production`:

```bash
VITE_SENTRY_DSN=https://xxxxx@o123456.ingest.sentry.io/xxxxx
VITE_ENVIRONMENT=production
VITE_RELEASE_VERSION=1.0.0
```

### Backend Integration

Already integrated in all services via `internal/monitoring/sentry.go`:

```go
import "user-service/internal/monitoring"

// Initialize on startup
monitoring.InitSentry(sentryDSN, environment, release)
defer monitoring.Flush()

// Capture errors
monitoring.CaptureError(err,
    map[string]string{"user_id": userID},
    map[string]interface{}{"request_data": data})

// Capture messages
monitoring.CaptureMessage("Important event",
    sentry.LevelWarning,
    map[string]string{"context": "value"})
```

### Frontend Integration

Already integrated in `frontend/src/lib/monitoring.ts`:

```typescript
import { initSentry, captureError, setUserContext } from "./lib/monitoring";

// Initialize in main.ts (already done)
initSentry(app, router, config);

// Capture errors
try {
  await riskyOperation();
} catch (error) {
  captureError(error as Error, { userId: user.id });
}

// Set user context after login
setUserContext({
  id: user.id,
  email: user.email,
  userType: user.userType,
});
```

### Features Enabled

**Backend:**

- Exception tracking
- Stack traces
- Performance monitoring (20% sample rate)
- Environment tagging
- Release tracking
- Custom tags and extras

**Frontend:**

- Exception tracking
- Performance monitoring (10% sample rate)
- Session replay (10% normal, 100% on error)
- User context
- Breadcrumbs (navigation, console, network)

### Testing Sentry

**Backend:**

```bash
# Trigger error
curl -X POST http://localhost:8080/api/qr-codes \
  -H "Content-Type: application/json" \
  -d '{"url":"invalid"}'

# Check Sentry dashboard for error
```

**Frontend:**

```javascript
// Add test button
<button @click="throwTestError">Test Sentry</button>

function throwTestError() {
  throw new Error('This is a test error')
}
```

### Sentry Dashboard

Access at https://sentry.io/organizations/your-org/issues/

**Key Metrics:**

- Error rate
- Affected users
- Stack traces
- Breadcrumbs
- Performance metrics

## 2. Structured Logging

### Implementation

All services use structured JSON logging via `middleware/logging.go`:

```json
{
  "timestamp": "2026-01-17T10:30:45Z",
  "method": "POST",
  "path": "/api/qr-codes",
  "status": 201,
  "duration_ms": 45,
  "ip": "192.168.1.1",
  "user_agent": "Mozilla/5.0...",
  "request_id": "20260117103045-abc123"
}
```

### Log Aggregation

#### Option 1: CloudWatch Logs (AWS)

```bash
# Install CloudWatch agent
wget https://s3.amazonaws.com/amazoncloudwatch-agent/ubuntu/amd64/latest/amazon-cloudwatch-agent.deb
sudo dpkg -i amazon-cloudwatch-agent.deb

# Configure log group
aws logs create-log-group --log-group-name /qr-dragonfly/user-service
```

**Configure in /opt/aws/amazon-cloudwatch-agent/etc/config.json:**

```json
{
  "logs": {
    "logs_collected": {
      "files": {
        "collect_list": [
          {
            "file_path": "/var/log/user-service/*.log",
            "log_group_name": "/qr-dragonfly/user-service",
            "log_stream_name": "{instance_id}"
          }
        ]
      }
    }
  }
}
```

#### Option 2: ELK Stack (Self-Hosted)

**docker-compose.monitoring.yml:**

```yaml
version: "3.8"

services:
  elasticsearch:
    image: elasticsearch:8.11.0
    environment:
      - discovery.type=single-node
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
    ports:
      - "9200:9200"
    volumes:
      - elasticsearch-data:/usr/share/elasticsearch/data

  logstash:
    image: logstash:8.11.0
    ports:
      - "5000:5000"
    volumes:
      - ./logstash/pipeline:/usr/share/logstash/pipeline
    depends_on:
      - elasticsearch

  kibana:
    image: kibana:8.11.0
    ports:
      - "5601:5601"
    environment:
      - ELASTICSEARCH_HOSTS=http://elasticsearch:9200
    depends_on:
      - elasticsearch

volumes:
  elasticsearch-data:
```

**Logstash pipeline (logstash/pipeline/logstash.conf):**

```
input {
  tcp {
    port => 5000
    codec => json_lines
  }
}

filter {
  json {
    source => "message"
  }

  date {
    match => ["timestamp", "ISO8601"]
    target => "@timestamp"
  }
}

output {
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    index => "qr-dragonfly-%{+YYYY.MM.dd}"
  }
}
```

**Ship logs to Logstash:**

```go
// In main.go
logstashAddr := os.Getenv("LOGSTASH_ADDR")
if logstashAddr != "" {
    conn, _ := net.Dial("tcp", logstashAddr)
    log.SetOutput(conn)
}
```

#### Option 3: Datadog

```bash
# Install Datadog agent
DD_API_KEY=xxxxx DD_SITE="datadoghq.com" bash -c "$(curl -L https://s3.amazonaws.com/dd-agent/scripts/install_script.sh)"

# Configure log collection
sudo vi /etc/datadog-agent/conf.d/go.d/conf.yaml
```

```yaml
logs:
  - type: file
    path: /var/log/qr-dragonfly/*.log
    service: qr-dragonfly
    source: go
    sourcecategory: custom
```

### Log Queries

**CloudWatch Insights:**

```sql
# Error rate by endpoint
fields @timestamp, path, status
| filter status >= 400
| stats count() by path
| sort count desc

# Slow requests
fields @timestamp, path, duration_ms
| filter duration_ms > 1000
| sort duration_ms desc

# Requests by IP
fields @timestamp, ip, method, path
| stats count() by ip
| sort count desc
```

**Elasticsearch:**

```json
GET /qr-dragonfly-*/_search
{
  "query": {
    "bool": {
      "must": [
        { "range": { "status": { "gte": 400 } } },
        { "range": { "@timestamp": { "gte": "now-1h" } } }
      ]
    }
  },
  "aggs": {
    "error_paths": {
      "terms": { "field": "path.keyword" }
    }
  }
}
```

## 3. Application Performance Monitoring (APM)

### Option 1: Sentry Performance

Already enabled with `TracesSampleRate: 0.2`

**Features:**

- Transaction tracing
- Database query performance
- HTTP request duration
- Custom spans

**Add custom spans:**

```go
import "github.com/getsentry/sentry-go"

span := sentry.StartSpan(ctx, "database.query")
defer span.Finish()

// ... database operation ...

span.SetData("rows_affected", rowsAffected)
```

### Option 2: Datadog APM

```go
import "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

// Initialize
tracer.Start(
    tracer.WithService("user-service"),
    tracer.WithEnv("production"),
)
defer tracer.Stop()

// Trace HTTP handlers
http.Handle("/api/endpoint",
    tracer.WrapHandler(handler, "web.request", "/api/endpoint"))
```

### Option 3: New Relic

```go
import "github.com/newrelic/go-agent/v3/newrelic"

app, _ := newrelic.NewApplication(
    newrelic.ConfigAppName("user-service"),
    newrelic.ConfigLicense(os.Getenv("NEW_RELIC_LICENSE_KEY")),
)

// Instrument transactions
txn := app.StartTransaction("handleRequest")
defer txn.End()
```

### Key Metrics to Monitor

**Backend:**

- Request duration (p50, p95, p99)
- Error rate
- Throughput (requests/second)
- Database query time
- Memory usage
- CPU usage
- Goroutine count

**Frontend:**

- Page load time
- Time to interactive (TTI)
- First contentful paint (FCP)
- API request duration
- JavaScript errors
- Bundle size

## 4. Uptime Monitoring

### Option 1: UptimeRobot (Free)

1. Go to https://uptimerobot.com
2. Add HTTP(s) monitors:
   - `https://your-domain.com`
   - `https://api.your-domain.com/healthz`
3. Configure alerts (email, SMS, Slack)
4. Set check interval (1-5 minutes)

### Option 2: Pingdom

```bash
# Configure monitors via API
curl -X POST https://api.pingdom.com/api/3.1/checks \
  -u "username:password" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "QR Dragonfly API",
    "type": "http",
    "host": "api.your-domain.com",
    "url": "/healthz"
  }'
```

### Option 3: Custom Health Checks

**Health check endpoint** (already exists):

```go
// /healthz endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
    // Check database
    if err := db.Ping(); err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]string{
            "status": "unhealthy",
            "error": "database unreachable"
        })
        return
    }

    json.NewEncoder(w).Encode(map[string]string{
        "status": "healthy",
        "version": version
    })
}
```

**Monitor script:**

```bash
#!/bin/bash
# monitor.sh

ENDPOINT="https://api.your-domain.com/healthz"
SLACK_WEBHOOK="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"

while true; do
    status=$(curl -s -o /dev/null -w "%{http_code}" $ENDPOINT)

    if [ $status -ne 200 ]; then
        curl -X POST $SLACK_WEBHOOK \
            -H 'Content-Type: application/json' \
            -d "{\"text\":\"🚨 Health check failed: HTTP $status\"}"
    fi

    sleep 60
done
```

## 5. Alerting

### Critical Alerts

**Error Rate > 5%:**

```bash
# CloudWatch Alarm
aws cloudwatch put-metric-alarm \
  --alarm-name high-error-rate \
  --alarm-description "Alert when error rate exceeds 5%" \
  --metric-name 4xxErrors \
  --namespace AWS/ApplicationELB \
  --statistic Sum \
  --period 300 \
  --threshold 5 \
  --comparison-operator GreaterThanThreshold \
  --evaluation-periods 1 \
  --alarm-actions arn:aws:sns:region:account:topic
```

**Response Time > 2s:**

```bash
aws cloudwatch put-metric-alarm \
  --alarm-name slow-response \
  --metric-name TargetResponseTime \
  --namespace AWS/ApplicationELB \
  --statistic Average \
  --period 60 \
  --threshold 2 \
  --comparison-operator GreaterThanThreshold \
  --evaluation-periods 2
```

### Notification Channels

**Slack Integration:**

```bash
# Create Slack incoming webhook
# https://api.slack.com/messaging/webhooks

# Send alert
curl -X POST https://hooks.slack.com/services/YOUR/WEBHOOK \
  -H 'Content-Type: application/json' \
  -d '{
    "text": "🚨 High error rate detected",
    "attachments": [{
      "color": "danger",
      "fields": [
        {"title": "Service", "value": "user-service", "short": true},
        {"title": "Error Rate", "value": "8.5%", "short": true}
      ]
    }]
  }'
```

**PagerDuty Integration:**

```go
import "github.com/PagerDuty/go-pagerduty"

client := pagerduty.NewClient(apiKey)
event := pagerduty.V2Event{
    RoutingKey: integrationKey,
    Action:     "trigger",
    Payload: &pagerduty.V2Payload{
        Summary:  "High error rate",
        Severity: "error",
        Source:   "user-service",
    },
}
client.ManageEvent(&event)
```

## 6. Dashboards

### Grafana Dashboard

**docker-compose.monitoring.yml:**

```yaml
grafana:
  image: grafana/grafana:10.2.0
  ports:
    - "3000:3000"
  environment:
    - GF_SECURITY_ADMIN_PASSWORD=admin
  volumes:
    - grafana-data:/var/lib/grafana
    - ./grafana/dashboards:/etc/grafana/provisioning/dashboards
```

**Dashboard JSON (grafana/dashboards/overview.json):**

```json
{
  "dashboard": {
    "title": "QR Dragonfly Overview",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{service}}"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total{status=~\"5..\"}[5m])"
          }
        ]
      },
      {
        "title": "Response Time (p95)",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, http_request_duration_seconds)"
          }
        ]
      }
    ]
  }
}
```

### Custom Metrics

**Expose Prometheus metrics:**

```go
import "github.com/prometheus/client_golang/prometheus"
import "github.com/prometheus/client_golang/prometheus/promhttp"

var (
    httpRequests = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    httpDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration",
        },
        []string{"method", "path"},
    )
)

func init() {
    prometheus.MustRegister(httpRequests)
    prometheus.MustRegister(httpDuration)
}

// Expose metrics endpoint
http.Handle("/metrics", promhttp.Handler())
```

## 7. Log Retention

### CloudWatch

```bash
# Set retention to 30 days
aws logs put-retention-policy \
  --log-group-name /qr-dragonfly/user-service \
  --retention-in-days 30
```

### Elasticsearch

```bash
# Delete logs older than 90 days
curator_cli delete_indices \
  --filter_list '[{"filtertype":"age","source":"creation_date","direction":"older","unit":"days","unit_count":90}]'
```

## 8. Cost Optimization

### Log Sampling

```go
// Sample 10% of successful requests
if status < 400 && rand.Float64() > 0.1 {
    return // Don't log
}
```

### Metric Aggregation

```go
// Send metrics in batches
type MetricBatch struct {
    metrics []Metric
    mu      sync.Mutex
}

func (b *MetricBatch) Add(m Metric) {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.metrics = append(b.metrics, m)

    if len(b.metrics) >= 100 {
        b.Flush()
    }
}
```

## 9. Monitoring Checklist

- [ ] Sentry configured for all services
- [ ] Structured logging enabled
- [ ] Log aggregation set up
- [ ] APM tool integrated
- [ ] Uptime monitors configured
- [ ] Alerting rules defined
- [ ] Dashboard created
- [ ] On-call rotation established
- [ ] Runbook documented
- [ ] Log retention policy set
- [ ] Cost budget defined
- [ ] Weekly metric reviews scheduled

## 10. Troubleshooting

### High Error Rate

1. Check Sentry for error patterns
2. Review recent deployments
3. Check dependency health
4. Review database performance
5. Check rate limiting logs

### Slow Performance

1. Check APM traces
2. Identify slow database queries
3. Review cache hit rates
4. Check external API latency
5. Profile CPU/memory usage

### Service Down

1. Check health endpoint
2. Review server logs
3. Verify database connectivity
4. Check resource utilization
5. Review recent config changes

## 11. Resources

- **Sentry Docs**: https://docs.sentry.io
- **Grafana Docs**: https://grafana.com/docs
- **Prometheus**: https://prometheus.io/docs
- **ELK Stack**: https://www.elastic.co/guide
- **CloudWatch**: https://docs.aws.amazon.com/cloudwatch
