# Admin Guard Rules (ERP)

## Mục tiêu

Ngăn thao tác chỉnh sửa gây lệch kết quả hoặc vi phạm fairness khi kỳ đã vào vùng đóng cược.

## Bảng áp dụng

- `game_periods`
- `bet_tickets`
- `bet_items`

## Rule khóa chỉnh sửa

### Period

Không cho update/delete khi:

- `status >= LOCKED`, hoặc
- `now >= bet_lock_at`

### Ticket / Item

Không cho update/delete khi period liên quan:

- `status >= LOCKED`, hoặc
- `now >= bet_lock_at`

## Hành vi UI/ERP

- Ẩn nút `Edit/Delete` khi không hợp lệ.
- Nếu vẫn gọi trực tiếp endpoint/update action:
  - backend model/service phải reject, không chỉ chặn bằng UI.

## Message gợi ý

- `Kỳ đã bước vào giai đoạn khóa cược, không thể chỉnh sửa.`
- `Vé cược thuộc kỳ đã khóa, thao tác bị từ chối.`
