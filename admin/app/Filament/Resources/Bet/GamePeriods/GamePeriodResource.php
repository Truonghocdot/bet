<?php

namespace App\Filament\Resources\Bet\GamePeriods;

use App\Filament\Resources\Bet\GamePeriods\Pages\CreateGamePeriod;
use App\Filament\Resources\Bet\GamePeriods\Pages\EditGamePeriod;
use App\Filament\Resources\Bet\GamePeriods\Pages\ListGamePeriods;
use App\Filament\Resources\Bet\GamePeriods\Schemas\GamePeriodForm;
use App\Filament\Resources\Bet\GamePeriods\Tables\GamePeriodsTable;
use App\Models\Bet\GamePeriod;
use BackedEnum;
use UnitEnum;
use App\Filament\Resources\BaseResource;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Table;

class GamePeriodResource extends BaseResource
{
    protected static ?string $model = GamePeriod::class;
    protected static bool $canDeleteRecords = false;
    protected static bool $canDeleteAnyRecords = false;
    protected static bool $canForceDeleteRecords = false;
    protected static bool $canForceDeleteAnyRecords = false;
    protected static bool $canRestoreRecords = false;
    protected static bool $canRestoreAnyRecords = false;
    protected static UnitEnum|string|null $navigationGroup = 'Cược';
    protected static ?string $navigationLabel = 'Kỳ game';
    public static function shouldRegisterNavigation(): bool
    {
        return true;
    }

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedRectangleStack;

    protected static ?string $recordTitleAttribute = 'period_no';

    protected static function abilityPrefix(): string
    {
        return 'bet.game-periods';
    }

    public static function form(Schema $schema): Schema
    {
        return GamePeriodForm::configure($schema);
    }

    public static function table(Table $table): Table
    {
        return GamePeriodsTable::configure($table);
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
            'index' => ListGamePeriods::route('/'),
            'create' => CreateGamePeriod::route('/create'),
            'edit' => EditGamePeriod::route('/{record}/edit'),
        ];
    }
}
