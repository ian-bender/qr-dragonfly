# click-service

A small Go HTTP service that provides **tracked redirect links** for QR codes.

It receives requests like `GET /r/{qrId}`, atomically increments a per-day stats row (total + per-hour counters), then redirects to the QR code's destination URL by looking it up from `qr-service`.

## Run

```bash
cd backend/click-service
export QR_SERVICE_BASE_URL=http://localhost:8080
go run ./cmd/server
```

By default, persistence is **in-memory**.

To persist clicks in Postgres, set `DATABASE_URL` (Heroku Postgres format):

```bash
export DATABASE_URL='postgres://user:pass@host:5432/dbname?sslmode=disable'
go run ./cmd/server
```

Defaults:

- `PORT=8082`
- `CORS_ALLOW_ORIGINS=http://localhost:5173` (comma-separated)
- `QR_SERVICE_BASE_URL=http://localhost:8080`

## Endpoints

- `GET /healthz` → `{ "status": "ok" }`
- `GET /r/{qrId}` → redirects (302) and records a click asynchronously
- `GET /api/clicks/{qrId}` → basic stats (all-time total + last click timestamp/country)
- `GET /api/clicks/{qrId}/daily?day=YYYY-MM-DD` → per-day stats object with per-hour click counts (UTC) and `regionCounts` JSON

## Region notes

This service captures the following headers when present (stored as the last click's country for the day):

- `CF-IPCountry` (Cloudflare)
- `X-Country` / `X-Geo-Country` (generic)

If you want full GeoIP (country/region/city), we can plug in a GeoIP DB/service later.
