<?php

namespace App\Filament\Resources\Users\Pages\Admins;

use App\Enum\User\RoleUser;
use App\Filament\Resources\Users\AdminUserResource;
use App\Filament\Resources\Users\Pages\CreateUser;

class CreateAdminUser extends CreateUser
{
    protected static string $resource = AdminUserResource::class;

    protected static function fixedRole(): ?RoleUser
    {
        return RoleUser::ADMIN;
    }
}
