<?php

namespace App\Filament\Resources\Affiliate\AffiliateProfiles\Pages;

use App\Filament\Resources\Affiliate\AffiliateProfiles\AffiliateProfileResource;
use Filament\Actions\CreateAction;
use Filament\Resources\Pages\ListRecords;

class ListAffiliateProfiles extends ListRecords
{
    protected static string $resource = AffiliateProfileResource::class;

    protected function getHeaderActions(): array
    {
        return [
            CreateAction::make(),
        ];
    }
}
