<?php

namespace App\Filament\Resources\Affiliate\AffiliateRewardSettings\Pages;

use App\Filament\Resources\Affiliate\AffiliateRewardSettings\AffiliateRewardSettingResource;
use Filament\Actions\CreateAction;
use Filament\Resources\Pages\ListRecords;

class ListAffiliateRewardSettings extends ListRecords
{
    protected static string $resource = AffiliateRewardSettingResource::class;

    public function getBreadcrumb(): string
    {
        return 'Danh sách cấu hình thưởng';
    }

    protected function getHeaderActions(): array
    {
        return [
            CreateAction::make(),
        ];
    }
}
