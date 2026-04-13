<?php

namespace App\Filament\Resources\System\NewsArticles\Pages;

use App\Filament\Resources\System\NewsArticles\NewsArticleResource;
use App\Models\Content\NewsArticle;
use App\Support\Media\WebpImageConverter;
use Filament\Resources\Pages\CreateRecord;
use Illuminate\Support\Str;

class CreateNewsArticle extends CreateRecord
{
    protected static string $resource = NewsArticleResource::class;

    protected function mutateFormDataBeforeCreate(array $data): array
    {
        $data['created_by'] = auth()->id();
        $data['cover_image_path'] = WebpImageConverter::convertPublicDiskPath($data['cover_image_path'] ?? null);

        $slugSeed = trim((string) ($data['slug'] ?? ''));
        if ($slugSeed === '') {
            $slugSeed = (string) ($data['title'] ?? '');
        }
        $slug = Str::slug($slugSeed);
        if ($slug === '') {
            $slug = 'tin-tuc-'.time();
        }
        if (NewsArticle::query()->where('slug', $slug)->exists()) {
            $slug .= '-'.time();
        }
        $data['slug'] = $slug;

        if (($data['is_published'] ?? false) && blank($data['published_at'] ?? null)) {
            $data['published_at'] = now();
        }

        return $data;
    }
}

