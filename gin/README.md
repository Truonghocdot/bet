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
- `AUTH_FORGOT_OTP_TTL=2m`
- `AUTH_FORGOT_OTP_COOLDOWN=2m`
- `AUTH_FORGOT_WINDOW=2m`
- `AUTH_FORGOT_LIMIT_IP=20`
- `AUTH_FORGOT_LIMIT_TARGET=10`
- `AUTH_LOGIN_FAIL_WINDOW=2m`
- `AUTH_LOGIN_FAIL_LIMIT_IP=30`
- `AUTH_LOGIN_FAIL_LIMIT_ACCOUNT=10`
- `AUTH_LOGIN_LOCK_DURATION=2m`
- `AUTH_REGISTER_WINDOW=2m`
- `AUTH_REGISTER_LIMIT_IP=20`
- `AUTH_REGISTER_LIMIT_EMAIL=10`
- `AUTH_REGISTER_LIMIT_PHONE=10`
- `PUBLIC_REGISTER_URL=http://localhost:3000/register`
- `GIN_INTERNAL_TOKEN`
- `PAYMENT_RECEIVING_ACCOUNTS_REDIS_KEY=shared:payment:receiving-accounts:v1`
