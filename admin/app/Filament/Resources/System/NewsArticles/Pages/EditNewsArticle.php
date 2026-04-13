<?php

namespace App\Filament\Resources\System\NewsArticles\Pages;

use App\Filament\Resources\System\NewsArticles\NewsArticleResource;
use App\Models\Content\NewsArticle;
use App\Support\Media\WebpImageConverter;
use Filament\Resources\Pages\EditRecord;
use Illuminate\Support\Str;

class EditNewsArticle extends EditRecord
{
    protected static string $resource = NewsArticleResource::class;

    protected function mutateFormDataBeforeSave(array $data): array
    {
        $data['cover_image_path'] = WebpImageConverter::convertPublicDiskPath($data['cover_image_path'] ?? null);

        $slugSeed = trim((string) ($data['slug'] ?? ''));
        if ($slugSeed === '') {
            $slugSeed = (string) ($data['title'] ?? '');
        }
        $slug = Str::slug($slugSeed);
        if ($slug === '') {
            $slug = 'tin-tuc-'.$this->record->id;
        }
        if (
            NewsArticle::query()
                ->where('slug', $slug)
                ->where('id', '!=', $this->record->id)
                ->exists()
        ) {
            $slug .= '-'.time();
        }
        $data['slug'] = $slug;

        if (($data['is_published'] ?? false) && blank($data['published_at'] ?? null)) {
            $data['published_at'] = now();
        }

        if (! ($data['is_published'] ?? false)) {
            $data['published_at'] = null;
        }

        return $data;
    }
}

