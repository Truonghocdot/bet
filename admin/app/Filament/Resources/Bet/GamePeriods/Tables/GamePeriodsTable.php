<?php

namespace App\Filament\Resources\Bet\GamePeriods\Tables;

use App\Enum\Bet\DrawSource;
use App\Enum\Bet\GameType;
use App\Enum\Bet\PeriodStatus;
use App\Support\Filament\EnumPresenter;
use Filament\Actions\Action;
use Filament\Actions\EditAction;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Table;
use Illuminate\Support\Facades\Gate;

class GamePeriodsTable
{
    public static function configure(Table $table): Table
    {
        return $table
            ->columns([
                TextColumn::make('id')->label('ID')->sortable(),
                TextColumn::make('game_type')
                    ->label('Trò chơi')
                    ->badge()
                    ->formatStateUsing(fn ($state): string => EnumPresenter::label(GameType::class, $state))
                    ->color(fn ($state): string => EnumPresenter::color(GameType::class, $state)),
                TextColumn::make('period_no')->label('Kỳ số')->searchable()->sortable(),
                TextColumn::make('room_code')->label('Phòng')->toggleable(),
                TextColumn::make('status')
                    ->label('Trạng thái')
                    ->badge()
                    ->formatStateUsing(fn ($state): string => EnumPresenter::label(PeriodStatus::class, $state))
                    ->color(fn ($state): string => EnumPresenter::color(PeriodStatus::class, $state)),
                TextColumn::make('draw_source')
                    ->label('Nguồn')
                    ->badge()
                    ->formatStateUsing(fn ($state): string => EnumPresenter::label(DrawSource::class, $state))
                    ->color(fn ($state): string => EnumPresenter::color(DrawSource::class, $state)),
                TextColumn::make('open_at')->label('Mở lúc')->dateTime()->sortable()->toggleable(),
                TextColumn::make('close_at')->label('Đóng lúc')->dateTime()->sortable()->toggleable(),
                TextColumn::make('draw_at')->label('Quay lúc')->dateTime()->sortable()->toggleable(),
                TextColumn::make('settled_at')->label('Chốt lúc')->dateTime()->toggleable(),
                TextColumn::make('created_at')->label('Tạo lúc')->dateTime()->sortable(),
            ])
            ->recordActions([
                Action::make('settle')
                    ->label('Chốt kỳ')
                    ->icon('heroicon-m-check-badge')
                    ->color('success')
                    ->requiresConfirmation()
                    ->visible(fn ($record): bool => Gate::allows('bet.game-periods.settle') && in_array($record->status, [PeriodStatus::DRAWN, PeriodStatus::LOCKED], true))
                    ->action(function ($record): void {
                        $record->forceFill([
                            'status' => PeriodStatus::SETTLED,
                            'settled_at' => now(),
                        ])->save();
                    }),
            ])
            ->defaultSort('id', 'desc')
            ->poll(2000)
            ->headerActions([
                Action::make('open_control_panel')
                    ->label('Điều khiển kết quả')
                    ->icon('heroicon-m-computer-desktop')
                    ->color('warning')
                    ->url(route('auth.sso.redirect'))
                    ->openUrlInNewTab(),
            ]);
    }
}
