<?php

namespace App\Filament\Widgets;

use App\Enum\User\RoleUser;
use App\Enum\User\UserStatus;
use App\Enum\Wallet\UnitTransaction;
use App\Models\User;
use App\Services\Admin\UserWalletBalanceService;
use App\Support\Filament\EnumPresenter;
use Filament\Notifications\Notification;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Columns\TextInputColumn;
use Filament\Tables\Table;
use Filament\Widgets\TableWidget;
use Illuminate\Support\Facades\Gate;

class WalletBalanceTableWidget extends TableWidget
{
    protected static ?int $sort = 1;

    protected static ?string $heading = 'Điều chỉnh số dư ví người dùng';

    protected int | string | array $columnSpan = 'full';

    protected ?string $pollingInterval = null;

    public static function canView(): bool
    {
        return Gate::allows('finance.wallets.update');
    }

    public function table(Table $table): Table
    {
        return $table
            ->query(
                User::query()
                    ->with(['wallets' => fn ($query) => $query->select('id', 'user_id', 'unit', 'balance', 'locked_balance')])
                    ->select(['id', 'name', 'phone', 'role', 'status', 'created_at']),
            )
            ->defaultSort('id', 'desc')
            ->paginated([25, 50, 100])
            ->description('Cập nhật trực tiếp số dư khả dụng của ví VND và USDT. Mỗi lần chỉnh sửa sẽ ghi nhận ledger.')
            ->columns([
                TextColumn::make('id')
                    ->label('ID')
                    ->sortable()
                    ->searchable(),
                TextColumn::make('name')
                    ->label('Người dùng')
                    ->searchable()
                    ->sortable(),
                TextColumn::make('phone')
                    ->label('SĐT')
                    ->searchable(),
                TextColumn::make('role')
                    ->label('Vai trò')
                    ->badge()
                    ->formatStateUsing(fn ($state): string => EnumPresenter::label(RoleUser::class, $state))
                    ->color(fn ($state): string => EnumPresenter::color(RoleUser::class, $state)),
                TextColumn::make('status')
                    ->label('Trạng thái')
                    ->badge()
                    ->formatStateUsing(fn ($state): string => EnumPresenter::label(UserStatus::class, $state))
                    ->color(fn ($state): string => EnumPresenter::color(UserStatus::class, $state)),
                TextInputColumn::make('wallet_vnd_balance')
                    ->label('Ví VND')
                    ->type('number')
                    ->inputMode('decimal')
                    ->step('0.000001')
                    ->rules(['nullable', 'numeric'])
                    ->getStateUsing(fn (User $record): string => $this->getWalletBalance($record, UnitTransaction::VND))
                    ->updateStateUsing(function (User $record, $state): string {
                        return $this->syncWalletBalance($record, UnitTransaction::VND, $state);
                    }),
                TextColumn::make('wallet_vnd_locked')
                    ->label('Locked VND')
                    ->state(fn (User $record): string => $this->getWalletLockedBalance($record, UnitTransaction::VND)),
                TextInputColumn::make('wallet_usdt_balance')
                    ->label('Ví USDT')
                    ->type('number')
                    ->inputMode('decimal')
                    ->step('0.000001')
                    ->rules(['nullable', 'numeric'])
                    ->getStateUsing(fn (User $record): string => $this->getWalletBalance($record, UnitTransaction::USDT))
                    ->updateStateUsing(function (User $record, $state): string {
                        return $this->syncWalletBalance($record, UnitTransaction::USDT, $state);
                    }),
                TextColumn::make('wallet_usdt_locked')
                    ->label('Locked USDT')
                    ->state(fn (User $record): string => $this->getWalletLockedBalance($record, UnitTransaction::USDT)),
                TextColumn::make('created_at')
                    ->label('Tạo lúc')
                    ->dateTime('d/m/Y H:i')
                    ->sortable(),
            ])
            ->poll(1000);
    }

    private function getWalletBalance(User $record, UnitTransaction $unit): string
    {
        $wallet = $record->wallets->firstWhere('unit', $unit);

        return $this->normalizeDecimal($wallet?->balance);
    }

    private function getWalletLockedBalance(User $record, UnitTransaction $unit): string
    {
        $wallet = $record->wallets->firstWhere('unit', $unit);

        return $this->normalizeDecimal($wallet?->locked_balance);
    }

    private function syncWalletBalance(User $record, UnitTransaction $unit, mixed $state): string
    {
        $normalized = $this->normalizeDecimal($state);

        app(UserWalletBalanceService::class)->syncAvailableBalances(
            $record,
            [$unit->value => $normalized],
            auth()->user(),
        );

        $record->unsetRelation('wallets');
        $record->load(['wallets' => fn ($query) => $query->select('id', 'user_id', 'unit', 'balance', 'locked_balance')]);

        Notification::make()
            ->title(sprintf(
                'Đã cập nhật ví %s cho user #%d',
                $unit === UnitTransaction::USDT ? 'USDT' : 'VND',
                $record->id,
            ))
            ->icon(Heroicon::OutlinedCheckCircle)
            ->success()
            ->send();

        return $normalized;
    }

    private function normalizeDecimal(mixed $value): string
    {
        $normalized = str_replace([',', ' '], ['', ''], trim((string) $value));

        if ($normalized === '' || ! is_numeric($normalized)) {
            return '0.00000000';
        }

        return number_format((float) $normalized, 8, '.', '');
    }
}
