<?php

namespace App\Filament\Resources\Bet\BetItems\Pages;

use App\Filament\Resources\Bet\BetItems\BetItemResource;
use Filament\Actions\DeleteAction;
use Filament\Resources\Pages\EditRecord;

class EditBetItem extends EditRecord
{
    protected static string $resource = BetItemResource::class;

    protected function getHeaderActions(): array
    {
        return [
            DeleteAction::make(),
        ];
    }
}
