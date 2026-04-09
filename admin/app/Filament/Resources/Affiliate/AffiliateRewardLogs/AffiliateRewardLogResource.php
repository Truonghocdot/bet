<?php

namespace App\Filament\Resources\Affiliate\AffiliateRewardLogs;

use App\Filament\Resources\Affiliate\AffiliateRewardLogs\Pages\CreateAffiliateRewardLog;
use App\Filament\Resources\Affiliate\AffiliateRewardLogs\Pages\EditAffiliateRewardLog;
use App\Filament\Resources\Affiliate\AffiliateRewardLogs\Pages\ListAffiliateRewardLogs;
use App\Filament\Resources\Affiliate\AffiliateRewardLogs\Schemas\AffiliateRewardLogForm;
use App\Filament\Resources\Affiliate\AffiliateRewardLogs\Tables\AffiliateRewardLogsTable;
use App\Models\Affiliate\AffiliateRewardLog;
use BackedEnum;
use UnitEnum;
use App\Filament\Resources\BaseResource;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Table;

class AffiliateRewardLogResource extends BaseResource
{
    protected static ?string $model = AffiliateRewardLog::class;
    protected static UnitEnum|string|null $navigationGroup = 'Tiếp thị liên kết';
    protected static ?string $navigationLabel = 'Lịch sử thưởng';

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedRectangleStack;

    protected static ?string $recordTitleAttribute = 'id';

    protected static function abilityPrefix(): string
    {
        return 'affiliate.affiliate-reward-logs';
    }

    public static function form(Schema $schema): Schema
    {
        return AffiliateRewardLogForm::configure($schema);
    }

    public static function table(Table $table): Table
    {
        return AffiliateRewardLogsTable::configure($table);
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
            'index' => ListAffiliateRewardLogs::route('/'),
            'create' => CreateAffiliateRewardLog::route('/create'),
            'edit' => EditAffiliateRewardLog::route('/{record}/edit'),
        ];
    }
}
