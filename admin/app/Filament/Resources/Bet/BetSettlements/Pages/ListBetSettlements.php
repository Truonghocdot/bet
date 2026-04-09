<?php

namespace App\Filament\Resources\Bet\BetSettlements\Pages;

use App\Filament\Resources\Bet\BetSettlements\BetSettlementResource;
use Filament\Actions\CreateAction;
use Filament\Resources\Pages\ListRecords;

class ListBetSettlements extends ListRecords
{
    protected static string $resource = BetSettlementResource::class;

    protected function getHeaderActions(): array
    {
        return [
            CreateAction::make(),
        ];
    }
}
