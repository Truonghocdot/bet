<?php

namespace App\Filament\Resources\Transaction\WithdrawalRequests\Tables;

use App\Enum\Wallet\LedgerDirection;
use App\Enum\Wallet\UnitTransaction;
use App\Enum\Transaction\WithdrawalStatus;
use App\Models\Transaction\WithdrawalRequest;
use App\Models\Wallet\WalletLedgerEntry;
use App\Services\Admin\WithdrawalWorkflowService;
use App\Support\Filament\EnumPresenter;
use Filament\Actions\Action;
use Filament\Actions\BulkActionGroup;
use Filament\Actions\DeleteBulkAction;
use Filament\Actions\ForceDeleteBulkAction;
use Filament\Actions\RestoreBulkAction;
use Filament\Forms\Components\DateTimePicker;
use Filament\Forms\Components\FileUpload;
use Filament\Forms\Components\Placeholder;
use Filament\Forms\Components\TextInput;
use Filament\Forms\Components\Textarea;
use Filament\Notifications\Notification;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Enums\RecordActionsPosition;
use Filament\Tables\Filters\SelectFilter;
use Filament\Tables\Filters\TrashedFilter;
use Filament\Tables\Table;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Gate;
use Illuminate\Support\HtmlString;
use Illuminate\Support\Str;
use Illuminate\Validation\ValidationException;

class WithdrawalRequestsTable
{
    public static function configure(Table $table): Table
    {
        return $table
            ->columns([
                TextColumn::make('id')->label('ID')->sortable(),
                TextColumn::make('user.name')->label('Người dùng')->searchable()->sortable(),
                TextColumn::make('wallet.id')->label('Ví')->sortable(),
                TextColumn::make('accountWithdrawalInfo.account_number')->label('Tài khoản rút')->searchable(),
                TextColumn::make('unit')
                    ->label('Đơn vị')
                    ->badge()
                    ->formatStateUsing(fn ($state): string => EnumPresenter::label(UnitTransaction::class, $state))
                    ->color(fn ($state): string => EnumPresenter::color(UnitTransaction::class, $state)),
                TextColumn::make('amount')->label('Số tiền')->money('VND')->sortable(),
                TextColumn::make('fee')->label('Phí')->money('VND')->sortable(),
                TextColumn::make('net_amount')->label('Thực nhận')->money('VND')->sortable(),
                TextColumn::make('status')
                    ->label('Trạng thái')
                    ->badge()
                    ->formatStateUsing(fn ($state): string => EnumPresenter::label(WithdrawalStatus::class, $state))
                    ->color(fn ($state): string => EnumPresenter::color(WithdrawalStatus::class, $state)),
                TextColumn::make('reviewed_at')->label('Duyệt lúc')->dateTime()->toggleable(),
                TextColumn::make('paid_at')->label('Chi trả lúc')->dateTime()->toggleable(),
                TextColumn::make('transfer_reference')->label('Mã tham chiếu')->toggleable(),
                TextColumn::make('created_at')->label('Tạo lúc')->dateTime()->sortable(),
            ])
            ->filters([
                TrashedFilter::make(),
                SelectFilter::make('status')
                    ->label('Trạng thái')
                    ->options(EnumPresenter::options(WithdrawalStatus::class)),
            ])
            ->recordActions([
                Action::make('edit_finance')
                    ->label('Sửa số tiền')
                    ->icon('heroicon-o-pencil-square')
                    ->color('gray')
                    ->visible(fn (): bool => Gate::allows('finance.withdrawal-requests.update'))
                    ->fillForm(fn (WithdrawalRequest $record): array => [
                        'amount' => (string) $record->amount,
                        'created_at' => $record->created_at,
                        'reviewed_at' => $record->reviewed_at,
                        'paid_at' => $record->paid_at,
                    ])
                    ->form([
                        Placeholder::make('current_locked_balance')
                            ->label('Số dư đang bị đóng băng')
                            ->content(fn (WithdrawalRequest $record): string => number_format((float) ($record->wallet?->locked_balance ?? 0), 0, ',', '.')),
                        TextInput::make('amount')
                            ->label('Số tiền')
                            ->numeric()
                            ->required()
                            ->step('0.000001')
                            ->helperText('Khi sửa amount, hệ thống chỉ đồng bộ số dư đang bị đóng băng của ví. Số dư khả dụng giữ nguyên.'),
                        DateTimePicker::make('created_at')
                            ->label('Thời gian tạo')
                            ->seconds(false)
                            ->required(),
                        DateTimePicker::make('reviewed_at')
                            ->label('Thời gian duyệt')
                            ->seconds(false),
                        DateTimePicker::make('paid_at')
                            ->label('Thời gian chi trả')
                            ->seconds(false),
                    ])
                    ->action(function (WithdrawalRequest $record, array $data): void {
                        self::updateWithdrawalRequestRecord($record, $data);

                        Notification::make()
                            ->title('Đã cập nhật yêu cầu rút')
                            ->success()
                            ->send();
                    }),
                Action::make('approve')
                    ->label('Duyệt')
                    ->icon('heroicon-m-check')
                    ->requiresConfirmation()
                    ->modalHeading('Duyệt yêu cầu rút')
                    ->modalSubmitActionLabel('Xác nhận duyệt')
                    ->form([
                        Placeholder::make('withdrawal_user')
                            ->label('Người dùng')
                            ->content(fn ($record): string => $record->user?->name ?? '—'),
                        Placeholder::make('withdrawal_account')
                            ->label('Tài khoản nhận')
                            ->content(fn ($record): string => $record->accountWithdrawalInfo?->account_number ?? '—'),
                        Placeholder::make('withdrawal_amount')
                            ->label('Số tiền rút')
                            ->content(fn ($record): string => number_format((float) $record->amount, 0, ',', '.').' đ'),
                        Placeholder::make('withdrawal_unit')
                            ->label('Đơn vị')
                            ->content(fn ($record): string => EnumPresenter::label(UnitTransaction::class, $record->unit)),
                        Placeholder::make('withdrawal_created_at')
                            ->label('Thời gian tạo')
                            ->content(fn ($record): string => $record->created_at?->format('d/m/Y H:i:s') ?? '—'),
                        Placeholder::make('withdrawal_payout_info')
                            ->label('Thông tin nhận')
                            ->content(fn ($record): HtmlString => new HtmlString(self::buildPayoutPreview($record)))
                            ->columnSpanFull(),
                    ])
                    ->visible(fn ($record): bool => Gate::allows('finance.withdrawal-requests.approve') && $record->status === WithdrawalStatus::PENDING)
                    ->action(fn ($record) => app(WithdrawalWorkflowService::class)->approve($record, auth()->user())),
                Action::make('reject')
                    ->label('Từ chối')
                    ->icon('heroicon-m-x-mark')
                    ->color('danger')
                    ->requiresConfirmation()
                    ->visible(fn ($record): bool => Gate::allows('finance.withdrawal-requests.reject') && ! in_array($record->status, [WithdrawalStatus::REJECTED, WithdrawalStatus::PAID, WithdrawalStatus::CANCELED], true))
                    ->form([
                        Textarea::make('reason')->label('Lý do từ chối')->required()->rows(3),
                    ])
                    ->action(function ($record, array $data): void {
                        app(WithdrawalWorkflowService::class)->reject($record,
                         /** @var App/Model/User */ 
                                                   auth()->user()
                         , $data['reason'] ?? null);
                    }),
                Action::make('markPaid')
                    ->label('Đã chi trả')
                    ->icon('heroicon-m-banknotes')
                    ->color('success')
                    ->requiresConfirmation()
                    ->visible(fn ($record): bool => Gate::allows('finance.withdrawal-requests.mark-paid') && $record->status === WithdrawalStatus::APPROVED)
                    ->form([
                        TextInput::make('transfer_reference')
                            ->label('Mã giao dịch/chứng từ')
                            ->required(),
                        FileUpload::make('proof_path')
                            ->label('Ảnh chứng từ')
                            ->disk('public')
                            ->directory('withdrawal-proofs')
                            ->preserveFilenames(),
                        Textarea::make('note')
                            ->label('Ghi chú')
                            ->rows(3),
                    ])
                    ->action(function ($record, array $data): void {
                        app(WithdrawalWorkflowService::class)->markPaid(
                            $record,
                            auth()->user(),
                            $data['transfer_reference'],
                            $data['proof_path'] ?? null,
                        );
                    }),
            ], position: RecordActionsPosition::BeforeColumns)
            ->defaultSort('id', 'desc')
            ->poll(2000)
            ->toolbarActions([
                BulkActionGroup::make([
                    DeleteBulkAction::make(),
                    ForceDeleteBulkAction::make(),
                    RestoreBulkAction::make(),
                ]),
            ]);
    }

    private static function updateWithdrawalRequestRecord(WithdrawalRequest $record, array $data): void
    {
        DB::transaction(function () use ($record, $data): void {
            $record->refresh();

            $currentAmount = self::normalizeDecimal($record->amount);
            $newAmount = self::normalizeDecimal($data['amount'] ?? $record->amount);

            if (bccomp($currentAmount, $newAmount, 8) !== 0) {
                self::applyWithdrawalAmountDelta($record, $currentAmount, $newAmount);
            }

            $record->forceFill([
                'amount' => $newAmount,
                'net_amount' => self::calculateNetAmount($newAmount, $record->fee),
                'created_at' => $data['created_at'] ?? $record->created_at,
                'reviewed_at' => $data['reviewed_at'] ?? $record->reviewed_at,
                'paid_at' => $data['paid_at'] ?? $record->paid_at,
            ])->save();
        });
    }

    private static function applyWithdrawalAmountDelta(WithdrawalRequest $record, string $currentAmount, string $newAmount): void
    {
        if (in_array($record->status, [WithdrawalStatus::REJECTED, WithdrawalStatus::CANCELED, WithdrawalStatus::PAID], true)) {
            throw ValidationException::withMessages([
                'amount' => 'Chỉ có thể sửa số tiền khi lệnh rút còn chờ duyệt hoặc đã duyệt nhưng chưa chi trả.',
            ]);
        }

        $wallet = $record->wallet()->lockForUpdate()->first();

        if (! $wallet) {
            throw ValidationException::withMessages([
                'amount' => 'Không tìm thấy ví liên kết để điều chỉnh chênh lệch.',
            ]);
        }

        $balanceBefore = self::normalizeDecimal($wallet->balance);
        $lockedBefore = self::normalizeDecimal($wallet->locked_balance);
        $delta = bcsub($newAmount, $currentAmount, 8);
        $lockedAfter = bcadd($lockedBefore, $delta, 8);

        if (bccomp($lockedAfter, '0.00000000', 8) < 0) {
            throw ValidationException::withMessages([
                'amount' => 'Số dư đang bị đóng băng không đủ để đồng bộ theo amount mới.',
            ]);
        }

        $wallet->forceFill([
            'locked_balance' => $lockedAfter,
        ])->save();

        WalletLedgerEntry::create([
            'wallet_id' => $wallet->id,
            'user_id' => $record->user_id,
            'direction' => LedgerDirection::NEUTRAL,
            'amount' => self::absoluteDecimal($delta),
            'balance_before' => $balanceBefore,
            'balance_after' => $balanceBefore,
            'reference_type' => WithdrawalRequest::class,
            'reference_id' => $record->id,
            'note' => 'Điều chỉnh locked_balance của yêu cầu rút thủ công từ Filament',
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

    private static function buildPayoutPreview(mixed $record): string
    {
        $info = $record->accountWithdrawalInfo;
        $unitValue = (int) ($record->unit?->value ?? $record->unit ?? 0);
        $provider = strtoupper((string) ($info?->provider_code ?? ''));
        $accountNumber = (string) ($info?->account_number ?? '');
        $netAmount = number_format((float) $record->net_amount, 0, ',', '.').' đ';

        if ($unitValue === UnitTransaction::USDT->value) {
            $qrUrl = $accountNumber !== '' 
                ? 'https://api.qrserver.com/v1/create-qr-code/?size=200x200&data='.rawurlencode($accountNumber)
                : null;

            if ($qrUrl) {
                return sprintf(
                    '<div style="display:flex;gap:14px;align-items:flex-start"><img src="%s" alt="qr" style="width:200px;height:200px;border-radius:10px;border:1px solid #e5e7eb;object-fit:cover"><div style="line-height:1.45"><strong>Ví USDT:</strong> %s<br><strong>Mạng:</strong> %s<br><strong>Số tiền rút:</strong> %s %s</div></div>',
                    e($qrUrl),
                    e($accountNumber),
                    e($provider !== '' ? $provider : '—'),
                    e((string) $record->net_amount),
                    e(EnumPresenter::label(UnitTransaction::class, $record->unit))
                );
            }

            return sprintf(
                '<div style="line-height:1.35"><strong>Ví:</strong> %s<br><strong>Mạng:</strong> %s<br><strong>Số tiền:</strong> %s %s</div>',
                e($accountNumber !== '' ? $accountNumber : '—'),
                e($provider !== '' ? $provider : '—'),
                e((string) $record->net_amount),
                e(EnumPresenter::label(UnitTransaction::class, $record->unit))
            );
        }

        $bankCode = Str::lower($provider);
        $qrUrl = null;
        if ($bankCode !== '' && $accountNumber !== '') {
            $qrUrl = sprintf(
                'https://img.vietqr.io/image/%s-%s-compact2.jpg?amount=%s&addInfo=%s',
                rawurlencode($bankCode),
                rawurlencode($accountNumber),
                rawurlencode((string) ((int) round((float) $record->net_amount))),
                rawurlencode('WD-'.$record->id)
            );
        }

        if ($qrUrl === null) {
            return sprintf(
                '<div style="line-height:1.35"><strong>Ngân hàng:</strong> %s<br><strong>STK:</strong> %s<br><strong>Thực nhận:</strong> %s</div>',
                e($provider !== '' ? $provider : '—'),
                e($accountNumber !== '' ? $accountNumber : '—'),
                e($netAmount)
            );
        }

        return sprintf(
            '<div style="display:flex;gap:14px;align-items:flex-start"><img src="%s" alt="qr" style="width:200px;height:200px;border-radius:10px;border:1px solid #e5e7eb;object-fit:cover"><div style="line-height:1.45"><strong>Ngân hàng:</strong> %s<br><strong>STK:</strong> %s<br><strong>Thực nhận:</strong> %s</div></div>',
            e($qrUrl),
            e($provider),
            e($accountNumber),
            e($netAmount)
        );
    }
}
