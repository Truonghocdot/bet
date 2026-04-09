<?php

namespace App\Filament\Resources\Affiliate\AffiliateRewardSettings;

use App\Filament\Resources\Affiliate\AffiliateRewardSettings\Pages\CreateAffiliateRewardSetting;
use App\Filament\Resources\Affiliate\AffiliateRewardSettings\Pages\EditAffiliateRewardSetting;
use App\Filament\Resources\Affiliate\AffiliateRewardSettings\Pages\ListAffiliateRewardSettings;
use App\Filament\Resources\Affiliate\AffiliateRewardSettings\Schemas\AffiliateRewardSettingForm;
use App\Filament\Resources\Affiliate\AffiliateRewardSettings\Tables\AffiliateRewardSettingsTable;
use App\Models\Affiliate\AffiliateRewardSetting;
use BackedEnum;
use UnitEnum;
use App\Filament\Resources\BaseResource;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Table;

class AffiliateRewardSettingResource extends BaseResource
{
    protected static ?string $model = AffiliateRewardSetting::class;
    protected static UnitEnum|string|null $navigationGroup = 'Tiếp thị liên kết';
    protected static ?string $navigationLabel = 'Cấu hình thưởng';

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedRectangleStack;

    protected static ?string $recordTitleAttribute = 'name';

    protected static function abilityPrefix(): string
    {
        return 'affiliate.affiliate-reward-settings';
    }

    public static function form(Schema $schema): Schema
    {
        return AffiliateRewardSettingForm::configure($schema);
    }

    public static function table(Table $table): Table
    {
        return AffiliateRewardSettingsTable::configure($table);
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
            'index' => ListAffiliateRewardSettings::route('/'),
            'create' => CreateAffiliateRewardSetting::route('/create'),
            'edit' => EditAffiliateRewardSetting::route('/{record}/edit'),
        ];
    }
}
