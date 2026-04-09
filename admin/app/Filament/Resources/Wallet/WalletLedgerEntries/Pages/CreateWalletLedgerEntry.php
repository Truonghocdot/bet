<?php

namespace App\Filament\Resources\Wallet\WalletLedgerEntries\Pages;

use App\Filament\Resources\Wallet\WalletLedgerEntries\WalletLedgerEntryResource;
use Filament\Resources\Pages\CreateRecord;

class CreateWalletLedgerEntry extends CreateRecord
{
    protected static string $resource = WalletLedgerEntryResource::class;
}
