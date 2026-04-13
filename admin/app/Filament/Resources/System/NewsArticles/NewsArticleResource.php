<?php

namespace App\Filament\Resources\System\NewsArticles;

use App\Filament\Resources\BaseResource;
use App\Filament\Resources\System\NewsArticles\Pages\CreateNewsArticle;
use App\Filament\Resources\System\NewsArticles\Pages\EditNewsArticle;
use App\Filament\Resources\System\NewsArticles\Pages\ListNewsArticles;
use App\Filament\Resources\System\NewsArticles\Schemas\NewsArticleForm;
use App\Filament\Resources\System\NewsArticles\Tables\NewsArticlesTable;
use App\Models\Content\NewsArticle;
use BackedEnum;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Table;
use UnitEnum;

class NewsArticleResource extends BaseResource
{
    protected static ?string $model = NewsArticle::class;
    protected static UnitEnum|string|null $navigationGroup = 'Thiết lập';
    protected static ?string $navigationLabel = 'Tin tức';
    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedNewspaper;
    protected static ?string $recordTitleAttribute = 'title';

    public static function shouldRegisterNavigation(): bool
    {
        return true;
    }

    protected static function abilityPrefix(): string
    {
        return 'system.news-articles';
    }

    public static function form(Schema $schema): Schema
    {
        return NewsArticleForm::configure($schema);
    }

    public static function table(Table $table): Table
    {
        return NewsArticlesTable::configure($table);
    }

    public static function getPages(): array
    {
        return [
            'index' => ListNewsArticles::route('/'),
            'create' => CreateNewsArticle::route('/create'),
            'edit' => EditNewsArticle::route('/{record}/edit'),
        ];
    }
}

