<?php

namespace App\Filament\Resources\System\Banners;

use App\Filament\Resources\BaseResource;
use App\Filament\Resources\System\Banners\Pages\CreateBanner;
use App\Filament\Resources\System\Banners\Pages\EditBanner;
use App\Filament\Resources\System\Banners\Pages\ListBanners;
use App\Filament\Resources\System\Banners\Schemas\BannerForm;
use App\Filament\Resources\System\Banners\Tables\BannersTable;
use App\Models\Content\Banner;
use BackedEnum;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Table;
use UnitEnum;

class BannerResource extends BaseResource
{
    protected static ?string $model = Banner::class;
    protected static UnitEnum|string|null $navigationGroup = 'Thiết lập';
    protected static ?string $navigationLabel = 'Banner trang chủ';
    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedPhoto;
    protected static ?string $recordTitleAttribute = 'id';

    public static function shouldRegisterNavigation(): bool
    {
        return true;
    }

    protected static function abilityPrefix(): string
    {
        return 'system.banners';
    }

    public static function form(Schema $schema): Schema
    {
        return BannerForm::configure($schema);
    }

    public static function table(Table $table): Table
    {
        return BannersTable::configure($table);
    }

    public static function getPages(): array
    {
        return [
            'index' => ListBanners::route('/'),
            'create' => CreateBanner::route('/create'),
            'edit' => EditBanner::route('/{record}/edit'),
        ];
    }
}

