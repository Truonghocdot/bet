# Bản Ghi Cần Insert

Tài liệu này chỉ rõ các bản ghi phải được insert trong transaction khi tạo mới user.

## 1) End User Register

Thứ tự insert khuyến nghị:

1. `users`
2. `wallets`
3. `affiliate_profiles`
4. `affiliate_referrals` nếu có referral code/link hợp lệ

### 1.1 `users` - bắt buộc

Insert bản ghi user gốc với các field tối thiểu:

- `name`
- `email`
- `phone` nếu có
- `password` đã hash
- `role = CLIENT`
- `status` theo policy đăng ký
- `email_verified_at` và `phone_verified_at` để `null` nếu chưa verify

### 1.2 `wallets` - bắt buộc theo cấu hình hiện tại

Trong hệ thống này, user mới nên được tạo sẵn các ví mặc định để các luồng nạp/rút/cược không bị thiếu record.

- 1 wallet `VND` - chỉ tạo 1 duy nhất cho đơn vị VND để người dùng chơi trực tiếp

Mỗi wallet:

- `user_id` = user mới
- `unit` = đơn vị tương ứng
- `balance = 0`
- `locked_balance = 0`
- `status = ACTIVE`

Nếu sau này tắt bớt unit nào đó thì chỉ insert các wallet đang được enable.

### 1.3 `affiliate_profiles` - bắt buộc

Mỗi user mới phải có 1 hồ sơ affiliate riêng để sinh:

- `ref_code`
- `ref_link`

Các field gợi ý khi insert:

- `user_id`
- `ref_code` sinh tự động, duy nhất
- `ref_link` sinh từ `ref_code`
- `status = PENDING` hoặc `ACTIVE` theo policy

### 1.4 `affiliate_referrals` - có điều kiện

Chỉ insert khi user mới đi vào từ referral code/link hợp lệ.

Trạng thái khi vừa register:

- `status = PENDING`

Các field chính:

- `affiliate_profile_id` của người mời
- `referrer_user_id`
- `referred_user_id`
- `affiliate_link_id` nếu đi qua link cụ thể
- `status = PENDING`

Khi user nạp đầu tiên >= 50.000 VND thì hệ thống mới update sang `QUALIFIED`.

### 1.5 Không insert khi register

Không tạo các record sau ở bước register:

- `transactions`
- `withdrawal_requests`
- `wallet_ledger_entries`
- `account_withdrawal_infos`
- `affiliate_reward_logs`

## 2) ERP Create User

Thứ tự insert khuyến nghị:

1. `users`
2. `wallets` nếu admin chọn tạo sẵn
3. `affiliate_profiles` nếu admin bật affiliate cho user này
4. `account_withdrawal_infos` nếu admin nhập sẵn tài khoản rút

### 2.1 `users` - bắt buộc

Insert user gốc với quyền ERP cho phép set thêm:

- `role`
- `status`
- `email_verified_at`
- `phone_verified_at`
- `last_login_at` nếu cần backfill

### 2.2 `wallets` - tùy chọn

Admin có thể tạo sẵn wallet nếu muốn user có thể thao tác ngay.

Khuyến nghị:

- tạo `VND`

### 2.3 `affiliate_profiles` - tùy chọn

Chỉ tạo nếu admin muốn kích hoạt affiliate profile ngay khi tạo user.

Lưu ý:

- `ref_code` phải được hệ thống tự sinh, duy nhất
- quản trị viên không nhập tay `ref_code`
- `ref_link` được tạo tự động từ `ref_code`

### 2.4 `account_withdrawal_infos` - tùy chọn

Chỉ tạo nếu user được nhập sẵn tài khoản rút ngay trong ERP.

## 3) Nguyên tắc transaction

Toàn bộ các insert trên phải chạy trong cùng một transaction.

Nếu một bản ghi insert thất bại:

- rollback toàn bộ
- không để user tồn tại một nửa trạng thái
- không để wallet/affiliate tạo lệch user
