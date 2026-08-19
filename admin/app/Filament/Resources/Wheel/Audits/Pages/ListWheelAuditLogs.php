<?php

namespace App\Filament\Resources\Wheel\Audits\Pages;

use App\Filament\Resources\Wheel\Audits\WheelAuditLogResource;
use Filament\Resources\Pages\ListRecords;

class ListWheelAuditLogs extends ListRecords
{
    protected static string $resource = WheelAuditLogResource::class;
}
