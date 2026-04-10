# Gin Core Service

Service nay la backend chinh cho:

- auth end-user
- join game
- place bet
- session tren connection
- outbox event publishing
- worker trigger sau nay

## Cau truc

```text
cmd/api
internal/app
internal/auth
internal/domain
internal/event/outbox
internal/platform/postgres
internal/repository/postgres
internal/service
internal/transport/http
internal/ws
```

## Endpoint scaffold

- `GET /healthz`
- `POST /v1/auth/register`
- `POST /v1/auth/login`
- `GET /v1/auth/me`
- `POST /v1/auth/forgot-password`
- `POST /v1/auth/forgot-password/verify-otp`
- `POST /v1/auth/reset-password`
- `POST /v1/games/{game}/join`
- `POST /v1/games/{game}/bets`

## Env toi thieu

- `DATABASE_URL`
- `REDIS_ADDR`
- `GATE_BASE_URL`
- `AUTH_TOKEN_SECRET`
- `AUTH_TOKEN_TTL=24h`
- `AUTH_FORGOT_OTP_TTL=5m`
- `AUTH_FORGOT_OTP_COOLDOWN=60s`
- `PUBLIC_REGISTER_URL=http://localhost:3000/register`
