# Enum Convention

## Cấu trúc thư mục

- `User/*`: enum cho user và phân quyền
- `Wallet/*`: enum cho ví và ledger
- `Transaction/*`: enum cho nạp/rút
- `Bet/*`: enum cho game round-based (Wingo/K3/Lottery)
- `Affiliate/*`: enum cho referral theo mốc
- `Payment/*`: enum cho tài khoản nhận tiền (bank/crypto)

## Quy tắc bắt buộc

- Cột `tinyint` trong DB phải map đúng enum trong `database/database.md`.
- Không hard-code số trạng thái trong service/controller.
- Khi đổi enum phải cập nhật đồng thời migration + docs + business logic.

## Affiliate referral business (đang áp dụng)

- Người được mời chỉ hợp lệ khi có nạp đầu tiên `>= 50.000 VND`.
- Thưởng theo mốc cấu hình trong `affiliate_reward_settings`.
- Ví dụ mốc:
  - 3 người hợp lệ -> 50.000 VND
  - 5 người hợp lệ -> 80.000 VND

Enum chính:

- `AffiliateReferralStatus`:
  - `PENDING`: chưa đạt điều kiện nạp
  - `QUALIFIED`: đã đạt điều kiện
  - `INVALID`: bị loại khỏi chương trình
- `AffiliateRewardStatus`:
  - `PENDING`: đã ghi nhận mốc nhưng chưa trả
  - `PAID`: đã cộng ví và ghi ledger
  - `CANCELED`: hủy chi trả

## Payment receiving account

- `PaymentReceivingAccountType`:
  - `BANK`: tài khoản ngân hàng
  - `CRYPTO`: ví tiền ảo / blockchain address
- `PaymentReceivingAccountStatus`:
  - `ACTIVE`: được hiển thị cho UI/API
  - `INACTIVE`: ẩn khỏi UI/API
