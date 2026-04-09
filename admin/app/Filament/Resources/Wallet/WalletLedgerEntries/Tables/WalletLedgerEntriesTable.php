<?php

namespace App\Filament\Resources\Wallet\WalletLedgerEntries\Tables;

use App\Enum\Wallet\LedgerDirection;
use App\Support\Filament\EnumPresenter;
use Filament\Actions\BulkActionGroup;
use Filament\Actions\DeleteBulkAction;
use Filament\Actions\EditAction;
use Filament\Actions\ForceDeleteBulkAction;
use Filament\Actions\RestoreBulkAction;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Filters\TrashedFilter;
use Filament\Tables\Table;

class WalletLedgerEntriesTable
{
    public static function configure(Table $table): Table
    {
        return $table
            ->columns([
                TextColumn::make('id')
                    ->label('ID')
                    ->sortable(),
                TextColumn::make('wallet.id')
                    ->label('Ví')
                    ->sortable(),
                TextColumn::make('user.name')
                    ->label('Người dùng')
                    ->searchable()
                    ->sortable(),
                TextColumn::make('direction')
                    ->label('Chiều')
                    ->badge()
                    ->formatStateUsing(fn ($state): string => EnumPresenter::label(LedgerDirection::class, $state))
                    ->color(fn ($state): string => EnumPresenter::color(LedgerDirection::class, $state)),
                TextColumn::make('amount')
                    ->label('Số tiền')
                    ->money('VND')
                    ->sortable(),
                TextColumn::make('balance_before')
                    ->label('Trước')
                    ->money('VND')
                    ->toggleable(),
                TextColumn::make('balance_after')
                    ->label('Sau')
                    ->money('VND')
                    ->toggleable(),
                TextColumn::make('reference_type')
                    ->label('Tham chiếu')
                    ->toggleable(),
                TextColumn::make('reference_id')
                    ->label('Mã tham chiếu')
                    ->toggleable(),
                TextColumn::make('created_at')
                    ->label('Tạo lúc')
                    ->dateTime()
                    ->sortable(),
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
