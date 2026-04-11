# Play View (Lottery-style) - Design Analysis

Nguồn:

- `view-play.png`
- `code.html`

Màn hình mô tả: **Win Go 30s** (room-based, round-based). Đây là màn chơi chính: xem số dư, room hiện tại, kỳ hiện tại, đếm ngược, kết quả gần đây, đặt cược theo cửa (màu/số/lớn nhỏ), và xem lịch sử.

## 1) Blocks và hành vi UI

### 1.1 TopAppBar

- Back: quay lại danh sách game / home.
- Title: `Win Go 30s` (tên biến thể).
- `Hỗ trợ`: mở hỗ trợ.

### 1.2 Balance panel

- `Số dư hiện tại` (VND): hiển thị + nút refresh.
- CTA: `Nạp tiền`, `Rút tiền`.

Backend cần:

- số dư ví VND (và trạng thái ví/user để có thể disable CTA nếu bị khóa)

### 1.3 Announcements marquee

- Dòng chạy chúc mừng thắng (marketing).

Backend:

- có thể dùng feed marketing curated/cached, không cần realtime.

### 1.4 Game selection tabs (time variants)

Tabs:

- `30 Giây` (active)
- `1 Phút`
- `3 Phút`
- `5 Phút`

Nghiệp vụ:

- mỗi tab là **một room cố định** của cùng `GameType = WINGO`.
- user switch tab thì chuyển sang room stream tương ứng, cập nhật `current_period` + countdown + recent results.

Mapping đề xuất:

- `wingo_30s`, `wingo_1m`, `wingo_3m`, `wingo_5m`
- `room_code` là khóa phòng, `period_no` là mã hiển thị của từng kỳ.

### 1.5 Countdown & results

- `Số kỳ`: `period_no` (ví dụ: `20240520342`)
- Countdown: mm:ss (trong mockup là `00:28`)
- `Kết quả gần đây`: 5 kết quả gần nhất (số 0..9 với màu nền theo rule)

Backend cần:

- `current_period`:
  - `id`
  - `period_no`
  - `open_at`, `bet_lock_at`, `draw_at`, `status`
  - server time (để client render countdown ổn định)
- `recent_results`: list kết quả gần nhất (tối thiểu 5)
- `GET /v1/play/rooms/{room_code}/state` và `history` là public; chỉ `bets` của user và đặt lệnh mới cần Bearer token.

Rule UI:

- khi `status = LOCKED` (đã khóa cược) thì disable bet buttons.

### 1.6 Betting area

Cửa cược trong UI:

- `COLOR`: `Xanh`, `Tím`, `Đỏ`
- `NUMBER`: 0..9
- `BIG_SMALL`: `LỚN`, `NHỎ`

Tooling UI:

- `Ngẫu nhiên`: chọn ngẫu nhiên cửa cược (thường là number/color)
- multiplier presets: `X1`, `X5`, `X10`, `X20`, `X50`
  - multiplier ảnh hưởng stake (stake = base_amount * multiplier)

Backend cần:

- odds/payout table cho từng cửa cược (để hiển thị odds nếu UI muốn).
- validation server-side:
  - option_type/option_key hợp lệ cho `WINGO`
  - chỉ cho đặt khi period `OPEN`
  - idempotency theo `request_id`/`ticket_no` (tránh double click)

### 1.7 History tabs

Tabs:

- `Lịch sử trò chơi`: lịch sử kết quả theo kỳ (public feed)
- `Biểu đồ`: chart (phase sau)
- `Lịch sử của tôi`: bet history của user

Trong mockup tab 1 có bảng:

- `Kỳ xổ` (period_no rút gọn)
- `Số` (result number)
- `Lớn nhỏ`
- `Màu sắc`

Backend cần:

- `GET game history`: list period results theo room
- `GET my bet history`: list ticket của user theo period/game

## 2) API tối thiểu (đề xuất cho Gin)

- `GET /v1/play/rooms`
- `GET /v1/play/rooms/{room_code}/state`
- `GET /v1/play/rooms/{room_code}/history`
- `GET /v1/play/rooms/{room_code}/bets`
- `POST /v1/play/rooms/{room_code}/bets`
  - body: `request_id`, `period_id`, `items[] { option_type, option_key, stake }`

## 3) Notes triển khai

- Play-view không cần realtime bắt buộc, nhưng countdown nên dựa trên `server_time` để không bị lệch.
- `recent_results` và `game history` nên cache ngắn.
- Bet placement phải idempotent theo `request_id`.
- `X-Connection-ID` nên được giữ xuyên suốt nếu hệ thống join/bet đang dùng phiên kết nối.
