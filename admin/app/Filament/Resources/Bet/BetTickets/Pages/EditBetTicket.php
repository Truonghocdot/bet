<?php

namespace App\Filament\Resources\Bet\BetTickets\Pages;

use App\Filament\Resources\Bet\BetTickets\BetTicketResource;
use Filament\Actions\DeleteAction;
use Filament\Resources\Pages\EditRecord;

class EditBetTicket extends EditRecord
{
    protected static string $resource = BetTicketResource::class;

    protected function getHeaderActions(): array
    {
        return [
            DeleteAction::make(),
        ];
    }
}
