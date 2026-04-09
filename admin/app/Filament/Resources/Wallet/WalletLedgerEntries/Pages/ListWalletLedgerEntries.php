<?php

namespace App\Filament\Resources\Wallet\WalletLedgerEntries\Pages;

use App\Filament\Resources\Wallet\WalletLedgerEntries\WalletLedgerEntryResource;
use Filament\Actions\CreateAction;
use Filament\Resources\Pages\ListRecords;

class ListWalletLedgerEntries extends ListRecords
{
    protected static string $resource = WalletLedgerEntryResource::class;

    protected function getHeaderActions(): array
    {
        return [
            CreateAction::make(),
        ];
    }
}
