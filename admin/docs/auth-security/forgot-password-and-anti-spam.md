# Forgot Password Và Chống Spam

## 1) Mục tiêu

Thiết kế một luồng auth đủ an toàn để:

- giảm spam `register`, `login`, `forgot password`
- không cho lộ thông tin tài khoản tồn tại hay không tồn tại
- cho phép reset password bằng OTP qua `email` hoặc `phone`
- tách rõ trách nhiệm:
  - `gin` xử lý nghiệp vụ auth chính
  - `gate` xử lý gửi `email/sms/push`

## 2) Nguyên tắc chung

- `gin` là nơi tạo yêu cầu OTP, xác minh OTP, đổi mật khẩu.
- `gate` chỉ là cổng gửi thông báo, không quyết định OTP đúng/sai.
- OTP chỉ được lưu dưới dạng `hash`, không lưu plaintext trong DB.
- Mọi luồng dễ bị abuse phải có:
  - `cooldown`
  - `rate limit`
  - `max attempts`
  - `idempotency`
- Response cho `forgot password` phải là generic:****
  - không trả lời trực tiếp là email/phone có tồn tại hay không.

## 3) Luồng chống spam

### 3.1 Register

Chặn theo các lớp:

- theo `IP`
- theo `email`
- theo `phone`
- theo `device fingerprint` nếu frontend có gửi

Rule khuyến nghị:

- tối đa `5` lần đăng ký / `15 phút` / `IP`
- tối đa `3` lần đăng ký / `24 giờ` / `email`
- tối đa `3` lần đăng ký / `24 giờ` / `phone`
- cooldown `60 giây` nếu vừa submit thất bại liên tiếp quá nhanh

Khi vượt ngưỡng:

- không tạo user
- trả message chung
- ghi log vào bảng auth limit audit nếu cần

### 3.2 Login

Chặn theo:

- `IP`
- `account` (`email` hoặc `phone`)
- cặp `IP + account`

Rule khuyến nghị:

- tối đa `10` lần login fail / `15 phút` / `IP`
- tối đa `5` lần login fail / `15 phút` / `account`
- nếu fail liên tiếp `>= 5` lần thì khóa mềm `15 phút` theo `account`

Lưu ý:

- chưa khóa `users.status`, vì đây là cơ chế chống brute force tạm thời
- lock kiểu này nên nằm ở `redis`

### 3.3 Forgot Password

Chặn theo:

- `IP`
- `target` (`email` hoặc `phone`)
- `purpose = RESET_PASSWORD`

Rule khuyến nghị:

- tối đa `5` yêu cầu OTP / `1 giờ` / `IP`
- tối đa `3` OTP đang gửi / `30 phút` / `target`
- cooldown gửi lại OTP: `60 giây`
- OTP hết hạn sau `5 phút`
- tối đa `5` lần nhập sai OTP / mỗi request

Nếu vượt ngưỡng:

- không gửi OTP mới
- vẫn trả response mềm, không làm lộ nội bộ

## 4) Luồng quên mật khẩu

### 4.1 Bước 1: Gửi yêu cầu quên mật khẩu

Input:

- `channel = EMAIL | PHONE`
- `account`

Xử lý:

1. chuẩn hóa `account`
2. áp rate limit theo `IP`, `account`, `channel`
3. tìm user theo `email` hoặc `phone`
4. dù có thấy user hay không, vẫn trả response chung
5. nếu user tồn tại:
   - tạo OTP mới
   - hủy các OTP `PENDING` cũ cùng `purpose`
   - lưu DB
   - đẩy command sang `gate` để gửi email/SMS

Response gợi ý:

- `Nếu tài khoản tồn tại, mã OTP đã được gửi.`

### 4.2 Bước 2: Xác minh OTP

Input:

- `account`
- `channel`
- `otp`

Xử lý:

1. tìm OTP `PENDING`
2. kiểm tra:
   - chưa hết hạn
   - chưa vượt số lần nhập sai
   - đúng `channel`
   - đúng `purpose = RESET_PASSWORD`
3. nếu sai:
   - tăng `attempt_count`
   - nếu quá giới hạn thì chuyển `status = LOCKED`
4. nếu đúng:
   - chuyển `status = VERIFIED`
   - set `verified_at`
   - sinh `reset_token` ngắn hạn hoặc dùng luôn `otp_request_id`

### 4.3 Bước 3: Đặt mật khẩu mới

Input:

- `otp_request_id` hoặc `reset_token`
- `new_password`

Xử lý:

1. kiểm tra OTP record đã `VERIFIED`
2. kiểm tra chưa `USED`, chưa hết hạn phiên reset
3. update `users.password`
4. set:
   - `used_at`
   - `status = USED`
5. hủy các OTP reset password còn `PENDING` / `VERIFIED` khác của user
6. xóa hoặc vô hiệu hóa access token cũ nếu sau này có session/token store

## 5) Nguyên tắc không lộ thông tin tài khoản

Các response sau phải là generic:

- gửi OTP quên mật khẩu
- resend OTP
- verify account bằng email/phone nếu sau này có

Không được trả:

- `email không tồn tại`
- `phone chưa đăng ký`
- `account bị khóa`

ở bước gửi OTP công khai.

Những thông tin đó chỉ nên ghi log nội bộ.

## 6) OTP channel

### EMAIL

Khi `channel = EMAIL`:

- bắt buộc user có `email`
- `gin` tạo OTP
- `gate` gửi email template reset password

### PHONE

Khi `channel = PHONE`:

- bắt buộc user có `phone`
- `gin` tạo OTP
- `gate` gửi SMS hoặc push nhà mạng

Lưu ý:

- chưa cần yêu cầu `email_verified_at` hoặc `phone_verified_at` trước khi gửi reset OTP
- nhưng sau này nếu muốn thắt chặt hơn có thể chỉ cho reset qua channel đã verify

## 7) Vai trò của Redis

Redis là lớp chính cho anti-spam runtime:

- `auth:cooldown:forgot:{channel}:{target}`
- `auth:cooldown:register:ip:{ip}`
- `auth:cooldown:login:account:{account}`
- `auth:rate:forgot:ip:{ip}:{window}`
- `auth:rate:forgot:target:{channel}:{target}:{window}`
- `auth:rate:login:ip:{ip}:{window}`
- `auth:rate:login:account:{account}:{window}`
- `auth:lock:login:{account}`

Nguyên tắc:

- Redis là lớp chặn nhanh
- DB là lớp audit và source of truth cho OTP

## 8) Vai trò của Database

DB giữ:

- OTP request
- trạng thái OTP
- số lần nhập sai
- audit gửi/xác minh/dùng OTP

Không nên chỉ lưu OTP trong Redis, vì:

- khó audit
- khó điều tra tranh chấp
- khó nối ERP sau này

## 9) Vai trò của Gate

`gate` không tự sinh OTP.

`gin` sẽ đẩy sang `gate` một notification command kiểu:

- `channel = email`
- `template = forgot-password-otp`
- `target = email user`
- `payload = { otp, expired_in, user_name }`

hoặc:

- `channel = sms`
- `template = forgot-password-otp`
- `target = phone user`
- `payload = { otp, expired_in }`

Điểm tách này giúp:

- auth logic không phụ thuộc nhà cung cấp mail/SMS
- đổi cổng gửi không phải sửa core auth

## 10) Bảng cần có

### `auth_otp_requests`

Lưu yêu cầu OTP cho quên mật khẩu và các mục đích xác minh sau này.

- `id`: bigint unsigned
- `user_id`: bigint unsigned nullable
- `channel`: tinyint
- `purpose`: tinyint
- `target`: varchar(255)
- `otp_hash`: varchar(255)
- `otp_last4`: varchar(10)
- `request_token`: varchar(100) unique
- `attempt_count`: int unsigned default `0`
- `max_attempts`: int unsigned default `5`
- `expires_at`: timestamp
- `verified_at`: timestamp nullable
- `used_at`: timestamp nullable
- `locked_at`: timestamp nullable
- `sent_at`: timestamp nullable
- `status`: tinyint
- `meta`: json nullable
- `created_at`: timestamp
- `updated_at`: timestamp

Index:

- unique(`request_token`)
- index(`user_id`, `purpose`, `status`)
- index(`channel`, `target`, `purpose`, `status`)
- index(`expires_at`, `status`)

### `auth_action_limits`

Bảng audit mềm cho abuse/rate-limit nếu muốn lưu lâu dài ngoài Redis.

- `id`: bigint unsigned
- `scope`: varchar(50)
- `action`: varchar(50)
- `subject_key`: varchar(255)
- `hit_count`: int unsigned
- `window_started_at`: timestamp
- `window_ended_at`: timestamp
- `last_hit_at`: timestamp
- `meta`: json nullable
- `created_at`: timestamp
- `updated_at`: timestamp

Lưu ý:

- bảng này không thay Redis
- chỉ dùng cho audit, BI, ERP hoặc điều tra abuse

## 11) Enum cần có

- `auth_otp_requests.channel`
  - `EMAIL = 1`
  - `PHONE = 2`

- `auth_otp_requests.purpose`
  - `RESET_PASSWORD = 1`
  - `VERIFY_EMAIL = 2`
  - `VERIFY_PHONE = 3`

- `auth_otp_requests.status`
  - `PENDING = 1`
  - `VERIFIED = 2`
  - `USED = 3`
  - `EXPIRED = 4`
  - `LOCKED = 5`
  - `CANCELLED = 6`

## 12) Điều kiện để bắt đầu code

Trước khi code, cần chốt 4 điểm:

1. OTP dài bao nhiêu số
2. reset qua `email`, `phone`, hay cho phép cả hai
3. có bắt buộc channel đã verify mới được reset không
4. `gate` sẽ gửi qua provider nào: mail, sms, hay cả hai

## 13) Khuyến nghị triển khai phase đầu

Phase đầu nên làm tối giản nhưng đúng:

- chống spam bằng `redis`
- lưu OTP bằng `DB`
- chỉ hỗ trợ `forgot password`
- gửi qua `gate`
- response generic
- chưa cần captcha ở phase đầu

Khi traffic tăng mới thêm:

- captcha
- device fingerprint mạnh hơn
- blacklist IP / ASN / country
- velocity rule nâng cao

