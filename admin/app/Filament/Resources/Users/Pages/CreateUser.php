<?php

namespace App\Filament\Resources\Users\Pages;

use App\Enum\User\RoleUser;
use App\Services\Admin\UserProvisioningService;
use Filament\Resources\Pages\CreateRecord;
use Illuminate\Database\Eloquent\Model;

abstract class CreateUser extends CreateRecord
{
    protected static function fixedRole(): ?RoleUser
    {
        return null;
    }

    protected static function forceAffiliateProfile(): bool
    {
        return false;
    }

    protected function mutateFormDataBeforeCreate(array $data): array
    {
        if ($role = static::fixedRole()) {
            $data['role'] = $role->value;
        }

        if (static::forceAffiliateProfile()) {
            $data['provision_affiliate_profile'] = true;
        }

        return $data;
    }

    protected function handleRecordCreation(array $data): Model
    {
        return app(UserProvisioningService::class)->createFromErp($data, auth()->user());
    }
}
