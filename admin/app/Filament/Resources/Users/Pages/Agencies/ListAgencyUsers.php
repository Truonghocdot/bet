<?php

namespace App\Filament\Resources\Users\Pages\Agencies;

use App\Filament\Resources\Users\AgencyUserResource;
use App\Filament\Resources\Users\Pages\ListUsers;

class ListAgencyUsers extends ListUsers
{
    protected static string $resource = AgencyUserResource::class;
}
