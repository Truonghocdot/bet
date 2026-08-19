<?php

namespace App\Filament\Resources\Wheel\Campaigns\Pages;

use App\Filament\Resources\Wheel\Campaigns\WheelCampaignResource;
use Filament\Resources\Pages\CreateRecord;

class CreateWheelCampaign extends CreateRecord
{
    protected static string $resource = WheelCampaignResource::class;

    protected function mutateFormDataBeforeCreate(array $data): array
    {
        $data['created_by'] = auth()->id();
        $data['duration_seconds'] = 300;
        $data['spin_duration_seconds'] = 5;

        return $data;
    }
}
