<?php

namespace App\Filament\Resources\Users\Pages\Agencies;

use App\Enum\User\RoleUser;
use App\Filament\Resources\Users\AgencyUserResource;
use App\Filament\Resources\Users\Pages\CreateUser;

class CreateAgencyUser extends CreateUser
{
    protected static string $resource = AgencyUserResource::class;

    protected static function fixedRole(): ?RoleUser
    {
        return RoleUser::AGENCY;
    }

    protected static function forceAffiliateProfile(): bool
    {
        return true;
    }
}
