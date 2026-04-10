# Gin Core Service

Service nay la backend chinh cho:

- auth end-user
- join game
- place bet
- tao deposit intent
- apply deposit tu webhook noi bo
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
- `POST /v1/deposits/vietqr/init`
- `POST /v1/deposits/usdt/init`
- `GET /v1/deposits/{client_ref}`
- `POST /internal/v1/deposits/apply`

## Env toi thieu

- `DATABASE_URL`
- `REDIS_ADDR`
- `REDIS_DB=2`
- `GATE_BASE_URL`
- `AUTH_TOKEN_SECRET`
- `AUTH_TOKEN_TTL=24h`
- `AUTH_FORGOT_OTP_TTL=5m`
- `AUTH_FORGOT_OTP_COOLDOWN=60s`
- `PUBLIC_REGISTER_URL=http://localhost:3000/register`
- `GIN_INTERNAL_TOKEN`
- `PAYMENT_RECEIVING_ACCOUNTS_REDIS_KEY=shared:payment:receiving-accounts:v1`
