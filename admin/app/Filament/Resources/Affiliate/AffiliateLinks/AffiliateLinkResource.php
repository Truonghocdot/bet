<?php

namespace App\Filament\Resources\Affiliate\AffiliateLinks;

use App\Filament\Resources\Affiliate\AffiliateLinks\Pages\CreateAffiliateLink;
use App\Filament\Resources\Affiliate\AffiliateLinks\Pages\EditAffiliateLink;
use App\Filament\Resources\Affiliate\AffiliateLinks\Pages\ListAffiliateLinks;
use App\Filament\Resources\Affiliate\AffiliateLinks\Schemas\AffiliateLinkForm;
use App\Filament\Resources\Affiliate\AffiliateLinks\Tables\AffiliateLinksTable;
use App\Models\Affiliate\AffiliateLink;
use BackedEnum;
use UnitEnum;
use App\Filament\Resources\BaseResource;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Table;

class AffiliateLinkResource extends BaseResource
{
    protected static ?string $model = AffiliateLink::class;
    protected static UnitEnum|string|null $navigationGroup = 'Tiếp thị liên kết';
    protected static ?string $navigationLabel = 'Liên kết theo dõi';

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedRectangleStack;

    protected static ?string $recordTitleAttribute = 'tracking_code';

    protected static function abilityPrefix(): string
    {
        return 'affiliate.affiliate-links';
    }

    public static function form(Schema $schema): Schema
    {
        return AffiliateLinkForm::configure($schema);
    }

    public static function table(Table $table): Table
    {
        return AffiliateLinksTable::configure($table);
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
            'index' => ListAffiliateLinks::route('/'),
            'create' => CreateAffiliateLink::route('/create'),
            'edit' => EditAffiliateLink::route('/{record}/edit'),
        ];
    }
}
