<?php

namespace App\Filament\Resources\System\NewsArticles\Pages;

use App\Filament\Resources\System\NewsArticles\NewsArticleResource;
use Filament\Resources\Pages\ListRecords;

class ListNewsArticles extends ListRecords
{
    protected static string $resource = NewsArticleResource::class;
}

