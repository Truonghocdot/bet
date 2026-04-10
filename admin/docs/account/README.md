# Account Screen (End-user) - Business Spec

Nguồn thiết kế:

- `practice/vue/docs/design/account/account.png`
- `practice/vue/docs/design/account/code.html`

Màn hình: **Tài khoản** (end-user, mobile). Tài liệu này mô tả nghiệp vụ + data contract để backend `gin` và ERP `admin` triển khai nhất quán.

## 1) Blocks và dữ liệu cần có

### 1.1 TopAppBar

- Icon `menu`: mở navigation drawer (front-end).
- Icon `send`: hành vi chưa định nghĩa trong thiết kế (tạm coi là "chia sẻ" hoặc "gửi liên hệ").

Backend: không yêu cầu.

### 1.2 User profile card

Hiển thị:

- `avatar` (url hoặc placeholder)
- `name`
- `id` hiển thị kiểu `ID: 889922` (user id hoặc user_code)
- `rank label` (ví dụ: `Thành viên Bạc`)
- `VIP level` (ví dụ: `VIP LV.4`)

Yêu cầu dữ liệu:

- `user.id` (hoặc `user.code` nếu muốn không lộ id thật)
- `user.name`
- `user.avatar_url` (optional)
- `user.vip_level` (derived)
- `user.rank_name` (derived)

Rule:

- `vip_level` và `rank_name` là dữ liệu derived theo tổng nạp/cược hoặc rule vận hành. Nếu chưa triển khai, trả về mặc định `VIP LV.0` và `Thành viên`.

### 1.3 Quick actions

Khối `Ví Của Tôi`:

- hiển thị tổng số dư theo unit mặc định (trong UI đang là VND, format `12.500.000đ`).

2 nút:

- `Nạp tiền` -> điều hướng sang luồng deposit.
- `Rút tiền` -> điều hướng sang luồng withdraw.

Yêu cầu dữ liệu:

- `wallet.balance_vnd` (hoặc wallet mặc định theo `unit=VND`)
- trạng thái ví (`ACTIVE/LOCKED`) để disable nút thao tác nếu bị khóa.

Rule:

- Nếu ví `LOCKED` hoặc user `SUSPENDED/BANNED` thì disable `Nạp/Rút` và hiển thị message phù hợp (front-end).

### 1.4 Menu list

#### a) Thông báo (Notifications)

Hiển thị:

- số lượng chưa đọc (badge đỏ, ví dụ `3`)

Tương tác:

- click -> mở màn list notifications.

Yêu cầu dữ liệu:

- `unread_count` cho current user.

Backend source-of-truth:

- dùng module `Notifications (In-app)` theo spec ở `admin/docs/notifications-and-news/README.md`.

#### b) Quy đổi quà (Redeem gifts)

Thiết kế có menu entry nhưng chưa có schema/flow trong DB hiện tại.

Nghiệp vụ tối thiểu (đề xuất để triển khai sau):

- user có `points` hoặc `reward_balance`
- danh mục `rewards` (voucher, quà, bonus)
- user đổi quà -> sinh record `reward_redemptions`
- ERP duyệt/chi trả nếu là quà vật lý, hoặc cộng ví nếu là bonus.

Giai đoạn này: để `TODO` (chưa triển khai).

#### c) Thống kê trò chơi (Game statistics)

Nghiệp vụ:

- hiển thị thống kê theo ngày/tuần/tháng:
  - tổng cược
  - tổng thắng/thua
  - net P/L
  - số vé cược

Giai đoạn này: có thể làm read-only từ bảng `bet_tickets` / `bet_items` / `bet_settlements`.

#### d) Ngôn ngữ (Language)

Hiển thị:

- ngôn ngữ hiện tại (ví dụ `Tiếng Việt`).

Nghiệp vụ tối thiểu:

- user chọn language -> lưu preference.

Storage (đề xuất):

- thêm `users.locale` (varchar(10)) hoặc table `user_settings`.

Giai đoạn này: nếu chưa có DB field, lưu localStorage phía client; backend chưa bắt buộc.

### 1.5 Trung tâm dịch vụ

#### a) Cài đặt (Settings)

Các hạng mục thường có:

- đổi mật khẩu
- liên kết email/phone + verify
- thiết lập bảo mật (2FA nếu có)

Vì hệ thống đã có OTP reset password (auth-security), phần Settings sẽ là nơi gọi các flow đó.

#### b) Góp ý (Feedback)

Nghiệp vụ:

- user gửi góp ý kèm nội dung + ảnh (optional)
- hệ thống lưu ticket để support xử lý.

Giai đoạn này: để `TODO` nếu chưa có ticket schema.

#### c) Hỗ trợ 24/7

Nghiệp vụ:

- mở livechat hoặc tạo ticket.

Giai đoạn này: có thể chỉ mở link external.

### 1.6 Đăng xuất

- Button `Đăng xuất tài khoản`.

Auth hiện tại của `gin` là access token tự ký, chưa có refresh/token blacklist.

Nghiệp vụ logout tối thiểu:

- client xóa token local.

Nâng cấp sau:

- thêm token version / blacklist nếu muốn revoke server-side.

## 2) API tối thiểu cần có ở Gin (đề xuất)

### 2.1 Account summary

`GET /v1/account/summary` (auth required)

Response:

- `user: { id, name, avatar_url?, vip_level?, rank_name? }`
- `wallet: { vnd_balance, vnd_locked_balance, status }`
- `notifications: { unread_count }`

### 2.2 Notifications

Theo spec riêng:

- `GET /v1/notifications`
- `POST /v1/notifications/{id}/read`

### 2.3 Deposit/Withdraw entrypoints

Đã có nghiệp vụ finance ở DB:

- danh sách tài khoản hệ thống nhận tiền nạp (`payment_receiving_accounts`)
- tạo transaction nạp/rút (`transactions`, `withdrawal_requests`)

Screen account chỉ cần điều hướng sang các màn tương ứng.

## 3) Quyền và ràng buộc

- user `status != ACTIVE`:
  - vẫn cho xem thông tin cơ bản
  - khóa thao tác nạp/rút/cược (tùy policy)
- wallet `status = LOCKED`:
  - khóa nạp/rút

## 4) TODO cho phase sau (khớp UI nhưng chưa có DB)

- `Quy đổi quà` (points/rewards)
- `Feedback/Ticketing`
- `Language preference` lưu DB
- `VIP/Rank` rule engine

