<?php

namespace App\Filament\Resources\Bet\BetItems;

use App\Filament\Resources\Bet\BetItems\Pages\CreateBetItem;
use App\Filament\Resources\Bet\BetItems\Pages\EditBetItem;
use App\Filament\Resources\Bet\BetItems\Pages\ListBetItems;
use App\Filament\Resources\Bet\BetItems\Schemas\BetItemForm;
use App\Filament\Resources\Bet\BetItems\Tables\BetItemsTable;
use App\Models\Bet\BetItem;
use BackedEnum;
use UnitEnum;
use App\Filament\Resources\BaseResource;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Table;

class BetItemResource extends BaseResource
{
    protected static ?string $model = BetItem::class;
    protected static UnitEnum|string|null $navigationGroup = 'Cược';
    protected static ?string $navigationLabel = 'Chi tiết cược';

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedRectangleStack;

    protected static ?string $recordTitleAttribute = 'id';

    protected static function abilityPrefix(): string
    {
        return 'bet.bet-items';
    }

    public static function form(Schema $schema): Schema
    {
        return BetItemForm::configure($schema);
    }

    public static function table(Table $table): Table
    {
        return BetItemsTable::configure($table);
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
            'index' => ListBetItems::route('/'),
            'create' => CreateBetItem::route('/create'),
            'edit' => EditBetItem::route('/{record}/edit'),
        ];
    }
}
