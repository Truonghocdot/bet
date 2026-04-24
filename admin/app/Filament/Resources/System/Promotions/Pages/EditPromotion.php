<?php

namespace App\Filament\Resources\System\Promotions\Pages;

use App\Filament\Resources\System\Promotions\PromotionResource;
use App\Support\Media\WebpImageConverter;
use Filament\Resources\Pages\EditRecord;

class EditPromotion extends EditRecord
{
    protected static string $resource = PromotionResource::class;

    protected function mutateFormDataBeforeSave(array $data): array
    {
        $data['image_path'] = WebpImageConverter::convertPublicDiskPath($data['image_path'] ?? null);

        return $data;
    }
}
