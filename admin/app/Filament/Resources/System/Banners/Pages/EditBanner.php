<?php

namespace App\Filament\Resources\System\Banners\Pages;

use App\Filament\Resources\System\Banners\BannerResource;
use App\Support\Media\WebpImageConverter;
use Filament\Resources\Pages\EditRecord;

class EditBanner extends EditRecord
{
    protected static string $resource = BannerResource::class;

    protected function mutateFormDataBeforeSave(array $data): array
    {
        $data['image_path'] = WebpImageConverter::convertPublicDiskPath($data['image_path'] ?? null);

        return $data;
    }
}

