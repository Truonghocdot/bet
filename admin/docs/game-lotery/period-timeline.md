# Period Timeline Theo Room

## Trường thời gian bắt buộc

- `open_at`
- `bet_lock_at` (`draw_at - 5s`)
- `draw_at`
- `settled_at`

## Lifecycle chuẩn

- `SCHEDULED`: đã tạo kỳ, chưa mở cược
- `OPEN`: đang nhận cược
- `LOCKED`: đã khóa nhận cược
- `DRAWN`: đã có kết quả
- `SETTLED`: đã chấm xong và kết thúc kỳ

## Mốc chuyển trạng thái

- `now >= open_at`: `SCHEDULED -> OPEN`
- `now >= bet_lock_at`: `OPEN -> LOCKED`
- `now >= draw_at`: `LOCKED -> DRAWN`
- settlement worker thành công: `DRAWN -> SETTLED`

## Rule 5 giây cuối

Từ `bet_lock_at` trở đi:

- API cược trả lỗi từ chối.
- ERP không cho chỉnh sửa:
  - `game_periods`
  - `bet_tickets`
  - `bet_items`

## Quy tắc độc lập room

- Mỗi `room_code` có chuỗi kỳ riêng.
- Kỳ cùng timestamp ở room khác nhau không liên quan nhau.
- Không dùng kỳ toàn cục cho tất cả room.
