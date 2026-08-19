<?php

namespace App\Filament\Resources\Wheel\Invitations\Pages;

use App\Filament\Resources\Wheel\Invitations\WheelInvitationResource;
use App\Services\Wheel\WheelCampaignService;
use Filament\Resources\Pages\CreateRecord;

class CreateWheelInvitation extends CreateRecord
{
    protected static string $resource = WheelInvitationResource::class;

    protected function mutateFormDataBeforeCreate(array $data): array
    {
        $data['status'] = 'draft';

        return $data;
    }

    protected function afterCreate(): void
    {
        app(WheelCampaignService::class)->snapshotInvitation($this->record);
    }
}
