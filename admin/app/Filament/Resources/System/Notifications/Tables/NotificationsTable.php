<?php

namespace App\Filament\Resources\System\Notifications\Tables;

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
                    ->formatStateUsing(fn ($state): string => EnumPresenter::label(NotificationStatus::class, $state))
                    ->color(fn ($state): string => EnumPresenter::color(NotificationStatus::class, $state)),
                TextColumn::make('audience')
                    ->label('Đối tượng')
                    ->badge()
                    ->formatStateUsing(fn ($state): string => EnumPresenter::label(\App\Enum\Notification\NotificationAudience::class, $state))
                    ->color(fn ($state): string => EnumPresenter::color(\App\Enum\Notification\NotificationAudience::class, $state)),
                TextColumn::make('target_users_count')
                    ->label('User đích')
                    ->counts('targetUsers')
                    ->sortable(),
                TextColumn::make('reads_count')
                    ->label('Đã đọc')
                    ->counts('reads')
                    ->sortable(),
                TextColumn::make('publish_at')->label('Phát hành')->dateTime()->sortable(),
                TextColumn::make('expires_at')->label('Hết hạn')->dateTime()->toggleable(),
                TextColumn::make('createdBy.name')->label('Tạo bởi')->toggleable(),
                TextColumn::make('created_at')->label('Tạo lúc')->dateTime()->sortable(),
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
                    ->visible(fn ($record): bool => $record->status !== NotificationStatus::PUBLISHED)
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
                    ->visible(fn ($record): bool => $record->status !== NotificationStatus::ARCHIVED)
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
}
