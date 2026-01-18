# Quota System Documentation

## Overview

The application enforces usage limits based on subscription tiers to prevent abuse and encourage upgrades.

## Quota Tiers

| Tier           | Active QR Codes | Total QR Codes | Price     |
| -------------- | --------------- | -------------- | --------- |
| **Free**       | 5               | 20             | $0/month  |
| **Basic**      | 50              | 200            | $9/month  |
| **Enterprise** | 2,000           | 10,000         | $99/month |
| **Admin**      | Unlimited       | Unlimited      | N/A       |

### Active vs Total QR Codes

- **Active QR Codes**: Currently enabled and scannable
- **Total QR Codes**: All QR codes created (active + inactive)

## Backend Implementation

### Quota Enforcement (qr-service)

Located in `backend/qr-service/internal/httpapi/router.go`:

```go
func quotaForUserType(userType string) quota {
    switch userType {
    case "basic":
        return quota{maxActive: 50, maxTotal: 200}
    case "enterprise":
        return quota{maxActive: 2000, maxTotal: 10000}
    case "admin":
        return quota{maxActive: 1_000_000_000, maxTotal: 1_000_000_000}
    case "free":
    default:
        return quota{maxActive: 5, maxTotal: 20}
    }
}
```

### Enforcement Points

#### 1. QR Code Creation

Checks **total** quota before creating:

```go
total, err := srv.Store.CountTotal()
if total >= qt.maxTotal {
    writeJSON(w, http.StatusForbidden, map[string]string{"error": "quota_total_exceeded"})
    return
}
```

If creating as **active**, also checks active quota:

```go
if requestedActive {
    active, err := srv.Store.CountActive()
    if active >= qt.maxActive {
        writeJSON(w, http.StatusForbidden, map[string]string{"error": "quota_active_exceeded"})
        return
    }
}
```

#### 2. QR Code Activation

When toggling inactive → active, checks active quota:

```go
if toggleToActive {
    active, err := srv.Store.CountActive()
    if active >= qt.maxActive {
        writeJSON(w, http.StatusForbidden, map[string]string{"error": "quota_active_exceeded"})
        return
    }
}
```

### Error Codes

| Error Code              | HTTP Status | Meaning                                           |
| ----------------------- | ----------- | ------------------------------------------------- |
| `quota_total_exceeded`  | 403         | Cannot create more QR codes (reached total limit) |
| `quota_active_exceeded` | 403         | Cannot activate QR code (reached active limit)    |
| `quota_check_failed`    | 500         | Database error while checking quotas              |

## Frontend Implementation

### Error Handling

The frontend already handles quota errors in `useQrCodes.ts`:

```typescript
function qrCodesErrorMessage(err: unknown): string {
    if (err instanceof ApiError) {
        const code = payload?.error
        switch (code) {
            case 'quota_total_exceeded':
                return 'You've reached your QR code limit for your plan.'
            case 'quota_active_exceeded':
                return 'You've reached your active QR code limit for your plan.'
        }
    }
}
```

### Display Current Usage

To show users their current usage, add to frontend:

#### 1. Add Quota API Endpoint

Add to `backend/qr-service/internal/httpapi/router.go`:

```go
// Add to router
mux.Handle("/api/qr-codes/quota", wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    userType := userTypeFromRequest(r)
    qt := quotaForUserType(userType)

    total, err := srv.Store.CountTotal()
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "count_failed"})
        return
    }

    active, err := srv.Store.CountActive()
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "count_failed"})
        return
    }

    writeJSON(w, http.StatusOK, map[string]interface{}{
        "userType": userType,
        "active": map[string]int{
            "current": active,
            "limit": qt.maxActive,
        },
        "total": map[string]int{
            "current": total,
            "limit": qt.maxTotal,
        },
    })
})))
```

#### 2. Add Frontend API Call

Create `frontend/src/api/qrCodes/quota.ts`:

```typescript
export interface QuotaResponse {
  userType: string;
  active: {
    current: number;
    limit: number;
  };
  total: {
    current: number;
    limit: number;
  };
}

export async function getQuota(): Promise<QuotaResponse> {
  return requestJson<QuotaResponse>({
    path: "/api/qr-codes/quota",
    method: "GET",
  });
}
```

#### 3. Display in UI

Add quota display component to HomePage:

```vue
<template>
  <div class="quotaDisplay">
    <div class="quotaItem">
      <span class="quotaLabel">Active QR Codes</span>
      <span class="quotaValue">
        {{ quota.active.current }} / {{ quota.active.limit }}
      </span>
      <div class="quotaBar">
        <div
          class="quotaProgress"
          :class="{
            warning: activePercentage > 80,
            danger: activePercentage >= 100,
          }"
          :style="{ width: `${Math.min(activePercentage, 100)}%` }"
        />
      </div>
    </div>

    <div class="quotaItem">
      <span class="quotaLabel">Total QR Codes</span>
      <span class="quotaValue">
        {{ quota.total.current }} / {{ quota.total.limit }}
      </span>
      <div class="quotaBar">
        <div
          class="quotaProgress"
          :class="{
            warning: totalPercentage > 80,
            danger: totalPercentage >= 100,
          }"
          :style="{ width: `${Math.min(totalPercentage, 100)}%` }"
        />
      </div>
    </div>

    <RouterLink
      v-if="activePercentage > 80 || totalPercentage > 80"
      to="/subscription"
      class="upgradeLink"
    >
      Upgrade your plan →
    </RouterLink>
  </div>
</template>

<style>
.quotaDisplay {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 24px;
}

.quotaItem {
  margin-bottom: 16px;
}

.quotaItem:last-of-type {
  margin-bottom: 0;
}

.quotaLabel {
  display: block;
  font-size: 14px;
  opacity: 0.7;
  margin-bottom: 4px;
}

.quotaValue {
  display: block;
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 8px;
}

.quotaBar {
  height: 8px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 4px;
  overflow: hidden;
}

.quotaProgress {
  height: 100%;
  background: linear-gradient(90deg, #3b82f6, #60a5fa);
  transition: width 0.3s ease, background 0.3s ease;
}

.quotaProgress.warning {
  background: linear-gradient(90deg, #f59e0b, #fbbf24);
}

.quotaProgress.danger {
  background: linear-gradient(90deg, #ef4444, #f87171);
}

.upgradeLink {
  display: inline-block;
  margin-top: 12px;
  color: #60a5fa;
  text-decoration: none;
  font-weight: 500;
}

.upgradeLink:hover {
  color: #93c5fd;
}
</style>
```

## Testing

### Unit Tests

The quota system has comprehensive tests in `backend/qr-service/internal/httpapi/quota_test.go`:

```bash
cd backend/qr-service
go test ./internal/httpapi -run TestQuota
```

### Manual Testing

#### Test Total Quota (Free Tier)

```bash
# Create 20 QR codes (free tier limit)
for i in {1..20}; do
  curl -X POST http://localhost:8080/api/qr-codes \
    -H "Content-Type: application/json" \
    -H "X-User-Type: free" \
    -d "{\"url\":\"https://example.com/$i\",\"label\":\"Test $i\"}"
done

# Try to create 21st (should fail)
curl -X POST http://localhost:8080/api/qr-codes \
  -H "Content-Type: application/json" \
  -H "X-User-Type: free" \
  -d '{"url":"https://example.com/21","label":"Test 21"}' | jq
# Expected: {"error":"quota_total_exceeded"}
```

#### Test Active Quota (Free Tier)

```bash
# Create 20 inactive QR codes
for i in {1..20}; do
  curl -X POST http://localhost:8080/api/qr-codes \
    -H "Content-Type: application/json" \
    -H "X-User-Type: free" \
    -d "{\"url\":\"https://example.com/$i\",\"label\":\"Test $i\",\"active\":false}"
done

# Try to activate 6 (free tier allows 5 active)
# Activate first 5 (should succeed)
# Activate 6th (should fail with quota_active_exceeded)
```

## Future Enhancements

### 1. Soft Limits with Grace Period

- Warn at 80% usage
- Allow temporary overage (24-48 hours)
- Email notifications

### 2. Usage Analytics

- Track quota usage over time
- Predict when users will hit limits
- Proactive upgrade suggestions

### 3. Custom Quotas

- Allow custom limits for enterprise customers
- Per-organization quotas
- Team member quotas

### 4. Usage-Based Pricing

- Pay per QR code over limit
- Pay per scan over limit
- Flexible billing

### 5. Quota Reset Options

- Monthly reset (archive old QR codes)
- Annual quotas
- Rolling window quotas
