<?php

namespace App\Filament\Resources\Transaction\Transactions\Tables;

use App\Enum\Transaction\TransactionStatus;
use App\Enum\Transaction\TypeTransaction;
use App\Enum\User\RoleUser;
use App\Enum\Wallet\LedgerDirection;
use App\Models\Transaction\Transaction;
use App\Models\Wallet\WalletLedgerEntry;
use App\Support\Filament\EnumPresenter;
use Filament\Actions\Action;
use Filament\Forms\Components\DateTimePicker;
use Filament\Forms\Components\TextInput;
use Filament\Notifications\Notification;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Enums\RecordActionsPosition;
use Filament\Tables\Filters\TrashedFilter;
use Filament\Tables\Table;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Gate;
use Illuminate\Validation\ValidationException;

class TransactionsTable
{
    public static function configure(Table $table): Table
    {
        return $table
            ->columns([
                TextColumn::make('id')->label('ID')->sortable(),
                TextColumn::make('user_id')->label('ID người dùng')->sortable()->searchable(),
                TextColumn::make('user.name')->label('Người dùng')->searchable()->sortable(),
                TextColumn::make('user.phone')->label('SDT')->searchable(),
                TextColumn::make('agency_id')
                    ->label('ID agency')
                    ->getStateUsing(function ($record): string {
                        if ($record->type !== TypeTransaction::DEPOSIT) {
                            return '—';
                        }

                        $agency = $record->user?->referredByReferral?->referrerUser;
                        if (! $agency || $agency->role !== RoleUser::AGENCY) {
                            return '—';
                        }

                        return (string) $agency->id;
                    }),
                TextColumn::make('agency_name')
                    ->label('Thuộc agency')
                    ->getStateUsing(function ($record): string {
                        if ($record->type !== TypeTransaction::DEPOSIT) {
                            return '—';
                        }

                        $agency = $record->user?->referredByReferral?->referrerUser;
                        if (! $agency || $agency->role !== RoleUser::AGENCY) {
                            return '—';
                        }

                        return $agency->name ?: ('Agency #'.$agency->id);
                    })
                    ->searchable(query: function ($query, string $search): void {
                        $query->whereHas('user.referredByReferral.referrerUser', function ($agencyQuery) use ($search): void {
                            $agencyQuery
                                ->where('role', RoleUser::AGENCY->value)
                                ->where('name', 'like', '%'.$search.'%');
                        });
                    }),
                TextColumn::make('type')
                    ->label('Loại')
                    ->badge()
                    ->formatStateUsing(fn ($state): string => EnumPresenter::label(TypeTransaction::class, $state))
                    ->color(fn ($state): string => EnumPresenter::color(TypeTransaction::class, $state)),
                TextColumn::make('unit')
                    ->label('Đơn vị')
                    ->badge()
                    ->formatStateUsing(fn ($state): string => EnumPresenter::label(\App\Enum\Wallet\UnitTransaction::class, $state))
                    ->color(fn ($state): string => EnumPresenter::color(\App\Enum\Wallet\UnitTransaction::class, $state)),
                TextColumn::make('amount')->label('Số tiền')->money('VND')->sortable(),
                TextColumn::make('client_ref')
                    ->label('Mã giao dịch CK')
                    ->searchable()
                    ->toggleable()
                    ->formatStateUsing(function ($state): string {
                        $value = trim((string) $state);

                        if ($value === '') {
                            return '—';
                        }

                        return str_starts_with($value, 'DEP-') ? substr($value, 4) : $value;
                    })
                    ->copyable()
                    ->fontFamily('mono'),
                TextColumn::make('status')
                    ->label('Trạng thái')
                    ->badge()
                    ->formatStateUsing(fn ($state): string => EnumPresenter::label(TransactionStatus::class, $state))
                    ->color(fn ($state): string => EnumPresenter::color(TransactionStatus::class, $state)),
                TextColumn::make('provider')->label('Nhà cung cấp')->toggleable(),
                TextColumn::make('approved_at')->label('Duyệt lúc')->dateTime()->toggleable(),
                TextColumn::make('created_at')->label('Tạo lúc')->dateTime()->sortable(),
            ])
            ->defaultSort('id', 'desc')
            ->poll(2000)
            ->filters([
                TrashedFilter::make(),
            ])
            ->recordActions([
                Action::make('edit_finance')
                    ->label('Sửa số tiền')
                    ->icon('heroicon-o-pencil-square')
                    ->color('gray')
                    ->visible(fn (): bool => Gate::allows('finance.transactions.update'))
                    ->fillForm(fn (Transaction $record): array => [
                        'amount' => (string) $record->amount,
                        'created_at' => $record->created_at,
                        'approved_at' => $record->approved_at,
                    ])
                    ->form([
                        TextInput::make('amount')
                            ->label('Số tiền')
                            ->numeric()
                            ->required()
                            ->step('0.000001'),
                        DateTimePicker::make('created_at')
                            ->label('Thời gian tạo')
                            ->seconds(false)
                            ->required(),
                        DateTimePicker::make('approved_at')
                            ->label('Thời gian duyệt')
                            ->seconds(false),
                    ])
                    ->action(function (Transaction $record, array $data): void {
                        self::updateTransactionRecord($record, $data);

                        Notification::make()
                            ->title('Đã cập nhật giao dịch')
                            ->success()
                            ->send();
                    }),
                Action::make('approve_deposit')
                    ->label('Duyệt nạp')
                    ->icon('heroicon-o-check-circle')
                    ->color('success')
                    ->requiresConfirmation()
                    ->modalHeading('Duyệt nạp tiền')
                    ->modalDescription('Bạn có chắc chắn đã nhận được tiền và muốn cộng vào ví người dùng?')
                    ->modalSubmitActionLabel('Xác nhận duyệt')
                    ->visible(fn ($record) => $record->type === TypeTransaction::DEPOSIT && $record->status === TransactionStatus::PENDING)
                    ->action(function ($record, \App\Services\Admin\DepositWorkflowService $service) {
                        if ($service->approve($record)) {
                            \Filament\Notifications\Notification::make()
                                ->title('Đã duyệt nạp tiền thành công')
                                ->success()
                                ->send();
                        } else {
                            \Filament\Notifications\Notification::make()
                                ->title('Có lỗi xảy ra khi duyệt')
                                ->danger()
                                ->send();
                        }
                    }),

                Action::make('reject_deposit')
                    ->label('Từ chối')
                    ->icon('heroicon-o-x-circle')
                    ->color('danger')
                    ->requiresConfirmation()
                    ->modalHeading('Từ chối nạp tiền')
                    ->modalDescription('Bạn có chắc chắn muốn từ chối yêu cầu nạp tiền này?')
                    ->modalSubmitActionLabel('Xác nhận từ chối')
                    ->visible(fn ($record) => $record->type === TypeTransaction::DEPOSIT && $record->status === TransactionStatus::PENDING)
                    ->schema([
                        TextInput::make('reason_failed')
                            ->label('Lý do từ chối')
                            ->placeholder('Nhập lý do (tùy chọn)'),
                    ])
                    ->action(function ($record, array $data, \App\Services\Admin\DepositWorkflowService $service) {
                        if ($service->reject($record, $data['reason_failed'] ?? null)) {
                            \Filament\Notifications\Notification::make()
                                ->title('Đã từ chối yêu cầu nạp tiền')
                                ->warning()
                                ->send();
                        }
                    }),
            ], position: RecordActionsPosition::BeforeColumns);
    }

    private static function updateTransactionRecord(Transaction $record, array $data): void
    {
        DB::transaction(function () use ($record, $data): void {
            $record->refresh();

            $currentAmount = self::normalizeDecimal($record->amount);
            $newAmount = self::normalizeDecimal($data['amount'] ?? $record->amount);

            if (bccomp($currentAmount, $newAmount, 8) !== 0) {
                self::applyTransactionAmountDelta($record, $currentAmount, $newAmount);
            }

            $record->forceFill([
                'amount' => $newAmount,
                'net_amount' => self::calculateNetAmount($newAmount, $record->fee),
                'created_at' => $data['created_at'] ?? $record->created_at,
                'approved_at' => $data['approved_at'] ?? $record->approved_at,
            ])->save();
        });
    }

    private static function applyTransactionAmountDelta(Transaction $record, string $currentAmount, string $newAmount): void
    {
        if (
            $record->status === TransactionStatus::COMPLETED
            && $record->type !== TypeTransaction::DEPOSIT
        ) {
            throw ValidationException::withMessages([
                'amount' => 'Không hỗ trợ sửa số tiền cho giao dịch đã hoàn tất này vì sẽ lệch sổ ví.',
            ]);
        }

        if (
            $record->status !== TransactionStatus::COMPLETED
            || $record->type !== TypeTransaction::DEPOSIT
        ) {
            return;
        }

        $wallet = $record->wallet()->lockForUpdate()->first();

        if (! $wallet) {
            throw ValidationException::withMessages([
                'amount' => 'Không tìm thấy ví liên kết để điều chỉnh chênh lệch.',
            ]);
        }

        $balanceBefore = self::normalizeDecimal($wallet->balance);
        $delta = bcsub($newAmount, $currentAmount, 8);
        $balanceAfter = bcadd($balanceBefore, $delta, 8);

        $wallet->forceFill([
            'balance' => $balanceAfter,
        ])->save();

        WalletLedgerEntry::create([
            'wallet_id' => $wallet->id,
            'user_id' => $record->user_id,
            'direction' => bccomp($delta, '0.00000000', 8) >= 0
                ? LedgerDirection::CREDIT
                : LedgerDirection::DEBIT,
            'amount' => self::absoluteDecimal($delta),
            'balance_before' => $balanceBefore,
            'balance_after' => $balanceAfter,
            'reference_type' => Transaction::class,
            'reference_id' => $record->id,
            'note' => 'Điều chỉnh số tiền giao dịch nạp thủ công từ Filament',
            'created_at' => now(),
        ]);
    }

    private static function calculateNetAmount(mixed $amount, mixed $fee): string
    {
        $normalizedAmount = self::normalizeDecimal($amount);
        $normalizedFee = self::normalizeDecimal($fee);

        if (bccomp($normalizedAmount, $normalizedFee, 8) < 0) {
            return '0.00000000';
        }

        return bcsub($normalizedAmount, $normalizedFee, 8);
    }

    private static function normalizeDecimal(mixed $value): string
    {
        $normalized = str_replace([',', ' '], ['', ''], trim((string) $value));

        if ($normalized === '' || ! is_numeric($normalized)) {
            return '0.00000000';
        }

        return number_format((float) $normalized, 8, '.', '');
    }

    private static function absoluteDecimal(string $value): string
    {
        return ltrim($value, '-') ?: '0.00000000';
    }
}
