<?php

namespace App\Filament\Resources\Bet\BetItems\Tables;

use App\Enum\Bet\BetItemResult;
use App\Enum\Bet\BetOptionType;
use App\Support\Filament\EnumPresenter;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Table;

class BetItemsTable
{
    public static function configure(Table $table): Table
    {
        return $table
            ->columns([
                TextColumn::make('id')->label('ID')->sortable(),
                TextColumn::make('ticket.ticket_no')->label('Vé')->searchable()->sortable(),
                TextColumn::make('period.period_no')->label('Kỳ')->searchable()->sortable(),
                TextColumn::make('option_type')->label('Kiểu')->badge()->formatStateUsing(fn ($state): string => EnumPresenter::label(BetOptionType::class, $state))->color(fn ($state): string => EnumPresenter::color(BetOptionType::class, $state)),
                TextColumn::make('option_key')->label('Mã')->searchable(),
                TextColumn::make('option_label')->label('Nhãn')->searchable(),
                TextColumn::make('odds_at_placement')->label('Tỷ lệ')->numeric()->sortable(),
                TextColumn::make('stake')->label('Tiền cược')->money('VND')->sortable(),
                TextColumn::make('result')->label('Kết quả')->badge()->formatStateUsing(fn ($state): string => EnumPresenter::label(BetItemResult::class, $state))->color(fn ($state): string => EnumPresenter::color(BetItemResult::class, $state)),
                TextColumn::make('payout_amount')->label('Tiền trả')->money('VND')->toggleable(),
                TextColumn::make('settled_at')->label('Chốt lúc')->dateTime()->toggleable(),
                TextColumn::make('created_at')->label('Tạo lúc')->dateTime()->sortable(),
            ])
            ;
    }
}
