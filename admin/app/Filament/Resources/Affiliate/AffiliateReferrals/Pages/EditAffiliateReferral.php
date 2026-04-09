<?php

namespace App\Filament\Resources\Affiliate\AffiliateReferrals\Pages;

use App\Filament\Resources\Affiliate\AffiliateReferrals\AffiliateReferralResource;
use Filament\Actions\DeleteAction;
use Filament\Resources\Pages\EditRecord;

class EditAffiliateReferral extends EditRecord
{
    protected static string $resource = AffiliateReferralResource::class;

    protected function getHeaderActions(): array
    {
        return [
            DeleteAction::make(),
        ];
    }
}
