<?php

namespace App\Filament\Resources\Affiliate\AffiliateReferrals;

use App\Filament\Resources\Affiliate\AffiliateReferrals\Pages\CreateAffiliateReferral;
use App\Filament\Resources\Affiliate\AffiliateReferrals\Pages\EditAffiliateReferral;
use App\Filament\Resources\Affiliate\AffiliateReferrals\Pages\ListAffiliateReferrals;
use App\Filament\Resources\Affiliate\AffiliateReferrals\Schemas\AffiliateReferralForm;
use App\Filament\Resources\Affiliate\AffiliateReferrals\Tables\AffiliateReferralsTable;
use App\Models\Affiliate\AffiliateReferral;
use BackedEnum;
use UnitEnum;
use App\Filament\Resources\BaseResource;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Table;

class AffiliateReferralResource extends BaseResource
{
    protected static ?string $model = AffiliateReferral::class;
    protected static bool $canCreateRecords = false;
    protected static bool $canUpdateRecords = false;
    protected static bool $canDeleteRecords = false;
    protected static bool $canDeleteAnyRecords = false;
    protected static bool $canForceDeleteRecords = false;
    protected static bool $canForceDeleteAnyRecords = false;
    protected static bool $canRestoreRecords = false;
    protected static bool $canRestoreAnyRecords = false;
    protected static UnitEnum|string|null $navigationGroup = 'Tiếp thị liên kết';
    protected static ?string $navigationLabel = 'Người được mời';

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedRectangleStack;

    protected static ?string $recordTitleAttribute = 'id';

    protected static function abilityPrefix(): string
    {
        return 'affiliate.affiliate-referrals';
    }

    public static function form(Schema $schema): Schema
    {
        return AffiliateReferralForm::configure($schema);
    }

    public static function table(Table $table): Table
    {
        return AffiliateReferralsTable::configure($table);
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
            'index' => ListAffiliateReferrals::route('/'),
            'create' => CreateAffiliateReferral::route('/create'),
            'edit' => EditAffiliateReferral::route('/{record}/edit'),
        ];
    }
}
