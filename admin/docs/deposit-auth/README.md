# Nghiệp Vụ Auth + Nạp Tiền

Tài liệu này ghi lại luồng nghiệp vụ cho:

- đăng ký / đăng nhập user end-user
- đăng ký có mã giới thiệu
- nạp USDT
- nạp VietQR qua Sepay
- webhook nạp tiền ở `gate` không cần xác thực

## 0. Domain & Ownership (ff789)

Domain chính:

- `ff789.club`: site FE chính (Vue)
- `api.ff789.club`: API chính (service `gin`)
- `admin.ff789.club`: ERP/Backoffice (Laravel `admin`)
- `gate.ff789.club`: gateway nhận webhook thanh toán và forward nội bộ
- (planned) 1 site Vue để control kết quả: subdomain TBD (ví dụ `control.ff789.club`)

Phân tách cấu hình payment trong ERP (source of truth):

- NowPayments:
  - nơi cấu hình: màn setting schema
    - `/home/truonghocdot/study/practice/admin/app/Filament/Pages/System/Schemas/ExchangeRatePageForm.php`
  - ý nghĩa: chứa thông tin kết nối/credential để `gate`/`gin` dùng khi làm việc với NowPayments (USDT)
- Sepay:
  - nơi cấu hình danh sách tài khoản nhận tiền:
    - `/home/truonghocdot/study/practice/admin/app/Filament/Resources/Payment/PaymentReceivingAccounts`
  - **chỉ** dùng cho Sepay (VietQR), không trộn NowPayments vào bảng này

Webhook:

- `gate.ff789.club` là nơi **nhận webhook thanh toán cho cả 2 hướng** (Sepay + NowPayments) và forward sang `api.ff789.club` (gin).
- Webhook inbound ở `gate` không cần xác thực (public callback), nhưng forward nội bộ từ `gate -> gin` nên có token nội bộ.

## 1. Auth end-user

### Đăng ký
- User tự đăng ký ở FE.
- Input tối thiểu:
  - `name`
  - `email`
  - `phone` nếu có
  - `password`
  - `ref_code` tùy chọn
- Khi đăng ký thành công:
  - tạo `users`
  - tạo `wallets` mặc định cho `VND` và `USDT`
  - tạo `affiliate_profiles`
  - nếu có `ref_code` hợp lệ thì tạo `affiliate_referrals`

### Đăng nhập
- Hỗ trợ 2 tab:
  - đăng nhập bằng `số điện thoại`
  - đăng nhập bằng `email`
- Backend dùng chung endpoint login, frontend chỉ đổi cách nhập.

## 2. Nạp tiền

### Nạp USDT
- Chỉ hỗ trợ 1 loại tiền ảo ở phase đầu: `USDT`.
- User chọn nạp USDT:
  - `gin` tạo invoice/payment session qua NowPayments (hoặc generate address theo policy sản phẩm)
  - config/credential NowPayments lấy từ ERP (xem mục 0)
  - trả về địa chỉ ví, mạng lưới và hướng dẫn chuyển tiền
  - tạo `transactions` trạng thái `PENDING`
- Khi provider xác nhận on-chain:
  - `gate` nhận webhook callback
  - `gate` đẩy payload đã chuẩn hóa sang `gin`
  - `gin` cập nhật transaction, cộng ví, ghi ledger

### Nạp VietQR qua Sepay
- User nhập số tiền VND.
- Hệ thống lấy ngẫu nhiên 1 tài khoản nhận tiền `BANK/VND` đang `ACTIVE`.
- `gin` tạo transaction `PENDING` với `client_ref`.
- `gate` là nơi nhận webhook callback từ Sepay.
- Webhook callback **không cần xác thực**.
- `gate` chuyển callback sang `gin` internal endpoint để apply giao dịch.

## 3. Account nhận tiền

- Danh sách account nhận tiền do quản trị viên cấu hình trong ERP.
- Chỉ account `ACTIVE` mới được hiển thị.
- Mỗi lần user mở màn nạp, hệ thống chọn ngẫu nhiên 1 account phù hợp theo `unit` và `type`.
- Cache runtime:
  - Redis shared key: `shared:payment:receiving-accounts:v1`
  - snapshot source: bảng `payment_receiving_accounts` (chỉ dùng cho Sepay/VietQR)

## 4. Webhook

- Webhook inbound ở `gate` **không cần xác thực**.
- Gate chỉ đóng vai trò:
  - nhận callback từ provider
  - chuẩn hóa payload
  - chuyển sang `gin`
- Endpoint internal giữa `gate` và `gin` có thể dùng token nội bộ.

## 5. Setup tối thiểu

1. Admin tạo `payment_receiving_accounts` (Sepay/VietQR).
2. Admin cấu hình NowPayments (USDT) ở `ExchangeRatePageForm` (mục 0).
3. Chạy command prime cache account nhận tiền:
   - `php artisan payment:prime-receiving-accounts`
4. Set Redis shared DB giống giữa `admin` và `gin`.
   - `gin` nên dùng `REDIS_DB=2` để đọc key `shared:payment:receiving-accounts:v1`
5. Cấu hình `gate` trỏ về `gin` internal endpoint.
   - `GIN_INTERNAL_BASE_URL=http://localhost:8081`
   - `GIN_INTERNAL_TOKEN=<secret>`
6. Cấu hình `Sepay webhook` trỏ về `gate`.
   - webhook inbound không cần xác thực

## 6. Checklist triển khai

- Đã migrate bảng `transactions` với:
  - `client_ref`
  - `receiving_account_id`
  - `meta`
- Đã tạo seed/config cho `payment_receiving_accounts` (Sepay)
- Đã cấu hình NowPayments ở ERP (`ExchangeRatePageForm`)
- Đã prime Redis shared key `shared:payment:receiving-accounts:v1`
- `gin` đã có:
  - `POST /v1/deposits/vietqr/init`
  - `POST /v1/deposits/usdt/init`
  - `GET /v1/deposits/{client_ref}`
  - `POST /internal/v1/deposits/apply`
- `gate` đã có webhook deposit public và forward nội bộ sang `gin`
