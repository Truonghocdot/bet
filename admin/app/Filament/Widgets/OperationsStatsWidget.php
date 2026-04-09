<?php

namespace App\Filament\Widgets;

use App\Enum\Bet\PeriodStatus;
use App\Enum\Transaction\TransactionStatus;
use App\Enum\Transaction\TypeTransaction;
use App\Enum\Transaction\WithdrawalStatus;
use App\Enum\User\RoleUser;
use App\Models\Bet\BetTicket;
use App\Models\Bet\GamePeriod;
use App\Models\Payment\VietQrBank;
use App\Models\Transaction\Transaction;
use App\Models\Transaction\WithdrawalRequest;
use App\Models\User;
use App\Services\Admin\ExchangeRateService;
use Filament\Support\Icons\Heroicon;
use Filament\Widgets\StatsOverviewWidget;
use Filament\Widgets\StatsOverviewWidget\Stat;

class OperationsStatsWidget extends StatsOverviewWidget
{
    protected static ?int $sort = 1;

    protected ?string $pollingInterval = '60s';

    protected ?string $heading = 'Tổng quan vận hành';

    protected ?string $description = 'Tỷ giá, doanh thu, lãi/lỗ và trạng thái xử lý hiện tại của ERP.';

    public static function canView(): bool
    {
        return auth()->check();
    }

    /**
     * @return array<Stat>
     */
    protected function getStats(): array
    {
        $exchangeRate = app(ExchangeRateService::class)->getSnapshot();

        $depositRevenue = (float) Transaction::query()
            ->where('type', TypeTransaction::DEPOSIT->value)
            ->where('status', TransactionStatus::COMPLETED->value)
            ->sum('net_amount');

        $withdrawalPaid = (float) WithdrawalRequest::query()
            ->where('status', WithdrawalStatus::PAID->value)
            ->sum('net_amount');

        $betStake = (float) BetTicket::query()->sum('stake');
        $betPayout = (float) BetTicket::query()->sum('actual_payout');
        $betProfitLoss = $betStake - $betPayout;

        $pendingWithdrawals = WithdrawalRequest::query()
            ->where('status', WithdrawalStatus::PENDING->value)
            ->count();

        $openPeriods = GamePeriod::query()
            ->where('status', PeriodStatus::OPEN->value)
            ->count();

        $clientCount = User::query()
            ->where('role', RoleUser::CLIENT->value)
            ->count();

        $bankCount = VietQrBank::query()->count();

        return [
            Stat::make('Người chơi', number_format($clientCount, 0, ',', '.'))
                ->description('Tài khoản khách hàng đang hoạt động')
                ->icon(Heroicon::OutlinedUsers)
                ->color('info'),

            Stat::make('Doanh thu nạp', $this->formatVnd($depositRevenue))
                ->description('Nạp thành công')
                ->icon(Heroicon::OutlinedArrowDownTray)
                ->color('success'),

            Stat::make('Tổng cược', $this->formatVnd($betStake))
                ->description('Tổng tiền đã vào cược')
                ->icon(Heroicon::OutlinedBanknotes)
                ->color('primary'),

            Stat::make('Tiền thắng trả', $this->formatVnd($betPayout))
                ->description('Đã trả cho vé thắng')
                ->icon(Heroicon::OutlinedArrowTrendingUp)
                ->color('warning'),

            Stat::make('Chi trả rút', $this->formatVnd($withdrawalPaid))
                ->description('Rút thủ công đã chi')
                ->icon(Heroicon::OutlinedArrowUpTray)
                ->color('danger'),

            Stat::make('Lãi/lỗ cược', $this->formatSignedVnd($betProfitLoss))
                ->description($betProfitLoss >= 0 ? 'Đang lãi' : 'Đang lỗ')
                ->icon(Heroicon::OutlinedScale)
                ->color($betProfitLoss >= 0 ? 'success' : 'danger'),

            Stat::make('Rút chờ duyệt', number_format($pendingWithdrawals, 0, ',', '.'))
                ->description('Yêu cầu đang đợi xử lý')
                ->icon(Heroicon::OutlinedClock)
                ->color($pendingWithdrawals > 0 ? 'warning' : 'success'),

            Stat::make('Kỳ đang mở', number_format($openPeriods, 0, ',', '.'))
                ->description('Kỳ game đang nhận cược')
                ->icon(Heroicon::OutlinedPlayCircle)
                ->color($openPeriods > 0 ? 'primary' : 'gray'),

            Stat::make('Tỷ giá USDT/VND', $this->formatVnd((float) $exchangeRate['rate']))
                ->description('Nguồn: '.($exchangeRate['source_name'] ?? 'manual'))
                ->icon(Heroicon::OutlinedCurrencyDollar)
                ->color('info'),

            Stat::make('Ngân hàng VietQR', number_format($bankCount, 0, ',', '.'))
                ->description('Danh mục ngân hàng đã đồng bộ')
                ->icon(Heroicon::OutlinedBuildingLibrary)
                ->color($bankCount > 0 ? 'success' : 'gray'),
        ];
    }

    private function formatVnd(float|int|string $value): string
    {
        return number_format((float) $value, 0, ',', '.') . ' VND';
    }

    private function formatSignedVnd(float|int|string $value): string
    {
        $numeric = (float) $value;
        $prefix = $numeric > 0 ? '+' : '';

        return $prefix . number_format($numeric, 0, ',', '.') . ' VND';
    }
}
