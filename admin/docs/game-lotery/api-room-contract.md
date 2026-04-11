# API Room Contract (Play)

## 1) GET `/v1/play/rooms`

Trả danh sách room hiện có để dựng tab và lobby.

Response chính:

- `code`
- `game_type`
- `duration_seconds`
- `bet_cutoff_seconds`
- `status`
- `sort_order`

## 2) GET `/v1/play/rooms/{room_code}/state`

Trả trạng thái room theo thời điểm hiện tại.

Response chính:

- `server_time`
- `room`
- `current_period`:
  - `id`
  - `period_no`
  - `status`
  - `open_at`
  - `bet_lock_at`
  - `draw_at`
- `recent_results` (list ngắn)

## 3) GET `/v1/play/rooms/{room_code}/history`

Trả lịch sử phân trang theo room.

Query:

- `page`
- `page_size`

## 4) POST `/v1/play/rooms/{room_code}/bets`

Payload:

- `request_id` (idempotency key)
- `period_id`
- `items[]`

Behavior:

- reject nếu `now >= bet_lock_at`
- reject nếu period không ở `OPEN`
- reject nếu room/period không khớp

## 5) Compat API cũ

Trong phase chuyển tiếp vẫn giữ:

- `v1/games/*`

Mapping:

- game type map về room mặc định:
  - `wingo -> wingo_1m`
  - `k3 -> k3_1m`
  - `lottery -> lottery_1m`
