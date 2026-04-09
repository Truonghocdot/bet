<?php

namespace App\Filament\Resources\Transaction\WithdrawalRequests\Tables;

use App\Enum\Transaction\WithdrawalStatus;
use App\Services\Admin\WithdrawalWorkflowService;
use App\Support\Filament\EnumPresenter;
use Filament\Actions\Action;
use Filament\Actions\BulkActionGroup;
use Filament\Actions\DeleteBulkAction;
use Filament\Actions\EditAction;
use Filament\Actions\ForceDeleteBulkAction;
use Filament\Actions\RestoreBulkAction;
use Filament\Forms\Components\FileUpload;
use Filament\Forms\Components\TextInput;
use Filament\Forms\Components\Textarea;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Filters\SelectFilter;
use Filament\Tables\Filters\TrashedFilter;
use Filament\Tables\Table;
use App\Enum\Wallet\UnitTransaction;
use Illuminate\Support\Facades\Gate;

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
                        app(WithdrawalWorkflowService::class)->reject($record, auth()->user(), $data['reason'] ?? null);
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
                EditAction::make(),
            ])
            ->toolbarActions([
                BulkActionGroup::make([
                    DeleteBulkAction::make(),
                    ForceDeleteBulkAction::make(),
                    RestoreBulkAction::make(),
                ]),
            ]);
    }
}
