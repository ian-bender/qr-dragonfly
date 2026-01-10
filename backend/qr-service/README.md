# qr-service

A small Go HTTP service that implements **QR code CRUD** (no user/account endpoints).

This is designed to be moved to its own repo later.

## Run

```bash
cd backend/qr-service
go run ./cmd/server
```

By default, persistence is **in-memory**.

To persist QR codes in Postgres, set `DATABASE_URL` (Heroku Postgres format):

```bash
export DATABASE_URL='postgres://user:pass@host:5431/dbname?sslmode=disable'
go run ./cmd/server
```

## Deploy to Heroku (Container)

This folder is ready for Heroku Container Registry deployment.

From `backend/qr-service`:

```bash
heroku login
heroku container:login

heroku create <your-qr-service-app>

# Create/attach Heroku Postgres (sets DATABASE_URL automatically)
heroku addons:create heroku-postgresql:essential-0

heroku config:set CORS_ALLOW_ORIGINS=https://<your-frontend-domain>

heroku container:push web
heroku container:release web
```

Heroku sets `PORT` automatically; the server binds to `:$PORT`.

Defaults:

- `PORT=8080`
- `CORS_ALLOW_ORIGINS=http://localhost:5173` (comma-separated)

## API

Base path: `/api/qr-codes`

- `GET /api/qr-codes/` → list
- `POST /api/qr-codes/` → create
- `GET /api/qr-codes/{id}/` → get
- `PATCH /api/qr-codes/{id}/` → update
- `DELETE /api/qr-codes/{id}/` → delete

### Create

`POST /api/qr-codes/`

Body:

```json
{ "label": "Landing", "url": "https://example.com" }
```

Response:

```json
{
  "id": "...",
  "label": "Landing",
  "url": "https://example.com",
  "createdAtIso": "2025-12-26T00:00:00Z"
}
```

## Notes

- If `DATABASE_URL` is set, the service stores QR codes in Postgres.
- If `DATABASE_URL` is not set, the service uses an in-memory store.
- The backend does **not** generate image QR data URLs; the frontend can continue using `qrcode`.
