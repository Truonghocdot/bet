<?php

namespace App\Filament\Pages\Finance;

use App\Filament\Widgets\WalletBalanceTableWidget;
use BackedEnum;
use Filament\Pages\Concerns\HasMaxWidth;
use Filament\Pages\Concerns\HasTopbar;
use Filament\Pages\Page;
use Filament\Support\Icons\Heroicon;
use Illuminate\Contracts\Support\Htmlable;
use Illuminate\Support\Facades\Gate;
use UnitEnum;

class WalletBalancePage extends Page
{
    use HasMaxWidth;
    use HasTopbar;

    protected static bool $isDiscovered = true;

    protected static ?string $slug = 'finance/wallet-balances';

    protected static UnitEnum|string|null $navigationGroup = 'Tài chính';

    protected static ?string $navigationLabel = 'Số dư ví';

    protected static ?string $title = 'Số dư ví';

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedBanknotes;

    protected static ?int $navigationSort = 6;

    public static function canAccess(): bool
    {
        return Gate::allows('finance.wallets.update');
    }

    public function getTitle(): string | Htmlable
    {
        return static::$title;
    }

    public function getHeading(): string | Htmlable | null
    {
        return static::$title;
    }

    public function getSubheading(): string | Htmlable | null
    {
        return 'Danh sách người dùng và số dư ví khả dụng để điều chỉnh nhanh VND / USDT.';
    }

    protected function getHeaderWidgets(): array
    {
        return [
            WalletBalanceTableWidget::class,
        ];
    }

    public function getHeaderWidgetsColumns(): int | array
    {
        return 1;
    }
}
