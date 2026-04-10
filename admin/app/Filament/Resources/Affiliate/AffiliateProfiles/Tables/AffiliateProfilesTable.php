<?php

namespace App\Filament\Resources\Affiliate\AffiliateProfiles\Tables;

use App\Enum\Affiliate\AffiliateProfileStatus;
use App\Support\Filament\EnumPresenter;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Filters\SelectFilter;
use Filament\Tables\Table;

class AffiliateProfilesTable
{
    public static function configure(Table $table): Table
    {
        return $table
            ->columns([
                TextColumn::make('id')->label('ID')->sortable(),
                TextColumn::make('user.name')->label('Người dùng')->searchable()->sortable(),
                TextColumn::make('ref_code')->label('Mã giới thiệu')->searchable()->sortable(),
                TextColumn::make('status')->label('Trạng thái')->badge()->formatStateUsing(fn ($state): string => EnumPresenter::label(AffiliateProfileStatus::class, $state))->color(fn ($state): string => EnumPresenter::color(AffiliateProfileStatus::class, $state)),
                TextColumn::make('created_at')->label('Tạo lúc')->dateTime()->sortable(),
            ])
            ->filters([
                SelectFilter::make('status')
                    ->label('Trạng thái')
                    ->options(EnumPresenter::options(AffiliateProfileStatus::class)),
            ]);
    }
}
