<?php

namespace App\Filament\Resources\System\Banners\Pages;

use App\Filament\Resources\System\Banners\BannerResource;
use App\Support\Media\WebpImageConverter;
use Filament\Resources\Pages\CreateRecord;

class CreateBanner extends CreateRecord
{
    protected static string $resource = BannerResource::class;

    protected function mutateFormDataBeforeCreate(array $data): array
    {
        $data['created_by'] = auth()->id();
        $data['image_path'] = WebpImageConverter::convertPublicDiskPath($data['image_path'] ?? null);

        return $data;
    }
}

