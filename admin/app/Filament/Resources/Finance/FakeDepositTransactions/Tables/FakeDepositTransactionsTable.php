<?php

namespace App\Filament\Resources\Finance\FakeDepositTransactions\Tables;

use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Table;
use Illuminate\Support\Str;

class FakeDepositTransactionsTable
{
    public static function configure(Table $table): Table
    {
        return $table
            ->columns([
                TextColumn::make('id')
                    ->label('STT')
                    ->sortable(),
                TextColumn::make('masked_code')
                    ->label('ID')
                    ->searchable(),
                TextColumn::make('masked_phone')
                    ->label('Số điện thoại')
                    ->searchable(),
                TextColumn::make('status_label')
                    ->label('Trạng thái')
                    ->badge()
                    ->color(fn (string $state): string => Str::contains(Str::lower($state), 'thành công') ? 'success' : 'warning'),
                TextColumn::make('created_at')
                    ->label('Thời gian')
                    ->dateTime(format: 'd/m/Y H:i:s', timezone: config('app.timezone', 'Asia/Ho_Chi_Minh'))
                    ->sortable(),
            ])
            ->defaultSort('id', 'desc')
            ->poll('2s')
            ->paginated([25, 50, 100]);
    }
}
