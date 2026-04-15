<?php

namespace App\Filament\Resources\Users\Pages\Clients;

use App\Enum\User\RoleUser;
use App\Filament\Resources\Users\ClientUserResource;
use App\Filament\Resources\Users\Pages\CreateUser;

class CreateClientUser extends CreateUser
{
    protected static string $resource = ClientUserResource::class;

    protected static function fixedRole(): ?RoleUser
    {
        return RoleUser::CLIENT;
    }
}
