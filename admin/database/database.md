# Database Design - Wingo / K3 / Lottery + Affiliate Referral

## 1) Mục tiêu triển khai

Tài liệu này mô tả schema DB cho:

- ví tiền, nạp/rút
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
