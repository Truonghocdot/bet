<?php

namespace App\Filament\Resources\Users\RelationManagers;

use App\Filament\Resources\Bet\BetTickets\BetTicketResource;
use Filament\Resources\RelationManagers\RelationManager;

class BetTicketsRelationManager extends RelationManager
{
    protected static string $relationship = 'gameTickets';
    protected static ?string $relatedResource = BetTicketResource::class;
    protected static ?string $title = 'Vé cược';
}
