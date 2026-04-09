<?php

namespace App\Filament\Resources\Bet\GamePeriods\Pages;

use App\Filament\Resources\Bet\GamePeriods\GamePeriodResource;
use Filament\Actions\DeleteAction;
use Filament\Resources\Pages\EditRecord;

class EditGamePeriod extends EditRecord
{
    protected static string $resource = GamePeriodResource::class;

    protected function getHeaderActions(): array
    {
        return [
            DeleteAction::make(),
        ];
    }
}
