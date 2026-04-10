<?php

namespace App\Filament\Resources\Bet\BetTickets\Tables;

use App\Enum\Bet\BetStatus;
use App\Enum\Bet\BetTicketType;
use App\Enum\Bet\GameType;
use App\Enum\Wallet\UnitTransaction;
use App\Support\Filament\EnumPresenter;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Table;

class BetTicketsTable
{
    public static function configure(Table $table): Table
    {
        return $table
            ->columns([
                TextColumn::make('id')->label('ID')->sortable(),
                TextColumn::make('ticket_no')->label('Mã vé')->searchable()->sortable(),
                TextColumn::make('user.name')->label('Người dùng')->searchable()->sortable(),
                TextColumn::make('game_type')->label('Trò chơi')->badge()->formatStateUsing(fn ($state): string => EnumPresenter::label(GameType::class, $state))->color(fn ($state): string => EnumPresenter::color(GameType::class, $state)),
                TextColumn::make('period.period_no')->label('Kỳ')->searchable()->sortable(),
                TextColumn::make('bet_type')->label('Loại vé')->badge()->formatStateUsing(fn ($state): string => EnumPresenter::label(BetTicketType::class, $state))->color(fn ($state): string => EnumPresenter::color(BetTicketType::class, $state)),
                TextColumn::make('unit')->label('Đơn vị')->badge()->formatStateUsing(fn ($state): string => EnumPresenter::label(UnitTransaction::class, $state))->color(fn ($state): string => EnumPresenter::color(UnitTransaction::class, $state)),
                TextColumn::make('stake')->label('Tiền cược')->money('VND')->sortable(),
                TextColumn::make('potential_payout')->label('Thắng dự kiến')->money('VND')->sortable(),
                TextColumn::make('actual_payout')->label('Thắng thực')->money('VND')->toggleable(),
                TextColumn::make('status')->label('Trạng thái')->badge()->formatStateUsing(fn ($state): string => EnumPresenter::label(BetStatus::class, $state))->color(fn ($state): string => EnumPresenter::color(BetStatus::class, $state)),
                TextColumn::make('settled_at')->label('Chốt lúc')->dateTime()->toggleable(),
                TextColumn::make('created_at')->label('Đặt lúc')->dateTime()->sortable(),
            ])
            ;
    }
}
