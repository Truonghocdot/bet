# Pub/Sub Cho Gin Game Backend

## Bối cảnh

- `Laravel` chỉ dùng cho `ERP / backoffice`
- `Gin` là backend chính xử lý:
  - vào game
  - đặt cược
  - đóng kỳ
  - quay kết quả
  - settlement
  - cập nhật ví
  - bắn realtime cho client
- mỗi `game` hiện tại chỉ có `1 room` duy nhất
- user `join room` rồi thao tác `đặt lệnh` trên chính connection đó

Vì đây là hệ thống có tiền, không được dùng `pub/sub` kiểu fire-and-forget cho nghiệp vụ cốt lõi.

## Kết luận ngắn

Phân tách làm 2 lớp:

1. `Durable event bus` cho nghiệp vụ quan trọng
- dùng `Redis Streams` hoặc `NATS JetStream`
- áp dụng cho:
  - `bet.placed`
  - `period.locked`
  - `draw.received`
  - `ticket.settlement.requested`
  - `ticket.settled`
  - `wallet.changed`
  - `withdrawal.requested`

2. `Realtime fan-out` cho UI
- dùng `Redis Pub/Sub` hoặc `WebSocket hub`
- áp dụng cho:
  - countdown kỳ game
  - kết quả mới
  - số người đang online
  - trạng thái room/game

Nói ngắn hơn:
- `tiền / cược / settlement` => phải đi qua `stream`
- `hiển thị realtime` => mới dùng `pub/sub`

## Vì sao không dùng thuần Redis Pub/Sub cho core betting

`Redis Pub/Sub` không giữ message.

Nếu worker chết hoặc reconnect chậm:
- message mất
- settlement lệch
- ví lệch
- khó replay

Với site bet, đây là rủi ro không chấp nhận được.

## Lựa chọn phù hợp nhất cho stack hiện tại

Nếu muốn triển khai nhanh và gọn ops:

- chọn `Redis Streams`

Lý do:
- bạn đã có Redis để cache / share data
- không phải dựng thêm Kafka/RabbitMQ ngay
- có `consumer group`
- có thể replay / ack / retry
- đủ tốt cho giai đoạn đầu và trung hạn

Chỉ nên nhảy sang `Kafka` hoặc `NATS JetStream` khi:
- lưu lượng rất lớn
- nhiều service độc lập
- cần retention / replay mạnh hơn

## Kiến trúc đề xuất

## Room model

Vì mỗi game hiện tại chỉ có 1 room:

- `WINGO` => 1 room
- `K3` => 1 room
- `LOTTERY` => 1 room

Không cần thiết kế `room_id` động quá sớm.

Giai đoạn này chỉ cần:

- `game_type`
- `period_id`
- `user_id`
- `connection_id`

Naming gợi ý:

- `rt:game:wingo`
- `rt:game:k3`
- `rt:game:lottery`

Nếu sau này mở nhiều room:
- mới tách tiếp `rt:game:{game}:room:{room}`

### 1. Command path

Client không publish trực tiếp.

Client tạo request join, sau đó giữ một connection sống với `Gin`.

Khuyến nghị:

1. `HTTP` để auth và lấy token phiên websocket
2. `WebSocket` cho `join game`, `place bet`, `receive result`, `receive wallet update`

Client gọi HTTP/WebSocket command vào `Gin`:

- `POST /games/{game}/bets`
- `POST /games/{game}/join`

`Gin` làm trong transaction:

1. validate request
2. lock ví / lock kỳ game nếu cần
3. tạo `bet_ticket`, `bet_item`
4. trừ hoặc lock tiền
5. ghi `wallet_ledger_entries`
6. ghi `outbox_events`
7. commit

Sau commit:
- publisher đẩy `outbox_events` ra `Redis Streams`

Điểm quan trọng:
- không publish event trước khi commit DB

## 2. Event stream đề xuất

### Stream nghiệp vụ

- `stream:game-events`
- `stream:wallet-events`
- `stream:withdrawal-events`

Hoặc gọn hơn giai đoạn đầu:

- `stream:domain-events`

Mỗi event có:

```json
{
  "event_id": "uuid",
  "event_name": "bet.placed",
  "aggregate_type": "bet_ticket",
  "aggregate_id": 10001,
  "user_id": 55,
  "game_type": "wingo",
  "period_id": 9001,
  "occurred_at": "2026-04-10T05:00:00+07:00",
  "payload": {}
}
```

### Channel realtime

- `rt:game:wingo`
- `rt:game:k3`
- `rt:game:lottery`
- `rt:period:9001`
- `rt:user:55`

Dùng để broadcast:
- countdown
- room state
- result
- ticket update cho đúng user

## 3. Consumer groups nên có

### `cg:settlement`
- nhận `draw.received`
- chấm ticket
- ghi `bet_settlements`
- cập nhật ví
- phát `ticket.settled`

### `cg:wallet-projection`
- cập nhật cache snapshot số dư
- đẩy event cho websocket gateway

### `cg:notification`
- gửi push/ws event cho user

### `cg:analytics`
- cập nhật số liệu dashboard, heatmap, room stats

## Luồng khi người dùng vào game

### A. User mở room/game

1. client gọi `POST /games/{game}/join`
2. Gin trả snapshot:
- period hiện tại
- countdown
- odds
- game status
- balance user
3. nếu join thành công, server giữ `connection session` trong memory:
- `connection_id`
- `user_id`
- `game_type`
- `joined_at`
- `last_seq`
- `current_period_id`
4. server subscribe user vào:
- `rt:game:{game}`
- `rt:user:{user_id}`

Mục tiêu:
- lần đầu lấy bằng snapshot
- sau đó cập nhật bằng realtime

Không dùng stream cho bước này vì đây là luồng hiển thị.

## Session trên connection

Khi user đã join, connection có thể giữ `session context` để mọi lệnh đi trên cùng một flow:

- `connection_id`
- `user_id`
- `game_type`
- `joined_period_id`
- `last_client_request_id`
- `authenticated_at`

Điểm rất quan trọng:

- chỉ giữ `context`
- không giữ `source of truth`

Không nên coi connection là nơi giữ dữ liệu gốc cho:

- balance
- ticket
- trạng thái settlement
- kết quả kỳ

Lý do:

- socket có thể đứt
- pod có thể restart
- scale ngang sẽ lệch state

Connection chỉ là:

- nơi nhận command
- nơi trả ack
- nơi stream realtime về client

## Command model trong connection

Sau khi join, client gửi message trực tiếp trên socket:

```json
{
  "type": "bet.place",
  "request_id": "uuid",
  "game_type": "wingo",
  "payload": {
    "period_id": 9001,
    "bet_type": "single",
    "items": []
  }
}
```

Server trả ack trên chính connection:

```json
{
  "type": "bet.place.ack",
  "request_id": "uuid",
  "success": true,
  "ticket_id": 10001
}
```

Và sau đó có thể push event tiếp:

```json
{
  "type": "ticket.updated",
  "ticket_id": 10001,
  "status": "pending"
}
```

Đây là cách đúng nếu bạn muốn:
- user join room
- đặt lệnh trong connection
- mọi phản hồi đều đi trong cùng một flow kết nối

Nhưng phần commit nghiệp vụ vẫn phải xuống DB ngay khi nhận command.

## Luồng khi người dùng đặt cược

### B. User bấm đặt cược

1. client gửi message `bet.place` trên connection đã join
2. Gin validate:
- connection đã `join` đúng game
- game đang `OPEN`
- period chưa `LOCKED`
- odds còn hiệu lực
- balance đủ
 - `request_id` chưa xử lý trước đó
3. Gin mở DB transaction
4. insert:
- `bet_tickets`
- `bet_items`
- `wallet_ledger_entries`
- update `wallets`
- insert `outbox_events`
5. commit
6. publisher bắn event `bet.placed` vào stream
7. server trả `ack` ngay trên chính connection
8. realtime channel bắn update game/user nếu cần

Điểm then chốt:
- response cho user đi trên cùng websocket flow
- pub/sub chỉ xử lý hậu quả và fan-out
- DB vẫn là source of truth

## Luồng quay số và settlement

### C. Kết quả game về

1. cron / worker / provider adapter ghi kết quả kỳ
2. transaction DB:
- update `game_periods.status = DRAWN`
- ghi raw result
- insert outbox `draw.received`
3. settlement worker consume `draw.received`
4. worker chấm toàn bộ ticket trong kỳ
5. update:
- `bet_items`
- `bet_tickets`
- `bet_settlements`
- `wallets`
- `wallet_ledger_entries`
6. publish:
- `ticket.settled`
- `wallet.changed`
- realtime `period settled`

## Outbox pattern

Nên có bảng:

- `outbox_events`

Gợi ý cột:

- `id`
- `event_id`
- `event_name`
- `aggregate_type`
- `aggregate_id`
- `payload`
- `status`
- `available_at`
- `published_at`
- `retry_count`
- `created_at`

Lý do:
- bảo đảm `DB commit` và `event publish` không lệch nhau
- retry an toàn
- dễ audit

## Idempotency

Bắt buộc có ở các bước:

- đặt cược
- settlement
- cập nhật ví
- payout

Nên có:

- `request_id` từ client khi đặt cược
- `event_id` duy nhất
- unique key cho `wallet ledger`
- unique key cho `bet settlement`

Ví dụ:
- 1 ticket không được settle 2 lần cùng một phase
- 1 request đặt cược không được insert 2 ticket

## WebSocket layer

Gin không nên để mỗi worker tự giữ state socket phức tạp.

Tối giản:
- một `ws hub`
- map `user_id -> connections`
- map `game_type -> subscribers`
- map `connection_id -> session context`

Worker không push trực tiếp vào socket.
Worker chỉ publish realtime event vào Redis Pub/Sub.
`ws hub` mới là consumer và fan-out về client.

Như vậy:
- worker stateless hơn
- scale ngang dễ hơn

## Mapping responsibility

### Laravel ERP
- đọc DB
- quản trị user
- quản trị giao dịch
- quản trị period, settlement manual
- không xử lý core realtime

### Gin Core
- auth end user
- join game session
- place bet
- period engine
- result intake
- settlement
- wallet mutation
- realtime gateway / websocket

## Gói Go nên dùng

Nếu triển khai theo hướng này, bộ tối thiểu:

- `github.com/gin-gonic/gin`
- `github.com/redis/go-redis/v9`
- `github.com/gorilla/websocket`
- `github.com/google/uuid`
- driver DB bạn đang chọn (`pgx` hoặc `gorm` / `sqlc`)

Nếu đi production nghiêm túc, tôi khuyên:
- query write side dùng transaction chặt
- queue/event publisher tách package riêng
- websocket hub tách package riêng

## Thiết kế package trong Gin

Gợi ý:

```text
internal/
  app/
  transport/http/
  transport/ws/
  domain/
    bet/
    wallet/
    period/
    affiliate/
  repository/
  service/
  event/
    outbox/
    stream/
    realtime/
  worker/
    settlement/
    projection/
    notification/
```

## Cách triển khai theo từng phase

### Phase 1
- DB transaction chuẩn
- `outbox_events`
- publisher nền
- `Redis Streams`
- WebSocket hub đơn giản

### Phase 2
- consumer group settlement
- retry / dead-letter logic
- user channel / room channel rõ ràng

### Phase 3
- split service nếu tải lớn
- cân nhắc `NATS JetStream`

## Quyết định đề xuất

Nếu làm ngay với repo này:

1. `Gin` cấp token và mở websocket session cho user
2. user gửi `join game`
3. `Gin` giữ `connection session context`
4. user gửi `bet.place` trên chính connection đó
5. `Gin` ghi DB + outbox trong cùng transaction
6. publisher đẩy sang `Redis Streams`
7. settlement worker consume từ stream
8. realtime UI dùng `Redis Pub/Sub + WebSocket`

Đây là phương án cân bằng nhất giữa:
- đúng nghiệp vụ tiền
- triển khai nhanh
- ít hạ tầng
- đủ đường mở rộng sau này
