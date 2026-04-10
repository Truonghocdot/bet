# Database Design - Wingo / K3 / Lottery + Affiliate Referral

## 1) Mục tiêu triển khai

Tài liệu này mô tả schema DB cho:

- ví tiền, nạp/rút
- tỷ giá USDT/VND dùng chung cho Laravel và Gin
- game cược theo kỳ quay: `Wingo`, `K3`, `Lottery`
- settlement và audit tài chính
- affiliate referral theo mốc số người mời hợp lệ

## 2) Quy ước chung

- PK/FK: `bigint unsigned`
- tiền: `decimal(20,8)` hoặc `decimal(14,6)` cho odds
- trạng thái/loại: `tinyint` + enum ở `app/Enum/*`
- mọi bảng nghiệp vụ có `created_at`, `updated_at`
- bảng cần truy vết có `deleted_at`

## 3) Map Enum bắt buộc

### User / Wallet / Transaction

- `users.role` -> `App\Enum\User\RoleUser`
- `users.status` -> `App\Enum\User\UserStatus`
- `auth_otp_requests.channel` -> `App\Enum\Auth\OtpChannel`
- `auth_otp_requests.purpose` -> `App\Enum\Auth\OtpPurpose`
- `auth_otp_requests.status` -> `App\Enum\Auth\OtpStatus`
- `wallets.unit` -> `App\Enum\Wallet\UnitTransaction`
- `wallets.status` -> `App\Enum\Wallet\WalletStatus`
- `wallet_ledger_entries.direction` -> `App\Enum\Wallet\LedgerDirection`
- `transactions.type` -> `App\Enum\Transaction\TypeTransaction`
- `transactions.status` -> `App\Enum\Transaction\TransactionStatus`
- `withdrawal_requests.status` -> `App\Enum\Transaction\WithdrawalStatus`

### Bet

- `game_periods.game_type` -> `App\Enum\Bet\GameType`
- `game_periods.status` -> `App\Enum\Bet\PeriodStatus`
- `game_periods.draw_source` -> `App\Enum\Bet\DrawSource`
- `bet_tickets.bet_type` -> `App\Enum\Bet\BetTicketType`
- `bet_tickets.status` -> `App\Enum\Bet\BetStatus`
- `bet_items.option_type` -> `App\Enum\Bet\BetOptionType`
- `bet_items.result` -> `App\Enum\Bet\BetItemResult`
- `bet_settlements.settlement_type` -> `App\Enum\Bet\SettlementType`

### Affiliate Referral

- `affiliate_profiles.status` -> `App\Enum\Affiliate\AffiliateProfileStatus`
- `affiliate_links.status` -> `App\Enum\Affiliate\AffiliateLinkStatus`
- `affiliate_referrals.status` -> `App\Enum\Affiliate\AffiliateReferralStatus`
- `affiliate_reward_logs.status` -> `App\Enum\Affiliate\AffiliateRewardStatus`

### Payment Receiving Account

- `payment_receiving_accounts.type` -> `App\Enum\Payment\PaymentReceivingAccountType`
- `payment_receiving_accounts.status` -> `App\Enum\Payment\PaymentReceivingAccountStatus`

## 4) Core tables

### `users`

Người dùng toàn hệ thống.

- `id`: bigint unsigned
- `name`: varchar(100)
- `email`: varchar(255)
- `phone`: varchar(20) nullable
- `password`: varchar(255)
- `role`: tinyint
- `status`: tinyint
- `email_verified_at`: timestamp nullable
- `phone_verified_at`: timestamp nullable
- `last_login_at`: timestamp nullable
- `created_at`: timestamp
- `updated_at`: timestamp
- `deleted_at`: timestamp nullable

Index:

- unique(`email`)
- unique(`phone`)
- index(`role`, `status`)

### `auth_otp_requests`

Lưu yêu cầu OTP cho các mục đích auth, trước mắt dùng cho `forgot password`.

- `id`: bigint unsigned
- `user_id`: bigint unsigned nullable
- `channel`: tinyint
- `purpose`: tinyint
- `target`: varchar(255)
- `otp_hash`: varchar(255)
- `otp_last4`: varchar(10)
- `request_token`: varchar(100)
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

Constraint/index:

- unique(`request_token`)
- index(`user_id`, `purpose`, `status`)
- index(`channel`, `target`, `purpose`, `status`)
- index(`expires_at`, `status`)

Nghiệp vụ:

- Chỉ lưu `otp_hash`, không lưu OTP plaintext.
- Mỗi lần tạo OTP reset password mới phải hủy các OTP `PENDING` cũ cùng user/channel/purpose.
- OTP dùng một lần:
  - sau khi đổi mật khẩu thành công phải chuyển `status = USED`
  - set `used_at`
- Nếu nhập sai quá `max_attempts` thì chuyển `status = LOCKED`.
- Response public cho bước gửi OTP phải là generic, không để lộ tài khoản có tồn tại hay không.

### `auth_action_limits`

Bảng audit mềm cho anti-spam / abuse.

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

Index:

- index(`scope`, `action`)
- index(`subject_key`, `action`)
- index(`window_ended_at`)

Nghiệp vụ:

- Bảng này không thay thế Redis rate limit.
- Redis là lớp chặn nhanh runtime.
- DB chỉ dùng để audit, phân tích abuse, hiển thị ERP nếu cần.

### `wallets`

Ví theo từng đơn vị tiền của user.

- `id`: bigint unsigned
- `user_id`: bigint unsigned
- `unit`: tinyint
- `balance`: decimal(20,8)
- `locked_balance`: decimal(20,8)
- `status`: tinyint
- `created_at`: timestamp
- `updated_at`: timestamp

Constraint:

- unique(`user_id`, `unit`)

### `wallet_ledger_entries`

Sổ cái ví để audit toàn bộ biến động số dư.

- `id`: bigint unsigned
- `wallet_id`: bigint unsigned
- `user_id`: bigint unsigned
- `direction`: tinyint
- `amount`: decimal(20,8)
- `balance_before`: decimal(20,8)
- `balance_after`: decimal(20,8)
- `reference_type`: varchar(50)
- `reference_id`: bigint unsigned nullable
- `note`: varchar(255) nullable
- `created_at`: timestamp

Index:

- index(`wallet_id`, `created_at`)
- index(`user_id`, `created_at`)
- index(`reference_type`, `reference_id`)

## 5) Deposit / Withdraw

### `transactions`

Giao dịch nạp/rút.

- `id`: bigint unsigned
- `user_id`: bigint unsigned
- `wallet_id`: bigint unsigned
- `unit`: tinyint
- `type`: tinyint
- `amount`: decimal(20,8)
- `fee`: decimal(20,8) default `0`
- `net_amount`: decimal(20,8)
- `status`: tinyint
- `provider`: varchar(50) nullable
- `provider_txn_id`: varchar(100) nullable
- `reason_failed`: text nullable
- `approved_by`: bigint unsigned nullable
- `approved_at`: timestamp nullable
- `created_at`: timestamp
- `updated_at`: timestamp
- `deleted_at`: timestamp nullable

### `withdrawal_requests`

Yêu cầu rút tiền.

- `id`: bigint unsigned
- `user_id`: bigint unsigned
- `wallet_id`: bigint unsigned
- `account_withdrawal_info_id`: bigint unsigned
- `unit`: tinyint
- `amount`: decimal(20,8)
- `fee`: decimal(20,8) default `0`
- `net_amount`: decimal(20,8)
- `status`: tinyint
- `reason_rejected`: text nullable
- `reviewed_by`: bigint unsigned nullable
- `reviewed_at`: timestamp nullable
- `created_at`: timestamp
- `updated_at`: timestamp
- `deleted_at`: timestamp nullable

### `account_withdrawal_infos`

Thông tin tài khoản nhận rút của user.

- `id`: bigint unsigned
- `user_id`: bigint unsigned
- `unit`: tinyint
- `provider_code`: varchar(50)
- `account_name`: varchar(255)
- `account_number`: varchar(255)
- `is_default`: boolean
- `created_at`: timestamp
- `updated_at`: timestamp
- `deleted_at`: timestamp nullable

### `payment_receiving_accounts`

Danh sách tài khoản hệ thống dùng để nhận tiền nạp từ user.  
Đây là bảng cấu hình nội bộ cho admin, không phải tài khoản của user.

- `id`: bigint unsigned
- `code`: varchar(50)
- `name`: varchar(100)
- `type`: tinyint
- `unit`: tinyint
- `provider_code`: varchar(50) nullable
- `account_name`: varchar(255) nullable
- `account_number`: varchar(255) nullable
- `wallet_address`: varchar(255) nullable
- `network`: varchar(50) nullable
- `qr_code_path`: varchar(255) nullable
- `instructions`: text nullable
- `status`: tinyint
- `is_default`: boolean
- `sort_order`: int unsigned
- `created_at`: timestamp
- `updated_at`: timestamp
- `deleted_at`: timestamp nullable

Constraint/index:

- unique(`code`)
- index(`type`, `unit`, `status`)
- index(`is_default`, `status`)

Nghiệp vụ:

- `type = BANK`:
  - `provider_code`, `account_name`, `account_number` là dữ liệu chính.
  - `unit` thường là `VND`.
  - `wallet_address` và `network` để `null`.
- `type = CRYPTO`:
  - `wallet_address`, `network` là dữ liệu chính.
  - `unit` thường là `USDT`.
  - `account_number` có thể để `null` hoặc dùng như mã tham chiếu nội bộ nếu cần.
- Chỉ `status = ACTIVE` mới được API/public UI hiển thị.
- `is_default = true` dùng để chọn sẵn tài khoản mặc định cho từng `unit` hoặc từng `type`.
- `sort_order` dùng để sắp xếp hiển thị khi có nhiều tài khoản nhận tiền.

### `vietqr_banks`

Danh mục ngân hàng Việt Nam lấy từ API VietQR.  
Đây là bảng reference để dùng chung cho admin, API và các service khác.  
Service Laravel sẽ đồng bộ dữ liệu từ nguồn `https://api.vietqr.io/v2/banks`, lưu vào DB, sau đó prime sang cache và Redis raw JSON.

- `id`: bigint unsigned
- `source_id`: int unsigned
- `code`: varchar(50)
- `name`: varchar(255)
- `short_name`: varchar(100)
- `bin`: varchar(20)
- `logo`: varchar(255) nullable
- `transfer_supported`: boolean
- `lookup_supported`: boolean
- `support`: tinyint nullable
- `raw_payload`: json nullable
- `synced_at`: timestamp nullable
- `created_at`: timestamp
- `updated_at`: timestamp

Constraint/index:

- unique(`source_id`)
- unique(`code`)
- unique(`bin`)
- index(`short_name`)
- index(`transfer_supported`, `lookup_supported`)

Nghiệp vụ:

- API VietQR trả về danh sách ngân hàng ở `data[]`.
- `raw_payload` giữ nguyên object gốc để tương thích khi API bổ sung field mới.
- Cron `banks:sync-vietqr` sẽ:
  - gọi API VietQR
  - upsert toàn bộ ngân hàng vào DB
  - xóa các ngân hàng không còn trong danh sách nguồn
  - prime cache Laravel
  - ghi raw JSON sang Redis key dùng chung cho service khác
- Laravel cache key: `admin:vietqr:banks:snapshot`
- Redis shared key raw JSON: `shared:vietqr:banks`
- Redis connection dùng chung: `shared`
- Service khác, bao gồm Gin, chỉ cần đọc raw JSON từ Redis key này là dùng lại được ngay, không phụ thuộc serializer của Laravel.
- TTL cache mặc định: 24 giờ, phù hợp với danh mục ngân hàng ít thay đổi.

## 5) Tỷ giá USDT/VND

### `exchange_rate_settings`

Bảng cấu hình 1 dòng cho tỷ giá `USDT/VND`.  
Đây là nguồn sự thật trong DB, còn Redis / Cache chỉ là lớp runtime để đọc nhanh và chia sẻ sang service Gin.

- `id`: bigint unsigned
- `code`: varchar(50)
- `base_currency`: varchar(10)
- `quote_currency`: varchar(10)
- `rate`: decimal(20,8)
- `source_rate`: decimal(20,8) nullable
- `auto_sync`: boolean
- `source_name`: varchar(100) nullable
- `last_synced_at`: timestamp nullable
- `updated_by`: bigint unsigned nullable
- `note`: text nullable
- `created_at`: timestamp
- `updated_at`: timestamp

Constraint:

- unique(`code`)

Nghiệp vụ:

- Chỉ có 1 dòng `code = USDT_VND`.
- `rate` là tỷ giá áp dụng thực tế cho toàn hệ thống.
- `source_rate` là tỷ giá lấy từ nguồn provider hoặc nguồn đồng bộ gần nhất.
- Nếu `auto_sync = true`, cron có thể ghi đè `rate` bằng `source_rate`.
- Nếu `auto_sync = false`, cron chỉ cập nhật `source_rate`, `source_name`, `last_synced_at`, còn `rate` giữ nguyên để admin chỉnh tay.
- Khi admin chỉnh trong Filament page:
  - cập nhật DB
  - đẩy snapshot vào cache Laravel
  - ghi raw JSON sang Redis key dùng chung cho Gin
- Laravel cache key: `admin:exchange-rate:usdt-vnd:snapshot`
- Redis shared key raw JSON: `shared:exchange-rate:usdt-vnd`
- Redis connection dùng chung: `shared`
- Gin chỉ đọc raw JSON từ Redis key này, không dùng serializer / cache store của Laravel.
- Nếu Redis key bị mất, service Laravel phải tự prime lại từ DB khi page hoặc command truy cập.
- Command đồng bộ: `rates:sync-usdt-vnd`
- Lịch chạy: mỗi 5 phút, không cho chạy chồng nhau.

## 6) Bet flow chi tiết

### 6.1 Enum nghiệp vụ Bet

#### `GameType`

- `WINGO`: game theo kết quả số, màu, lớn/nhỏ.
- `K3`: game theo tổng/xúc xắc.
- `LOTTERY`: game theo kết quả xổ số.

Note:

- `game_type` quyết định parser kết quả, rule cược, và công thức payout.
- Không dùng chung một hàm settlement nếu rule game khác nhau.

#### `PeriodStatus`

- `SCHEDULED`: đã tạo kỳ, chưa mở cược.
- `OPEN`: đang nhận cược.
- `LOCKED`: đã khóa cược, chờ kết quả.
- `DRAWN`: đã có kết quả raw.
- `SETTLED`: đã chấm xong và cập nhật ví.
- `CANCELED`: kỳ bị hủy.

Transition chuẩn:

- `SCHEDULED -> OPEN -> LOCKED -> DRAWN -> SETTLED`
- `SCHEDULED/OPEN/LOCKED -> CANCELED`

Note:

- Không cho phép `SETTLED` quay lại `OPEN`.
- Không ghi đè `result_payload` nếu kỳ đã `SETTLED`, trừ case rollback có audit riêng.

#### `DrawSource`

- `AUTO`: lấy tự động từ provider hoặc job.
- `MANUAL`: admin nhập tay.
- `IMPORTED`: import batch.

Note:

- Phải lưu `draw_source` + `result_hash` để audit nguồn kết quả.

#### `BetTicketType`

- `SINGLE`: một lựa chọn.
- `MULTI`: nhiều lựa chọn trong một ticket.

Note:

- `MULTI` phải quy định rõ cách tính odds tổng: multiply hay custom rule.

#### `BetStatus`

- `PENDING`: chưa chấm.
- `WON`: thắng.
- `LOST`: thua.
- `VOID`: hoàn cược.
- `HALF_WON`: thắng nửa.
- `HALF_LOST`: thua nửa.
- `CANCELED`: hủy vé.
- `CASHED_OUT`: tất toán sớm.

Note:

- Trạng thái terminal: `WON`, `LOST`, `VOID`, `HALF_WON`, `HALF_LOST`, `CANCELED`, `CASHED_OUT`.
- Ticket terminal không được chỉnh stake/odds.

#### `BetOptionType`

- `NUMBER`: cược số cụ thể.
- `BIG_SMALL`: lớn/nhỏ.
- `ODD_EVEN`: chẵn/lẻ.
- `COLOR`: màu.
- `SUM`: tổng.
- `COMBINATION`: tổ hợp.

Note:

- `option_key` phải chuẩn hóa theo `game_type`.
- Ví dụ Wingo: `big`, `small`, `odd`, `even`, `green`, `red`, `violet`, `number_0..9`.

#### `BetItemResult`

- `PENDING`: chưa chấm.
- `WON`: thắng.
- `LOST`: thua.
- `VOID`: hoàn cược.
- `HALF_WON`: thắng nửa.
- `HALF_LOST`: thua nửa.

Note payout:

- `WON`: trả đủ theo odds.
- `HALF_WON`: trả một phần theo rule game.
- `VOID`: hoàn stake.
- `LOST`: mất stake.

#### `SettlementType`

- `AUTO`: worker tự động chấm theo `result_payload`.
- `MANUAL`: admin chấm tay hoặc override có kiểm soát.
- `ROLLBACK`: hoàn tác settlement trước đó để chấm lại.

Note xử lý:

- `AUTO`: chỉ chạy khi period `DRAWN`, phải idempotent theo `period_id`.
- `MANUAL`: bắt buộc có `note` và `settled_by`.
- `ROLLBACK`: tạo bút toán đảo ví/ledger, không sửa lịch sử cũ.

### 6.2 `game_periods`

Mỗi record là một kỳ quay của một game.

- `id`: bigint unsigned
- `game_type`: tinyint
- `period_no`: varchar(50)
- `room_code`: varchar(30) nullable
- `open_at`: timestamp
- `close_at`: timestamp
- `draw_at`: timestamp
- `settled_at`: timestamp nullable
- `status`: tinyint
- `draw_source`: tinyint nullable
- `result_payload`: json nullable
- `result_hash`: varchar(255) nullable
- `created_at`: timestamp
- `updated_at`: timestamp

Constraint/index:

- unique(`game_type`, `period_no`)
- index(`status`, `draw_at`)
- index(`game_type`, `close_at`)

Nghiệp vụ:

- Scheduler tạo kỳ kế tiếp khi kỳ hiện tại sắp hết hạn.
- Hết thời gian cược: `OPEN -> LOCKED`.
- Chỉ cho phép settle khi `status = DRAWN`.

### 6.3 `bet_tickets`

Đại diện đơn cược cấp order.

- `id`: bigint unsigned
- `ticket_no`: varchar(40)
- `user_id`: bigint unsigned
- `wallet_id`: bigint unsigned
- `unit`: tinyint
- `game_type`: tinyint
- `period_id`: bigint unsigned
- `bet_type`: tinyint
- `stake`: decimal(20,8)
- `total_odds`: decimal(14,6)
- `potential_payout`: decimal(20,8)
- `actual_payout`: decimal(20,8) nullable
- `status`: tinyint
- `placed_ip`: varchar(45) nullable
- `placed_device`: varchar(100) nullable
- `settled_at`: timestamp nullable
- `created_at`: timestamp
- `updated_at`: timestamp

Constraint/index:

- unique(`ticket_no`)
- index(`user_id`, `created_at`)
- index(`period_id`, `status`)
- index(`game_type`, `created_at`)

Nghiệp vụ:

- Tạo ticket phải chạy trong DB transaction với lock wallet.
- Khi đặt cược: trừ `balance`, cộng `locked_balance`, ghi ledger `bet_stake`.
- Không cho sửa ticket sau khi `status != PENDING`.

### 6.4 `bet_items`

Chi tiết từng dòng cược trong ticket.

- `id`: bigint unsigned
- `ticket_id`: bigint unsigned
- `period_id`: bigint unsigned
- `option_type`: tinyint
- `option_key`: varchar(100)
- `option_label`: varchar(150)
- `odds_at_placement`: decimal(12,4)
- `stake`: decimal(20,8)
- `result`: tinyint
- `payout_amount`: decimal(20,8) nullable
- `result_payload`: json nullable
- `settled_at`: timestamp nullable
- `created_at`: timestamp
- `updated_at`: timestamp

Nghiệp vụ:

- `odds_at_placement` là snapshot cố định, không đổi theo odds runtime.
- `result_payload` lưu dữ liệu chấm chi tiết để debug tranh chấp.

### 6.5 `bet_settlements`

Audit bất biến cho mỗi lần settle hoặc rollback.

- `id`: bigint unsigned
- `ticket_id`: bigint unsigned
- `period_id`: bigint unsigned
- `settlement_type`: tinyint
- `before_status`: tinyint
- `after_status`: tinyint
- `payout_amount`: decimal(20,8)
- `profit_loss`: decimal(20,8)
- `note`: varchar(255) nullable
- `settled_by`: bigint unsigned nullable
- `created_at`: timestamp

Index:

- index(`ticket_id`, `created_at`)
- index(`period_id`, `created_at`)

Nghiệp vụ theo `settlement_type`:

- `AUTO`:
  - Input: `period_id`, `result_payload` đã finalized.
  - Process: chấm từng `bet_items` -> tổng `actual_payout` -> update `bet_tickets`.
  - Wallet: giảm `locked_balance`, cộng `balance` theo payout.
  - Ledger: tạo record `bet_settlement`.
- `MANUAL`:
  - Chỉ admin/staff có quyền.
  - Bắt buộc ghi `note` lý do override.
  - Vẫn phải đi qua cùng pipeline wallet + ledger như AUTO.
- `ROLLBACK`:
  - Chỉ cho ticket đã có settlement trước đó.
  - Tạo settlement record mới với type `ROLLBACK`.
  - Tạo bút toán đảo (`reverse`) trong ledger, không update cứng ledger cũ.
  - Đưa ticket về trạng thái `PENDING` hoặc trạng thái trước settle theo thiết kế.

## 7) Affiliate Referral

### 7.1 Rule nghiệp vụ

- User A mời User B bằng `ref_code`.
- User B chỉ được tính là mời hợp lệ khi có giao dịch nạp `COMPLETED` đầu tiên >= `50.000 VND`.
- Thưởng được tính theo mốc số người mời hợp lệ, không tính % doanh thu.
- Ví dụ:
  - đạt 3 người hợp lệ -> thưởng 50.000 VND
  - đạt 5 người hợp lệ -> thưởng 80.000 VND

### 7.2 `affiliate_profiles`

- `id`: bigint unsigned
- `user_id`: bigint unsigned
- `ref_code`: varchar(50)
- `ref_link`: varchar(255)
- `status`: tinyint
- `created_at`: timestamp
- `updated_at`: timestamp

### 7.3 `affiliate_links`

- `id`: bigint unsigned
- `affiliate_profile_id`: bigint unsigned
- `campaign_name`: varchar(100)
- `tracking_code`: varchar(100)
- `landing_url`: varchar(255)
- `status`: tinyint
- `created_at`: timestamp
- `updated_at`: timestamp

### 7.4 `affiliate_referrals`

Lưu từng user được mời và trạng thái đạt điều kiện.

- `id`: bigint unsigned
- `affiliate_profile_id`: bigint unsigned
- `referrer_user_id`: bigint unsigned
- `referred_user_id`: bigint unsigned
- `affiliate_link_id`: bigint unsigned nullable
- `first_deposit_transaction_id`: bigint unsigned nullable
- `first_deposit_amount`: decimal(20,8) nullable
- `qualified_at`: timestamp nullable
- `status`: tinyint
- `created_at`: timestamp
- `updated_at`: timestamp

Constraint:

- unique(`referred_user_id`)

### 7.5 `affiliate_reward_settings`

Bảng cấu hình mốc thưởng theo số người mời hợp lệ.

- `id`: bigint unsigned
- `name`: varchar(100)
- `required_qualified_referrals`: int unsigned
- `reward_amount`: decimal(20,8)
- `unit`: tinyint
- `is_active`: boolean
- `effective_from`: timestamp nullable
- `effective_to`: timestamp nullable
- `note`: varchar(255) nullable
- `created_at`: timestamp
- `updated_at`: timestamp

Seed gợi ý:

- `{required_qualified_referrals: 3, reward_amount: 50000}`
- `{required_qualified_referrals: 5, reward_amount: 80000}`

### 7.6 `affiliate_reward_logs`

Lưu lịch sử trả thưởng để không trả trùng mốc.

- `id`: bigint unsigned
- `affiliate_profile_id`: bigint unsigned
- `referrer_user_id`: bigint unsigned
- `setting_id`: bigint unsigned
- `required_qualified_referrals`: int unsigned
- `actual_qualified_referrals`: int unsigned
- `reward_amount`: decimal(20,8)
- `unit`: tinyint
- `status`: tinyint
- `wallet_ledger_entry_id`: bigint unsigned nullable
- `granted_at`: timestamp nullable
- `created_at`: timestamp
- `updated_at`: timestamp

Constraint:

- unique(`affiliate_profile_id`, `setting_id`)

## 8) Luồng xử lý affiliate referral

### 8.1 Gắn người mời

- Khi đăng ký, nhận `ref_code`.
- Tạo `affiliate_referrals` với `status = PENDING`.

### 8.2 Xét điều kiện hợp lệ

- Khi `transactions` có bản ghi nạp `COMPLETED` đầu tiên của user được mời:
  - nếu `unit = VND` và `amount >= 50.000`:
    - `status = QUALIFIED`
    - lưu `first_deposit_transaction_id`
    - lưu `first_deposit_amount`
    - lưu `qualified_at`
  - nếu nhỏ hơn 50k: giữ `PENDING`

### 8.3 Tính thưởng theo mốc

- Đếm số `QUALIFIED` của mỗi affiliate profile.
- So với `affiliate_reward_settings` đang active.
- Mốc nào đạt mà chưa có `affiliate_reward_logs` thì tạo log + cộng ví:
  - cộng `wallet.balance`
  - ghi `wallet_ledger_entries` với `reference_type = affiliate_referral_reward`
  - cập nhật `affiliate_reward_logs.status = PAID`

## 9) Lưu ý kỹ thuật bắt buộc

- Toàn bộ cộng/trừ ví dùng DB transaction + lock row wallet (`FOR UPDATE`).
- Place bet/settle phải idempotent bằng key nghiệp vụ (`ticket_no`, `period_id`).
- Settlement job phải có lock theo `period_id` để tránh chạy song song.
- Chỉ xét giao dịch nạp `COMPLETED` cho affiliate.
- Điều kiện tối thiểu 50k nên để config hoặc env, không hard-code trong service.
- Auth anti-spam phải có cả Redis key runtime và DB audit nếu hệ thống cần điều tra abuse.
- OTP reset password chỉ lưu hash, không lưu plaintext trong DB.
- Luồng `forgot password` phải trả response generic để tránh lộ user enumeration.
