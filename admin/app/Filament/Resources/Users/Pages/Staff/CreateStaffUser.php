<?php

namespace App\Filament\Resources\Users\Pages\Staff;

use App\Enum\User\RoleUser;
use App\Filament\Resources\Users\Pages\CreateUser;
use App\Filament\Resources\Users\StaffUserResource;

class CreateStaffUser extends CreateUser
{
    protected static string $resource = StaffUserResource::class;

    protected static function fixedRole(): ?RoleUser
    {
        return RoleUser::STAFF;
    }

    protected static function forceAffiliateProfile(): bool
    {
        return true;
    }
}
