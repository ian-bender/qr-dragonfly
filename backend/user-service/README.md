# user-service (AWS Cognito)

A small Go HTTP service for user operations backed by AWS Cognito User Pools.

## What it provides

- `POST /api/users/register` – Sign up a user (email + password + optional name + optional `userType`)
- `POST /api/users/login` – Password auth, sets HttpOnly cookies (`access_token`, `id_token`, `refresh_token` when provided)
- `POST /api/users/logout` – Global sign-out (best-effort) + clears cookies
- `GET /api/users/me` – Returns current user based on `access_token` cookie

Admin endpoints (optional; guarded by `X-Admin-Key: $ADMIN_API_KEY`):

- `GET /api/users` – List users
- `GET /api/users/{id}` – Get user by username
- `POST /api/users` – Create user (suppresses Cognito email) and optionally set permanent password (+ optional `userType`)
- `PATCH /api/users/{id}` – Update email/name/userType, set password, enable/disable
- `DELETE /api/users/{id}` – Delete user

## Configure

Required:

- `AWS_REGION`
- `COGNITO_USER_POOL_ID`
- `COGNITO_CLIENT_ID`

Also required at runtime:

- AWS credentials with permission to call Cognito User Pools (e.g. `cognito-idp:*` for the operations you use). Locally, the AWS SDK can load credentials from your shell env vars, an AWS profile in `~/.aws`, or other standard AWS credential providers.

Optional:

- `COGNITO_CLIENT_SECRET` (if your App Client has a secret)
- `CORS_ALLOW_ORIGINS` (default `http://localhost:5173`)
- `PORT` (default `8081`)
- `ADMIN_API_KEY` (enables admin endpoints)
- `COOKIE_SECURE` (default `false` for localhost)
- `COOKIE_SAMESITE` (`Lax` default; supports `Lax`, `Strict`, `None`)

## Run

```bash
cd backend/user-service
export AWS_REGION=us-east-1
export COGNITO_USER_POOL_ID=us-east-1_TppiqQvor
export COGNITO_CLIENT_ID=...

# If you aren't using an AWS profile, export credentials too:
# export AWS_ACCESS_KEY_ID=...
# export AWS_SECRET_ACCESS_KEY=...
# export AWS_SESSION_TOKEN=...

go run ./cmd/server
```

If your local Go toolchain is older than what the AWS SDK requires, use Docker instead:

```bash
docker build -t user-service:local .
docker run --rm -p 8081:8081 \
	-e AWS_REGION=us-east-1 \
	-e COGNITO_USER_POOL_ID=us-east-1_TppiqQvor \
	-e COGNITO_CLIENT_ID=... \
	-e AWS_ACCESS_KEY_ID=... \
	-e AWS_SECRET_ACCESS_KEY=... \
	-e AWS_SESSION_TOKEN=... \
	user-service:local
```

## Deploy to Heroku (Container)

This folder is ready for Heroku Container Registry deployment.

From `backend/user-service`:

```bash
heroku login
heroku container:login

heroku create <your-user-service-app>

# Required
heroku config:set AWS_REGION=us-east-1
heroku config:set COGNITO_USER_POOL_ID=us-east-1_...
heroku config:set COGNITO_CLIENT_ID=...

# Optional
heroku config:set COGNITO_CLIENT_SECRET=...
heroku config:set ADMIN_API_KEY=...
heroku config:set CORS_ALLOW_ORIGINS=https://<your-frontend-domain>
heroku config:set COOKIE_SECURE=true
heroku config:set COOKIE_SAMESITE=None

heroku container:push web
heroku container:release web
```

Heroku sets `PORT` automatically; the server binds to `:$PORT`.

## Notes

- The frontend uses `fetch(..., { credentials: 'include' })`, so this service uses HttpOnly cookies for tokens.
- For production cross-site use, you’ll likely need `COOKIE_SAMESITE=None` and `COOKIE_SECURE=true`.

## User type (Cognito custom attribute)

This service supports a simple user tier/type via a Cognito custom attribute:

- Attribute name in Cognito: `custom:user_type`
- JSON field in this API: `userType`
- Allowed values: `free`, `basic`, `enterprise`, `admin`

Important Cognito constraint:

- You must create the custom attribute in your User Pool schema first (e.g. an attribute named `user_type` of type String, marked mutable). Cognito will expose it to the API as `custom:user_type`.

Security note:

- `POST /api/users/register` will default `userType` to `free` and rejects `admin`. Only the admin endpoints can set `userType=admin`.

```

```
