<?php

namespace App\Filament\Resources\Transaction\Transactions\Tables;

use App\Enum\Transaction\TransactionStatus;
use App\Enum\Transaction\TypeTransaction;
use App\Enum\User\RoleUser;
use App\Support\Filament\EnumPresenter;
use Filament\Actions\Action;
use Filament\Forms\Components\TextInput;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Enums\RecordActionsPosition;
use Filament\Tables\Filters\TrashedFilter;
use Filament\Tables\Table;

class TransactionsTable
{
    public static function configure(Table $table): Table
    {
        return $table
            ->columns([
                TextColumn::make('id')->label('ID')->sortable(),
                TextColumn::make('user_id')->label('ID người dùng')->sortable()->searchable(),
                TextColumn::make('user.name')->label('Người dùng')->searchable()->sortable(),
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
}
