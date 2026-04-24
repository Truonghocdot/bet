<?php

namespace App\Filament\Resources\System\Promotions\Pages;

use App\Filament\Resources\System\Promotions\PromotionResource;
use App\Support\Media\WebpImageConverter;
use Filament\Resources\Pages\CreateRecord;

class CreatePromotion extends CreateRecord
{
    protected static string $resource = PromotionResource::class;

    protected function mutateFormDataBeforeCreate(array $data): array
    {
        $data['created_by'] = auth()->id();
        $data['image_path'] = WebpImageConverter::convertPublicDiskPath($data['image_path'] ?? null);

        return $data;
    }
}
