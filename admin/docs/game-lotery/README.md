# Game Lottery (Wingo/K3/Lottery) - Business Spec

Tài liệu này mô tả nghiệp vụ cho nhóm game **lottery/round-based** trong hệ thống:

- `WINGO`
- `K3`
- `LOTTERY` (bao gồm 5D/3D tuỳ thiết kế)

Mục tiêu: chuẩn hóa cửa cược (`option_type`, `option_key`), chuẩn hóa kết quả (`result_payload`), và mô tả pipeline đặt cược -> khóa -> draw -> settlement -> payout.

DB reference:

- `game_periods`, `bet_tickets`, `bet_items`, `bet_settlements` trong `admin/database/database.md`.
- Enums trong `admin/app/Enum/Bet/*`.

## 1) Khái niệm chung

### 1.1 Variant/Room

Một game có thể có nhiều **variant** theo thời lượng kỳ:

- Wingo: `30s`, `1m`, `3m`, `5m`

Implementation:

- vẫn giữ `game_type = WINGO`
- phân biệt variant bằng `room_code` (khuyến nghị) hoặc encode trong `period_no`

Rule:

- mỗi variant có **1 room duy nhất** (theo yêu cầu hiện tại).

### 1.2 Vòng đời Period

Theo `PeriodStatus`:

- `SCHEDULED -> OPEN -> LOCKED -> DRAWN -> SETTLED`
- có thể `CANCELED`

Điểm chặn:

- chỉ nhận bet khi `OPEN`
- khi `LOCKED` thì reject đặt cược
- settlement chỉ chạy khi `DRAWN`

### 1.3 Ticket vs Items

- `bet_tickets`: order cấp user (1 ticket có thể single hoặc multi)
- `bet_items`: các dòng cược (color/number/big_small/odd_even/...)

Khuyến nghị cho UI kiểu play-view:

- mỗi lần user chọn 1 cửa và stake -> tạo `SINGLE ticket` 1 item
- nếu UI cho đặt nhiều cửa một lần -> `MULTI ticket` nhiều item

## 2) Chuẩn hóa cửa cược (option_type/option_key)

Enum mapping:

- `BetOptionType::NUMBER`
- `BetOptionType::BIG_SMALL`
- `BetOptionType::ODD_EVEN`
- `BetOptionType::COLOR`
- `BetOptionType::SUM`
- `BetOptionType::COMBINATION`

### 2.1 WINGO

UI play-view thể hiện:

- `COLOR`: `green`, `red`, `violet`
- `NUMBER`: `0..9`
- `BIG_SMALL`: `big`, `small`

`option_key` chuẩn:

- NUMBER: `number_0` .. `number_9`
- COLOR: `green`, `red`, `violet`
- BIG_SMALL: `big`, `small`
- ODD_EVEN (nếu mở): `odd`, `even`

### 2.2 K3

K3 là xúc xắc (3 dice) thường có:

- SUM: cược tổng 3..18
- ODD_EVEN: chẵn/lẻ (tổng)
- BIG_SMALL: lớn/nhỏ (tổng >= 11 là lớn, <= 10 là nhỏ) (rule phải chốt)
- COMBINATION: bộ 3 số (triples), đôi, hoặc “any triple”

`option_key` gợi ý:

- SUM: `sum_3` .. `sum_18`
- ODD_EVEN: `odd`, `even`
- BIG_SMALL: `big`, `small`
- COMBINATION:
  - `triple_any`
  - `triple_1_1_1` .. `triple_6_6_6`
  - `pair_1_1`, ...

### 2.3 LOTTERY (5D/3D)

Lottery kiểu 5D thường:

- NUMBER: chọn dãy số (ví dụ 5 chữ số)
- SUM: tổng chữ số
- ODD_EVEN/BIG_SMALL theo tổng hoặc theo chữ số cuối
- COMBINATION theo rule

Vì rule rất đa dạng, khuyến nghị:

- bắt buộc có `lottery_variant` (3D/4D/5D) trong `room_code`
- `option_key` encode rõ:
  - `pick5_01234`
  - `last_digit_7`
  - `sum_23`

## 3) Kết quả (result_payload) chuẩn hóa

`game_periods.result_payload` nên có shape chuẩn theo game_type, để settlement worker không phải đoán.

### 3.1 WINGO result_payload

Ví dụ:

- `number`: 0..9
- `color`: `green|red|violet`
- `big_small`: `big|small`
- `odd_even`: `odd|even`

### 3.2 K3 result_payload

- `dice`: `[d1, d2, d3]` (1..6)
- `sum`: 3..18
- `big_small`: `big|small`
- `odd_even`: `odd|even`
- `is_triple`: bool

### 3.3 LOTTERY result_payload

- `digits`: string (ví dụ `"01234"`)
- `sum`: int
- `last_digit`: int
- thêm field theo variant nếu cần

## 4) Đặt cược (place bet) - rule bắt buộc

### 4.1 Validation

- user `ACTIVE`, wallet `ACTIVE`
- period tồn tại, đúng `game_type` + `room_code` variant
- period `status = OPEN`
- amount > 0, không vượt max/min (config)
- item `option_type/option_key` hợp lệ với game_type

### 4.2 Idempotency

Mỗi request đặt cược phải có `request_id` (UUID hoặc string unique) để chống double-click:

- nếu nhận lại `request_id` đã xử lý -> trả response cũ

### 4.3 Wallet locking

Trong 1 DB transaction:

- lock wallet row (`FOR UPDATE`)
- trừ `wallet.balance`
- cộng `wallet.locked_balance`
- insert `bet_tickets` + `bet_items`
- ghi `wallet_ledger_entries` với `reference_type = bet_stake`

## 5) Settlement/Payout

### 5.1 Settlement trigger

- auto worker chạy khi period `DRAWN`
- admin manual override dùng `SettlementType::MANUAL`
- rollback dùng `SettlementType::ROLLBACK`

### 5.2 Payout logic (khung)

Settlement phải:

- chấm từng `bet_item`:
  - map `option_type/option_key` vs `result_payload`
  - set `result` (`WON/LOST/VOID/...`)
  - tính `payout_amount`
- tổng hợp về ticket:
  - set `bet_tickets.status`, `actual_payout`, `settled_at`
- wallet:
  - giảm `locked_balance` theo stake
  - cộng `balance` theo payout
- ledger:
  - ghi `bet_settlement`
- audit:
  - insert `bet_settlements` (immutable)

## 6) Mapping nhanh sang UI play-view (Wingo)

UI actions:

- chọn tab time (30s/1m/3m/5m) -> chọn `room_code`
- hiển thị `period_no`, countdown, recent results
- đặt cược:
  - `COLOR`: green/red/violet
  - `NUMBER`: number_0..number_9
  - `BIG_SMALL`: big/small
- multipliers X1..X50:
  - chỉ là UI helper để tính stake
  - backend chỉ nhận stake cuối

## 7) TODO cần chốt trước khi code game đầy đủ

- odds table cho từng game/option_key (cấu hình DB hay hardcode?)
- rule Wingo màu cho từng số (mapping number->color)
- rule `big_small`, `odd_even` cho Wingo/K3/Lottery
- LOTTERY variant cụ thể (3D/4D/5D) và format `digits`

