<?php

namespace App\Filament\Resources\Finance\FakeWithdrawTransactions\Pages;

use App\Filament\Resources\Finance\FakeWithdrawTransactions\FakeWithdrawTransactionResource;
use Filament\Resources\Pages\ListRecords;

class ListFakeWithdrawTransactions extends ListRecords
{
    protected static string $resource = FakeWithdrawTransactionResource::class;

    protected function getHeaderActions(): array
    {
        return [];
    }
}
