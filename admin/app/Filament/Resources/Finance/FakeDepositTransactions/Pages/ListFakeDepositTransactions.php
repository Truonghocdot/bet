<?php

namespace App\Filament\Resources\Finance\FakeDepositTransactions\Pages;

use App\Filament\Resources\Finance\FakeDepositTransactions\FakeDepositTransactionResource;
use Filament\Resources\Pages\ListRecords;

class ListFakeDepositTransactions extends ListRecords
{
    protected static string $resource = FakeDepositTransactionResource::class;

    protected function getHeaderActions(): array
    {
        return [];
    }
}
