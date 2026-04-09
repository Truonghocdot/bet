<?php

namespace App\Filament\Resources\Bet\BetTickets;

use App\Filament\Resources\Bet\BetTickets\Pages\CreateBetTicket;
use App\Filament\Resources\Bet\BetTickets\Pages\EditBetTicket;
use App\Filament\Resources\Bet\BetTickets\Pages\ListBetTickets;
use App\Filament\Resources\Bet\BetTickets\Schemas\BetTicketForm;
use App\Filament\Resources\Bet\BetTickets\Tables\BetTicketsTable;
use App\Models\Bet\BetTicket;
use BackedEnum;
use UnitEnum;
use App\Filament\Resources\BaseResource;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Table;

class BetTicketResource extends BaseResource
{
    protected static ?string $model = BetTicket::class;
    protected static UnitEnum|string|null $navigationGroup = 'Cược';
    protected static ?string $navigationLabel = 'Vé cược';

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedRectangleStack;

    protected static ?string $recordTitleAttribute = 'ticket_no';

    protected static function abilityPrefix(): string
    {
        return 'bet.bet-tickets';
    }

    public static function form(Schema $schema): Schema
    {
        return BetTicketForm::configure($schema);
    }

    public static function table(Table $table): Table
    {
        return BetTicketsTable::configure($table);
    }

    public static function getRelations(): array
    {
        return [
            //
        ];
    }

    public static function getPages(): array
    {
        return [
            'index' => ListBetTickets::route('/'),
            'create' => CreateBetTicket::route('/create'),
            'edit' => EditBetTicket::route('/{record}/edit'),
        ];
    }
}
