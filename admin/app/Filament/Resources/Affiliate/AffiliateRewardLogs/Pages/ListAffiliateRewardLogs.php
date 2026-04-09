<?php

namespace App\Filament\Resources\Affiliate\AffiliateRewardLogs\Pages;

use App\Filament\Resources\Affiliate\AffiliateRewardLogs\AffiliateRewardLogResource;
use Filament\Actions\CreateAction;
use Filament\Resources\Pages\ListRecords;

class ListAffiliateRewardLogs extends ListRecords
{
    protected static string $resource = AffiliateRewardLogResource::class;

    protected function getHeaderActions(): array
    {
        return [
            CreateAction::make(),
        ];
    }
}
