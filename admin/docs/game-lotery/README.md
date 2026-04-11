# Game Lottery - Room Engine 24/7 Spec

Tài liệu trung tâm cho nghiệp vụ lottery/round-based:

- `WINGO`
- `K3`
- `LOTTERY` (`5D` trong phase hiện tại)

Mục tiêu phase này:

- mọi tab time trên màn play là một `room_code` cố định
- engine chạy 24/7, không phụ thuộc việc có người chơi
- khóa lệnh ở 5 giây cuối kỳ: user không đặt cược được, admin cũng không được sửa

DB reference:

- `game_rooms`, `game_periods`, `bet_tickets`, `bet_items`, `bet_settlements`, `game_round_histories`
- source schema: `admin/database/database.md`

## 1) Quy tắc cứng

- `room_code` là mã cứng, không tạo động.
- mỗi room có chuỗi kỳ quay riêng, độc lập room khác.
- lifecycle chuẩn: `SCHEDULED -> OPEN -> LOCKED -> DRAWN -> SETTLED`.
- `bet_lock_at = draw_at - 5s`.
- từ `bet_lock_at`:
  - user: reject đặt cược mới.
  - admin ERP: không được update kỳ, vé và chi tiết vé.

## 2) Tài liệu con bắt buộc đọc cùng

- [Room Catalog](./room-catalog.md)
- [Period Timeline](./period-timeline.md)
- [Play Room Business](./play-room-business.md)
- [Engine Workflow](./engine-workflow.md)
- [API Room Contract](./api-room-contract.md)
- [Admin Guard Rules](./admin-guard-rules.md)

## 3) Chuẩn option/result theo game

Enums sử dụng:

- `BetOptionType`
- `BetItemResult`
- `BetStatus`
- `PeriodStatus`
- `DrawSource`

Wingo:

- `COLOR`: `green|red|violet`
- `NUMBER`: `number_0..number_9`
- `BIG_SMALL`: `big|small`

K3:

- `SUM`: `sum_3..sum_18`
- `ODD_EVEN`: `odd|even`
- `BIG_SMALL`: `big|small`
- `COMBINATION`: `triple_any` hoặc cụ thể theo quy tắc sản phẩm

Lottery (5D):

- `NUMBER`: `pick5_xxxxx`
- `SUM`: `sum_n`
- `ODD_EVEN/BIG_SMALL`: theo rule variant

## 4) Rule validation cược (áp dụng cho API cũ và mới)

- user/wallet phải `ACTIVE`.
- period phải thuộc đúng room và `status = OPEN`.
- từ `now >= bet_lock_at` thì reject.
- request phải idempotent bằng `request_id`.

## 5) Compat policy

- thêm API room mới cho play-view.
- giữ `v1/games/*` trong giai đoạn chuyển tiếp, map nội bộ sang room mặc định của game.
