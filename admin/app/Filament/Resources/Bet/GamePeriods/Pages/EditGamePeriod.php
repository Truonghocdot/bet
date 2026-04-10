<?php

namespace App\Filament\Resources\Bet\GamePeriods\Pages;

use App\Filament\Resources\Bet\GamePeriods\GamePeriodResource;
use Filament\Resources\Pages\EditRecord;

class EditGamePeriod extends EditRecord
{
    protected static string $resource = GamePeriodResource::class;
    protected static ?string $title = 'Hồ sơ kỳ game';
}
