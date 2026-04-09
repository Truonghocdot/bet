<?php

namespace App\Filament\Resources\Bet\BetSettlements\Tables;

use App\Enum\Bet\BetStatus;
use App\Enum\Bet\SettlementType;
use App\Support\Filament\EnumPresenter;
use Filament\Actions\BulkActionGroup;
use Filament\Actions\DeleteBulkAction;
use Filament\Actions\EditAction;
use Filament\Actions\ForceDeleteBulkAction;
use Filament\Actions\RestoreBulkAction;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Filters\TrashedFilter;
use Filament\Tables\Table;

class BetSettlementsTable
{
    public static function configure(Table $table): Table
    {
        return $table
            ->columns([
                TextColumn::make('id')->label('ID')->sortable(),
                TextColumn::make('ticket.ticket_no')->label('Vé')->searchable()->sortable(),
                TextColumn::make('period.period_no')->label('Kỳ')->searchable()->sortable(),
                TextColumn::make('settlement_type')->label('Kiểu chốt')->badge()->formatStateUsing(fn ($state): string => EnumPresenter::label(SettlementType::class, $state))->color(fn ($state): string => EnumPresenter::color(SettlementType::class, $state)),
                TextColumn::make('before_status')->label('Trước')->badge()->formatStateUsing(fn ($state): string => EnumPresenter::label(BetStatus::class, $state))->color(fn ($state): string => EnumPresenter::color(BetStatus::class, $state)),
                TextColumn::make('after_status')->label('Sau')->badge()->formatStateUsing(fn ($state): string => EnumPresenter::label(BetStatus::class, $state))->color(fn ($state): string => EnumPresenter::color(BetStatus::class, $state)),
                TextColumn::make('payout_amount')->label('Tiền trả')->money('VND')->sortable(),
                TextColumn::make('profit_loss')->label('Lãi lỗ')->money('VND')->sortable(),
                TextColumn::make('settledBy.name')->label('Chốt bởi')->searchable()->toggleable(),
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
