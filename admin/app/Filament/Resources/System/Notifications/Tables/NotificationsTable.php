<?php

namespace App\Filament\Resources\System\Notifications\Tables;

use App\Enum\Notification\NotificationAudience;
use App\Enum\Notification\NotificationStatus;
use App\Support\Filament\EnumPresenter;
use Filament\Actions\Action;
use Filament\Actions\BulkActionGroup;
use Filament\Actions\CreateAction;
use Filament\Actions\DeleteBulkAction;
use Filament\Actions\EditAction;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Filters\SelectFilter;
use Filament\Tables\Table;

class NotificationsTable
{
    public static function configure(Table $table): Table
    {
        return $table
            ->columns([
                TextColumn::make('id')->label('ID')->sortable(),
                TextColumn::make('title')
                    ->label('Tiêu đề')
                    ->searchable()
                    ->sortable()
                    ->limit(80),
                TextColumn::make('status')
                    ->label('Trạng thái')
                    ->badge()
                    ->formatStateUsing(fn($state): string => EnumPresenter::label(NotificationStatus::class, $state))
                    ->color(fn($state): string => EnumPresenter::color(NotificationStatus::class, $state)),
                TextColumn::make('audience')
                    ->label('Đối tượng')
                    ->badge()
                    ->formatStateUsing(fn($state): string => EnumPresenter::label(\App\Enum\Notification\NotificationAudience::class, $state))
                    ->color(fn($state): string => EnumPresenter::color(\App\Enum\Notification\NotificationAudience::class, $state)),
                TextColumn::make('target_users_count')
                    ->label('User đích')
                    ->counts('targetUsers')
                    ->badge()
                    ->color(fn ($record): string => (int) ($record->audience?->value ?? $record->audience) === NotificationAudience::ALL->value ? 'gray' : 'info')
                    ->formatStateUsing(function ($state, $record): string {
                        if ((int) ($record->audience?->value ?? $record->audience) === NotificationAudience::ALL->value) {
                            return 'Toàn bộ';
                        }

                        return self::formatPeopleCount($state);
                    })
                    ->sortable(),
                TextColumn::make('reads_count')
                    ->label('Đã đọc')
                    ->counts('reads')
                    ->badge()
                    ->color('success')
                    ->formatStateUsing(fn ($state): string => self::formatPeopleCount($state))
                    ->sortable(),
                TextColumn::make('pending_response_targets_count')
                    ->label('Chờ phản hồi')
                    ->counts('pendingResponseTargets')
                    ->badge()
                    ->color('warning')
                    ->formatStateUsing(fn ($state, $record): string => $record->supportsResponseTracking() ? self::formatPeopleCount($state) : '—')
                    ->sortable(),
                TextColumn::make('confirmed_response_targets_count')
                    ->label('Đã xác nhận')
                    ->counts('confirmedResponseTargets')
                    ->badge()
                    ->color('success')
                    ->formatStateUsing(fn ($state, $record): string => $record->supportsResponseTracking() ? self::formatPeopleCount($state) : '—')
                    ->sortable(),
                TextColumn::make('canceled_response_targets_count')
                    ->label('Đã hủy')
                    ->counts('canceledResponseTargets')
                    ->badge()
                    ->color('danger')
                    ->formatStateUsing(fn ($state, $record): string => $record->supportsResponseTracking() ? self::formatPeopleCount($state) : '—')
                    ->sortable(),
                TextColumn::make('publish_at')
                    ->label('Phát hành')
                    ->dateTime(format: 'd/m/Y H:i', timezone: config('app.timezone', 'Asia/Ho_Chi_Minh'))
                    ->sortable(),
                TextColumn::make('expires_at')
                    ->label('Hết hạn')
                    ->dateTime(format: 'd/m/Y H:i', timezone: config('app.timezone', 'Asia/Ho_Chi_Minh'))
                    ->toggleable(),
                TextColumn::make('createdBy.name')->label('Tạo bởi')->toggleable(),
                TextColumn::make('created_at')
                    ->label('Tạo lúc')
                    ->  dateTime(format: 'd/m/Y H:i', timezone: config('app.timezone', 'Asia/Ho_Chi_Minh'))
                    ->sortable(),
            ])
            ->filters([
                SelectFilter::make('status')
                    ->label('Trạng thái')
                    ->options(EnumPresenter::options(NotificationStatus::class)),
                SelectFilter::make('audience')
                    ->label('Đối tượng')
                    ->options(EnumPresenter::options(\App\Enum\Notification\NotificationAudience::class)),
            ])
            ->recordActions([
                Action::make('publish')
                    ->label('Phát hành')
                    ->icon('heroicon-m-paper-airplane')
                    ->color('success')
                    ->requiresConfirmation()
                    ->visible(fn($record): bool => $record->status !== NotificationStatus::PUBLISHED)
                    ->action(function ($record): void {
                        $record->status = NotificationStatus::PUBLISHED;
                        if (blank($record->publish_at)) {
                            $record->publish_at = now();
                        }
                        $record->save();
                    }),
                Action::make('archive')
                    ->label('Lưu trữ')
                    ->icon('heroicon-m-archive-box')
                    ->color('warning')
                    ->requiresConfirmation()
                    ->visible(fn($record): bool => $record->status !== NotificationStatus::ARCHIVED)
                    ->action(function ($record): void {
                        $record->status = NotificationStatus::ARCHIVED;
                        $record->save();
                    }),
                EditAction::make(),
            ])
            ->headerActions([
                CreateAction::make()
                    ->label('Tạo thông báo')
                    ->icon('heroicon-m-plus'),
            ])
            ->defaultSort('id', 'desc')
            ->poll(2000)
            ->toolbarActions([
                BulkActionGroup::make([
                    DeleteBulkAction::make(),
                ]),
            ]);
    }

    private static function formatPeopleCount(mixed $state): string
    {
        return number_format((int) $state).' người';
    }
}
