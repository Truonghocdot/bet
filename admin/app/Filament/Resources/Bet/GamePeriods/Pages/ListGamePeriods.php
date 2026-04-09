<?php

namespace App\Filament\Resources\Bet\GamePeriods\Pages;

use App\Filament\Resources\Bet\GamePeriods\GamePeriodResource;
use Filament\Actions\CreateAction;
use Filament\Resources\Pages\ListRecords;

class ListGamePeriods extends ListRecords
{
    protected static string $resource = GamePeriodResource::class;

    protected function getHeaderActions(): array
    {
        return [
            CreateAction::make(),
        ];
    }
}
