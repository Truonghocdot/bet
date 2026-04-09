<?php

namespace App\Filament\Resources\Affiliate\AffiliateProfiles\Pages;

use App\Filament\Resources\Affiliate\AffiliateProfiles\AffiliateProfileResource;
use Filament\Actions\DeleteAction;
use Filament\Resources\Pages\EditRecord;

class EditAffiliateProfile extends EditRecord
{
    protected static string $resource = AffiliateProfileResource::class;

    protected function getHeaderActions(): array
    {
        return [
            DeleteAction::make(),
        ];
    }
}
