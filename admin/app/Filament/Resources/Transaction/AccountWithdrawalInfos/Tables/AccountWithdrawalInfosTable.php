<?php

namespace App\Filament\Resources\Transaction\AccountWithdrawalInfos\Tables;

use App\Enum\Wallet\UnitTransaction;
use App\Support\Filament\EnumPresenter;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Columns\IconColumn;
use Filament\Tables\Filters\TrashedFilter;
use Filament\Tables\Table;

class AccountWithdrawalInfosTable
{
    public static function configure(Table $table): Table
    {
        return $table
            ->columns([
                TextColumn::make('id')->label('ID')->sortable(),
                TextColumn::make('user.name')->label('Người dùng')->searchable()->sortable(),
                TextColumn::make('unit')
                    ->label('Đơn vị')
                    ->badge()
                    ->formatStateUsing(fn ($state): string => EnumPresenter::label(UnitTransaction::class, $state))
                    ->color(fn ($state): string => EnumPresenter::color(UnitTransaction::class, $state)),
                TextColumn::make('provider_code')->label('Nhà cung cấp')->sortable(),
                TextColumn::make('account_name')->label('Chủ tài khoản')->searchable(),
                TextColumn::make('account_number')->label('Số tài khoản')->searchable(),
                IconColumn::make('is_default')->label('Mặc định')->boolean(),
                TextColumn::make('created_at')->label('Tạo lúc')->dateTime()->sortable(),
            ])
            ->filters([
                TrashedFilter::make(),
            ]);
    }
}
