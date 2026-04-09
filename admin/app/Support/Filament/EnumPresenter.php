<?php

namespace App\Support\Filament;

use App\Enum\Affiliate\AffiliateLinkStatus;
use App\Enum\Affiliate\AffiliateProfileStatus;
use App\Enum\Affiliate\AffiliateReferralStatus;
use App\Enum\Affiliate\AffiliateRewardStatus;
use App\Enum\Bet\BetItemResult;
use App\Enum\Bet\BetOptionType;
use App\Enum\Bet\BetStatus;
use App\Enum\Bet\BetTicketType;
use App\Enum\Bet\DrawSource;
use App\Enum\Bet\GameType;
use App\Enum\Bet\PeriodStatus;
use App\Enum\Bet\SettlementType;
use App\Enum\Payment\PaymentReceivingAccountStatus;
use App\Enum\Payment\PaymentReceivingAccountType;
use App\Enum\Transaction\TransactionStatus;
use App\Enum\Transaction\TypeTransaction;
use App\Enum\Transaction\WithdrawalStatus;
use App\Enum\User\RoleUser;
use App\Enum\User\UserStatus;
use App\Enum\Wallet\LedgerDirection;
use App\Enum\Wallet\UnitTransaction;
use App\Enum\Wallet\WalletStatus;
use BackedEnum;
use UnitEnum;

class EnumPresenter
{
    public static function options(string $enumClass): array
    {
        $options = [];

        foreach ($enumClass::cases() as $case) {
            $options[$case->value] = static::label($enumClass, $case);
        }

        return $options;
    }

    public static function label(string $enumClass, BackedEnum|UnitEnum|int|string|null $state): string
    {
        $value = static::normalize($state);
        $caseName = static::caseName($enumClass, $value);

        return static::labels($enumClass)[$caseName] ?? (string) $value;
    }

    public static function color(string $enumClass, BackedEnum|UnitEnum|int|string|null $state): string
    {
        $value = static::normalize($state);
        $caseName = static::caseName($enumClass, $value);

        return static::colors($enumClass)[$caseName] ?? 'gray';
    }

    private static function normalize(BackedEnum|UnitEnum|int|string|null $state): int|string|null
    {
        if ($state instanceof BackedEnum) {
            return $state->value;
        }

        if ($state instanceof UnitEnum) {
            return $state->name;
        }

        return $state;
    }

    private static function caseName(string $enumClass, int|string|null $value): ?string
    {
        if ($value === null) {
            return null;
        }

        foreach ($enumClass::cases() as $case) {
            if ((string) $case->value === (string) $value) {
                return $case->name;
            }
        }

        return null;
    }

    private static function labels(string $enumClass): array
    {
        return match ($enumClass) {
            RoleUser::class => [
                'ADMIN' => 'Quản trị viên',
                'CLIENT' => 'Người chơi',
                'STAFF' => 'Nhân sự vận hành',
                'AGENCY' => 'Đại lý',
            ],
            UserStatus::class => [
                'ACTIVE' => 'Hoạt động',
                'SUSPENDED' => 'Tạm khóa',
                'BANNED' => 'Chặn',
            ],
            WalletStatus::class => [
                'ACTIVE' => 'Hoạt động',
                'LOCKED' => 'Khóa',
            ],
            LedgerDirection::class => [
                'CREDIT' => 'Cộng',
                'DEBIT' => 'Trừ',
            ],
            UnitTransaction::class => [
                'VND' => 'VND',
                'USDT' => 'USDT',
            ],
            TransactionStatus::class => [
                'PENDING' => 'Chờ xử lý',
                'CONFIRMED' => 'Đã xác nhận',
                'COMPLETED' => 'Hoàn thành',
                'FAILED' => 'Thất bại',
                'CANCELED' => 'Đã hủy',
            ],
            TypeTransaction::class => [
                'DEPOSIT' => 'Nạp',
                'WITHDRAW' => 'Rút',
            ],
            WithdrawalStatus::class => [
                'PENDING' => 'Chờ duyệt',
                'APPROVED' => 'Đã duyệt',
                'REJECTED' => 'Từ chối',
                'CANCELED' => 'Đã hủy',
                'PAID' => 'Đã chi trả',
            ],
            PaymentReceivingAccountType::class => [
                'BANK' => 'Ngân hàng',
                'CRYPTO' => 'Tiền ảo',
            ],
            PaymentReceivingAccountStatus::class => [
                'ACTIVE' => 'Hiển thị',
                'INACTIVE' => 'Ẩn',
            ],
            GameType::class => [
                'WINGO' => 'Wingo',
                'K3' => 'K3',
                'LOTTERY' => 'Lottery',
            ],
            PeriodStatus::class => [
                'SCHEDULED' => 'Đã lên lịch',
                'OPEN' => 'Mở cược',
                'LOCKED' => 'Đã khóa',
                'DRAWN' => 'Đã ra kết quả',
                'SETTLED' => 'Đã chốt',
                'CANCELED' => 'Đã hủy',
            ],
            DrawSource::class => [
                'AUTO' => 'Tự động',
                'MANUAL' => 'Thủ công',
                'IMPORTED' => 'Import',
            ],
            BetTicketType::class => [
                'SINGLE' => 'Đơn',
                'MULTI' => 'Nhiều lựa chọn',
            ],
            BetOptionType::class => [
                'NUMBER' => 'Số',
                'BIG_SMALL' => 'Lớn/nhỏ',
                'ODD_EVEN' => 'Chẵn/lẻ',
                'COLOR' => 'Màu',
                'SUM' => 'Tổng',
                'COMBINATION' => 'Tổ hợp',
            ],
            BetStatus::class => [
                'PENDING' => 'Chờ chấm',
                'WON' => 'Thắng',
                'LOST' => 'Thua',
                'VOID' => 'Hoàn cược',
                'HALF_WON' => 'Thắng nửa',
                'HALF_LOST' => 'Thua nửa',
                'CANCELED' => 'Đã hủy',
                'CASHED_OUT' => 'Tất toán sớm',
            ],
            BetItemResult::class => [
                'PENDING' => 'Chờ chấm',
                'WON' => 'Thắng',
                'LOST' => 'Thua',
                'VOID' => 'Hoàn cược',
                'HALF_WON' => 'Thắng nửa',
                'HALF_LOST' => 'Thua nửa',
            ],
            SettlementType::class => [
                'AUTO' => 'Tự động',
                'MANUAL' => 'Thủ công',
                'ROLLBACK' => 'Hoàn tác',
            ],
            AffiliateProfileStatus::class => [
                'PENDING' => 'Chờ duyệt',
                'ACTIVE' => 'Hoạt động',
                'SUSPENDED' => 'Tạm khóa',
            ],
            AffiliateLinkStatus::class => [
                'ACTIVE' => 'Hoạt động',
                'PAUSED' => 'Tạm dừng',
            ],
            AffiliateReferralStatus::class => [
                'PENDING' => 'Chưa đủ điều kiện',
                'QUALIFIED' => 'Đủ điều kiện',
                'INVALID' => 'Không hợp lệ',
            ],
            AffiliateRewardStatus::class => [
                'PENDING' => 'Chờ trả',
                'PAID' => 'Đã trả',
                'CANCELED' => 'Đã hủy',
            ],
            default => [],
        };
    }

    private static function colors(string $enumClass): array
    {
        return match ($enumClass) {
            RoleUser::class => [
                'ADMIN' => 'danger',
                'CLIENT' => 'success',
                'STAFF' => 'warning',
                'AGENCY' => 'info',
            ],
            UserStatus::class => [
                'ACTIVE' => 'success',
                'SUSPENDED' => 'warning',
                'BANNED' => 'danger',
            ],
            WalletStatus::class => [
                'ACTIVE' => 'success',
                'LOCKED' => 'danger',
            ],
            LedgerDirection::class => [
                'CREDIT' => 'success',
                'DEBIT' => 'danger',
            ],
            UnitTransaction::class => [
                'VND' => 'info',
                'USDT' => 'success',
            ],
            TransactionStatus::class => [
                'PENDING' => 'warning',
                'CONFIRMED' => 'info',
                'COMPLETED' => 'success',
                'FAILED' => 'danger',
                'CANCELED' => 'gray',
            ],
            TypeTransaction::class => [
                'DEPOSIT' => 'success',
                'WITHDRAW' => 'danger',
            ],
            WithdrawalStatus::class => [
                'PENDING' => 'warning',
                'APPROVED' => 'info',
                'REJECTED' => 'danger',
                'CANCELED' => 'gray',
                'PAID' => 'success',
            ],
            PaymentReceivingAccountType::class => [
                'BANK' => 'info',
                'CRYPTO' => 'success',
            ],
            PaymentReceivingAccountStatus::class => [
                'ACTIVE' => 'success',
                'INACTIVE' => 'gray',
            ],
            GameType::class => [
                'WINGO' => 'success',
                'K3' => 'warning',
                'LOTTERY' => 'info',
            ],
            PeriodStatus::class => [
                'SCHEDULED' => 'gray',
                'OPEN' => 'success',
                'LOCKED' => 'warning',
                'DRAWN' => 'info',
                'SETTLED' => 'success',
                'CANCELED' => 'danger',
            ],
            DrawSource::class => [
                'AUTO' => 'success',
                'MANUAL' => 'warning',
                'IMPORTED' => 'info',
            ],
            BetTicketType::class => [
                'SINGLE' => 'info',
                'MULTI' => 'warning',
            ],
            BetOptionType::class => [
                'NUMBER' => 'info',
                'BIG_SMALL' => 'warning',
                'ODD_EVEN' => 'info',
                'COLOR' => 'danger',
                'SUM' => 'success',
                'COMBINATION' => 'warning',
            ],
            BetStatus::class => [
                'PENDING' => 'warning',
                'WON' => 'success',
                'LOST' => 'danger',
                'VOID' => 'gray',
                'HALF_WON' => 'info',
                'HALF_LOST' => 'warning',
                'CANCELED' => 'gray',
                'CASHED_OUT' => 'info',
            ],
            BetItemResult::class => [
                'PENDING' => 'warning',
                'WON' => 'success',
                'LOST' => 'danger',
                'VOID' => 'gray',
                'HALF_WON' => 'info',
                'HALF_LOST' => 'warning',
            ],
            SettlementType::class => [
                'AUTO' => 'success',
                'MANUAL' => 'warning',
                'ROLLBACK' => 'danger',
            ],
            AffiliateProfileStatus::class => [
                'PENDING' => 'warning',
                'ACTIVE' => 'success',
                'SUSPENDED' => 'danger',
            ],
            AffiliateLinkStatus::class => [
                'ACTIVE' => 'success',
                'PAUSED' => 'gray',
            ],
            AffiliateReferralStatus::class => [
                'PENDING' => 'warning',
                'QUALIFIED' => 'success',
                'INVALID' => 'danger',
            ],
            AffiliateRewardStatus::class => [
                'PENDING' => 'warning',
                'PAID' => 'success',
                'CANCELED' => 'gray',
            ],
            default => [],
        };
    }
}
