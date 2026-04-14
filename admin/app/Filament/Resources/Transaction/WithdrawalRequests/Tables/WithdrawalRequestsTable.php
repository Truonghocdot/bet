<?php

namespace App\Filament\Resources\Transaction\WithdrawalRequests\Tables;

use App\Enum\Transaction\WithdrawalStatus;
use App\Services\Admin\WithdrawalWorkflowService;
use App\Support\Filament\EnumPresenter;
use Filament\Actions\Action;
use Filament\Actions\BulkActionGroup;
use Filament\Actions\DeleteBulkAction;
use Filament\Actions\ForceDeleteBulkAction;
use Filament\Actions\RestoreBulkAction;
use Filament\Forms\Components\FileUpload;
use Filament\Forms\Components\Placeholder;
use Filament\Forms\Components\TextInput;
use Filament\Forms\Components\Textarea;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Filters\SelectFilter;
use Filament\Tables\Filters\TrashedFilter;
use Filament\Tables\Table;
use App\Enum\Wallet\UnitTransaction;
use Filament\Tables\Enums\RecordActionsPosition;
use Illuminate\Support\Facades\Gate;
use Illuminate\Support\HtmlString;
use Illuminate\Support\Str;

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

    private static function buildPayoutPreview(mixed $record): string
    {
        $info = $record->accountWithdrawalInfo;
        $unitValue = (int) ($record->unit?->value ?? $record->unit ?? 0);
        $provider = strtoupper((string) ($info?->provider_code ?? ''));
        $accountNumber = (string) ($info?->account_number ?? '');
        $netAmount = number_format((float) $record->net_amount, 0, ',', '.').' đ';

        if ($unitValue === UnitTransaction::USDT->value) {
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
