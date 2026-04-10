# Gate Ingress Service

Service nay dung cho:

- webhook nap tien tu provider
- webhook callback doi soat
- trigger thong bao
- trigger email

`gate` khong xu ly core game logic.
Nhiem vu cua no la:

- nhan request ngoai vao
- validate chu ky / payload
- doi payload sang event noi bo
- day sang queue / outbox / service khac

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
