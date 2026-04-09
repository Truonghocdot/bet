<?php

namespace App\Filament\Resources\Affiliate\AffiliateReferrals\Tables;

use App\Enum\Affiliate\AffiliateReferralStatus;
use App\Support\Filament\EnumPresenter;
use Filament\Actions\BulkActionGroup;
use Filament\Actions\DeleteBulkAction;
use Filament\Actions\EditAction;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Filters\SelectFilter;
use Filament\Tables\Table;

class AffiliateReferralsTable
{
    public static function configure(Table $table): Table
    {
        return $table
            ->columns([
                TextColumn::make('id')->label('ID')->sortable(),
                TextColumn::make('affiliateProfile.ref_code')->label('Hồ sơ')->searchable()->sortable(),
                TextColumn::make('referrerUser.name')->label('Người mời')->searchable()->sortable(),
                TextColumn::make('referredUser.name')->label('Người được mời')->searchable()->sortable(),
                TextColumn::make('first_deposit_amount')->label('Nạp đầu')->money('VND')->toggleable(),
                TextColumn::make('qualified_at')->label('Đạt điều kiện lúc')->dateTime()->toggleable(),
                TextColumn::make('status')->label('Trạng thái')->badge()->formatStateUsing(fn ($state): string => EnumPresenter::label(AffiliateReferralStatus::class, $state))->color(fn ($state): string => EnumPresenter::color(AffiliateReferralStatus::class, $state)),
                TextColumn::make('created_at')->label('Tạo lúc')->dateTime()->sortable(),
            ])
            ->filters([
                SelectFilter::make('status')
                    ->label('Trạng thái')
                    ->options(EnumPresenter::options(AffiliateReferralStatus::class)),
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
