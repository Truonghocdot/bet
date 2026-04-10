<?php

namespace App\Filament\Resources\Transaction\Transactions\Pages;

use App\Filament\Resources\Transaction\Transactions\TransactionResource;
use Filament\Resources\Pages\ListRecords;

class ListTransactions extends ListRecords
{
    protected static string $resource = TransactionResource::class;
}
