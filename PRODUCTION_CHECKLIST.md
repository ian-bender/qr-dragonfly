# Production Readiness Checklist

Quick reference for making the app production-ready.

## ✅ 1. Database Persistence (COMPLETE)

- [x] PostgreSQL stores implemented for QR service
- [x] PostgreSQL stores implemented for Click service
- [x] Auto-migration on startup
- [x] Connection pooling configured
- [ ] **TODO: Set DATABASE_URL environment variable**
- [ ] **TODO: Set up automated backups**
- [ ] **TODO: Configure monitoring**

**Action Items:**

```bash
# 1. Create PostgreSQL database
createdb qrdragonfly

# 2. Set environment variable
export DATABASE_URL="postgresql://user:password@localhost:5432/qrdragonfly"

# 3. Restart services (will auto-migrate)
docker-compose up -d
```

See [DATABASE_SETUP.md](./DATABASE_SETUP.md) for detailed instructions.

## ✅ 2. Quota Enforcement (COMPLETE)

- [x] Quota limits defined per tier
- [x] Total QR code quota enforced on creation
- [x] Active QR code quota enforced on creation/activation
- [x] Proper error messages returned
- [x] Frontend error handling implemented
- [x] Unit tests written
- [ ] **TODO: Display current usage in UI**
- [ ] **TODO: Add upgrade prompts at 80% usage**

**Action Items:**

1. Add quota endpoint to backend (see QUOTA_SYSTEM.md)
2. Add quota display component to frontend
3. Add upgrade CTAs when approaching limits

See [QUOTA_SYSTEM.md](./QUOTA_SYSTEM.md) for implementation guide.

## ⚠️ 3. Email Configuration (NEEDS SETUP)

- [ ] **TODO: Verify domain in AWS SES**
- [ ] **TODO: Configure DNS records (SPF, DKIM, DMARC)**
- [ ] **TODO: Update Cognito to use SES**
- [ ] **TODO: Customize email templates**
- [ ] **TODO: Test email delivery**
- [ ] **TODO: Request SES production access**
- [ ] **TODO: Set up bounce/complaint monitoring**

**Action Items:**

1. Go to AWS SES Console
2. Verify your domain
3. Add DNS records provided by SES
4. Update Cognito User Pool email config
5. Customize verification/reset templates
6. Test with real email addresses

See [EMAIL_CONFIGURATION.md](./EMAIL_CONFIGURATION.md) for detailed instructions.

## Additional Priority Items

### Security (High Priority)

- [ ] Add rate limiting middleware
- [ ] Implement CSRF protection
- [ ] Add security headers
- [ ] Enable HTTPS/TLS
- [ ] Rotate secrets regularly
- [ ] Add input sanitization
- [ ] Implement session timeout

### Monitoring (High Priority)

- [ ] Set up error tracking (Sentry/Rollbar)
- [ ] Add application monitoring (DataDog/New Relic)
- [ ] Configure log aggregation
- [ ] Set up uptime monitoring
- [ ] Create alerting rules
- [ ] Add performance monitoring

### User Experience (Medium Priority)

- [ ] Add loading states everywhere
- [ ] Implement toast notifications
- [ ] Add confirmation dialogs for destructive actions
- [ ] Create onboarding tutorial
- [ ] Add help documentation
- [ ] Improve error messages

### Features (Medium Priority)

- [ ] Admin dashboard
- [ ] Account deletion (GDPR)
- [ ] Data export (GDPR)
- [ ] Two-factor authentication
- [ ] Session management
- [ ] Invoice history

### Infrastructure (Low Priority)

- [ ] Set up CI/CD pipeline
- [ ] Implement blue-green deployments
- [ ] Configure auto-scaling
- [ ] Set up CDN
- [ ] Multi-region deployment
- [ ] Disaster recovery plan

## Quick Start Production Deployment

### 1. Environment Variables

Create `.env.production`:

```bash
# Database
DATABASE_URL=postgresql://user:password@host:5432/dbname

# Cognito
AWS_REGION=us-east-1
COGNITO_USER_POOL_ID=us-east-1_xxxxx
COGNITO_CLIENT_ID=xxxxxxxxxxxxx
COGNITO_CLIENT_SECRET=xxxxxxxxxxxxx

# Stripe
STRIPE_SECRET_KEY=sk_live_xxxxx
STRIPE_WEBHOOK_SECRET=whsec_xxxxx
STRIPE_BASIC_PRICE_ID=price_xxxxx
STRIPE_ENTERPRISE_PRICE_ID=price_xxxxx
STRIPE_SUCCESS_URL=https://yourdomain.com/subscription?success=true
STRIPE_CANCEL_URL=https://yourdomain.com/subscription
STRIPE_PORTAL_RETURN_URL=https://yourdomain.com/account

# CORS
CORS_ALLOW_ORIGINS=https://yourdomain.com

# Security
COOKIE_SECURE=true
COOKIE_SAMESITE=Strict
ADMIN_API_KEY=generate-a-secure-random-key

# Ports (if deploying to different servers)
QR_SERVICE_PORT=8080
CLICK_SERVICE_PORT=8082
USER_SERVICE_PORT=8081
```

### 2. Deploy Backend Services

```bash
# Build
docker-compose -f docker-compose.prod.yml build

# Deploy
docker-compose -f docker-compose.prod.yml up -d

# Verify
docker-compose -f docker-compose.prod.yml ps
docker-compose -f docker-compose.prod.yml logs
```

### 3. Deploy Frontend

```bash
cd frontend
npm run build

# Deploy to hosting (choose one):
# - Vercel: vercel --prod
# - Netlify: netlify deploy --prod
# - S3 + CloudFront: aws s3 sync dist/ s3://your-bucket/
```

### 4. Verify Deployment

```bash
# Check health endpoints
curl https://api.yourdomain.com/healthz

# Test authentication
curl -X POST https://api.yourdomain.com/api/users/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password"}'

# Test QR creation (with auth cookie)
curl -X POST https://api.yourdomain.com/api/qr-codes \
  -H "Content-Type: application/json" \
  -b "access_token=xxx" \
  -d '{"url":"https://example.com","label":"Test"}'
```

### 5. Set Up Monitoring

```bash
# Add Sentry (example)
# Frontend
npm install @sentry/vue

# Backend
go get github.com/getsentry/sentry-go

# Configure in code with your DSN
```

## Testing Checklist

- [ ] User registration flow
- [ ] Email verification
- [ ] Login/logout
- [ ] Password reset
- [ ] QR code creation
- [ ] QR code editing
- [ ] QR code deletion
- [ ] QR code activation/deactivation
- [ ] Click tracking
- [ ] Analytics viewing
- [ ] Subscription purchase
- [ ] Subscription management
- [ ] Quota enforcement
- [ ] Error handling
- [ ] Mobile responsiveness

## Security Audit

- [ ] No secrets in code
- [ ] Environment variables used correctly
- [ ] HTTPS enforced
- [ ] CORS configured properly
- [ ] SQL injection prevented
- [ ] XSS prevented
- [ ] CSRF protection enabled
- [ ] Rate limiting active
- [ ] Password requirements enforced
- [ ] Session timeout configured
- [ ] Audit logging enabled

## Performance Checks

- [ ] Database indexes optimized
- [ ] API response times < 200ms
- [ ] Frontend bundle size < 500KB
- [ ] Images optimized
- [ ] CDN configured
- [ ] Caching headers set
- [ ] Database connection pooling
- [ ] No N+1 queries

## Compliance

- [ ] Terms of Service published
- [ ] Privacy Policy published
- [ ] Cookie Policy published
- [ ] GDPR compliance (EU users)
- [ ] CCPA compliance (CA users)
- [ ] Cookie consent banner
- [ ] Data export feature
- [ ] Account deletion feature

## Timeline Estimate

| Phase                       | Duration     | Priority |
| --------------------------- | ------------ | -------- |
| Database + Monitoring Setup | 1 day        | Critical |
| Email Configuration         | 2 days       | High     |
| Security Hardening          | 3 days       | High     |
| UI Improvements             | 3 days       | Medium   |
| Testing & QA                | 2 days       | High     |
| Documentation               | 1 day        | Medium   |
| **Total**                   | **~2 weeks** | -        |

## Support Resources

- [DATABASE_SETUP.md](./DATABASE_SETUP.md) - PostgreSQL configuration
- [QUOTA_SYSTEM.md](./QUOTA_SYSTEM.md) - Quota implementation
- [EMAIL_CONFIGURATION.md](./EMAIL_CONFIGURATION.md) - Email setup
- [README.md](./README.md) - General documentation
- [docker-compose.yml](./docker-compose.yml) - Local development
