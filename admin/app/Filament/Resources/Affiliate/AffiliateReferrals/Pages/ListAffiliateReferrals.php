<?php

namespace App\Filament\Resources\Affiliate\AffiliateReferrals\Pages;

use App\Filament\Resources\Affiliate\AffiliateReferrals\AffiliateReferralResource;
use Filament\Actions\CreateAction;
use Filament\Resources\Pages\ListRecords;

class ListAffiliateReferrals extends ListRecords
{
    protected static string $resource = AffiliateReferralResource::class;

    protected function getHeaderActions(): array
    {
        return [
            CreateAction::make(),
        ];
    }
}
