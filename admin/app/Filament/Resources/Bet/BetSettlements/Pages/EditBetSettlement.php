<?php

namespace App\Filament\Resources\Bet\BetSettlements\Pages;

use App\Filament\Resources\Bet\BetSettlements\BetSettlementResource;
use Filament\Actions\DeleteAction;
use Filament\Resources\Pages\EditRecord;

class EditBetSettlement extends EditRecord
{
    protected static string $resource = BetSettlementResource::class;

    protected function getHeaderActions(): array
    {
        return [
            DeleteAction::make(),
        ];
    }
}
