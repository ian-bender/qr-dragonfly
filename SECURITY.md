# Security Implementation Guide

## Overview

Comprehensive security measures implemented across the application.

## 1. Rate Limiting

### Implementation

Token bucket rate limiter per IP address with configurable rates:

- **User Service**: 100 requests/minute per IP
- **QR Service**: 200 requests/minute per IP
- **Click Service**: 500 requests/minute per IP

### Configuration

Located in `backend/*/internal/middleware/ratelimit.go`

**Adjust limits in main.go:**

```go
// Example: 50 requests per 30 seconds
rateLimiter := middleware.NewRateLimiter(50, 30*time.Second)
```

### Response

When rate limited:

```json
HTTP 429 Too Many Requests
Retry-After: 1m0s
{"error":"rate_limit_exceeded"}
```

### IP Detection

Checks headers in order:

1. `X-Forwarded-For` (if behind proxy)
2. `X-Real-IP`
3. `RemoteAddr` (fallback)

### Testing

```bash
# Test rate limit
for i in {1..150}; do
  curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8081/api/users/me
done
# Should see 200s then 429s
```

## 2. Security Headers

### Headers Applied

All responses include:

```
Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; ...
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: geolocation=(), microphone=(), camera=(), payment=()
```

### Content Security Policy (CSP)

Configured to allow:

- Scripts/styles from same origin + inline
- Images from same origin, data URLs, and HTTPS
- Stripe API connections
- No iframes (clickjacking protection)

**Customize CSP** in `middleware/security.go`:

```go
w.Header().Set("Content-Security-Policy",
    "default-src 'self'; "+
    "script-src 'self' 'unsafe-inline' https://cdn.example.com; "+
    "connect-src 'self' https://api.example.com")
```

### HSTS (HTTP Strict Transport Security)

- Forces HTTPS for 1 year
- Includes subdomains
- Preload list enabled

**Only active when TLS is detected** (prevents issues in development).

### Testing

```bash
curl -I https://your-domain.com | grep -E "X-|Content-Security"
```

## 3. CSRF Protection

### Implementation

Double-submit cookie pattern:

- Cookie: `csrf_token`
- Header: `X-CSRF-Token`

### Protected Methods

- POST
- PUT
- PATCH
- DELETE

### Exemptions

- GET, HEAD, OPTIONS (read-only)
- Webhook endpoints (validated via signature)

### Frontend Integration

**Set CSRF token:**

```typescript
// After login, get CSRF token from cookie
const csrfToken = document.cookie
  .split("; ")
  .find((row) => row.startsWith("csrf_token="))
  ?.split("=")[1];

// Include in all API requests
fetch("/api/endpoint", {
  method: "POST",
  headers: {
    "X-CSRF-Token": csrfToken,
    "Content-Type": "application/json",
  },
  body: JSON.stringify(data),
});
```

### Generate CSRF Token

**In user-service/internal/httpapi/router.go:**

```go
import "crypto/rand"
import "encoding/base64"

func generateCSRFToken() string {
    b := make([]byte, 32)
    rand.Read(b)
    return base64.URLEncoding.EncodeToString(b)
}

// After successful login:
csrfToken := generateCSRFToken()
http.SetCookie(w, &http.Cookie{
    Name:     "csrf_token",
    Value:    csrfToken,
    HttpOnly: false, // Accessible to JavaScript
    Secure:   true,
    SameSite: http.SameSiteStrictMode,
    MaxAge:   3600,
})
```

### Testing

```bash
# Request without CSRF token (should fail)
curl -X POST http://localhost:8081/api/qr-codes \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com"}'
# Expected: 403 {"error":"csrf_token_missing"}

# Request with invalid token (should fail)
curl -X POST http://localhost:8081/api/qr-codes \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: invalid" \
  -b "csrf_token=different" \
  -d '{"url":"https://example.com"}'
# Expected: 403 {"error":"csrf_token_invalid"}
```

## 4. Input Sanitization

### URL Validation

**QR Service** validates all URLs:

```go
func isValidHTTPURL(rawURL string) bool {
    parsed, err := url.Parse(strings.TrimSpace(rawURL))
    if err != nil {
        return false
    }

    // Must be HTTPS
    if parsed.Scheme != "https" {
        return false
    }

    // Must have hostname
    if parsed.Host == "" {
        return false
    }

    return true
}
```

**Rejects:**

- Non-HTTPS URLs (http://, ftp://, javascript:)
- URLs without hostnames
- Malformed URLs

### Email Validation

**User Service** validates emails:

```go
import "net/mail"

func isValidEmail(email string) bool {
    _, err := mail.ParseAddress(email)
    return err == nil
}
```

### SQL Injection Protection

**Using GORM ORM** - automatically parameterizes queries:

```go
// Safe - GORM uses prepared statements
db.Where("email = ?", userInput).Find(&users)

// NEVER DO THIS - vulnerable to SQL injection
db.Raw("SELECT * FROM users WHERE email = '" + userInput + "'")
```

### String Trimming

All user input is trimmed:

```go
email := strings.TrimSpace(req.Email)
label := strings.TrimSpace(req.Label)
```

## 5. XSS Prevention

### Frontend

**Vue.js automatically escapes** all template interpolations:

```vue
<!-- Safe - automatically escaped -->
<p>{{ userInput }}</p>

<!-- UNSAFE - bypasses escaping -->
<div v-html="userInput"></div>
```

### Backend

**JSON encoding** automatically escapes dangerous characters:

```go
// Safe - json.Marshal escapes <, >, &, etc.
json.NewEncoder(w).Encode(data)
```

### Content-Type Headers

Always set proper Content-Type:

```go
w.Header().Set("Content-Type", "application/json")
```

### Testing for XSS

```bash
# Try injecting script tag
curl -X POST http://localhost:8080/api/qr-codes \
  -H "Content-Type: application/json" \
  -d '{"label":"<script>alert(1)</script>","url":"https://example.com"}'

# Verify response is escaped
curl http://localhost:8080/api/qr-codes | jq
```

## 6. Authentication Security

### Password Requirements

**Cognito enforces:**

- Minimum 8 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one number
- At least one special character

### Session Management

**JWT tokens** with:

- Short lifespan (1 hour)
- HttpOnly cookies (not accessible to JavaScript)
- Secure flag (HTTPS only in production)
- SameSite=Strict/Lax

### Token Refresh

```typescript
// Implement refresh token logic
async function refreshAccessToken() {
  const refreshToken = getCookie("refresh_token");
  const response = await fetch("/api/users/refresh", {
    method: "POST",
    body: JSON.stringify({ refreshToken }),
  });
  return response.json();
}
```

## 7. Additional Security Measures

### Prevent Path Traversal

```go
// Sanitize file paths
func safePath(userPath string) (string, error) {
    clean := filepath.Clean(userPath)
    if strings.Contains(clean, "..") {
        return "", errors.New("invalid path")
    }
    return clean, nil
}
```

### Prevent Open Redirects

```go
// Validate redirect URLs
func isValidRedirect(redirectURL string) bool {
    allowed := []string{
        "https://your-domain.com",
        "http://localhost:5173",
    }

    for _, base := range allowed {
        if strings.HasPrefix(redirectURL, base) {
            return true
        }
    }
    return false
}
```

### Secrets Management

**Never commit secrets:**

```bash
# .gitignore
.env
.env.local
.env.production
*.pem
*.key
```

**Use environment variables:**

```go
secretKey := os.Getenv("SECRET_KEY")
if secretKey == "" {
    log.Fatal("SECRET_KEY not set")
}
```

### Audit Logging

**Log security events:**

```go
log.Printf("security_event: user=%s action=%s ip=%s",
    userID, "password_change", ip)
```

## 8. Production Checklist

- [ ] Enable HTTPS/TLS with valid certificate
- [ ] Set `COOKIE_SECURE=true`
- [ ] Set `COOKIE_SAMESITE=Strict`
- [ ] Rotate secrets regularly
- [ ] Enable rate limiting on all services
- [ ] Configure CSP for your domain
- [ ] Implement CSRF protection
- [ ] Set up Web Application Firewall (WAF)
- [ ] Enable database encryption at rest
- [ ] Use VPC/private networks
- [ ] Implement IP whitelisting for admin endpoints
- [ ] Set up DDoS protection (Cloudflare, AWS Shield)
- [ ] Regular security audits
- [ ] Dependency scanning (Dependabot, Snyk)

## 9. Testing Security

### Security Scanning Tools

```bash
# OWASP ZAP
docker run -t owasp/zap2docker-stable zap-baseline.py -t https://your-domain.com

# Nikto web scanner
nikto -h https://your-domain.com

# nmap port scanning
nmap -sV your-domain.com

# SSL/TLS testing
testssl.sh https://your-domain.com
```

### Penetration Testing Checklist

- [ ] SQL injection attempts
- [ ] XSS payloads in all inputs
- [ ] CSRF token bypass
- [ ] Rate limit testing
- [ ] Authentication bypass
- [ ] Session hijacking
- [ ] Path traversal
- [ ] Open redirect
- [ ] Header injection
- [ ] CORS misconfiguration

## 10. Incident Response

### If Compromised

1. **Immediate:**

   - Rotate all secrets/credentials
   - Invalidate all active sessions
   - Enable IP blocking if needed
   - Scale down affected services

2. **Investigation:**

   - Review access logs
   - Check for data exfiltration
   - Identify attack vector
   - Assess damage scope

3. **Recovery:**

   - Patch vulnerabilities
   - Restore from clean backups
   - Deploy fixes
   - Monitor for re-compromise

4. **Post-Mortem:**
   - Document incident
   - Update security procedures
   - Notify affected users (if required)
   - Implement additional controls

## 11. Security Contacts

- **Security Email**: security@your-domain.com
- **Bug Bounty**: hackerone.com/your-company
- **Responsible Disclosure**: 90 days before public disclosure

## 12. Compliance

### GDPR

- [ ] Data encryption
- [ ] Right to deletion
- [ ] Data portability
- [ ] Consent management
- [ ] Breach notification (72 hours)

### PCI DSS (if handling payments)

- [ ] Stripe handles card data (PCI compliant)
- [ ] Never store card numbers
- [ ] Secure transmission
- [ ] Access controls

### SOC 2

- [ ] Access logging
- [ ] Encryption in transit/at rest
- [ ] Incident response plan
- [ ] Regular audits
