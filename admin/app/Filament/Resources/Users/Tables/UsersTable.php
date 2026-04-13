<?php

namespace App\Filament\Resources\Users\Tables;

use App\Enum\User\RoleUser;
use App\Enum\User\UserStatus;
use App\Support\Filament\EnumPresenter;
use Filament\Actions\BulkActionGroup;
use Filament\Actions\DeleteBulkAction;
use Filament\Actions\EditAction;
use Filament\Actions\ForceDeleteBulkAction;
use Filament\Actions\RestoreBulkAction;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Filters\TrashedFilter;
use Filament\Tables\Table;

class UsersTable
{
    public static function configure(Table $table): Table
    {
        return $table
            ->columns([
                TextColumn::make('id')
                    ->label('ID')
                    ->sortable(),
                TextColumn::make('name')
                    ->label('Họ và tên')
                    ->searchable()
                    ->sortable(),
                TextColumn::make('email')
                    ->label('Email')
                    ->searchable()
                    ->copyable()
                    ->toggleable(),
                TextColumn::make('phone')
                    ->label('Số điện thoại')
                    ->searchable()
                    ->copyable()
                    ->toggleable(),
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
                TextColumn::make('email_verified_at')
                    ->label('Email xác minh')
                    ->dateTime()
                    ->toggleable(),
                TextColumn::make('phone_verified_at')
                    ->label('SĐT xác minh')
                    ->dateTime()
                    ->toggleable(),
                TextColumn::make('last_login_at')
                    ->label('Đăng nhập cuối')
                    ->dateTime()
                    ->toggleable(),
                TextColumn::make('wallets_count')
                    ->label('Số ví')
                    ->counts('wallets')
                    ->sortable(),
                TextColumn::make('transactions_count')
                    ->label('Số giao dịch')
                    ->counts('transactions')
                    ->sortable(),
                TextColumn::make('withdrawal_requests_count')
                    ->label('Số lệnh rút')
                    ->counts('withdrawalRequests')
                    ->sortable(),
                TextColumn::make('game_tickets_count')
                    ->label('Số vé cược')
                    ->counts('gameTickets')
                    ->sortable(),
                TextColumn::make('created_at')
                    ->label('Tạo lúc')
                    ->dateTime()
                    ->sortable(),
            ])
            ->filters([
                TrashedFilter::make(),
            ])
            ->defaultSort('id', 'desc')
            ->poll(2000)
            ->recordActions([
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
