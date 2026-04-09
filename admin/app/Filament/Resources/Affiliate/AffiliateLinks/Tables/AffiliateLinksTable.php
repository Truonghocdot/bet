<?php

namespace App\Filament\Resources\Affiliate\AffiliateLinks\Tables;

use App\Enum\Affiliate\AffiliateLinkStatus;
use App\Support\Filament\EnumPresenter;
use Filament\Actions\BulkActionGroup;
use Filament\Actions\DeleteBulkAction;
use Filament\Actions\EditAction;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Filters\SelectFilter;
use Filament\Tables\Table;

class AffiliateLinksTable
{
    public static function configure(Table $table): Table
    {
        return $table
            ->columns([
                TextColumn::make('id')->label('ID')->sortable(),
                TextColumn::make('affiliateProfile.ref_code')->label('Hồ sơ')->searchable()->sortable(),
                TextColumn::make('campaign_name')->label('Chiến dịch')->searchable()->sortable(),
                TextColumn::make('tracking_code')->label('Mã tracking')->searchable()->sortable(),
                TextColumn::make('status')->label('Trạng thái')->badge()->formatStateUsing(fn ($state): string => EnumPresenter::label(AffiliateLinkStatus::class, $state))->color(fn ($state): string => EnumPresenter::color(AffiliateLinkStatus::class, $state)),
                TextColumn::make('created_at')->label('Tạo lúc')->dateTime()->sortable(),
            ])
            ->filters([
                SelectFilter::make('status')
                    ->label('Trạng thái')
                    ->options(EnumPresenter::options(AffiliateLinkStatus::class)),
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
