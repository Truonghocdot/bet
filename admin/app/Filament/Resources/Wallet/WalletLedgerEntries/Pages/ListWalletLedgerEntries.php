<?php

namespace App\Filament\Resources\Wallet\WalletLedgerEntries\Pages;

use App\Filament\Resources\Wallet\WalletLedgerEntries\WalletLedgerEntryResource;
use Filament\Resources\Pages\ListRecords;

class ListWalletLedgerEntries extends ListRecords
{
    protected static string $resource = WalletLedgerEntryResource::class;
}
