<?php

namespace App\Filament\Resources\Affiliate\AffiliateRewardSettings\Pages;

use App\Filament\Resources\Affiliate\AffiliateRewardSettings\AffiliateRewardSettingResource;
use Filament\Actions\CreateAction;
use Filament\Resources\Pages\ListRecords;

class ListAffiliateRewardSettings extends ListRecords
{
    protected static string $resource = AffiliateRewardSettingResource::class;

    protected function getHeaderActions(): array
    {
        return [
            CreateAction::make(),
        ];
    }
}
