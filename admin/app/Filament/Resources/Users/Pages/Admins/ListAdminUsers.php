<?php

namespace App\Filament\Resources\Users\Pages\Admins;

use App\Filament\Resources\Users\AdminUserResource;
use App\Filament\Resources\Users\Pages\ListUsers;

class ListAdminUsers extends ListUsers
{
    protected static string $resource = AdminUserResource::class;
}
