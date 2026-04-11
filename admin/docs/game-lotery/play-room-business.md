# Nghiệp Vụ Màn Chơi Theo Room

Tài liệu này là nguồn chuẩn cho `PlayView` và các màn room game cùng cấu trúc.

## Mục tiêu

- Mỗi tab thời gian là một `room_code` cố định.
- UI đọc trạng thái phòng thật từ `gin`.
- `game period` chạy độc lập theo room, không phụ thuộc traffic người chơi.
- Khóa lệnh ở **5 giây cuối** trước `draw_at`.
- Khi kỳ đã vào vùng khóa, **admin cũng không được chỉnh sửa** các bản ghi liên quan.

## Mô hình room

### Quy ước room code

- `wingo_30s`
- `wingo_1m`
- `wingo_3m`
- `wingo_5m`
- `k3_1m`
- `k3_3m`
- `k3_5m`
- `k3_10m`
- `lottery_1m`
- `lottery_3m`
- `lottery_5m`
- `lottery_10m`

### Nguyên tắc

- `room_code` là định danh phòng, không thay đổi theo từng kỳ.
- `period_no` là mã hiển thị của kỳ.
- `current_period.id` là khóa định danh thật phải dùng khi đặt lệnh.

## Luồng hiển thị

### 1. Tải trạng thái phòng

FE gọi:

- `GET /v1/play/rooms/{room_code}/state`

Endpoint public, không cần Bearer token.

Phản hồi tối thiểu:

- `server_time`
- `room`
- `current_period`
- `recent_results`

FE dùng:

- `server_time` để đồng bộ countdown
- `current_period.period_no` để hiển thị kỳ hiện tại
- `current_period.draw_at` để đếm ngược
- `recent_results` để hiển thị kết quả gần nhất

### 2. Tải lịch sử

FE gọi:

- `GET /v1/play/rooms/{room_code}/history`
- `GET /v1/play/rooms/{room_code}/bets`

`history` public.
`bets` cần đăng nhập.

`history` là lịch sử kết quả room.  
`bets` là lịch sử của người chơi hiện tại.

### 3. Đặt lệnh

FE gọi:

- `POST /v1/play/rooms/{room_code}/bets`

Endpoint cần đăng nhập.

Payload:

- `request_id`
- `period_id`
- `items[]`

Quy ước:

- `period_id` phải lấy từ `current_period.id`
- `period_id` không được lấy từ `period_no`
- `request_id` là idempotency key để tránh double click
- `X-Connection-ID` phải được giữ xuyên suốt vòng đời join/bet nếu phiên đã được tạo

## Rule khóa lệnh

- `bet_lock_at = draw_at - 5s`
- Khi `now >= bet_lock_at`:
  - user không thể đặt thêm lệnh
  - admin không được chỉnh sửa kỳ/lệnh liên quan
  - FE phải disable action đặt lệnh

## Trạng thái kỳ

- `SCHEDULED`: kỳ đã tạo, chưa mở cược
- `OPEN`: đang nhận cược
- `LOCKED`: đã khóa cược
- `DRAWN`: đã có kết quả raw
- `SETTLED`: đã chốt ví/xử lý payout
- `CANCELED`: kỳ bị hủy

## Những điểm đang dùng trong FE hiện tại

- Route `/play/:game` vẫn được giữ để vào phòng game theo game type.
- FE chọn `room_code` theo biến thể đang active:
  - ví dụ `wingo + 30s => wingo_30s`
- FE vẫn có thể giữ layout cũ của app mobile, nhưng state không còn lấy từ mock.

## Những lỗi cần tránh

- Không dùng `period_no` làm `period_id`.
- Không dùng countdown local tĩnh nếu đã có `server_time`.
- Không cho UI tiếp tục bet khi room đã khóa, dù `current_period.status` chưa kịp refresh.
- Không để admin sửa record khi kỳ đã bước vào vùng khóa.
