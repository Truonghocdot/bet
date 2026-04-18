<?php

namespace App\Filament\Resources\Users\Pages\Admins;

use App\Filament\Resources\Users\AdminUserResource;
use App\Filament\Resources\Users\Pages\EditUser;

class EditAdminUser extends EditUser
{
    protected static string $resource = AdminUserResource::class;
}
