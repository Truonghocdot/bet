<?php

namespace App\Filament\Resources\Affiliate\AffiliateProfiles;

use App\Filament\Resources\Affiliate\AffiliateProfiles\Pages\CreateAffiliateProfile;
use App\Filament\Resources\Affiliate\AffiliateProfiles\Pages\EditAffiliateProfile;
use App\Filament\Resources\Affiliate\AffiliateProfiles\Pages\ListAffiliateProfiles;
use App\Filament\Resources\Affiliate\AffiliateProfiles\Schemas\AffiliateProfileForm;
use App\Filament\Resources\Affiliate\AffiliateProfiles\Tables\AffiliateProfilesTable;
use App\Models\Affiliate\AffiliateProfile;
use BackedEnum;
use UnitEnum;
use App\Filament\Resources\BaseResource;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Table;

class AffiliateProfileResource extends BaseResource
{
    protected static ?string $model = AffiliateProfile::class;
    protected static bool $canCreateRecords = false;
    protected static bool $canUpdateRecords = false;
    protected static bool $canDeleteRecords = false;
    protected static bool $canDeleteAnyRecords = false;
    protected static bool $canForceDeleteRecords = false;
    protected static bool $canForceDeleteAnyRecords = false;
    protected static bool $canRestoreRecords = false;
    protected static bool $canRestoreAnyRecords = false;
    protected static UnitEnum|string|null $navigationGroup = 'Tiếp thị liên kết';
    protected static ?string $navigationLabel = 'Hồ sơ affiliate';

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedRectangleStack;

    protected static ?string $recordTitleAttribute = 'ref_code';

    protected static function abilityPrefix(): string
    {
        return 'affiliate.affiliate-profiles';
    }

    public static function form(Schema $schema): Schema
    {
        return AffiliateProfileForm::configure($schema);
    }

    public static function table(Table $table): Table
    {
        return AffiliateProfilesTable::configure($table);
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
            'index' => ListAffiliateProfiles::route('/'),
            'create' => CreateAffiliateProfile::route('/create'),
            'edit' => EditAffiliateProfile::route('/{record}/edit'),
        ];
    }
}
