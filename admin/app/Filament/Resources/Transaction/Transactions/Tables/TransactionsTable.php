<?php

namespace App\Filament\Resources\Transaction\Transactions\Tables;

use App\Enum\Transaction\TransactionStatus;
use App\Enum\Transaction\TypeTransaction;
use App\Support\Filament\EnumPresenter;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Filters\TrashedFilter;
use Filament\Tables\Table;

class TransactionsTable
{
    public static function configure(Table $table): Table
    {
        return $table
            ->columns([
                TextColumn::make('id')->label('ID')->sortable(),
                TextColumn::make('user.name')->label('Người dùng')->searchable()->sortable(),
                TextColumn::make('wallet.id')->label('Ví')->sortable(),
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
                TextColumn::make('fee')->label('Phí')->money('VND')->sortable(),
                TextColumn::make('net_amount')->label('Thực nhận')->money('VND')->sortable(),
                TextColumn::make('status')
                    ->label('Trạng thái')
                    ->badge()
                    ->formatStateUsing(fn ($state): string => EnumPresenter::label(TransactionStatus::class, $state))
                    ->color(fn ($state): string => EnumPresenter::color(TransactionStatus::class, $state)),
                TextColumn::make('provider')->label('Nhà cung cấp')->toggleable(),
                TextColumn::make('provider_txn_id')->label('Mã nhà cung cấp')->toggleable(),
                TextColumn::make('approved_at')->label('Duyệt lúc')->dateTime()->toggleable(),
                TextColumn::make('created_at')->label('Tạo lúc')->dateTime()->sortable(),
            ])
            ->filters([
                TrashedFilter::make(),
            ]);
    }
}
