<?php

namespace App\Filament\Resources\Affiliate\AffiliateRewardSettings\Tables;

use App\Enum\Wallet\UnitTransaction;
use App\Support\Filament\EnumPresenter;
use Filament\Actions\BulkActionGroup;
use Filament\Actions\DeleteBulkAction;
use Filament\Actions\EditAction;
use Filament\Tables\Columns\IconColumn;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Filters\TernaryFilter;
use Filament\Tables\Table;

class AffiliateRewardSettingsTable
{
    public static function configure(Table $table): Table
    {
        return $table
            ->columns([
                TextColumn::make('id')->label('ID')->sortable(),
                TextColumn::make('name')->label('Tên')->searchable()->sortable(),
                TextColumn::make('required_qualified_referrals')->label('Số người')->sortable(),
                TextColumn::make('reward_amount')->label('Tiền thưởng')->money('VND')->sortable(),
                TextColumn::make('unit')->label('Đơn vị')->badge()->formatStateUsing(fn ($state): string => EnumPresenter::label(UnitTransaction::class, $state))->color(fn ($state): string => EnumPresenter::color(UnitTransaction::class, $state)),
                IconColumn::make('is_active')->label('Kích hoạt')->boolean(),
                TextColumn::make('effective_from')->label('Hiệu lực từ')->dateTime()->toggleable(),
                TextColumn::make('effective_to')->label('Hiệu lực đến')->dateTime()->toggleable(),
                TextColumn::make('created_at')->label('Tạo lúc')->dateTime()->sortable(),
            ])
            ->filters([
                TernaryFilter::make('is_active')->label('Kích hoạt'),
            ])
            ->recordActions([
                EditAction::make(),
            ])
            ->toolbarActions([
                BulkActionGroup::make([
                    DeleteBulkAction::make(),
                ]),
            ]);
    }
}
