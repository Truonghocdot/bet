<?php

namespace App\Filament\Resources\Payment\PaymentReceivingAccounts\Tables;

use App\Enum\Payment\PaymentReceivingAccountStatus;
use App\Enum\Payment\PaymentReceivingAccountType;
use App\Enum\Wallet\UnitTransaction;
use App\Support\Filament\EnumPresenter;
use Filament\Actions\BulkActionGroup;
use Filament\Actions\DeleteBulkAction;
use Filament\Actions\EditAction;
use Filament\Actions\ForceDeleteBulkAction;
use Filament\Actions\RestoreBulkAction;
use Filament\Tables\Columns\IconColumn;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Filters\TrashedFilter;
use Filament\Tables\Table;

class PaymentReceivingAccountsTable
{
    public static function configure(Table $table): Table
    {
        return $table
            ->columns([
                TextColumn::make('id')->label('ID')->sortable(),
                TextColumn::make('code')->label('Mã')->searchable()->sortable(),
                TextColumn::make('name')->label('Tên')->searchable()->sortable(),
                TextColumn::make('type')
                    ->label('Loại')
                    ->badge()
                    ->formatStateUsing(fn ($state): string => EnumPresenter::label(PaymentReceivingAccountType::class, $state))
                    ->color(fn ($state): string => EnumPresenter::color(PaymentReceivingAccountType::class, $state)),
                TextColumn::make('unit')
                    ->label('Đơn vị')
                    ->badge()
                    ->formatStateUsing(fn ($state): string => EnumPresenter::label(UnitTransaction::class, $state))
                    ->color(fn ($state): string => EnumPresenter::color(UnitTransaction::class, $state)),
                TextColumn::make('status')
                    ->label('Trạng thái')
                    ->badge()
                    ->formatStateUsing(fn ($state): string => EnumPresenter::label(PaymentReceivingAccountStatus::class, $state))
                    ->color(fn ($state): string => EnumPresenter::color(PaymentReceivingAccountStatus::class, $state)),
                IconColumn::make('is_default')->label('Mặc định')->boolean(),
                TextColumn::make('sort_order')->label('Thứ tự')->sortable(),
                TextColumn::make('created_at')->label('Tạo lúc')->dateTime()->sortable(),
            ])
            ->filters([
                TrashedFilter::make(),
            ])
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
