# Security & Monitoring - Quick Setup Guide

This guide covers the implementation of security hardening and monitoring/logging.

## What Was Implemented

### ✅ Security (Point 4)

1. **Rate Limiting**

   - Token bucket algorithm per IP address
   - Configurable limits per service
   - Automatic cleanup of old visitors
   - Location: `backend/*/internal/middleware/ratelimit.go`

2. **Security Headers**

   - Content Security Policy (CSP)
   - X-Frame-Options (clickjacking protection)
   - X-Content-Type-Options (MIME sniffing)
   - Strict-Transport-Security (HSTS)
   - X-XSS-Protection
   - Permissions-Policy
   - Location: `backend/*/internal/middleware/security.go`

3. **CSRF Protection**

   - Double-submit cookie pattern
   - Protects POST/PUT/PATCH/DELETE
   - Webhook exemptions
   - Location: `backend/*/internal/middleware/security.go`

4. **Input Sanitization**

   - URL validation (HTTPS only)
   - Email validation
   - String trimming
   - Already implemented in services

5. **SQL Injection Protection**

   - GORM ORM with prepared statements
   - Already implemented

6. **XSS Prevention**
   - Vue.js automatic escaping
   - JSON encoding in backend
   - Already implemented

### ✅ Monitoring & Logging (Point 5)

1. **Structured Logging**

   - JSON format logs
   - Request ID tracking
   - Duration, status, IP tracking
   - Location: `backend/*/internal/middleware/logging.go`

2. **Error Tracking (Sentry)**

   - Backend integration for all services
   - Frontend Vue integration
   - Session replay for errors
   - Performance monitoring
   - Location: `backend/*/internal/monitoring/sentry.go`, `frontend/src/lib/monitoring.ts`

3. **Request Tracing**
   - X-Request-Id header propagation
   - Duration tracking
   - Error context

## Installation

### 1. Install Dependencies

**Backend:**

```bash
cd backend/user-service && go mod tidy
cd ../qr-service && go mod tidy
cd ../click-service && go mod tidy
```

**Frontend:**

```bash
cd frontend
npm install
```

### 2. Configure Sentry

1. Create account at https://sentry.io
2. Create projects for each service
3. Copy DSNs to environment variables

### 3. Set Environment Variables

Copy and update `.env.example`:

```bash
cp .env.example .env

# Required for monitoring
SENTRY_DSN=your-backend-sentry-dsn
ENVIRONMENT=development
RELEASE_VERSION=1.0.0

# Frontend (in .env.local)
VITE_SENTRY_DSN=your-frontend-sentry-dsn
VITE_ENVIRONMENT=development
```

### 4. Start Services

```bash
docker-compose up --build
```

## Middleware Order

Middleware is applied in this order (outermost to innermost):

1. **CORS** - Handle cross-origin requests
2. **Security Headers** - Add security headers to all responses
3. **Structured Logging** - Log all requests/responses
4. **Rate Limiting** - Throttle excessive requests
5. **Router** - Route to handlers

## Rate Limits

| Service       | Limit   | Window   |
| ------------- | ------- | -------- |
| User Service  | 100 req | 1 minute |
| QR Service    | 200 req | 1 minute |
| Click Service | 500 req | 1 minute |

**To adjust**, edit `cmd/server/main.go`:

```go
rateLimiter := middleware.NewRateLimiter(200, time.Minute)
```

## Testing Security

### Test Rate Limiting

```bash
# Flood requests
for i in {1..150}; do
  curl -s http://localhost:8081/healthz
done
# Should see rate limit errors after 100 requests
```

### Test Security Headers

```bash
curl -I http://localhost:8081/healthz
# Check for X-Frame-Options, CSP, etc.
```

### Test Sentry

**Backend:**

```go
monitoring.CaptureError(errors.New("Test error"), nil, nil)
```

**Frontend:**

```typescript
import { captureError } from "./lib/monitoring";
captureError(new Error("Test error"));
```

Check Sentry dashboard for errors.

### Test Structured Logs

```bash
# View logs in JSON format
docker-compose logs user-service | tail -20

# Sample output:
# {"timestamp":"2026-01-17T10:30:45Z","method":"POST","path":"/api/users/login","status":200,"duration_ms":45,"ip":"172.17.0.1","request_id":"20260117103045-abc123"}
```

## Log Aggregation (Optional)

### CloudWatch Logs

```bash
# Configure log group
aws logs create-log-group --log-group-name /qr-dragonfly/user-service

# Ship logs (configure in main.go or use agent)
```

### ELK Stack

```bash
cd monitoring
docker-compose -f docker-compose.monitoring.yml up -d

# Access Kibana at http://localhost:5601
```

### Datadog

```bash
# Install agent
DD_API_KEY=xxxxx bash -c "$(curl -L https://s3.amazonaws.com/dd-agent/scripts/install_script.sh)"

# Configure in /etc/datadog-agent/conf.d/go.d/conf.yaml
```

## Alerting Setup

### Sentry Alerts

1. Go to Sentry project settings
2. Alerts → Create Alert Rule
3. Configure:
   - Condition: Error rate > 5% in 5 minutes
   - Action: Send to Slack/Email/PagerDuty

### Uptime Monitoring

1. Sign up at https://uptimerobot.com
2. Add monitors:
   - `https://your-domain.com`
   - `https://api.your-domain.com/healthz`
3. Set check interval to 5 minutes
4. Configure email/SMS alerts

## Production Checklist

### Security

- [ ] Set `COOKIE_SECURE=true`
- [ ] Set `COOKIE_SAMESITE=Strict`
- [ ] Enable HTTPS/TLS
- [ ] Configure rate limits for production traffic
- [ ] Review CSP for your domains
- [ ] Generate strong `ADMIN_API_KEY`
- [ ] Rotate secrets regularly

### Monitoring

- [ ] Sentry DSN configured for all services
- [ ] Log aggregation set up (CloudWatch/ELK/Datadog)
- [ ] Uptime monitors configured
- [ ] Alert rules defined
- [ ] Slack/PagerDuty integration
- [ ] Dashboard created
- [ ] On-call rotation established
- [ ] Log retention policy set

## Documentation

Comprehensive guides available:

- **[SECURITY.md](./SECURITY.md)** - Full security implementation details
- **[MONITORING.md](./MONITORING.md)** - Complete monitoring guide
- **[DATABASE_SETUP.md](./DATABASE_SETUP.md)** - Database configuration
- **[QUOTA_SYSTEM.md](./QUOTA_SYSTEM.md)** - Quota enforcement
- **[EMAIL_CONFIGURATION.md](./EMAIL_CONFIGURATION.md)** - Email setup
- **[PRODUCTION_CHECKLIST.md](./PRODUCTION_CHECKLIST.md)** - Full production checklist

## Common Issues

### Rate Limit Too Strict

```go
// Increase limit
rateLimiter := middleware.NewRateLimiter(500, time.Minute)
```

### Sentry Not Capturing Errors

- Verify DSN is correct
- Check `ENVIRONMENT` is set
- Ensure `monitoring.InitSentry()` is called
- Call `monitoring.Flush()` on shutdown

### Logs Not Appearing

- Check log format is JSON
- Verify log aggregation config
- Check retention policies
- Review IAM permissions (CloudWatch)

### CORS Errors After Security Headers

- Verify CORS middleware is outermost
- Check `CORS_ALLOW_ORIGINS` includes your frontend URL
- Ensure `AllowCredentials: true` is set

## Next Steps

1. **Week 1**: Deploy with security measures, monitor for false positives
2. **Week 2**: Set up log aggregation and dashboards
3. **Week 3**: Configure alerting rules
4. **Week 4**: Fine-tune rate limits and security policies

## Support

For issues:

1. Check documentation files
2. Review Sentry error details
3. Check structured logs for request IDs
4. Contact security@your-domain.com for security concerns

## Cost Estimates

| Service         | Free Tier       | Paid Tier              |
| --------------- | --------------- | ---------------------- |
| Sentry          | 5K errors/month | $26/month (50K errors) |
| Uptime Robot    | 50 monitors     | $7/month (unlimited)   |
| CloudWatch Logs | 5GB ingestion   | $0.50/GB               |
| Datadog         | 14-day trial    | $15/host/month         |

**Estimated Total**: $0-50/month depending on scale
