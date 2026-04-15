<?php

namespace App\Filament\Resources\Users\Pages\Staff;

use App\Filament\Resources\Users\Pages\ListUsers;
use App\Filament\Resources\Users\StaffUserResource;

class ListStaffUsers extends ListUsers
{
    protected static string $resource = StaffUserResource::class;
}
