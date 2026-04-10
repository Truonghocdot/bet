# Home Screen (End-user) - Business Spec

Nguồn thiết kế:

- `practice/vue/docs/design/home/homepage.png`
- `practice/vue/docs/design/home/code.html`

Màn hình: **Trang chủ** (end-user, mobile). Tài liệu này mô tả nghiệp vụ + data contract để backend `gin` và ERP `admin` triển khai nhất quán.

## 1) Blocks và dữ liệu cần có

### 1.1 TopAppBar

Hiển thị:

- icon `menu`: mở navigation drawer (front-end).
- `brand`: FF789.
- số dư nhanh: `VNĐ 1,250,000`.
- icon `send`: hành vi chưa định nghĩa trong thiết kế (tạm coi là "chia sẻ" hoặc "gửi liên hệ").

Yêu cầu dữ liệu:

- `wallet.vnd_balance` (format theo locale phía client).

Rule:

- Nếu user chưa đăng nhập: ẩn số dư hoặc hiển thị CTA `Đăng nhập`.

### 1.2 Hero promotion banner

Hiển thị banner khuyến mãi chính (ví dụ: "Thưởng nạp lần đầu đến 100% giá trị").

Yêu cầu dữ liệu:

- `banner.image_url`
- `banner.title`
- `banner.subtitle`
- `banner.cta_text`
- `banner.cta_link` hoặc `action` (đi đến màn nạp)
- `banner.active_from/active_to` (lọc theo thời gian)

Nguồn dữ liệu (đề xuất):

- ERP quản trị danh sách promotions/banners.
- `gin` chỉ đọc để trả về Home.

### 1.3 Quick categories (hàng icon 5 mục)

Các nút:

- `Xổ số`
- `Casino`
- `Bắn cá`
- `Thể thao`
- `Game bài`

Nghiệp vụ:

- điều hướng tới danh mục game tương ứng.

Yêu cầu dữ liệu:

- danh mục hiển thị có thể hardcode ở client (phase đầu), hoặc lấy từ server.

Ghi chú:

- Trong hệ thống DB hiện tại tập trung `Wingo/K3/Lottery`. Các danh mục khác (Casino/Bắn cá/Thể thao/Game bài) nếu chỉ là UI placeholder thì đánh dấu `coming soon`.

### 1.4 Trò chơi nổi bật (Featured games)

Hiển thị 4 card:

- `Win Go`
- `K3`
- `5D Lotre`
- `Trx Win`

Mỗi card có:

- icon
- title
- mô tả ngắn

Nghiệp vụ:

- click card -> vào màn game tương ứng và gửi `join`.

Yêu cầu dữ liệu:

- danh sách featured games (server có thể trả để điều khiển thứ tự/ẩn hiện).
- mapping `game_code`:
  - `wingo` (đã có trong `gin`)
  - `k3` (đã có trong `gin`)
  - `lottery` hoặc `5d_lottery` (cần chốt naming)
  - `trx_win` (chưa có trong enum hiện tại; nếu chưa làm thì để `coming soon`)

Rule:

- Nếu game đang bảo trì: card vẫn hiển thị nhưng disable và show status.

### 1.5 Thông tin trúng thưởng (Winning feed)

Hiển thị list "Thông tin trúng thưởng" (marketing feed), gồm:

- avatar
- nickname đã mask (ví dụ `User***821`, `Linh***9x`)
- mô tả (ví dụ: `Vừa rút tiền thành công`, `K3 Sicbo Master`, ...)
- số tiền thắng `+25,000,000đ`
- tag game/room (ví dụ `Win Go 1m`, `K3 Lotre`, `SABA Sport`)

Nghiệp vụ:

- Không phải realtime bắt buộc. Có thể polling hoặc load khi vào Home.
- Dữ liệu phải **ẩn danh** (mask) và không được lộ thông tin nhận dạng thật.

Nguồn dữ liệu (đề xuất):

- Phase đầu: data "curated" (seed/cấu hình) để tránh kéo phức tạp settlement.
- Phase sau: generate từ `bet_settlements` hoặc `withdrawal_requests PAID` nhưng phải:
  - mask username
  - giới hạn số record
  - không lộ user_id/email/phone

Rule:

- Không bao giờ hiển thị chính xác tên thật nếu user chưa bật public profile.
- Bắt buộc có rate-limit/cache vì home traffic cao.

### 1.6 Bottom navigation

Các tab:

- `Trang chủ` (active)
- `Hoạt động`
- `Tiếp thị`
- `Kiếm 150k`
- `Tôi`

Backend:

- `Hoạt động`: thường là lịch sử bet/transaction.
- `Tiếp thị`: affiliate hub (ref link, referrals, reward).
- `Kiếm 150k`: có thể là landing cho affiliate reward setting.
- `Tôi`: màn account (đã có spec).

### 1.7 Floating action button (Quick deposit)

- FAB icon `add_card` ở góc phải dưới.
- Nghiệp vụ: mở nhanh màn nạp.

Backend: không yêu cầu.

## 2) API tối thiểu cần có ở Gin (đề xuất)

### 2.1 Home summary

`GET /v1/home` (public hoặc auth-optional)

Response gợi ý:

- `auth`:
  - `is_authenticated` boolean
- `wallet` (nếu auth):
  - `vnd_balance`
- `hero_banner`:
  - `title`, `subtitle`, `image_url`, `cta`
- `categories`:
  - list `{ key, title, icon, route, enabled }`
- `featured_games`:
  - list `{ game_code, title, subtitle, enabled, maintenance_message? }`
- `winning_feed`:
  - list `{ display_name, avatar_url?, description, amount, tag }`

Caching:

- `hero_banner`, `featured_games`, `categories`, `winning_feed` nên cache.
- `wallet.vnd_balance` không cache (hoặc cache rất ngắn).

## 3) Ràng buộc và TODO

- Naming game codes cần chốt để khớp `gin`:
  - `wingo`, `k3`, `lottery` đã có
  - `5D Lotre` và `Trx Win` hiện chỉ là UI; nếu muốn backend thật cần enum + period/settlement riêng
- Promotions/banners chưa có schema trong DB hiện tại -> cần thiết kế bảng `promotions/banners` sau.
- Winning feed:
  - phase đầu nên curated/cached
  - phase sau mới derive từ settlement/withdrawal và phải ẩn danh + giới hạn hiển thị

