# image-code

Monorepo workspace:

- `frontend/` – Vue 3 + TypeScript + Vite app
- `backend/qr-service/` – Go QR CRUD service (supports Postgres via `DATABASE_URL`)
- `backend/user-service/` – Go user service backed by AWS Cognito

## Run frontend

```bash
cd frontend
npm run dev
```

Vite runs at `http://localhost:5173`.

## Run backends locally (Docker)

From the repo root:

```bash
docker compose up -d --build
```

- QR service: `http://localhost:8080/healthz`
- Click service: `http://localhost:8082/healthz`
- User service: `http://localhost:8081/healthz` (requires Cognito env vars; see below)
- Postgres is mapped to host `localhost:5433` to avoid conflicts with a local Postgres on `5432`.

To run `user-service`, copy `.env.example` → `.env` and set your Cognito values (required by `user-service` on startup).
`ADMIN_API_KEY` is optional; if unset, admin endpoints return `admin_disabled`.

Then rerun:

```bash
docker compose up -d --build
```
