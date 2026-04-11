# Engine Workflow 24/7

## Chế độ chạy

- chạy process riêng: `cmd/engine`
- loop nền theo tick ngắn (1s)
- không phụ thuộc có user online hay không

## Pipeline mỗi tick

1. nạp danh sách room `ACTIVE`
2. đảm bảo mỗi room luôn có kỳ hiện tại và kỳ kế tiếp
3. chạy transition theo thời gian:
   - mở kỳ (`SCHEDULED -> OPEN`)
   - khóa cược (`OPEN -> LOCKED`)
   - quay kết quả (`LOCKED -> DRAWN`)
   - settlement (`DRAWN -> SETTLED`)

## Draw source (phase hiện tại)

- pseudo-random nội bộ backend
- lưu `result_payload` + `draw_source=AUTO`

## Idempotency và concurrency

- dùng Redis lock theo `room_code` hoặc `period_id` để chống xử lý trùng đa instance.
- mọi step update phải có điều kiện status hiện tại trong SQL (`where status = ...`).

## Settlement tối thiểu

- kỳ không có vé cược vẫn phải chuyển `SETTLED`.
- vẫn ghi audit kết quả vào `game_round_histories`.

## Retry

- nếu draw/settlement lỗi: giữ kỳ ở status hiện tại để tick sau retry.
- không rollback trạng thái thành trạng thái cũ.
