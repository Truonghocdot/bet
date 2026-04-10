<?php

namespace App\Filament\Resources\Affiliate\AffiliateRewardLogs\Tables;

use App\Enum\Affiliate\AffiliateRewardStatus;
use App\Enum\Wallet\UnitTransaction;
use App\Support\Filament\EnumPresenter;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Filters\SelectFilter;
use Filament\Tables\Table;

class AffiliateRewardLogsTable
{
    public static function configure(Table $table): Table
    {
        return $table
            ->columns([
                TextColumn::make('id')->label('ID')->sortable(),
                TextColumn::make('affiliateProfile.ref_code')->label('Hồ sơ')->searchable()->sortable(),
                TextColumn::make('referrerUser.name')->label('Người mời')->searchable()->sortable(),
                TextColumn::make('setting.name')->label('Cấu hình')->searchable()->sortable(),
                TextColumn::make('reward_amount')->label('Tiền thưởng')->money('VND')->sortable(),
                TextColumn::make('unit')->label('Đơn vị')->badge()->formatStateUsing(fn ($state): string => EnumPresenter::label(UnitTransaction::class, $state))->color(fn ($state): string => EnumPresenter::color(UnitTransaction::class, $state)),
                TextColumn::make('status')->label('Trạng thái')->badge()->formatStateUsing(fn ($state): string => EnumPresenter::label(AffiliateRewardStatus::class, $state))->color(fn ($state): string => EnumPresenter::color(AffiliateRewardStatus::class, $state)),
                TextColumn::make('granted_at')->label('Trả thưởng lúc')->dateTime()->toggleable(),
                TextColumn::make('created_at')->label('Tạo lúc')->dateTime()->sortable(),
            ])
            ->filters([
                SelectFilter::make('status')
                    ->label('Trạng thái')
                    ->options(EnumPresenter::options(AffiliateRewardStatus::class)),
            ]);
    }
}
