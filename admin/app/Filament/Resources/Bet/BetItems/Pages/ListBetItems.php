<?php

namespace App\Filament\Resources\Bet\BetItems\Pages;

use App\Filament\Resources\Bet\BetItems\BetItemResource;
use Filament\Actions\CreateAction;
use Filament\Resources\Pages\ListRecords;

class ListBetItems extends ListRecords
{
    protected static string $resource = BetItemResource::class;

    protected function getHeaderActions(): array
    {
        return [
            CreateAction::make(),
        ];
    }
}
