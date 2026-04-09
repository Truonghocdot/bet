<?php

namespace App\Filament\Resources\Wallet\WalletLedgerEntries\Pages;

use App\Filament\Resources\Wallet\WalletLedgerEntries\WalletLedgerEntryResource;
use Filament\Actions\DeleteAction;
use Filament\Resources\Pages\EditRecord;

class EditWalletLedgerEntry extends EditRecord
{
    protected static string $resource = WalletLedgerEntryResource::class;

    protected function getHeaderActions(): array
    {
        return [
            DeleteAction::make(),
        ];
    }
}
