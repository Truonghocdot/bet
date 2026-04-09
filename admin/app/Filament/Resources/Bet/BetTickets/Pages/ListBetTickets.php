<?php

namespace App\Filament\Resources\Bet\BetTickets\Pages;

use App\Filament\Resources\Bet\BetTickets\BetTicketResource;
use Filament\Actions\CreateAction;
use Filament\Resources\Pages\ListRecords;

class ListBetTickets extends ListRecords
{
    protected static string $resource = BetTicketResource::class;

    protected function getHeaderActions(): array
    {
        return [
            CreateAction::make(),
        ];
    }
}
