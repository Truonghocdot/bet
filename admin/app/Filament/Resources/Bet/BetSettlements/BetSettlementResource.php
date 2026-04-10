<?php

namespace App\Filament\Resources\Bet\BetSettlements;

use App\Filament\Resources\Bet\BetSettlements\Pages\CreateBetSettlement;
use App\Filament\Resources\Bet\BetSettlements\Pages\EditBetSettlement;
use App\Filament\Resources\Bet\BetSettlements\Pages\ListBetSettlements;
use App\Filament\Resources\Bet\BetSettlements\Schemas\BetSettlementForm;
use App\Filament\Resources\Bet\BetSettlements\Tables\BetSettlementsTable;
use App\Models\Bet\BetSettlement;
use BackedEnum;
use UnitEnum;
use App\Filament\Resources\BaseResource;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Table;

class BetSettlementResource extends BaseResource
{
    protected static ?string $model = BetSettlement::class;
    protected static bool $canCreateRecords = false;
    protected static bool $canUpdateRecords = false;
    protected static bool $canDeleteRecords = false;
    protected static bool $canDeleteAnyRecords = false;
    protected static bool $canForceDeleteRecords = false;
    protected static bool $canForceDeleteAnyRecords = false;
    protected static bool $canRestoreRecords = false;
    protected static bool $canRestoreAnyRecords = false;
    protected static UnitEnum|string|null $navigationGroup = 'Cược';
    protected static ?string $navigationLabel = 'Chốt cược';

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedRectangleStack;

    protected static ?string $recordTitleAttribute = 'id';

    protected static function abilityPrefix(): string
    {
        return 'bet.bet-settlements';
    }

    public static function form(Schema $schema): Schema
    {
        return BetSettlementForm::configure($schema);
    }

    public static function table(Table $table): Table
    {
        return BetSettlementsTable::configure($table);
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
            'index' => ListBetSettlements::route('/'),
            'create' => CreateBetSettlement::route('/create'),
            'edit' => EditBetSettlement::route('/{record}/edit'),
        ];
    }
}
