<?php

namespace App\Filament\Resources\Affiliate\AffiliateRewardSettings\Pages;

use App\Filament\Resources\Affiliate\AffiliateRewardSettings\AffiliateRewardSettingResource;
use Filament\Actions\DeleteAction;
use Filament\Resources\Pages\EditRecord;

class EditAffiliateRewardSetting extends EditRecord
{
    protected static string $resource = AffiliateRewardSettingResource::class;

    protected function getHeaderActions(): array
    {
        return [
            DeleteAction::make(),
        ];
    }
}
