<?php

namespace App\Filament\Resources\System\Promotions;

use App\Filament\Resources\BaseResource;
use App\Filament\Resources\System\Banners\Schemas\BannerForm;
use App\Filament\Resources\System\Banners\Tables\BannersTable;
use App\Filament\Resources\System\Promotions\Pages\CreatePromotion;
use App\Filament\Resources\System\Promotions\Pages\EditPromotion;
use App\Filament\Resources\System\Promotions\Pages\ListPromotions;
use App\Models\Content\Banner;
use BackedEnum;
use Illuminate\Database\Eloquent\Builder;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Table;
use UnitEnum;

class PromotionResource extends BaseResource
{
    protected static ?string $model = Banner::class;
    protected static UnitEnum|string|null $navigationGroup = 'Thiết lập';
    protected static ?string $navigationLabel = 'Khuyến mãi';
    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedPhoto;
    protected static ?string $recordTitleAttribute = 'id';

    public static function shouldRegisterNavigation(): bool
    {
        return true;
    }

    protected static function abilityPrefix(): string
    {
        return 'system.promotions';
    }

    public static function form(Schema $schema): Schema
    {
        return BannerForm::configure(
            $schema,
            placement: 'promotion',
            sectionTitle: 'Hình ảnh khuyến mãi',
            imageLabel: 'Ảnh khuyến mãi',
        );
    }

    public static function table(Table $table): Table
    {
        return BannersTable::configure(
            $table,
            imageLabel: 'Ảnh khuyến mãi',
            createLabel: 'Tạo ảnh khuyến mãi',
        );
    }

    public static function getEloquentQuery(): Builder
    {
        return parent::getEloquentQuery()->where('placement', 'promotion');
    }

    public static function getPages(): array
    {
        return [
            'index' => ListPromotions::route('/'),
            'create' => CreatePromotion::route('/create'),
            'edit' => EditPromotion::route('/{record}/edit'),
        ];
    }
}
