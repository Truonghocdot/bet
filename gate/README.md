# Gate Ingress Service

Service nay dung cho:

- webhook nap tien tu provider
- webhook callback doi soat
- trigger thong bao
- trigger email

`gate` khong xu ly core game logic.
Nhiem vu cua no la:

- Nhận request ngoài vào
- webhook nạp tiền không cần xác thực
- đổi payload sang event nội bộ
- Đẩy sang service nội bộ (gin) để apply giao dịch

## Cấu trúc

```text
cmd/webhooks
internal/app
internal/domain/event
internal/service
internal/transport/http
```

## Endpoint scaffold

- `GET /healthz`
- `POST /v1/webhooks/deposits/{provider}`
- `POST /v1/notifications/email`
- `POST /v1/notifications/push`

## Env toi thieu

- `HTTP_ADDR=:8082`
- `GIN_INTERNAL_BASE_URL=http://localhost:8081`
- `GIN_INTERNAL_TOKEN`

## Env TC-Gaming

- `TCG_ENABLED=false`
- `TCG_BASE_URL=`
- `TCG_HTTP_TIMEOUT=30s`
- `TCG_MERCHANT_CODE=`
- `TCG_MERCHANT_DES_KEY=`
- `TCG_MERCHANT_SIGN_KEY=`
- `TCG_REPORT_FTP_HOST=`
- `TCG_REPORT_FTP_PORT=21`
- `TCG_REPORT_FTP_USERNAME=`
- `TCG_REPORT_FTP_PASSWORD=`
- `TCG_REPORT_FTP_BASE_DIR=`
- `TCG_GAME_LIST_SYNC_ENABLED=false`
- `TCG_GAME_LIST_SYNC_INTERVAL=5m`
- `TCG_GAME_LIST_REDIS_KEY=shared:tcg:game-list:v1`
- `TCG_GAME_LIST_PRODUCT_TYPES=3,4` (fallback cũ nếu không dùng map theo game type)
- `TCG_GAME_LIST_PLATFORM=all`
- `TCG_GAME_LIST_CLIENT_TYPE=all`
- `TCG_GAME_LIST_TYPES=RNG,FISH,LIVE,PVP` (fallback cũ nếu không dùng map theo game type)
- `TCG_GAME_LIST_PRODUCTS_RNG=275,16,141,55,140,98,243,212`
- `TCG_GAME_LIST_PRODUCTS_FISH=79,16,141,55,140,43,138,243,212`
- `TCG_GAME_LIST_PRODUCTS_LIVE=4,112,3,79,27,177,43,39,93,258,118,272`
- `TCG_GAME_LIST_PRODUCTS_PVP=140,43,212`
- `TCG_GAME_LIST_PRODUCTS_SPORTS=104,65,68,131,174,202,132`
- `TCG_GAME_LIST_PRODUCTS_ELOTT=`
- `TCG_GAME_LIST_LANGUAGE=VI`
- `TCG_GAME_LIST_PAGE=0`
- `TCG_GAME_LIST_PAGE_SIZE=0`
