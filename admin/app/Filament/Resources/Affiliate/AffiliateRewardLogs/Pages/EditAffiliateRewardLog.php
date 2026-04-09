<?php

namespace App\Filament\Resources\Affiliate\AffiliateRewardLogs\Pages;

use App\Filament\Resources\Affiliate\AffiliateRewardLogs\AffiliateRewardLogResource;
use Filament\Actions\DeleteAction;
use Filament\Resources\Pages\EditRecord;

class EditAffiliateRewardLog extends EditRecord
{
    protected static string $resource = AffiliateRewardLogResource::class;

    protected function getHeaderActions(): array
    {
        return [
            DeleteAction::make(),
        ];
    }
}
