<?php

namespace App\Filament\Resources\Affiliate\AffiliateLinks\Pages;

use App\Filament\Resources\Affiliate\AffiliateLinks\AffiliateLinkResource;
use Filament\Actions\DeleteAction;
use Filament\Resources\Pages\EditRecord;

class EditAffiliateLink extends EditRecord
{
    protected static string $resource = AffiliateLinkResource::class;

    protected function getHeaderActions(): array
    {
        return [
            DeleteAction::make(),
        ];
    }
}
