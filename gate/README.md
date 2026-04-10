# Gate Ingress Service

Service nay dung cho:

- webhook nap tien tu provider
- webhook callback doi soat
- trigger thong bao
- trigger email

`gate` khong xu ly core game logic.
Nhiem vu cua no la:

- nhan request ngoai vao
- webhook nap tien khong can xac thuc
- doi payload sang event noi bo
- day sang service noi bo (gin) de apply giao dich

## Cau truc

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
